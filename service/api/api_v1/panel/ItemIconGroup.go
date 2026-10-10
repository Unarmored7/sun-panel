package panel

import (
	"encoding/json"
	"strings"
	"sun-panel/api/api_v1/common/apiData/commonApiStructs"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/models"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type ItemIconGroup struct {
}

// Optional settings preserve visibility when older clients edit only a title.
type groupEditRequest struct {
	ID              uint    `json:"id"`
	Title           string  `json:"title"`
	Icon            string  `json:"icon"`
	Description     string  `json:"description"`
	Sort            *int    `json:"sort"`
	PublicVisitHide *int    `json:"publicVisitHide"`
	CardStyle       *string `json:"cardStyle"`
}

func (a *ItemIconGroup) Edit(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	req := groupEditRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorCode(c, -1, "Invalid group request", nil)
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" || utf8.RuneCountInString(req.Title) > 50 {
		apiReturn.ErrorCode(c, -1, "Group title must contain 1 to 50 characters", nil)
		return
	}
	if req.PublicVisitHide != nil && *req.PublicVisitHide != 0 && *req.PublicVisitHide != 1 {
		apiReturn.ErrorCode(c, -1, "Invalid group visibility", nil)
		return
	}
	if req.CardStyle != nil && *req.CardStyle != "" {
		var style struct {
			Style *int `json:"style"`
		}
		if len(*req.CardStyle) > 2000 || json.Unmarshal([]byte(*req.CardStyle), &style) != nil ||
			(style.Style != nil && (*style.Style < -1 || *style.Style > 1)) {
			apiReturn.ErrorCode(c, -1, "Invalid card style", nil)
			return
		}
	}
	group := models.ItemIconGroup{Title: req.Title, Icon: req.Icon, Description: req.Description, UserId: userInfo.ID}
	if req.ID != 0 {
		if err := global.Db.Where("id=? AND user_id=?", req.ID, userInfo.ID).First(&group).Error; err != nil {
			apiReturn.ErrorCode(c, 1005, "Group is not available", nil)
			return
		}
	}
	group.Title = req.Title
	group.Icon = req.Icon
	group.Description = req.Description
	if req.Sort != nil {
		group.Sort = *req.Sort
	}
	if req.PublicVisitHide != nil {
		group.PublicVisitHide = *req.PublicVisitHide
	}
	if req.CardStyle != nil {
		group.CardStyle = *req.CardStyle
	}
	var err error
	if req.ID == 0 {
		err = global.Db.Create(&group).Error
	} else {
		err = global.Db.Model(&group).Where("user_id=?", userInfo.ID).
			Select("title", "icon", "description", "sort", "public_visit_hide", "card_style").Updates(&group).Error
	}
	if err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	apiReturn.SuccessData(c, group)
}

func (a *ItemIconGroup) GetList(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	groups := []models.ItemIconGroup{}
	public := base.GetCurrentVisitMode(c) == base.VISIT_MODE_PUBLIC
	err := global.Db.Transaction(func(tx *gorm.DB) error {
		query := tx.Order("sort,created_at").Where("user_id=?", userInfo.ID)
		if public {
			query = query.Where("public_visit_hide=0 OR public_visit_hide IS NULL")
		}
		if err := query.Find(&groups).Error; err != nil {
			return err
		}
		// Reading an empty public page must never create a group or move private cards.
		if len(groups) == 0 && !public {
			defaultGroup := models.ItemIconGroup{Title: "APP", UserId: userInfo.ID, Icon: "material-symbols:ad-group-outline"}
			if err := tx.Create(&defaultGroup).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.ItemIcon{}).Where("user_id=? AND (item_icon_group_id=0 OR item_icon_group_id IS NULL)", userInfo.ID).
				Update("item_icon_group_id", defaultGroup.ID).Error; err != nil {
				return err
			}
			groups = append(groups, defaultGroup)
		}
		return nil
	})
	if err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	apiReturn.SuccessListData(c, groups, 0)
}

func (a *ItemIconGroup) Deletes(c *gin.Context) {
	req := commonApiStructs.RequestDeleteIds[uint]{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	userInfo, _ := base.GetCurrentUserInfo(c)

	var count int64
	if err := global.Db.Model(&models.ItemIconGroup{}).Where(" user_id=?", userInfo.ID).Count(&count).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	} else {
		var selected int64
		if err := global.Db.Model(&models.ItemIconGroup{}).Where("user_id=? AND id in ?", userInfo.ID, req.Ids).Count(&selected).Error; err != nil {
			apiReturn.ErrorDatabase(c, err.Error())
			return
		}
		if selected >= count && count > 0 {
			apiReturn.ErrorCode(c, 1201, "At least one must be retained", nil)
			return
		}

	}

	txErr := global.Db.Transaction(func(tx *gorm.DB) error {
		mitemIcon := models.ItemIcon{}
		if err := tx.Delete(&models.ItemIconGroup{}, "id in ? AND user_id=?", req.Ids, userInfo.ID).Error; err != nil {
			return err
		}

		if err := mitemIcon.DeleteByItemIconGroupIds(tx, userInfo.ID, req.Ids); err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		apiReturn.ErrorDatabase(c, txErr.Error())
		return
	}

	apiReturn.Success(c)
}

// 保存排序
func (a *ItemIconGroup) SaveSort(c *gin.Context) {
	req := commonApiStructs.SortRequest{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)

	transactionErr := global.Db.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		for _, v := range req.SortItems {
			if err := tx.Model(&models.ItemIconGroup{}).Where("user_id=? AND id=?", userInfo.ID, v.Id).Update("sort", v.Sort).Error; err != nil {
				// 返回任何错误都会回滚事务
				return err
			}
		}

		// 返回 nil 提交事务
		return nil
	})

	if transactionErr != nil {
		apiReturn.ErrorDatabase(c, transactionErr.Error())
		return
	}

	apiReturn.Success(c)
}
