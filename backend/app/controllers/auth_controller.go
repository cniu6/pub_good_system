package controllers

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/config"
	crypto_rand "crypto/rand"
	"fst/backend/internal/db"
	"fst/backend/utils"
	"log"
	"math/big"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthController struct{}

// getLangFromRequest 从请求中获取语言参数
// 优先使用请求体中的lang，其次使用Accept-Language请求头，默认en-US
func getLangFromRequest(c *gin.Context, reqLang string) string {
	lang := reqLang
	if lang == "" {
		// 从请求头获取
		lang = c.GetHeader("Accept-Language")
	}
	if lang == "" {
		lang = "en-US"
	}
	// 标准化语言代码
	if strings.Contains(lang, "zh") {
		lang = "zh-CN"
	} else {
		lang = "en-US"
	}
	return lang
}

func isNonProductionMode() bool {
	return !config.IsProductionMode()
}

func registrationAllowed() bool {
	if services.GlobalSettingsService != nil {
		return services.GlobalSettingsService.GetBoolWithDefault("allow_register", true)
	}

	setting, err := models.GetSettingByKey("allow_register")
	if err != nil {
		return true
	}

	value := strings.TrimSpace(setting.Value)
	if value == "" {
		return true
	}

	return value == "1" || strings.EqualFold(value, "true")
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Code     string `json:"code" binding:"required"`
}

type LoginRequest struct {
	UserName string `json:"userName"`
	Username string `json:"username"`
	Password string `json:"password" binding:"required"`
}

type SendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Lang  string `json:"lang"`
}

type ResetEmailRequest struct {
	Email string `json:"email" binding:"required"`
	Lang  string `json:"lang"`
}

type ResetPasswordConfirmRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// Register 注册新用户 (已迁移到 public.AuthController)
// Deprecated: 请使用 /api/v1/public/register
func (ctrl *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入，防止SQL注入和XSS攻击
	req.Username = utils.Clean_XSS(req.Username)
	req.Email = utils.Clean_XSS(req.Email)
	req.Code = utils.Clean_XSS(req.Code)
	// 密码不需要过滤（会被哈希处理）

	if !registrationAllowed() {
		utils.Fail(c, 403, "Registration is disabled")
		return
	}

	// 验证用户名格式 (3-50长度,大小写+下划线)
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)
	if !usernameRegex.MatchString(req.Username) {
		utils.Fail(c, 400, "Username must be 3-50 characters long and contain only letters, numbers, and underscores")
		return
	}

	// 检查用户名是否已存在
	if _, err := models.GetUserByUsername(req.Username); err == nil {
		utils.Fail(c, 400, "Username already exists")
		return
	}
	// 检查邮箱是否已存在
	if _, err := models.GetUserByEmail(req.Email); err == nil {
		utils.Fail(c, 400, "Email already exists")
		return
	}

	// Geetest validation (runtime config)
	geetestConfig := services.GetGlobalGeetestRuntimeConfig()
	if geetestConfig.Enabled {
		geetestReq := utils.GeetestValidateRequest{
			LotNumber:     c.GetHeader("X-Geetest-Lot-Number"),
			CaptchaOutput: c.GetHeader("X-Geetest-Captcha-Output"),
			PassToken:     c.GetHeader("X-Geetest-Pass-Token"),
			GenTime:       c.GetHeader("X-Geetest-Gen-Time"),
			CaptchaID:     c.GetHeader("X-Geetest-Captcha-Id"),
		}

		valid, err := utils.ValidateGeetest(geetestConfig.CaptchaID, geetestConfig.CaptchaKey, geetestReq)
		if err != nil || !valid {
			utils.Fail(c, 403, "Captcha validation failed")
			return
		}
	}

	consumed, err := models.ConsumeVerificationCode(req.Email, req.Code, "register")
	if err != nil || !consumed {
		utils.Fail(c, 400, "Invalid or expired verification code")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		if utils.IsPasswordValidationError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		utils.Fail(c, 500, "Failed to hash password")
		return
	}
	user := &models.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Role:     "user",
		Status:   1,
	}

	if err := models.CreateUser(user); err != nil {
		log.Printf("[AUTH] failed to create user: %v", err)
		utils.Fail(c, 500, "Failed to create user")
		return
	}

	if err := models.DeleteVerificationCodesByContact(req.Email, "register"); err != nil && isNonProductionMode() {
		log.Printf("[AUTH] cleanup register verification codes failed: %v", err)
	}

	utils.Success(c, gin.H{"message": "User registered successfully"})
}

