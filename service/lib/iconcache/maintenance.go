package iconcache

import (
	"encoding/json"
	"gorm.io/gorm"
	"sun-panel/models"
	"time"
)

// Discard abandoned URL previews after a day, retaining all live references.
// This also collects previews even when no later bookmark deletion occurs.
func (c *Cache) StartMaintenance(db *gorm.DB) {
	go func() {
		for {
			ReferenceMu.Lock()
			sources, paths, err := references(db)
			if err == nil {
				c.Prune(sources, paths, nil)
			}
			ReferenceMu.Unlock()
			time.Sleep(time.Hour)
		}
	}()
}

func references(db *gorm.DB) ([]string, []string, error) {
	var cards []models.ItemIcon
	var configs []models.UserConfig
	var modules []models.ModuleConfig
	var users []models.User
	var notices []models.Notice
	var settings []models.SystemSetting
	for _, table := range []interface{}{&cards, &configs, &modules, &users, &notices, &settings} {
		// Optional tables (notice is never migrated) would otherwise block every run.
		if !db.Migrator().HasTable(table) {
			continue
		}
		if err := db.Find(table).Error; err != nil {
			return nil, nil, err
		}
	}
	raw := []string{}
	for _, card := range cards {
		raw = append(raw, card.IconJson, card.ExpandParam)
	}
	for _, config := range configs {
		raw = append(raw, config.PanelJson, config.SearchEngineJson)
	}
	for _, module := range modules {
		raw = append(raw, module.ValueJson)
	}
	for _, user := range users {
		raw = append(raw, user.HeadImage)
	}
	for _, notice := range notices {
		raw = append(raw, notice.Content)
	}
	for _, setting := range settings {
		raw = append(raw, setting.ConfigValue)
	}
	values := []string{}
	var walk func(interface{})
	walk = func(value interface{}) {
		switch v := value.(type) {
		case string:
			values = append(values, v)
		case map[string]interface{}:
			for _, child := range v {
				walk(child)
			}
		case []interface{}:
			for _, child := range v {
				walk(child)
			}
		}
	}
	for _, value := range raw {
		var parsed interface{}
		if json.Unmarshal([]byte(value), &parsed) == nil {
			walk(parsed)
		} else {
			values = append(values, value)
		}
	}
	return values, values, nil
}
