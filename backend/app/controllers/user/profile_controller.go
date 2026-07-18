package user

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"
	"log"

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
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Gender     *uint8 `json:"gender"`
	Birthday   *int64 `json:"birthday"`
	Motto      string `json:"motto"`
	Mobile     string `json:"mobile"`
	BackGround string `json:"back_ground"`
	Language   string `json:"language"`
	Country    string `json:"country"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ProfileRealnameSummary struct {
	HasVerification bool   `json:"hasVerification"`
	ID              uint64 `json:"id,omitempty"`
	Status          uint8  `json:"status,omitempty"`
	RealName        string `json:"realName,omitempty"`
	CertificateType uint8  `json:"certificateType,omitempty"`
	CertificateNo   string `json:"certificateNo,omitempty"`
	SubmittedAt     *int64 `json:"submittedAt,omitempty"`
	ReviewedAt      *int64 `json:"reviewedAt,omitempty"`
	RejectReason    string `json:"rejectReason,omitempty"`
}

type ProfileResponse struct {
	ID            uint64                  `json:"id"`
	GroupID       uint64                  `json:"groupId"`
	Username      string                  `json:"username"`
	UserName      string                  `json:"userName"`
	Email         string                  `json:"email"`
	Nickname      string                  `json:"nickname"`
	Avatar        string                  `json:"avatar"`
	BackGroundRaw string                  `json:"back_ground"`
	BackGround    string                  `json:"backGround"`
	Gender        uint8                   `json:"gender"`
	Birthday      *int64                  `json:"birthday"`
	Motto         string                  `json:"motto"`
	Mobile        string                  `json:"mobile"`
	Money         float64                 `json:"money"`
	Score         int64                   `json:"score"`
	Level         uint64                  `json:"level"`
	Role          string                  `json:"role"`
	Status        uint8                   `json:"status"`
	Language      string                  `json:"language"`
	Country       string                  `json:"country"`
	LoginFailure  uint8                   `json:"loginFailure"`
	JoinTime      *int64                  `json:"joinTime"`
	JoinIP        string                  `json:"joinIp"`
	LastLoginTime *int64                  `json:"lastLoginTime"`
	LastLoginIP   string                  `json:"lastLoginIp"`
	UpdateTime    *int64                  `json:"updateTime"`
	CreateTime    *int64                  `json:"createTime"`
	Realname      ProfileRealnameSummary  `json:"realname"`
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
// @Success 200 {object} utils.Response{data=ProfileResponse}
// @Router /api/v1/user/profile [get]
func (ctrl *ProfileController) GetProfile(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	user, err := ctrl.user_svc.GetByID(uid)
	if err != nil {
		utils.Fail(c, 404, "User not found")
		return
	}

	// 隐藏敏感信息
	user.Password = ""

	realnameSummary := ProfileRealnameSummary{
		HasVerification: false,
	}
	if verification, err := models.GetRealnameVerificationByUserID(uid); err == nil && verification != nil {
		realnameSummary = ProfileRealnameSummary{
			HasVerification: true,
			ID:              verification.ID,
			Status:          verification.Status,
			RealName:        verification.RealName,
			CertificateType: verification.CertificateType,
			CertificateNo:   verification.CertificateNo,
			SubmittedAt:     verification.SubmittedAt,
			ReviewedAt:      verification.ReviewedAt,
			RejectReason:    verification.RejectReason,
		}
	}

	utils.Success(c, ProfileResponse{
		ID:            user.ID,
		GroupID:       user.GroupId,
		Username:      user.Username,
		UserName:      user.Username,
		Email:         user.Email,
		Nickname:      user.Nickname,
		Avatar:        user.Avatar,
		BackGroundRaw: user.BackGround,
		BackGround:    user.BackGround,
		Gender:        user.Gender,
		Birthday:      user.Birthday,
		Motto:         user.Motto,
		Mobile:        user.Mobile,
		Money:         user.Money,
		Score:         user.Score,
		Level:         user.Level,
		Role:          user.Role,
		Status:        user.Status,
		Language:      user.Language,
		Country:       user.Country,
		LoginFailure:  user.LoginFailure,
		JoinTime:      user.JoinTime,
		JoinIP:        user.JoinIp,
		LastLoginTime: user.LastLoginTime,
		LastLoginIP:   user.LastLoginIp,
		UpdateTime:    user.UpdateTime,
		CreateTime:    user.CreateTime,
		Realname:      realnameSummary,
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
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
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
	req.BackGround = utils.Clean_XSS(req.BackGround)
	req.Language = utils.Clean_XSS(req.Language)
	req.Country = utils.Clean_XSS(req.Country)

	// URL 字段校验：仅允许 http/https 协议
	if req.Avatar != "" && !utils.ValidateURL(req.Avatar) {
		utils.Fail(c, 400, "头像 URL 仅支持 http/https 协议")
		return
	}
	if req.BackGround != "" && !utils.ValidateURL(req.BackGround) {
		utils.Fail(c, 400, "背景图 URL 仅支持 http/https 协议")
		return
	}

	// 长度限制
	if len(req.Nickname) > 50 {
		utils.Fail(c, 400, "昵称长度不能超过50个字符")
		return
	}
	if len(req.Avatar) > 500 {
		utils.Fail(c, 400, "头像URL长度不能超过500个字符")
		return
	}
	if len(req.BackGround) > 500 {
		utils.Fail(c, 400, "背景图URL长度不能超过500个字符")
		return
	}
	if len(req.Motto) > 200 {
		utils.Fail(c, 400, "个性签名长度不能超过200个字符")
		return
	}

	// 构建更新请求
	update_req := &services.UserUpdateRequest{
		ID:       uid,
		Gender:   req.Gender,
		Birthday: req.Birthday,
	}
	update_req.Nickname = &req.Nickname
	update_req.Avatar = &req.Avatar
	update_req.Motto = &req.Motto
	update_req.Mobile = &req.Mobile
	update_req.BackGround = &req.BackGround
	update_req.Language = &req.Language
	update_req.Country = &req.Country

	if err := ctrl.user_svc.Update(update_req); err != nil {
		if services.IsClientError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		log.Printf("[PROFILE] update profile failed for user_id=%d: %v", uid, err)
		utils.Fail(c, 500, "Failed to update profile")
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
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	if err := ctrl.auth_svc.ChangePassword(uid, req.OldPassword, req.NewPassword); err != nil {
		if utils.IsPasswordValidationError(err) || services.IsClientError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		log.Printf("[PROFILE] change password failed for user_id=%d: %v", uid, err)
		utils.Fail(c, 500, "Failed to change password")
		return
	}

	utils.Success(c, gin.H{"message": "Password changed successfully"})
}

// ========================================
// 用户设置相关请求结构体
// ========================================

type UpdateSettingsRequest struct {
	Language    string `json:"language"`
	Theme       string `json:"theme"`
	NotifyEmail *bool  `json:"notify_email"`
}

// ========================================
// 用户设置
// ========================================

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
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	user, err := ctrl.user_svc.GetByID(uid)
	if err != nil {
		utils.Fail(c, 404, "User not found")
		return
	}

	// 从 user_settings 表获取设置
	settings, _ := models.GetUserSettings(user.ID)

	// 合并默认值
	language := user.Language
	if language == "" {
		language = "zh-CN"
	}
	theme := "light"
	notify_email := true
	if settings != nil {
		if settings.Theme != "" {
			theme = settings.Theme
		}
		notify_email = settings.NotifyEmail
	}

	utils.Success(c, gin.H{
		"language":     language,
		"theme":        theme,
		"notify_email": notify_email,
		"email":        user.Email,
		"mobile":       user.Mobile,
		"country":      user.Country,
	})
}

// UpdateSettings 更新用户设置
// @Summary 更新用户设置
// @Description 更新当前用户的个人设置（语言、主题、通知偏好等）
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateSettingsRequest true "设置信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/settings [put]
func (ctrl *ProfileController) UpdateSettings(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 更新 users 表中的 language 字段
	if req.Language != "" {
		update_req := &services.UserUpdateRequest{
			ID: uid,
		}
		lang := utils.Clean_XSS(req.Language)
		update_req.Language = &lang
		if err := ctrl.user_svc.Update(update_req); err != nil {
			utils.Fail(c, 500, "Failed to update settings")
			return
		}
	}

	// 更新 user_settings 表
	settings, _ := models.GetUserSettings(uid)
	if settings == nil {
		// 创建默认设置
		settings = &models.UserSettings{
			UserID:      uid,
			Theme:       "light",
			NotifyEmail: true,
		}
	}
	if req.Theme != "" {
		settings.Theme = req.Theme
	}
	if req.NotifyEmail != nil {
		settings.NotifyEmail = *req.NotifyEmail
	}

	if err := models.SaveUserSettings(settings); err != nil {
		utils.Fail(c, 500, "Failed to save settings")
		return
	}

	utils.Success(c, gin.H{"message": "Settings updated successfully"})
}

// ========================================
// 头像
// ========================================

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
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
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

	// URL 校验
	if !utils.ValidateURL(req.Avatar) {
		utils.Fail(c, 400, "头像 URL 仅支持 http/https 协议")
		return
	}
	if len(req.Avatar) > 500 {
		utils.Fail(c, 400, "头像URL长度不能超过500个字符")
		return
	}

	update_req := &services.UserUpdateRequest{
		ID: uid,
	}
	update_req.Avatar = &req.Avatar

	if err := ctrl.user_svc.Update(update_req); err != nil {
		utils.Fail(c, 500, "Failed to update avatar")
		return
	}

	utils.Success(c, gin.H{"message": "Avatar updated successfully", "avatar": req.Avatar})
}

// ========================================
// 注册路由
// ========================================

// RegisterRoutes 注册用户中心路由
func (ctrl *ProfileController) RegisterRoutes(group *gin.RouterGroup) {
	// 个人信息
	group.GET("/profile", ctrl.GetProfile)
	group.PUT("/profile", ctrl.UpdateProfile)
	group.GET("/apikey", ctrl.GetApiKey)

	// 密码
	group.PUT("/password", ctrl.ChangePassword)

	// 设置
	group.GET("/settings", ctrl.GetSettings)
	group.PUT("/settings", ctrl.UpdateSettings)

	// 头像
	group.PUT("/avatar", ctrl.UpdateAvatar)

	// 统计
	group.GET("/stats", ctrl.GetUserStats)

	verificationGroup := group.Group("")
	verificationGroup.Use(middleware.UserRateLimitMiddleware(1, 3))
	verificationGroup.POST("/email/send-code", ctrl.SendEmailChangeCode)
	verificationGroup.POST("/email/verify", ctrl.VerifyEmailChange)
	verificationGroup.POST("/phone/send-code", ctrl.SendPhoneChangeCode)
	verificationGroup.POST("/phone/verify", ctrl.VerifyPhoneChange)

	// 账号注销
	group.POST("/deactivate", ctrl.DeactivateAccount)

	// 会话管理
	group.GET("/sessions", ctrl.GetSessions)
	group.DELETE("/sessions/:id", ctrl.RevokeSession)
	group.POST("/sessions/revoke-all", ctrl.RevokeAllSessions)

	// API Key
	group.POST("/resetapikey", ctrl.ResetApiKey)

	// 余额/积分日志
	group.GET("/money-logs", ctrl.GetMoneyLogs)
	group.GET("/score-logs", ctrl.GetScoreLogs)

	// 本人操作日志 / API 访问日志（强制按当前 user_id 过滤）
	group.GET("/logs", ctrl.ListMyOperationLogs)
	group.GET("/logs/:id", ctrl.GetMyOperationLogDetail)
	group.GET("/api-logs", ctrl.ListMyAPILogs)
	group.GET("/api-logs/:id", ctrl.GetMyAPILogDetail)

	// 仪表盘
	group.GET("/dashboard", ctrl.GetDashboard)
}