// SendRegisterCode 发送注册验证码
func (ctrl *AuthController) SendRegisterCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入，防止SQL注入和XSS攻击
	req.Email = utils.Clean_XSS(req.Email)
	req.Lang = utils.Clean_XSS(req.Lang)
	if !registrationAllowed() {
		utils.Fail(c, 403, "Registration is disabled")
		return
	}

	// 生成6位验证码（使用 crypto/rand）
	code := generateSecureCodeLegacy()

	// 存储验证码到数据库，有效期可配置（分钟）
	expiresAt := time.Now().Add(time.Duration(config.GlobalConfig.RegisterCodeExpireMinutes) * time.Minute)
	err := models.CreateVerificationCode(req.Email, code, "register", expiresAt)
	if err != nil {
		log.Printf("[AUTH] failed to save register verification code: %v", err)
		utils.Fail(c, 500, "Failed to generate verification code")
		return
	}

	// 获取语言（优先请求体，其次请求头，默认英文）
	lang := getLangFromRequest(c, req.Lang)

	expireMinStr := fmt.Sprintf("%d", config.GlobalConfig.RegisterCodeExpireMinutes)
	emailSvc := services.NewEmailService()
	subject, body, err := emailSvc.RenderTemplateMail("register_code", lang, map[string]string{
		"code":           code,
		"expire_minutes": expireMinStr,
	})
	if err != nil {
		log.Printf("[AUTH] failed to render register email template: %v", err)
		utils.Fail(c, 500, "Failed to render email template")
		return
	}

	// 如果配置了SMTP，发送真实邮件
	if config.GlobalConfig.SMTPHost != "" {
		err := utils.SendEmail(utils.EmailMessage{
			To:      req.Email,
			Subject: subject,
			Body:    body,
		})

		// 记录邮件日志
		status := 1
		errMsg := ""
		if err != nil {
			status = 0
			errMsg = err.Error()
		}
		// 异步记录日志
		go func(email, subj, content string, st int, em string) {
			_ = models.CreateEmailLog(email, subj, content, "register_code", st, em)
		}(req.Email, subject, body, status, errMsg)

		if err != nil {
			// 如果发送失败，但在开发环境，我们可以返回验证码方便调试
			if isNonProductionMode() {
				log.Printf("[AUTH] failed to send register email: %v", err)
				utils.Fail(c, 500, "Failed to send email")
				return
			}
			log.Printf("[AUTH] failed to send register email: %v", err)
			utils.Fail(c, 500, "Failed to send email")
			return
		}
	} else {
		// 没有配置SMTP，在开发模式下直接返回验证码
		if isNonProductionMode() {
			log.Printf("[AUTH] SMTP not configured for register email delivery")
			utils.Fail(c, 500, "Failed to send email")
			return
		}
		// 生产模式下，如果没有配置SMTP，返回错误
		utils.Fail(c, 500, "Failed to send email")
		return
	}

	utils.Success(c, gin.H{"message": "Verification code sent"})
}

