package panel

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"sun-panel/global"
	"sun-panel/lib/iconcache"
	"sun-panel/models"
	"sun-panel/models/datatype"
	"sync"
	"time"
)

// Cards saved without an icon, name or description (imports, or a browser
// lookup that failed or was skipped) get them discovered on the server; the
// next page load shows them.
var iconFill = struct {
	sync.Mutex
	wg      sync.WaitGroup
	pending map[uint]bool
	retry   map[string]time.Time
	slots   chan struct{}
}{pending: map[uint]bool{}, retry: map[string]time.Time{}, slots: make(chan struct{}, 2)}

func missingIcon(icon datatype.ItemIconIconInfo) bool {
	return strings.TrimSpace(icon.Src) == "" && strings.TrimSpace(icon.Text) == ""
}

// Names that are only the address fallback from the browser or saveCard.
func hostnameTitle(card models.ItemIcon) bool {
	u, err := url.Parse(card.Url)
	if err != nil {
		return false
	}
	host := u.Hostname()
	title := strings.TrimSpace(card.Title)
	return title == "" || title == host || title == clipRunes(strings.TrimPrefix(host, "www."), 20)
}

func incompleteCard(card models.ItemIcon) bool {
	return missingIcon(card.Icon) || hostnameTitle(card) || strings.TrimSpace(card.Description) == ""
}

func clipRunes(value string, limit int) string {
	if runes := []rune(value); len(runes) > limit {
		return string(runes[:limit])
	}
	return value
}

func fillMissingIcons(cards []models.ItemIcon) {
	cache := iconcache.Default
	if cache == nil || global.Db == nil {
		return
	}
	now := time.Now()
	iconFill.Lock()
	defer iconFill.Unlock()
	for site, until := range iconFill.retry {
		if now.After(until) {
			delete(iconFill.retry, site)
		}
	}
	for _, card := range cards {
		if card.ID == 0 || !incompleteCard(card) || iconFill.pending[card.ID] || len(iconFill.pending) >= 256 || now.Before(iconFill.retry[card.Url]) {
			continue
		}
		iconFill.pending[card.ID] = true
		iconFill.wg.Add(1)
		go fillIcon(cache, card.ID, card.UserId, card.Url)
	}
}

func fillIcon(cache *iconcache.Cache, id, user uint, site string) {
	defer iconFill.wg.Done()
	defer func() {
		iconFill.Lock()
		delete(iconFill.pending, id)
		iconFill.Unlock()
	}()
	iconFill.slots <- struct{}{}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	info, err := cache.Discover(ctx, site)
	cancel()
	<-iconFill.slots
	// Successful lookups also wait: a site without a description is not re-read on every save.
	iconFill.Lock()
	iconFill.retry[site] = time.Now().Add(30 * time.Minute)
	iconFill.Unlock()
	if err != nil && info.Title == "" {
		return
	}
	cardMu.Lock()
	defer cardMu.Unlock()
	// The card may have been edited or deleted while the site was queried.
	var card models.ItemIcon
	if global.Db.Where("id=? AND user_id=?", id, user).First(&card).Error != nil || card.Url != site {
		return
	}
	json.Unmarshal([]byte(card.IconJson), &card.Icon)
	changes := map[string]interface{}{}
	if missingIcon(card.Icon) && info.OriginalIconURL != "" {
		icon, _ := json.Marshal(datatype.ItemIconIconInfo{ItemType: 2, Src: info.OriginalIconURL})
		changes["icon_json"] = string(icon)
	}
	if hostnameTitle(card) && info.Title != "" && info.Title != card.Title {
		changes["title"] = clipRunes(info.Title, 20)
	}
	if strings.TrimSpace(card.Description) == "" && info.Description != "" {
		changes["description"] = clipRunes(info.Description, 100)
	}
	if len(changes) == 0 {
		return
	}
	if err := global.Db.Model(&models.ItemIcon{}).Where("id=? AND user_id=?", id, user).UpdateColumns(changes).Error; err != nil && global.Logger != nil {
		global.Logger.Error("Saving discovered site details failed: ", err)
	}
}
