package admin

import (
	"fst/backend/app/models"
	"fst/backend/utils"
	"log"

	"github.com/gin-gonic/gin"
)

// UserLevelController 用户等级能力管理
type UserLevelController struct{}

func NewUserLevelController() *UserLevelController {
	return &UserLevelController{}
}

// List GET /user-levels
func (ctrl *UserLevelController) List(c *gin.Context) {
	list, err := models.ListUserLevelCaps()
	if err != nil {
		log.Printf("[ADMIN][UserLevel] list failed: %v", err)
		utils.Fail(c, 500, "获取用户等级失败")
		return
	}
	utils.Success(c, gin.H{"list": list})
}

type updateUserLevelBody struct {
	Level         uint64 `json:"level" binding:"required"`
	Name          string `json:"name"`
	AllowAPIKey   *bool  `json:"allow_api_key"`
	AllowRecharge *bool  `json:"allow_recharge"`
	AllowWithdraw *bool  `json:"allow_withdraw"`
	MenuFlags     string `json:"menu_flags"`
}

// Update PUT /user-levels
func (ctrl *UserLevelController) Update(c *gin.Context) {
	var req updateUserLevelBody
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	if req.Level == 0 {
		utils.Fail(c, 400, "无效等级")
		return
	}
	allowAPIKey := true
	allowRecharge := true
	allowWithdraw := true
	if req.AllowAPIKey != nil {
		allowAPIKey = *req.AllowAPIKey
	}
	if req.AllowRecharge != nil {
		allowRecharge = *req.AllowRecharge
	}
	if req.AllowWithdraw != nil {
		allowWithdraw = *req.AllowWithdraw
	}
	menuFlags := req.MenuFlags
	if menuFlags == "" {
		menuFlags = "{}"
	}
	if err := models.UpdateUserLevelCap(req.Level, req.Name, allowAPIKey, allowRecharge, allowWithdraw, menuFlags); err != nil {
		log.Printf("[ADMIN][UserLevel] update failed level=%d: %v", req.Level, err)
		utils.Fail(c, 500, "更新失败")
		return
	}
	cap, _ := models.GetUserLevelCap(req.Level)
	utils.Success(c, gin.H{"item": cap})
}

// RegisterRoutes 注册用户等级路由（管理端 AdminOnly 已保护）
func (ctrl *UserLevelController) RegisterRoutes(adminGroup *gin.RouterGroup) {
	g := adminGroup.Group("/user-levels")
	{
		g.GET("", ctrl.List)
		g.PUT("", ctrl.Update)
	}
}