// Login 用户登录 (已迁移到 public.AuthController)
// Deprecated: 请使用 /api/v1/public/login
func (ctrl *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入，防止SQL注入和XSS攻击
	req.UserName = utils.Clean_XSS(req.UserName)
	req.Username = utils.Clean_XSS(req.Username)
	// 密码不需要过滤（会被哈希处理）

	username := req.UserName
	if username == "" {
		username = req.Username
	}
	if username == "" {
		utils.Fail(c, 400, "username is required")
		return
	}

	geetestConfig := services.GetGlobalGeetestRuntimeConfig()
	if geetestConfig.Enabled {
		geetestReq := utils.GeetestValidateRequest{
			LotNumber:     c.GetHeader("X-Geetest-Lot-Number"),
			CaptchaOutput: c.GetHeader("X-Geetest-Captcha-Output"),
			PassToken:     c.GetHeader("X-Geetest-Pass-Token"),
			GenTime:       c.GetHeader("X-Geetest-Gen-Time"),
			CaptchaID:     c.GetHeader("X-Geetest-Captcha-Id"),
		}

		valid, err := utils.ValidateGeetest(geetestConfig.CaptchaID, geetestConfig.CaptchaKey, geetestReq)
		if err != nil || !valid {
			utils.Fail(c, 403, "Captcha validation failed")
			return
		}
	}

	// 支持用户名或邮箱登录
	user, err := models.GetUserByUsernameOrEmail(username)
	if err != nil {
		if isNonProductionMode() {
			log.Printf("[AUTH] legacy login failed: invalid account")
		}
		utils.Fail(c, 401, "Invalid account or password")
		return
	}

	// 检查账户是否被锁定（基于锁定时间和失败次数）
	now := time.Now().Unix()
	if user.LockUntil != nil && *user.LockUntil > now {
		// 账户仍在锁定期内
		remainingMinutes := (*user.LockUntil - now) / 60
		utils.Fail(c, 403, fmt.Sprintf("Account is locked. Please try again in %d minutes", remainingMinutes))
		return
	}
	if user.LockUntil != nil && *user.LockUntil <= now {
		// 锁定已过期，清除锁定状态
		_, _ = db.DB.Exec("UPDATE users SET lock_until = NULL WHERE id = ?", user.ID)
	}
	if int(user.LoginFailure) >= config.GlobalConfig.LoginMaxFailureCount {
		// 失败次数达到阈值，但锁定时间已过期，允许尝试（如果密码错误会重新锁定）
		// 这里不阻止，让密码验证来决定
	}

	if user.Status == 0 {
		utils.Fail(c, 403, "Account is inactive")
		return
	}

	// 验证密码
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		// 密码错误，增加失败计数（如果达到阈值会自动锁定）
		_ = models.IncrementLoginFailure(user.ID, config.GlobalConfig.LoginMaxFailureCount, config.GlobalConfig.LoginLockDurationMinutes)
		if isNonProductionMode() {
			log.Printf("[AUTH] legacy login failed: password mismatch for user_id=%d", user.ID)
		}
		utils.Fail(c, 401, "Invalid account or password")
		return
	}

	// 密码正确，登录成功
	// 获取客户端IP
	clientIP := c.ClientIP()
	if clientIP == "" {
		clientIP = c.GetHeader("X-Forwarded-For")
		if clientIP == "" {
			clientIP = c.GetHeader("X-Real-IP")
		}
	}
	if clientIP == "" {
		clientIP = "unknown"
	}

	// 更新登录信息（最后登录时间、IP，重置失败次数）
	if err := models.UpdateLoginInfo(user.ID, clientIP); err != nil {
		if isNonProductionMode() {
			log.Printf("[AUTH] failed to update legacy login info: %v", err)
		}
		// 不阻止登录，只记录错误
	}

	accessTTL := time.Duration(config.GlobalConfig.JWTAccessExpire) * time.Second
	refreshTTL := time.Duration(config.GlobalConfig.JWTRefreshExpire) * time.Second
	nowUnix := time.Now().Unix()

	accessToken, err := utils.GenerateTokenWithTTL(user.ID, user.Role, accessTTL)
	if err != nil {
		utils.Fail(c, 500, "Failed to generate access token")
		return
	}
	refreshToken, err := utils.GenerateRefreshTokenWithTTL(user.ID, refreshTTL)
	if err != nil {
		utils.Fail(c, 500, "Failed to generate refresh token")
		return
	}

	utils.Success(c, gin.H{
		"id":           user.ID,
		"userName":     user.Username,
		"email":        user.Email,
		"role":         []string{user.Role},
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"expiresAt":    nowUnix + int64(accessTTL.Seconds()),
	})
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (ctrl *AuthController) UpdateToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	claims, err := utils.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		utils.Fail(c, 401, "Invalid or expired refresh token")
		return
	}

	accessTTL := time.Duration(config.GlobalConfig.JWTAccessExpire) * time.Second
	refreshTTL := time.Duration(config.GlobalConfig.JWTRefreshExpire) * time.Second
	nowUnix := time.Now().Unix()

	user, err := models.GetUserByID(claims.UserID)
	if err != nil {
		utils.Fail(c, 401, "User not found")
		return
	}
	accessToken, err := utils.GenerateTokenWithTTL(user.ID, user.Role, accessTTL)
	if err != nil {
		utils.Fail(c, 500, "Failed to generate access token")
		return
	}
	refreshToken, err := utils.GenerateRefreshTokenWithTTL(user.ID, refreshTTL)
	if err != nil {
		utils.Fail(c, 500, "Failed to generate refresh token")
		return
	}
	utils.Success(c, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"expiresAt":    nowUnix + int64(accessTTL.Seconds()),
	})
}

func (ctrl *AuthController) GetUserRoutes(c *gin.Context) {
	utils.Success(c, []interface{}{})
}

