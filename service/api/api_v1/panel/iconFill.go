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
	queueFill(cards, false)
}

// Icons that may need replacing on an explicit refill: legacy online icons load
// from a third-party CDN, and remote images can disappear from their sites.
func checkableIcon(icon datatype.ItemIconIconInfo) bool {
	src := strings.ToLower(strings.TrimSpace(icon.Src))
	return icon.ItemType == 3 || (icon.ItemType == 2 && (strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://")))
}

// refillCards re-checks every card the user owns: missing details are looked up
// and dead or online icons are replaced, ignoring the failed-lookup backoff.
func refillCards(cards []models.ItemIcon) int {
	return queueFill(cards, true)
}

func queueFill(cards []models.ItemIcon, refill bool) int {
	cache := iconcache.Default
	if cache == nil || global.Db == nil {
		return 0
	}
	now := time.Now()
	iconFill.Lock()
	defer iconFill.Unlock()
	for site, until := range iconFill.retry {
		if now.After(until) {
			delete(iconFill.retry, site)
		}
	}
	queued := 0
	for _, card := range cards {
		wanted := incompleteCard(card) || (refill && checkableIcon(card.Icon))
		if card.ID == 0 || !wanted || iconFill.pending[card.ID] || len(iconFill.pending) >= 256 || (!refill && now.Before(iconFill.retry[card.Url])) {
			continue
		}
		iconFill.pending[card.ID] = true
		iconFill.wg.Add(1)
		queued++
		go fillIcon(cache, card.ID, card.UserId, card.Url, refill)
	}
	return queued
}

func fillIcon(cache *iconcache.Cache, id, user uint, site string, refill bool) {
	defer iconFill.wg.Done()
	defer func() {
		iconFill.Lock()
		delete(iconFill.pending, id)
		iconFill.Unlock()
	}()
	iconFill.slots <- struct{}{}
	defer func() { <-iconFill.slots }()
	var before models.ItemIcon
	if global.Db.Where("id=? AND user_id=?", id, user).First(&before).Error != nil {
		return
	}
	json.Unmarshal([]byte(before.IconJson), &before.Icon)
	// A refill replaces an online icon, or a remote image that no longer downloads.
	replaceIcon := false
	if refill && checkableIcon(before.Icon) {
		replaceIcon = before.Icon.ItemType == 3
		if !replaceIcon {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_, err := cache.Fetch(ctx, before.Icon.Src)
			cancel()
			replaceIcon = err != nil
		}
	}
	if !replaceIcon && !incompleteCard(before) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	info, err := cache.Discover(ctx, site)
	cancel()
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
	if (missingIcon(card.Icon) || (replaceIcon && card.IconJson == before.IconJson)) && info.OriginalIconURL != "" {
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
	if replaceIcon && changes["icon_json"] != nil {
		// The dead or online icon is no longer referenced by this card; cardMu is held,
		// as for edits, so cleanup cannot race with a save.
		cleanupRemovedIcons([]models.ItemIcon{before})
	}
}
