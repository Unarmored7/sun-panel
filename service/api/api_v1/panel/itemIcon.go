package panel

import (
	"context"
	"errors"
	"sun-panel/api/api_v1/common/apiData/commonApiStructs"
	"sun-panel/api/api_v1/common/apiData/panelApiStructs"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/lib/iconcache"
	"sun-panel/lib/remoteimage"
	"sun-panel/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type ItemIcon struct {
}

func (a *ItemIcon) Edit(c *gin.Context) {
	user, _ := base.GetCurrentUserInfo(c)
	var req models.ItemIcon
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	cardMu.Lock()
	defer cardMu.Unlock()
	old := []models.ItemIcon{}
	err := global.Db.Transaction(func(tx *gorm.DB) error { return saveCard(tx, &req, user.ID, &old) })
	if err != nil {
		if err == errCardAccess {
			apiReturn.ErrorCode(c, 1005, err.Error(), nil)
		} else {
			apiReturn.Error(c, err.Error())
		}
		return
	}
	cleanupRemovedIcons(old)
	readCardIcon(&req)
	apiReturn.SuccessData(c, req)
}

func (a *ItemIcon) AddMultiple(c *gin.Context) {
	user, _ := base.GetCurrentUserInfo(c)
	var req []models.ItemIcon
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	cardMu.Lock()
	defer cardMu.Unlock()
	old := []models.ItemIcon{}
	err := global.Db.Transaction(func(tx *gorm.DB) error {
		// Validate the complete import before touching any rows.
		for i := range req {
			if req[i].ID != 0 {
				return errors.New("New cards cannot specify an ID")
			}
			if err := validateCardOwner(tx, &req[i], user.ID); err != nil {
				return err
			}
		}
		for i := range req {
			if err := saveCard(tx, &req[i], user.ID, &old); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if err == errCardAccess {
			apiReturn.ErrorCode(c, 1005, err.Error(), nil)
		} else {
			apiReturn.Error(c, err.Error())
		}
		return
	}
	cleanupRemovedIcons(old)
	apiReturn.SuccessData(c, req)
}

// // 获取详情
// func (a *ItemIcon) GetInfo(c *gin.Context) {
// 	req := systemApiStructs.AiDrawGetInfoReq{}

// 	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
// 		apiReturn.ErrorParamFomat(c, err.Error())
// 		return
// 	}

// 	userInfo, _ := base.GetCurrentUserInfo(c)

// 	aiDraw := models.AiDraw{}
// 	aiDraw.ID = req.ID
// 	if err := aiDraw.GetInfo(global.Db); err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			apiReturn.Error(c, "不存在记录")
// 			return
// 		}
// 		apiReturn.ErrorDatabase(c, err.Error())
// 		return
// 	}

// 	if userInfo.ID != aiDraw.UserID {
// 		apiReturn.ErrorNoAccess(c)
// 		return
// 	}

// 	apiReturn.SuccessData(c, aiDraw)
// }

func (a *ItemIcon) GetListByGroupId(c *gin.Context) {
	cardMu.Lock()
	defer cardMu.Unlock()
	req := models.ItemIcon{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)
	var group models.ItemIconGroup
	query := global.Db.Where("id=? AND user_id=?", req.ItemIconGroupId, userInfo.ID)
	if base.GetCurrentVisitMode(c) == base.VISIT_MODE_PUBLIC {
		query = query.Where("public_visit_hide=0 OR public_visit_hide IS NULL")
	}
	if err := query.First(&group).Error; err != nil {
		apiReturn.ErrorCode(c, 1005, "Group is not available", nil)
		return
	}
	itemIcons := []models.ItemIcon{}

	if err := global.Db.Order("sort ,created_at").Find(&itemIcons, "item_icon_group_id = ? AND user_id=?", req.ItemIconGroupId, userInfo.ID).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	for k := range itemIcons {
		readCardIcon(&itemIcons[k])
	}

	apiReturn.SuccessListData(c, itemIcons, 0)
}

func (a *ItemIcon) Deletes(c *gin.Context) {
	var req commonApiStructs.RequestDeleteIds[uint]
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	user, _ := base.GetCurrentUserInfo(c)
	cardMu.Lock()
	defer cardMu.Unlock()
	var old []models.ItemIcon
	err := global.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id IN ? AND user_id=?", req.Ids, user.ID).Find(&old).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.ItemIcon{}).Where("id IN ? AND user_id=?", req.Ids, user.ID).Update("url_key", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&models.ItemIcon{}, "id IN ? AND user_id=?", req.Ids, user.ID).Error
	})
	if err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	cleanupRemovedIcons(old)
	apiReturn.Success(c)
}

// 保存排序
func (a *ItemIcon) SaveSort(c *gin.Context) {
	req := panelApiStructs.ItemIconSaveSortRequest{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)

	transactionErr := global.Db.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		for _, v := range req.SortItems {
			if err := tx.Model(&models.ItemIcon{}).Where("user_id=? AND id=? AND item_icon_group_id=?", userInfo.ID, v.Id, req.ItemIconGroupId).Update("sort", v.Sort).Error; err != nil {
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

// 支持获取并直接下载对方网站图标到服务器
func (a *ItemIcon) GetSiteFavicon(c *gin.Context) {
	var req panelApiStructs.ItemIconGetSiteFaviconReq
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	normalized, err := canonicalURL(req.Url)
	if err != nil {
		apiReturn.Error(c, err.Error())
		return
	}
	if _, err := remoteimage.ValidateURL(normalized); err != nil {
		apiReturn.Error(c, err.Error())
		return
	}
	if iconcache.Default == nil {
		apiReturn.Error(c, "Icon cache is unavailable")
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 12*time.Second)
	defer cancel()
	info, _ := iconcache.Default.Discover(ctx, normalized)
	// An unreachable site's URL remains usable; the UI falls back to its initial.
	apiReturn.SuccessData(c, info)
}