// SendResetEmail 发送重置密码邮件
func (ctrl *AuthController) SendResetEmail(c *gin.Context) {
	var req ResetEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入，防止SQL注入和XSS攻击
	req.Email = utils.Clean_XSS(req.Email)
	req.Lang = utils.Clean_XSS(req.Lang)

	// 检查邮箱是否存在
	user, err := models.GetUserByUsernameOrEmail(req.Email)
	if err != nil || user == nil {
		// 为了安全，即使邮箱不存在也提示发送成功，避免枚举
		utils.Success(c, gin.H{"message": "If the email exists, a reset code has been sent"})
		return
	}

	// 生成6位验证码（使用 crypto/rand）
	code := generateSecureCodeLegacy()

	// 存储验证码到数据库，15分钟有效期（重置密码链接需要更长时间）
	expiresAt := time.Now().Add(15 * time.Minute)
	err = models.CreateVerificationCode(user.Email, code, "reset_password", expiresAt)
	if err != nil {
		log.Printf("[AUTH] failed to save reset password code: %v", err)
		// 即使失败也返回成功，避免邮箱枚举攻击
		utils.Success(c, gin.H{"message": "If the email exists, a reset code has been sent"})
		return
	}

	// 构造链接 (假设前端路由是 /login/reset-password-confirm)
	// 从系统设置读取前端地址
	frontendURL := ""
	if s, err := models.GetSettingByKey("frontend_url"); err == nil && s.Value != "" {
		frontendURL = strings.TrimRight(s.Value, "/")
	}
	if frontendURL == "" {
		if isNonProductionMode() {
			frontendURL = "http://localhost:5173"
		} else {
			log.Printf("[AUTH] frontend_url missing while preparing legacy reset password email")
			utils.Success(c, gin.H{"message": "If the email exists, a reset code has been sent"})
			return
		}
	}
	resetLink := fmt.Sprintf("%s/#/login/reset-password-confirm?email=%s&token=%s", frontendURL, user.Email, code)

	// 获取语言（优先请求体，其次请求头，默认英文）
	lang := getLangFromRequest(c, req.Lang)

	tpl, err := models.GetEmailTemplate("reset_password", lang)
	var subject, body string
	if err == nil && tpl != nil {
		subject = strings.ReplaceAll(tpl.Subject, "{app_name}", config.GlobalConfig.AppName)
		body = strings.ReplaceAll(tpl.Content, "{code}", code)
		body = strings.ReplaceAll(body, "{link}", resetLink)
		body = strings.ReplaceAll(body, "{app_name}", config.GlobalConfig.AppName)
	} else {
		subject = fmt.Sprintf("【%s】密码重置请求", config.GlobalConfig.AppName)
		body = fmt.Sprintf("请点击以下链接重置密码：<br><a href=\"%s\">%s</a><br>或者使用验证码：%s<br>有效期15分钟。", resetLink, resetLink, code)
	}

	if config.GlobalConfig.SMTPHost != "" {
		err := utils.SendEmail(utils.EmailMessage{
			To:      user.Email,
			Subject: subject,
			Body:    body,
		})

		// 记录邮件日志
		status := 1
		errMsg := ""
		if err != nil {
			status = 0
			errMsg = err.Error()
		}
		// 异步记录日志
		go func(email, subj, content string, st int, em string) {
			_ = models.CreateEmailLog(email, subj, content, "reset_password", st, em)
		}(user.Email, subject, body, status, errMsg)

		if err != nil {
			log.Printf("[AUTH] failed to send reset password email: %v", err)
			utils.Success(c, gin.H{"message": "If the email exists, a reset code has been sent"})
			return
		}
	} else {
		if isNonProductionMode() {
			log.Printf("[AUTH] SMTP not configured for reset password email delivery")
			utils.Success(c, gin.H{"message": "If the email exists, a reset code has been sent"})
			return
		}
		// 生产模式下，如果没有配置SMTP，返回错误
		utils.Success(c, gin.H{"message": "If the email exists, a reset code has been sent"})
		return
	}

	utils.Success(c, gin.H{"message": "Reset email sent"})
}

// ResetPasswordConfirm 确认重置密码
func (ctrl *AuthController) ResetPasswordConfirm(c *gin.Context) {
	var req ResetPasswordConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入，防止SQL注入和XSS攻击
	req.Email = utils.Clean_XSS(req.Email)
	req.Code = utils.Clean_XSS(req.Code)
	// 密码不需要过滤（会被哈希处理）

	consumed, err := models.ConsumeVerificationCode(req.Email, req.Code, "reset_password")
	if err != nil || !consumed {
		utils.Fail(c, 400, "Invalid or expired reset token")
		return
	}

	// 重置成功后：清理该邮箱所有重置密码验证码
	_ = models.DeleteVerificationCodesByContact(req.Email, "reset_password")

	user, err := models.GetUserByEmail(req.Email)
	if err != nil {
		utils.Fail(c, 400, "User not found")
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		if utils.IsPasswordValidationError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		utils.Fail(c, 500, "Failed to hash password")
		return
	}
	user.Password = hashedPassword

	// 更新密码 (需要 models 支持 UpdateUser)
	// 这里直接写SQL更新
	err = models.UpdatePassword(user.ID, hashedPassword)
	if err != nil {
		utils.Fail(c, 500, "Failed to update password")
		return
	}

	utils.Success(c, gin.H{"message": "Password reset successfully"})
}

// generateSecureCodeLegacy 生成6位数字验证码（使用 crypto/rand）
func generateSecureCodeLegacy() string {
	n, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(1000000))
	if err != nil {
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		return fmt.Sprintf("%06d", rnd.Intn(1000000))
	}
	return fmt.Sprintf("%06d", n.Int64())
}
