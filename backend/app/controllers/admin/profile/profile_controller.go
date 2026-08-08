package profile

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// ProfileController 管理端个人设置（改密）
type ProfileController struct {
	authSvc *services.AuthService
}

func NewProfileController() *ProfileController {
	return &ProfileController{authSvc: services.NewAuthService()}
}

func adminUID(c *gin.Context) (uint64, bool) {
	v, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok && id > 0
}

// Me GET /me 当前管理员简要信息
// @Summary 当前管理员信息
// @Tags Admin-设置
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/me [get]
func (ctrl *ProfileController) Me(c *gin.Context) {
	uid, ok := adminUID(c)
	if !ok {
		utils.Fail(c, 401, "Unauthorized")
		return
	}
	user, err := models.GetUserByID(uid)
	if err != nil || user == nil {
		utils.Fail(c, 404, "用户不存在")
		return
	}
	utils.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"nickname": user.Nickname,
		"role":     user.Role,
	})
}

// ChangePassword PUT /me/password
// @Summary 修改管理员密码
// @Tags Admin-设置
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/me/password [put]
func (ctrl *ProfileController) ChangePassword(c *gin.Context) {
	uid, ok := adminUID(c)
	if !ok {
		utils.Fail(c, 401, "Unauthorized")
		return
	}
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	if len(req.NewPassword) < 6 {
		utils.Fail(c, 400, "新密码至少 6 位")
		return
	}
	if err := ctrl.authSvc.ChangePassword(uid, req.OldPassword, req.NewPassword); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.SuccessMsg(c, "密码已修改，请重新登录", nil)
}

// RegisterRoutes 注册个人设置路由
func (ctrl *ProfileController) RegisterRoutes(adminGroup *gin.RouterGroup) {
	me := adminGroup.Group("/me")
	me.Use(middleware.SimpleLogMiddleware("管理员个人设置"))
	{
		me.GET("", ctrl.Me)
		me.PUT("/password", ctrl.ChangePassword)
	}
}
