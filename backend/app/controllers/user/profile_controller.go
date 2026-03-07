package user

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"
	"math/rand"
	"strings"
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
		"id":            user.ID,
		"username":      user.Username,
		"email":         user.Email,
		"nickname":      user.Nickname,
		"avatar":        user.Avatar,
		"back_ground":   user.BackGround,
		"gender":        user.Gender,
		"birthday":      user.Birthday,
		"motto":         user.Motto,
		"mobile":        user.Mobile,
		"money":         user.Money,
		"score":         user.Score,
		"level":         user.Level,
		"role":          user.Role,
		"status":        user.Status,
		"language":      user.Language,
		"country":       user.Country,
		"apikey":        user.Apikey,
		"joinTime":      user.JoinTime,
		"joinIp":        user.JoinIp,
		"lastLoginTime": user.LastLoginTime,
		"lastLoginIp":   user.LastLoginIp,
		"updateTime":    user.UpdateTime,
		"createTime":    user.CreateTime,
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
	req.BackGround = utils.Clean_XSS(req.BackGround)
	req.Language = utils.Clean_XSS(req.Language)
	req.Country = utils.Clean_XSS(req.Country)

	// 构建更新请求
	update_req := &services.UserUpdateRequest{
		ID:         user_id.(uint64),
		Nickname:   req.Nickname,
		Avatar:     req.Avatar,
		Gender:     req.Gender,
		Birthday:   req.Birthday,
		Motto:      req.Motto,
		Mobile:     req.Mobile,
		BackGround: req.BackGround,
		Language:   req.Language,
		Country:    req.Country,
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

// ========================================
// 用户设置相关请求结构体
// ========================================

type UpdateSettingsRequest struct {
	Language    string `json:"language"`
	Theme       string `json:"theme"`
	NotifyEmail *bool  `json:"notify_email"`
}

type SendEmailCodeRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
	Lang     string `json:"lang"`
}

type VerifyEmailChangeRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
}

type SendPhoneCodeRequest struct {
	NewMobile string `json:"new_mobile" binding:"required"`
}

