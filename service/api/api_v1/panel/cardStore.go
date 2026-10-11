package panel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"net/url"
	"strings"
	"sun-panel/lib/iconcache"
	"sun-panel/models"
)

// Serializes mutations and reference cleanup, including concurrent imports.
var cardMu = &iconcache.ReferenceMu
var errCardAccess = errors.New("Card or group is not available")

func canonicalURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") || len(raw) > 1000 {
		return "", errors.New("Invalid HTTP(S) URL")
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	if (u.Scheme == "https" && u.Port() == "443") || (u.Scheme == "http" && u.Port() == "80") {
		u.Host = strings.TrimSuffix(u.Host, ":"+u.Port())
	}
	if u.Path == "/" {
		u.Path = ""
		u.RawPath = ""
	}
	return u.String(), nil
}

func cardURLKey(raw string) string {
	normalized, err := canonicalURL(raw)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

func validateCardOwner(tx *gorm.DB, card *models.ItemIcon, user uint) error {
	var group models.ItemIconGroup
	if tx.Where("id=? AND user_id=?", card.ItemIconGroupId, user).First(&group).Error != nil {
		return errCardAccess
	}
	if card.ID != 0 {
		var previous models.ItemIcon
		if tx.Where("id=? AND user_id=?", card.ID, user).First(&previous).Error != nil {
			return errCardAccess
		}
	}
	return nil
}

// Legacy rows have NULL keys. Resolve duplicates when that URL is next saved;
// startup never changes existing bookmarks or their source URLs.
func saveCard(tx *gorm.DB, card *models.ItemIcon, user uint, old *[]models.ItemIcon) error {
	if err := validateCardOwner(tx, card, user); err != nil {
		return err
	}
	normalized, err := canonicalURL(card.Url)
	if err != nil {
		return err
	}
	key := cardURLKey(normalized)
	var possible []models.ItemIcon
	if err := tx.Where("user_id=? AND (url_key=? OR url_key IS NULL OR id=?)", user, key, card.ID).Order("updated_at DESC,id DESC").Find(&possible).Error; err != nil {
		return err
	}
	var selected models.ItemIcon
	for _, previous := range possible {
		matches := cardURLKey(previous.Url) == key
		if previous.ID == card.ID || (card.ID == 0 && selected.ID == 0 && matches) {
			selected = previous
		}
		if matches || previous.ID == card.ID {
			*old = append(*old, previous)
		}
	}
	if card.ID == 0 && selected.ID != 0 {
		card.ID = selected.ID
	}
	for _, previous := range possible {
		if previous.ID != card.ID && cardURLKey(previous.Url) == key {
			if err := tx.Model(&models.ItemIcon{}).Where("id=? AND user_id=?", previous.ID, user).Update("url_key", nil).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.ItemIcon{}, "id=? AND user_id=?", previous.ID, user).Error; err != nil {
				return err
			}
		}
	}
	card.UserId = user
	card.User = models.User{}
	card.Url = normalized
	card.URLKey = &key
	card.Icon.CachedSrc = ""
	if strings.TrimSpace(card.Title) == "" {
		u, _ := url.Parse(normalized)
		card.Title = u.Hostname()
	}
	icon, err := json.Marshal(card.Icon)
	if err != nil {
		return err
	}
	card.IconJson = string(icon)
	if card.ID == 0 {
		if card.Sort == 0 {
			card.Sort = 9999
		}
		return tx.Create(card).Error
	}
	card.CreatedAt = selected.CreatedAt
	if card.Sort == 0 {
		card.Sort = selected.Sort
	}
	// Extensions from newer databases survive ordinary bookmark edits.
	card.CardType = selected.CardType
	card.OnlyName = selected.OnlyName
	card.BackgroundColor = selected.BackgroundColor
	card.ExpandParam = selected.ExpandParam
	return tx.Model(&models.ItemIcon{}).Where("id=? AND user_id=?", card.ID, user).
		Select("IconJson", "Title", "Url", "LanUrl", "Description", "OpenMethod", "Sort", "ItemIconGroupId", "URLKey").Updates(card).Error
}

func readCardIcon(card *models.ItemIcon) {
	json.Unmarshal([]byte(card.IconJson), &card.Icon)
	card.Icon.CachedSrc = iconcache.Default.Resolve(card.Icon.Src)
}
