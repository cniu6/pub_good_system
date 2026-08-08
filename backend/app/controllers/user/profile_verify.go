package user

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	"fst/backend/utils"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SendEmailCodeRequest 发送邮箱验证码请求
type SendEmailCodeRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
	Lang     string `json:"lang"`
}
//@name 发送邮箱验证码请求

type VerifyEmailChangeRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
}
//@name 验证邮箱变更请求

// SendPhoneCodeRequest 发送手机验证码请求
type SendPhoneCodeRequest struct {
	NewMobile string `json:"new_mobile" binding:"required"`
}
//@name 发送手机验证码请求

type VerifyPhoneChangeRequest struct {
	NewMobile string `json:"new_mobile" binding:"required"`
	Code      string `json:"code" binding:"required"`
}
//@name 验证手机号变更请求

// DeactivateAccountRequest 注销账号请求
type DeactivateAccountRequest struct {
	Password string `json:"password" binding:"required"`
	Reason   string `json:"reason"`
}
//@name 注销账号请求

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
	cfg := services.GetGlobalGeetestRuntimeConfig()
	if err := utils.ValidateGeetestFromHeaders(c, cfg.CaptchaID, cfg.CaptchaKey, cfg.Enabled); err != nil {
		utils.Fail(c, 403, "Captcha validation failed")
		return
	}

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

	var req SendEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	req.NewEmail = utils.Clean_XSS(req.NewEmail)

	// 检查邮箱是否已被使用
	existing, _ := models.GetUserByEmail(req.NewEmail)
	if existing != nil && existing.ID != uid {
		utils.Fail(c, 400, "Email already in use")
		return
	}
	hasRecentCode, err := models.HasRecentVerificationCode(req.NewEmail, "change_email", time.Now().Add(-time.Minute))
	if err != nil {
		utils.Fail(c, 500, "Failed to check verification cooldown")
		return
	}
	if hasRecentCode {
		utils.Fail(c, 429, "Please wait before requesting another verification code")
		return
	}

	// 检查邮箱验证码功能是否启用
	verifyConfig := services.GetGlobalVerifyConfig()
	if !verifyConfig.EmailEnabled {
		// 验证码功能关闭，直接更新邮箱
		update_req := &services.UserUpdateRequest{ID: uid}
		update_req.Email = &req.NewEmail
		if err := ctrl.user_svc.Update(update_req); err != nil {
			if services.IsClientError(err) {
				utils.Fail(c, 400, err.Error())
				return
			}
			log.Printf("[PROFILE] direct email change failed for user_id=%d: %v", uid, err)
			utils.Fail(c, 500, "Failed to change email")
			return
		}
		if !config.IsProductionMode() {
			log.Printf("[PROFILE] email verification disabled; direct email change applied for user_id=%d", uid)
		}
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
	email_svc := services.GetEmailService()
	if !email_svc.IsEmailConfigured() {
		utils.Fail(c, 500, "Email service not configured")
		return
	}

	lang := getLangFromRequest(c, req.Lang)
	vars := map[string]string{
		"code":           code,
		"expire_minutes": "15",
	}
	if err := email_svc.SendTemplateEmailWithUser(uid, req.NewEmail, "change_email", lang, vars); err != nil {
		// 降级：尝试用 register_code 模板
		if err2 := email_svc.SendTemplateEmailWithUser(uid, req.NewEmail, "register_code", lang, vars); err2 != nil {
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
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	var req VerifyEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	req.NewEmail = utils.Clean_XSS(req.NewEmail)
	req.Code = utils.Clean_XSS(req.Code)

	consumed, err := models.ConsumeVerificationCode(req.NewEmail, req.Code, "change_email")
	if err != nil || !consumed {
		utils.Fail(c, 400, "Invalid or expired verification code")
		return
	}
	if err := models.DeleteVerificationCodesByContact(req.NewEmail, "change_email"); err != nil {
		log.Printf("[PROFILE] cleanup change_email verification codes failed: email=%s err=%v", req.NewEmail, err)
	}

	// 更新邮箱
	update_req := &services.UserUpdateRequest{
		ID: uid,
	}
	update_req.Email = &req.NewEmail
	if err := ctrl.user_svc.Update(update_req); err != nil {
		if services.IsClientError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		log.Printf("[PROFILE] verify email change failed for user_id=%d: %v", uid, err)
		utils.Fail(c, 500, "Failed to change email")
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
	cfg := services.GetGlobalGeetestRuntimeConfig()
	if err := utils.ValidateGeetestFromHeaders(c, cfg.CaptchaID, cfg.CaptchaKey, cfg.Enabled); err != nil {
		utils.Fail(c, 403, "Captcha validation failed")
		return
	}

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

	var req SendPhoneCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	req.NewMobile = utils.Clean_XSS(req.NewMobile)
	normalizedMobile, err := utils.NormalizeAndValidateMobile(req.NewMobile, services.GetGlobalMobileCNOnly())
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	req.NewMobile = normalizedMobile
	user, err := ctrl.user_svc.GetByID(uid)
	if err != nil {
		utils.Fail(c, 404, "User not found")
		return
	}

	// 检查手机号是否已被使用
	if req.NewMobile != "" {
		existing, _ := models.GetUserByMobile(req.NewMobile)
		if existing != nil && existing.ID != uid {
			utils.Fail(c, 400, "Phone number already in use")
			return
		}
	}
	hasRecentCode, err := models.HasRecentVerificationCode(req.NewMobile, "change_phone", time.Now().Add(-time.Minute))
	if err != nil {
		utils.Fail(c, 500, "Failed to check verification cooldown")
		return
	}
	if hasRecentCode {
		utils.Fail(c, 429, "Please wait before requesting another verification code")
		return
	}

	// 检查短信验证码功能是否启用
	verifyConfig := services.GetGlobalVerifyConfig()
	if !verifyConfig.SMSEnabled {
		// 验证码功能关闭，直接更新手机号
		update_req := &services.UserUpdateRequest{ID: uid}
		update_req.Mobile = &req.NewMobile
		if err := ctrl.user_svc.Update(update_req); err != nil {
			if services.IsClientError(err) {
				utils.Fail(c, 400, err.Error())
				return
			}
			log.Printf("[PROFILE] direct phone change failed for user_id=%d: %v", uid, err)
			utils.Fail(c, 500, "Failed to change phone")
			return
		}
		if !config.IsProductionMode() {
			log.Printf("[PROFILE] sms verification disabled; direct phone change applied for user_id=%d", uid)
		}
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
	if services.GlobalSMSService == nil {
		if delErr := models.DeleteVerificationCodesByContact(req.NewMobile, "change_phone"); delErr != nil {
			log.Printf("[PROFILE] 清理验证码失败 mobile=%s type=change_phone: %v", models.MaskPhone(req.NewMobile), delErr)
		}
		utils.Fail(c, 500, "SMS service unavailable")
		return
	}
	providerName := services.GlobalSMSService.GetProviderName()
	if providerName == "none" || (providerName != "console" && !services.GlobalSMSService.IsConfigured()) || (providerName == "console" && config.IsProductionMode()) {
		if delErr := models.DeleteVerificationCodesByContact(req.NewMobile, "change_phone"); delErr != nil {
			log.Printf("[PROFILE] 清理验证码失败 mobile=%s type=change_phone: %v", models.MaskPhone(req.NewMobile), delErr)
		}
		utils.Fail(c, 500, "SMS service not configured")
		return
	}

	// 优先使用用户个人语言设置，避免短信内容被固定为中文。
	smsLang := user.Language
	if strings.TrimSpace(smsLang) == "" {
		smsLang = "zh-CN"
	}
	smsTemplateParams := map[string]string{
		"__template_name":  "bind_phone",
		"__template_order": "code,expire",
		"__user_id":        fmt.Sprintf("%d", uid),
	}
	if err := services.GlobalSMSService.SendCode(req.NewMobile, code, 10, smsTemplateParams, smsLang); err != nil {
		log.Printf("[SMS] failed to send code to %s via %s: %v", models.MaskPhone(req.NewMobile), providerName, err)
		if delErr := models.DeleteVerificationCodesByContact(req.NewMobile, "change_phone"); delErr != nil {
			log.Printf("[PROFILE] 清理验证码失败 mobile=%s type=change_phone: %v", models.MaskPhone(req.NewMobile), delErr)
		}
		utils.Fail(c, 500, "Failed to send verification code")
		return
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
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	var req VerifyPhoneChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	req.NewMobile = utils.Clean_XSS(req.NewMobile)
	req.Code = utils.Clean_XSS(req.Code)

	normalizedMobile, err := utils.NormalizeAndValidateMobile(req.NewMobile, services.GetGlobalMobileCNOnly())
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	req.NewMobile = normalizedMobile

	consumed, err := models.ConsumeVerificationCode(req.NewMobile, req.Code, "change_phone")
	if err != nil || !consumed {
		utils.Fail(c, 400, "Invalid or expired verification code")
		return
	}
	if err := models.DeleteVerificationCodesByContact(req.NewMobile, "change_phone"); err != nil {
		log.Printf("[PROFILE] cleanup change_phone verification codes failed: mobile=%s err=%v", models.MaskPhone(req.NewMobile), err)
	}
	update_req := &services.UserUpdateRequest{
		ID: uid,
	}
	update_req.Mobile = &req.NewMobile
	if err := ctrl.user_svc.Update(update_req); err != nil {
		if services.IsClientError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		log.Printf("[PROFILE] verify phone change failed for user_id=%d: %v", uid, err)
		utils.Fail(c, 500, "Failed to change phone")
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
	// 检查系统设置是否允许注销账号
	if !services.GetGlobalAllowDeleteAccount() {
		utils.Fail(c, 403, "Account deletion is currently disabled")
		return
	}

	// 极验验证
	cfg := services.GetGlobalGeetestRuntimeConfig()
	if err := utils.ValidateGeetestFromHeaders(c, cfg.CaptchaID, cfg.CaptchaKey, cfg.Enabled); err != nil {
		utils.Fail(c, 403, "Captcha validation failed")
		return
	}

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

	var req DeactivateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

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
