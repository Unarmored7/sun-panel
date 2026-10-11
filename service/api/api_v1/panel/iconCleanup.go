package panel

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sun-panel/global"
	"sun-panel/lib/iconcache"
	"sun-panel/models"
)

// Only files referenced by removed/replaced cards are immediate deletion candidates.
// Settings and other users' cards can share those same assets.
func cleanupCardIcons(removed []models.ItemIcon) error {
	if len(removed) == 0 {
		return nil
	}
	var cards []models.ItemIcon
	if err := global.Db.Find(&cards).Error; err != nil {
		return err
	}
	sources := []string{}
	paths := []string{}
	candidates := []string{}
	references := []string{}
	for _, card := range cards {
		json.Unmarshal([]byte(card.IconJson), &card.Icon)
		sources = append(sources, card.Icon.Src)
		paths = append(paths, card.Icon.Src)
		references = append(references, card.IconJson, card.ExpandParam)
	}
	for _, card := range removed {
		json.Unmarshal([]byte(card.IconJson), &card.Icon)
		if card.Icon.ItemType == 2 && card.Icon.Src != "" {
			candidates = append(candidates, card.Icon.Src)
		}
	}
	// Optional tables are absent in small embedded deployments/tests.
	if global.Db.Migrator().HasTable(&models.UserConfig{}) {
		var cfgs []models.UserConfig
		if err := global.Db.Find(&cfgs).Error; err != nil {
			return err
		}
		for _, cfg := range cfgs {
			references = append(references, cfg.PanelJson, cfg.SearchEngineJson)
		}
	}
	if global.Db.Migrator().HasTable(&models.ModuleConfig{}) {
		var cfgs []models.ModuleConfig
		if err := global.Db.Find(&cfgs).Error; err != nil {
			return err
		}
		for _, cfg := range cfgs {
			references = append(references, cfg.ValueJson)
		}
	}
	if global.Db.Migrator().HasTable(&models.SystemSetting{}) {
		var settings []models.SystemSetting
		if err := global.Db.Find(&settings).Error; err != nil {
			return err
		}
		for _, setting := range settings {
			references = append(references, setting.ConfigValue)
		}
	}
	var users []models.User
	if err := global.Db.Find(&users).Error; err != nil {
		return err
	}
	for _, user := range users {
		references = append(references, user.HeadImage)
	}
	if global.Db.Migrator().HasTable(&models.Notice{}) {
		var notices []models.Notice
		if err := global.Db.Find(&notices).Error; err != nil {
			return err
		}
		for _, notice := range notices {
			references = append(references, notice.Content)
		}
	}
	// Parse embedded JSON strings so escaped slashes do not hide references.
	var addStrings func(interface{})
	addStrings = func(value interface{}) {
		switch v := value.(type) {
		case string:
			sources = append(sources, v)
			paths = append(paths, v)
			references = append(references, v)
		case map[string]interface{}:
			for _, child := range v {
				addStrings(child)
			}
		case []interface{}:
			for _, child := range v {
				addStrings(child)
			}
		}
	}
	for _, reference := range append([]string{}, references...) {
		var value interface{}
		if json.Unmarshal([]byte(reference), &value) == nil {
			addStrings(value)
		} else {
			sources = append(sources, reference)
			paths = append(paths, reference)
		}
	}
	for _, candidate := range candidates {
		for _, reference := range references {
			if strings.Contains(reference, candidate) {
				sources = append(sources, candidate)
				break
			}
		}
	}
	if err := iconcache.Default.Prune(sources, paths, candidates); err != nil {
		return err
	}
	if global.Config == nil {
		return nil
	}
	root := global.Config.GetValueString("base", "source_path")
	absolute, err := filepath.Abs(root)
	if err != nil || root == "" {
		return err
	}
	realRoot, err := filepath.EvalSymlinks(absolute)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	prefix := "/" + strings.Trim(filepath.ToSlash(root), "./") + "/"
	for _, candidate := range candidates {
		u, err := url.Parse(candidate)
		if err != nil || u.Host != "" || u.Scheme != "" {
			continue
		}
		publicPath := "/" + strings.TrimPrefix(strings.TrimPrefix(u.Path, "./"), "/")
		if !strings.HasPrefix(publicPath, prefix) {
			continue
		}
		used := false
		for _, reference := range references {
			if strings.Contains(reference, publicPath) {
				used = true
				break
			}
		}
		if !used && global.Db.Migrator().HasTable(&models.File{}) {
			owners := []uint{}
			for _, card := range removed {
				owners = append(owners, card.UserId)
			}
			var foreign int64
			if err := global.Db.Model(&models.File{}).Where("src IN ? AND user_id NOT IN ?", []string{publicPath, "." + publicPath, strings.TrimPrefix(publicPath, "/")}, owners).Count(&foreign).Error; err != nil {
				return err
			}
			used = foreign > 0
		}
		if used {
			continue
		}
		relative := strings.TrimPrefix(publicPath, prefix)
		if relative == "" || strings.Contains(relative, "\\") || filepath.Clean(relative) == ".." || strings.HasPrefix(filepath.Clean(relative), ".."+string(filepath.Separator)) {
			continue
		}
		file := filepath.Join(realRoot, filepath.FromSlash(relative))
		realFile, err := filepath.EvalSymlinks(file)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil {
			rel, err := filepath.Rel(realRoot, realFile)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				continue
			}
			info, err := os.Stat(realFile)
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				continue
			}
			if err := os.Remove(realFile); err != nil {
				return err
			}
		}
		if global.Db.Migrator().HasTable(&models.File{}) {
			if err := global.Db.Where("src IN ?", []string{publicPath, "." + publicPath, strings.TrimPrefix(publicPath, "/")}).Delete(&models.File{}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func cleanupRemovedIcons(cards []models.ItemIcon) {
	if err := cleanupCardIcons(cards); err != nil && global.Logger != nil {
		global.Logger.Error("Icon reference cleanup failed: ", err)
	}
}
