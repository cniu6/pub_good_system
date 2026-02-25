package user

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// ProfileController 用户个人中心控制器（需要登录）
type ProfileController struct {
	user_svc *services.UserService
	auth_svc *services.AuthService
}

// NewProfileController 创建个人中心控制器
func NewProfileController() *ProfileController {
	return &ProfileController{
		user_svc: services.NewUserService(),
		auth_svc: services.NewAuthService(),
	}
}

// ========================================
// 请求结构体
// ========================================

type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   uint8  `json:"gender"`
	Birthday *int64 `json:"birthday"`
	Motto    string `json:"motto"`
	Mobile   string `json:"mobile"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ========================================
// 控制器方法
// ========================================

// GetProfile 获取个人信息
// @Summary 获取个人信息
// @Description 获取当前登录用户的个人信息
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/profile [get]
func (ctrl *ProfileController) GetProfile(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	user, err := ctrl.user_svc.GetByID(user_id.(uint64))
	if err != nil {
		utils.Fail(c, 404, "User not found")
		return
	}

	// 隐藏敏感信息
	user.Password = ""

	utils.Success(c, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"nickname":   user.Nickname,
		"avatar":     user.Avatar,
		"gender":     user.Gender,
		"birthday":   user.Birthday,
		"motto":      user.Motto,
		"mobile":     user.Mobile,
		"role":       user.Role,
		"status":     user.Status,
		"createTime": user.CreateTime,
	})
}

// UpdateProfile 更新个人信息
// @Summary 更新个人信息
// @Description 更新当前登录用户的个人信息
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "更新信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/profile [put]
func (ctrl *ProfileController) UpdateProfile(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入
	req.Nickname = utils.Clean_XSS(req.Nickname)
	req.Avatar = utils.Clean_XSS(req.Avatar)
	req.Motto = utils.Clean_XSS(req.Motto)
	req.Mobile = utils.Clean_XSS(req.Mobile)

	// 构建更新请求
	update_req := &services.UserUpdateRequest{
		ID:       user_id.(uint64),
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Gender:   req.Gender,
		Birthday: req.Birthday,
		Motto:    req.Motto,
		Mobile:   req.Mobile,
	}

	if err := ctrl.user_svc.Update(update_req); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "Profile updated successfully"})
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "密码信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/password [put]
func (ctrl *ProfileController) ChangePassword(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	if err := ctrl.auth_svc.ChangePassword(user_id.(uint64), req.OldPassword, req.NewPassword); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "Password changed successfully"})
}

// GetRoutes 获取用户路由
// @Summary 获取用户路由
// @Description 获取当前用户可访问的路由列表
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/routes [get]
func (ctrl *ProfileController) GetRoutes(c *gin.Context) {
	// TODO: 根据用户角色返回对应的路由
	// 目前返回空数组
	utils.Success(c, []interface{}{})
}

// GetSettings 获取用户设置
// @Summary 获取用户设置
// @Description 获取当前用户的个人设置
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/settings [get]
func (ctrl *ProfileController) GetSettings(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	user, err := ctrl.user_svc.GetByID(user_id.(uint64))
	if err != nil {
		utils.Fail(c, 404, "User not found")
		return
	}

	// 返回用户设置（可以扩展更多设置项）
	utils.Success(c, gin.H{
		"language":   "zh-CN", // 默认语言，可扩展
		"theme":      "light", // 默认主题
		"email":      user.Email,
		"mobile":     user.Mobile,
		"notifyEmail": true, // 邮件通知开关
	})
}

// UpdateAvatar 更新头像
// @Summary 更新头像
// @Description 更新当前用户的头像URL
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "头像URL {avatar: \"url\"}"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/avatar [put]
func (ctrl *ProfileController) UpdateAvatar(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	var req struct {
		Avatar string `json:"avatar" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入
	req.Avatar = utils.Clean_XSS(req.Avatar)

	update_req := &services.UserUpdateRequest{
		ID:     user_id.(uint64),
		Avatar: req.Avatar,
	}

	if err := ctrl.user_svc.Update(update_req); err != nil {
		utils.Fail(c, 500, "Failed to update avatar")
		return
	}

	utils.Success(c, gin.H{"message": "Avatar updated successfully", "avatar": req.Avatar})
}

// GetUserStats 获取用户统计
// @Summary 获取用户统计
// @Description 获取当前用户的统计数据
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/stats [get]
func (ctrl *ProfileController) GetUserStats(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	user, err := ctrl.user_svc.GetByID(user_id.(uint64))
	if err != nil {
		utils.Fail(c, 404, "User not found")
		return
	}

	// 返回用户统计信息
	utils.Success(c, gin.H{
		"joinTime":      user.JoinTime,
		"lastLoginTime": user.LastLoginTime,
		"lastLoginIp":   user.LastLoginIp,
		"loginCount":    0, // TODO: 可以添加登录次数统计
		"daysJoined":    calculateDaysJoined(user.JoinTime),
	})
}

// calculateDaysJoined 计算加入天数
func calculateDaysJoined(join_time *int64) int {
	if join_time == nil {
		return 0
	}
	join := time.Unix(*join_time, 0)
	return int(time.Since(join).Hours() / 24)
}

// RegisterRoutes 注册用户中心路由
func (ctrl *ProfileController) RegisterRoutes(group *gin.RouterGroup) {
	// 个人信息
	group.GET("/profile", ctrl.GetProfile)
	group.PUT("/profile", ctrl.UpdateProfile)

	// 密码
	group.PUT("/password", ctrl.ChangePassword)

	// 路由
	group.GET("/routes", ctrl.GetRoutes)

	// 设置
	group.GET("/settings", ctrl.GetSettings)

	// 头像
	group.PUT("/avatar", ctrl.UpdateAvatar)

	// 统计
	group.GET("/stats", ctrl.GetUserStats)
}

// ========================================
// 辅助：确保导入
// ========================================

var _ = models.User{}
