package models

import (
	"gorm.io/gorm"
)

type ItemIconGroup struct {
	BaseModel
	OnlyName        string `gorm:"type:varchar(50)" json:"onlyName"`
	Icon            string `json:"icon"`
	Title           string `gorm:"type:varchar(50)" json:"title"`
	Description     string `gorm:"type:varchar(1000)" json:"description"`
	Sort            int    `gorm:"type:int(11)" json:"sort"`
	PublicVisitHide int    `gorm:"type:tinyint(1);not null;default:0" json:"publicVisitHide"`
	CardStyle       string `gorm:"type:text" json:"cardStyle"`
	UserId          uint   `json:"userId"`
	User            User   `json:"user"`
}

func (m *ItemIconGroup) DeleteByUserId(db *gorm.DB, userId uint) (err error) {
	err = db.Delete(&ItemIconGroup{}, "user_id = ?", userId).Error
	return
}
