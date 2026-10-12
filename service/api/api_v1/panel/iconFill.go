package panel

import (
	"context"
	"encoding/json"
	"strings"
	"sun-panel/global"
	"sun-panel/lib/iconcache"
	"sun-panel/models"
	"sun-panel/models/datatype"
	"sync"
	"time"
)

// Cards saved without an icon (imports, or a browser lookup that failed or was
// skipped) get one discovered on the server; the next page load shows it.
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
		if card.ID == 0 || !missingIcon(card.Icon) || iconFill.pending[card.ID] || len(iconFill.pending) >= 256 || now.Before(iconFill.retry[card.Url]) {
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
	if err != nil || info.OriginalIconURL == "" {
		iconFill.Lock()
		iconFill.retry[site] = time.Now().Add(30 * time.Minute)
		iconFill.Unlock()
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
	if !missingIcon(card.Icon) {
		return
	}
	icon, _ := json.Marshal(datatype.ItemIconIconInfo{ItemType: 2, Src: info.OriginalIconURL})
	if err := global.Db.Model(&models.ItemIcon{}).Where("id=? AND user_id=?", id, user).UpdateColumn("icon_json", string(icon)).Error; err != nil && global.Logger != nil {
		global.Logger.Error("Saving discovered icon failed: ", err)
	}
}
