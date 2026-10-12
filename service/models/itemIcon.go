package models

import (
	"sun-panel/models/datatype"

	"gorm.io/gorm"
)

type ItemIcon struct {
	BaseModel
	CardType        int                       `gorm:"type:tinyint(1);default:1" json:"cardType"`
	OnlyName        string                    `gorm:"type:varchar(200)" json:"onlyName"`
	BackgroundColor string                    `gorm:"type:varchar(20)" json:"backgroundColor"`
	ExpandParam     string                    `gorm:"type:text" json:"expandParam"`
	IconJson        string                    `gorm:"type:varchar(1000)" json:"-"`
	URLKey          *string                   `gorm:"type:char(64);uniqueIndex:idx_card_user_url" json:"-"`
	Icon            datatype.ItemIconIconInfo `gorm:"-" json:"icon"`
	Title           string                    `gorm:"type:varchar(50)" json:"title"`
	Url             string                    `gorm:"type:varchar(1000)" json:"url"`
	LanUrl          string                    `gorm:"type:varchar(1000)" json:"-"` // Retired LAN address, kept in the database only
	Description     string                    `gorm:"type:varchar(1000)" json:"description"`
	OpenMethod      int                       `gorm:"type:tinyint(1)" json:"openMethod"`
	Sort            int                       `gorm:"type:int(11)" json:"sort"`
	ItemIconGroupId int                       `json:"itemIconGroupId"`
	UserId          uint                      `gorm:"uniqueIndex:idx_card_user_url" json:"userId"`
	User            User                      `json:"-"`
}

func (m *ItemIcon) DeleteByItemIconGroupIds(db *gorm.DB, userId uint, itemIconGroupIds []uint) (err error) {
	if err = db.Model(&ItemIcon{}).Where("item_icon_group_id IN ? AND user_id=?", itemIconGroupIds, userId).Update("url_key", nil).Error; err != nil {
		return err
	}
	err = db.Delete(&ItemIcon{}, "item_icon_group_id in ? AND user_id=?", itemIconGroupIds, userId).Error
	return
}

func (m *ItemIcon) DeleteByUserId(db *gorm.DB, userId uint) (err error) {
	if err := db.Model(&ItemIcon{}).Where("user_id=?", userId).Update("url_key", nil).Error; err != nil {
		return err
	}
	return db.Delete(&ItemIcon{}, "user_id=?", userId).Error
}