type VerifyPhoneChangeRequest struct {
	NewMobile string `json:"new_mobile" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

type DeactivateAccountRequest struct {
	Password string `json:"password" binding:"required"`
	Reason   string `json:"reason"`
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

	user, err := ctrl.user_svc.GetByID(user_id.(uint64))
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

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	uid := user_id.(uint64)

	// 更新 users 表中的 language 字段
	if req.Language != "" {
		update_req := &services.UserUpdateRequest{
			ID:       uid,
			Language: utils.Clean_XSS(req.Language),
		}
		ctrl.user_svc.Update(update_req)
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

// ========================================
// 用户统计
// ========================================

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

	// 获取登录次数
	login_count, _ := models.GetUserLoginCount(user.ID)

	utils.Success(c, gin.H{
		"joinTime":      user.JoinTime,
		"lastLoginTime": user.LastLoginTime,
		"lastLoginIp":   user.LastLoginIp,
		"loginCount":    login_count,
		"daysJoined":    calculateDaysJoined(user.JoinTime),
		"money":         user.Money,
		"score":         user.Score,
		"level":         user.Level,
	})
}

// ========================================
// 邮箱变更验证码流程
// ========================================

// SendEmailChangeCode 发送修改邮箱验证码
// @Summary 发送修改邮箱验证码
// @Description 发送验证码到新邮箱以验证邮箱变更
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SendEmailCodeRequest true "新邮箱"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/email/send-code [post]
func (ctrl *ProfileController) SendEmailChangeCode(c *gin.Context) {
	// 极验验证
	if !validateGeetestFromRequest(c) {
		utils.Fail(c, 403, "Captcha validation failed")
		return
	}

	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	var req SendEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	req.NewEmail = utils.Clean_XSS(req.NewEmail)
	uid := user_id.(uint64)

	// 检查邮箱是否已被使用
	existing, _ := models.GetUserByEmail(req.NewEmail)
	if existing != nil && existing.ID != uid {
		utils.Fail(c, 400, "Email already in use")
		return
	}

	// 检查邮箱验证码功能是否启用
	verifyConfig := services.GetGlobalVerifyConfig()
	if !verifyConfig.EmailEnabled {
		// 验证码功能关闭，直接更新邮箱
		update_req := &services.UserUpdateRequest{ID: uid, Email: req.NewEmail}
		if err := ctrl.user_svc.Update(update_req); err != nil {
			utils.Fail(c, 500, err.Error())
			return
		}
		fmt.Printf("[DEV] Email verify disabled, directly changed email for user %d to %s\n", uid, req.NewEmail)
		utils.Success(c, gin.H{"message": "Email changed successfully (verification disabled)", "verified": true, "email": req.NewEmail})
		return
	}

	// 生成验证码
	code := generateCode()

	// 存储验证码（类型为 change_email）
	expires_at := time.Now().Add(15 * time.Minute)
	if err := models.CreateVerificationCode(req.NewEmail, code, "change_email", expires_at); err != nil {
		utils.Fail(c, 500, "Failed to generate verification code")
		return
	}

	// 发送邮件
	email_svc := services.NewEmailService()
	if !email_svc.IsEmailConfigured() {
		utils.Fail(c, 500, "Email service not configured")
		return
	}

	lang := getLangFromRequest(c, req.Lang)
	vars := map[string]string{
		"code":           code,
		"expire_minutes": "15",
	}
	if err := email_svc.SendTemplateEmail(req.NewEmail, "change_email", lang, vars); err != nil {
		// 降级：尝试用 register_code 模板
		if err2 := email_svc.SendTemplateEmail(req.NewEmail, "register_code", lang, vars); err2 != nil {
			utils.Fail(c, 500, "Failed to send verification email")
			return
		}
	}

	utils.Success(c, gin.H{"message": "Verification code sent to new email"})
}

// VerifyEmailChange 验证并修改邮箱
// @Summary 验证并修改邮箱
// @Description 使用验证码确认邮箱变更
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body VerifyEmailChangeRequest true "验证信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/email/verify [post]
func (ctrl *ProfileController) VerifyEmailChange(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	var req VerifyEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	req.NewEmail = utils.Clean_XSS(req.NewEmail)
	req.Code = utils.Clean_XSS(req.Code)

	// 验证验证码
	valid, code_id, err := models.VerifyCode(req.NewEmail, req.Code, "change_email")
	if err != nil || !valid {
		utils.Fail(c, 400, "Invalid or expired verification code")
		return
	}
	_ = models.MarkVerificationCodeAsUsed(code_id)
	_ = models.DeleteVerificationCodesByEmail(req.NewEmail, "change_email")

	// 更新邮箱
	update_req := &services.UserUpdateRequest{
		ID:    user_id.(uint64),
		Email: req.NewEmail,
	}
	if err := ctrl.user_svc.Update(update_req); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "Email changed successfully", "email": req.NewEmail})
}

// ========================================
// 手机变更验证码流程
// ========================================

// SendPhoneChangeCode 发送修改手机号验证码
// @Summary 发送修改手机号验证码
// @Description 发送验证码到新手机号（占位实现，需对接短信服务）
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SendPhoneCodeRequest true "新手机号"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/phone/send-code [post]
func (ctrl *ProfileController) SendPhoneChangeCode(c *gin.Context) {
	// 极验验证
	if !validateGeetestFromRequest(c) {
		utils.Fail(c, 403, "Captcha validation failed")
		return
	}

	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	var req SendPhoneCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	req.NewMobile = utils.Clean_XSS(req.NewMobile)
	uid := user_id.(uint64)

	// 检查短信验证码功能是否启用
	verifyConfig := services.GetGlobalVerifyConfig()
	if !verifyConfig.SMSEnabled {
		// 验证码功能关闭，直接更新手机号
		update_req := &services.UserUpdateRequest{ID: uid, Mobile: req.NewMobile}
		if err := ctrl.user_svc.Update(update_req); err != nil {
			utils.Fail(c, 500, err.Error())
			return
		}
		fmt.Printf("[DEV] SMS verify disabled, directly changed phone for user %d to %s\n", uid, req.NewMobile)
		utils.Success(c, gin.H{"message": "Phone changed successfully (verification disabled)", "verified": true, "mobile": req.NewMobile})
		return
	}

	// 生成验证码
	code := generateCode()

	// 存储验证码
	expires_at := time.Now().Add(10 * time.Minute)
	if err := models.CreateVerificationCode(req.NewMobile, code, "change_phone", expires_at); err != nil {
		utils.Fail(c, 500, "Failed to generate verification code")
		return
	}

	// 通过 SMS 服务发送验证码
	if services.GlobalSMSService != nil {
		if err := services.GlobalSMSService.SendCode(req.NewMobile, code, 10); err != nil {
			fmt.Printf("[SMS] Failed to send code to %s: %v\n", req.NewMobile, err)
		}
	} else {
		fmt.Printf("[DEV] Phone change code for %s: %s\n", req.NewMobile, code)
	}

	utils.Success(c, gin.H{"message": "Verification code sent"})
}

// VerifyPhoneChange 验证并修改手机号
// @Summary 验证并修改手机号
// @Description 使用验证码确认手机号变更
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body VerifyPhoneChangeRequest true "验证信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/phone/verify [post]
func (ctrl *ProfileController) VerifyPhoneChange(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	var req VerifyPhoneChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	req.NewMobile = utils.Clean_XSS(req.NewMobile)
	req.Code = utils.Clean_XSS(req.Code)

	// 验证验证码
	valid, code_id, err := models.VerifyCode(req.NewMobile, req.Code, "change_phone")
	if err != nil || !valid {
		utils.Fail(c, 400, "Invalid or expired verification code")
		return
	}
	_ = models.MarkVerificationCodeAsUsed(code_id)
	_ = models.DeleteVerificationCodesByEmail(req.NewMobile, "change_phone")

	// 更新手机号
	update_req := &services.UserUpdateRequest{
		ID:     user_id.(uint64),
		Mobile: req.NewMobile,
	}
	if err := ctrl.user_svc.Update(update_req); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "Phone number changed successfully", "mobile": req.NewMobile})
}

// ========================================
// 账号注销
// ========================================

// DeactivateAccount 注销账号
// @Summary 注销账号
// @Description 用户主动注销（软删除）自己的账号
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body DeactivateAccountRequest true "密码确认"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/deactivate [post]
func (ctrl *ProfileController) DeactivateAccount(c *gin.Context) {
	// 极验验证
	if !validateGeetestFromRequest(c) {
		utils.Fail(c, 403, "Captcha validation failed")
		return
	}

	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	var req DeactivateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	uid := user_id.(uint64)

	// 验证密码
	user, err := ctrl.user_svc.GetByID(uid)
	if err != nil {
		utils.Fail(c, 404, "User not found")
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		utils.Fail(c, 400, "Incorrect password")
		return
	}

	// 软删除用户
	if err := ctrl.user_svc.Delete(uid); err != nil {
		utils.Fail(c, 500, "Failed to deactivate account")
		return
	}

	utils.Success(c, gin.H{"message": "Account deactivated successfully"})
}

// ========================================
// 登录设备/会话管理
// ========================================

// GetSessions 获取用户登录会话列表
// @Summary 获取登录会话
// @Description 获取当前用户的所有活跃登录会话
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/sessions [get]
func (ctrl *ProfileController) GetSessions(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	sessions, err := models.GetUserSessions(user_id.(uint64))
	if err != nil {
		utils.Success(c, []interface{}{})
		return
	}

	utils.Success(c, sessions)
}

// RevokeSession 踢出指定会话
// @Summary 踢出会话
// @Description 撤销指定的登录会话
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "会话ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/sessions/:id [delete]
func (ctrl *ProfileController) RevokeSession(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	session_id := c.Param("id")
	if session_id == "" {
		utils.Fail(c, 400, "Session ID is required")
		return
	}

	if err := models.RevokeUserSession(user_id.(uint64), session_id); err != nil {
		utils.Fail(c, 500, "Failed to revoke session")
		return
	}

	utils.Success(c, gin.H{"message": "Session revoked successfully"})
}

// RevokeAllSessions 踢出所有其他会话
// @Summary 踢出所有其他会话
// @Description 撤销当前会话以外的所有登录会话
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/sessions/revoke-all [post]
func (ctrl *ProfileController) RevokeAllSessions(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	current_token := c.GetHeader("Authorization")
	if err := models.RevokeAllUserSessions(user_id.(uint64), current_token); err != nil {
		utils.Fail(c, 500, "Failed to revoke sessions")
		return
	}

	utils.Success(c, gin.H{"message": "All other sessions revoked"})
}

// ========================================
// API Key 管理
// ========================================

// ResetApiKey 重置API密钥
// @Summary 重置API密钥
// @Description 重置当前用户的API密钥
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/resetapikey [post]
func (ctrl *ProfileController) ResetApiKey(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	new_key, err := models.ResetUserApiKey(user_id.(uint64))
	if err != nil {
		utils.Fail(c, 500, "Failed to reset API key")
		return
	}

	utils.Success(c, gin.H{"apikey": new_key})
}

// ========================================
// 辅助函数
// ========================================

// calculateDaysJoined 计算加入天数
func calculateDaysJoined(join_time *int64) int {
	if join_time == nil {
		return 0
	}
	join := time.Unix(*join_time, 0)
	return int(time.Since(join).Hours() / 24)
}

// generateCode 生成6位数字验证码
func generateCode() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%06d", rnd.Intn(1000000))
}

// getLangFromRequest 从请求获取语言
func getLangFromRequest(c *gin.Context, req_lang string) string {
	lang := req_lang
	if lang == "" {
		lang = c.GetHeader("Accept-Language")
	}
	if lang == "" {
		lang = "zh-CN"
	}
	if strings.Contains(lang, "zh") {
		return "zh-CN"
	}
	return "en-US"
}

// validateGeetestFromRequest 从请求中提取极验参数并校验
// 返回 true 表示通过（或极验未启用），false 表示校验失败
func validateGeetestFromRequest(c *gin.Context) bool {
	geetestConfig := services.GetGlobalGeetestRuntimeConfig()
	if !geetestConfig.Enabled {
		return true
	}

	geetestReq := utils.GeetestValidateRequest{
		LotNumber:     c.GetHeader("X-Geetest-Lot-Number"),
		CaptchaOutput: c.GetHeader("X-Geetest-Captcha-Output"),
		PassToken:     c.GetHeader("X-Geetest-Pass-Token"),
		GenTime:       c.GetHeader("X-Geetest-Gen-Time"),
		CaptchaID:     c.GetHeader("X-Geetest-Captcha-Id"),
	}

	valid, err := utils.ValidateGeetest(geetestConfig.CaptchaID, geetestConfig.CaptchaKey, geetestReq)
	if err != nil || !valid {
		return false
	}
	return true
}

// ========================================
// 用户仪表盘
// ========================================

// GetDashboard 获取用户仪表盘数据
// @Summary 获取用户仪表盘
// @Description 返回用户统计概览、公告、快捷操作入口
// @Tags 用户中心
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/dashboard [get]
func (ctrl *ProfileController) GetDashboard(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}

	uid := user_id.(uint64)
	user, err := ctrl.user_svc.GetByID(uid)
	if err != nil {
		utils.Fail(c, 404, "User not found")
		return
	}

	login_count, _ := models.GetUserLoginCount(uid)

	// 公告列表（可扩展为从数据库读取）
	announcements := []gin.H{
		{"id": 1, "type": "info", "title": "系统维护通知", "content": "系统将于本周六凌晨进行维护升级", "time": time.Now().Unix()},
		{"id": 2, "type": "success", "title": "新功能上线", "content": "用户中心新增设备管理和账号安全功能", "time": time.Now().Unix()},
		{"id": 3, "type": "warning", "title": "安全提醒", "content": "请定期修改密码以保障账号安全", "time": time.Now().Unix()},
	}

	utils.Success(c, gin.H{
		"user": gin.H{
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"email":    user.Email,
			"role":     user.Role,
			"level":    user.Level,
		},
		"stats": gin.H{
			"money":      user.Money,
			"score":      user.Score,
			"level":      user.Level,
			"loginCount": login_count,
			"daysJoined": calculateDaysJoined(user.JoinTime),
		},
		"announcements": announcements,
	})
}

// ========================================
// 注册路由
// ========================================

// RegisterRoutes 注册用户中心路由
func (ctrl *ProfileController) RegisterRoutes(group *gin.RouterGroup) {
	// 个人信息
	group.GET("/profile", ctrl.GetProfile)
	group.PUT("/profile", ctrl.UpdateProfile)

	// 密码
	group.PUT("/password", ctrl.ChangePassword)

	// 设置
	group.GET("/settings", ctrl.GetSettings)
	group.PUT("/settings", ctrl.UpdateSettings)

	// 头像
	group.PUT("/avatar", ctrl.UpdateAvatar)

	// 统计
	group.GET("/stats", ctrl.GetUserStats)

	// 邮箱变更验证
	group.POST("/email/send-code", ctrl.SendEmailChangeCode)
	group.POST("/email/verify", ctrl.VerifyEmailChange)

	// 手机变更验证
	group.POST("/phone/send-code", ctrl.SendPhoneChangeCode)
	group.POST("/phone/verify", ctrl.VerifyPhoneChange)

	// 账号注销
	group.POST("/deactivate", ctrl.DeactivateAccount)

	// 会话管理
	group.GET("/sessions", ctrl.GetSessions)
	group.DELETE("/sessions/:id", ctrl.RevokeSession)
	group.POST("/sessions/revoke-all", ctrl.RevokeAllSessions)

	// API Key
	group.POST("/resetapikey", ctrl.ResetApiKey)

	// 仪表盘
	group.GET("/dashboard", ctrl.GetDashboard)
}

// ========================================
// 辅助：确保导入
// ========================================

var _ = models.User{}
