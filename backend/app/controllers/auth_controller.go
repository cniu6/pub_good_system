package controllers

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/internal/config"
	"fst/backend/internal/db"
	"fst/backend/utils"
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

// Register 注册新用户
// @Summary 用户注册
// @Description 注册一个新用户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /auth/register [post]
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

	// 验证验证码（从数据库获取）
	valid, codeID, err := models.VerifyCode(req.Email, req.Code, "register")
	if err != nil || !valid {
		utils.Fail(c, 400, "Invalid or expired verification code")
		return
	}
	// 标记验证码为已使用
	_ = models.MarkVerificationCodeAsUsed(codeID)

	// 注册成功后：清理该邮箱所有注册验证码（以及可能残留的旧记录）
	_ = models.DeleteVerificationCodesByEmail(req.Email, "register")

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

	// 极验人机验证
	if config.GlobalConfig.GeetestEnabled {
		geetestReq := utils.GeetestValidateRequest{
			LotNumber:     c.GetHeader("X-Geetest-Lot-Number"),
			CaptchaOutput: c.GetHeader("X-Geetest-Captcha-Output"),
			PassToken:     c.GetHeader("X-Geetest-Pass-Token"),
			GenTime:       c.GetHeader("X-Geetest-Gen-Time"),
			CaptchaID:     c.GetHeader("X-Geetest-Captcha-Id"),
		}

		valid, err := utils.ValidateGeetest(config.GlobalConfig.GeetestID, config.GlobalConfig.GeetestKey, geetestReq)
		if err != nil || !valid {
			utils.Fail(c, 403, "Captcha validation failed")
			return
		}
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
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
		fmt.Printf("[ERROR] Failed to create user: %v\n", err)
		if config.GlobalConfig.AppMode == "dev" {
			utils.Fail(c, 500, fmt.Sprintf("Failed to create user: %v", err))
			return
		}
		utils.Fail(c, 500, "Failed to create user")
		return
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

	// 生成6位验证码
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06d", rnd.Intn(1000000))

	// 存储验证码到数据库，有效期可配置（分钟）
	expiresAt := time.Now().Add(time.Duration(config.GlobalConfig.RegisterCodeExpireMinutes) * time.Minute)
	err := models.CreateVerificationCode(req.Email, code, "register", expiresAt)
	if err != nil {
		fmt.Printf("[ERROR] Failed to save verification code: %v\n", err)
		utils.Fail(c, 500, "Failed to generate verification code")
		return
	}

	// 获取语言（优先请求体，其次请求头，默认英文）
	lang := getLangFromRequest(c, req.Lang)

	tpl, err := models.GetEmailTemplate("register_code", lang)
	var subject, body string
	expireMinStr := fmt.Sprintf("%d", config.GlobalConfig.RegisterCodeExpireMinutes)

	if err == nil && tpl != nil {
		subject = strings.ReplaceAll(tpl.Subject, "{app_name}", config.GlobalConfig.AppName)
		body = strings.ReplaceAll(tpl.Content, "{code}", code)
		body = strings.ReplaceAll(body, "{app_name}", config.GlobalConfig.AppName)
		body = strings.ReplaceAll(body, "{expire_minutes}", expireMinStr)
	} else {
		// 降级使用默认硬编码内容
		if lang == "zh-CN" {
			subject = fmt.Sprintf("【%s】注册验证码", config.GlobalConfig.AppName)
			body = fmt.Sprintf("您的验证码是：%s，有效期%s分钟。", code, expireMinStr)
		} else {
			subject = fmt.Sprintf("[%s] Registration Code", config.GlobalConfig.AppName)
			body = fmt.Sprintf("Your code is: %s, valid for %s minutes.", code, expireMinStr)
		}
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
			if config.GlobalConfig.AppMode == "dev" {
				fmt.Printf("[DEV] Email send failed. Code: %s\n", code)
				fmt.Printf("[DEV] Error: %v\n", err)
				utils.Fail(c, 500, "Email send failed (Check server logs for code)")
				return
			}
			fmt.Printf("[ERROR] Failed to send email: %v\n", err)
			utils.Fail(c, 500, "Failed to send email: "+err.Error())
			return
		}
	} else {
		// 没有配置SMTP，在开发模式下直接返回验证码
		if config.GlobalConfig.AppMode != "production" {
			fmt.Printf("[DEV] SMTP not configured. Code: %s\n", code)
			utils.Fail(c, 500, "SMTP not configured (Check server logs for code)")
			return
		}
		// 生产模式下，如果没有配置SMTP，返回错误
		utils.Fail(c, 500, "SMTP service not configured")
		return
	}

	utils.Success(c, gin.H{"message": "Verification code sent"})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录并获取 Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /auth/login [post]
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

	if config.GlobalConfig.GeetestEnabled {
		geetestReq := utils.GeetestValidateRequest{
			LotNumber:     c.GetHeader("X-Geetest-Lot-Number"),
			CaptchaOutput: c.GetHeader("X-Geetest-Captcha-Output"),
			PassToken:     c.GetHeader("X-Geetest-Pass-Token"),
			GenTime:       c.GetHeader("X-Geetest-Gen-Time"),
			CaptchaID:     c.GetHeader("X-Geetest-Captcha-Id"),
		}

		valid, err := utils.ValidateGeetest(config.GlobalConfig.GeetestID, config.GlobalConfig.GeetestKey, geetestReq)
		if err != nil || !valid {
			utils.Fail(c, 403, "Captcha validation failed")
			return
		}
	}

	// 支持用户名或邮箱登录
	user, err := models.GetUserByUsernameOrEmail(username)
	if err != nil {
		if config.GlobalConfig.AppMode == "dev" {
			fmt.Printf("[LOGIN-DEBUG] user not found for '%s': %v\n", username, err)
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
		if config.GlobalConfig.AppMode == "dev" {
			pwdPrefix := ""
			pwdLen := len(user.Password)
			if pwdLen >= 4 {
				pwdPrefix = user.Password[0:4]
			}
			fmt.Printf("[LOGIN-DEBUG] password mismatch for user '%s' (id=%d), hash_prefix=%s, hash_len=%d\n",
				user.Username, user.ID, pwdPrefix, pwdLen)
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
		if config.GlobalConfig.AppMode == "dev" {
			fmt.Printf("[LOGIN-DEBUG] Failed to update login info: %v\n", err)
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
	refreshToken, err := utils.GenerateTokenWithTTL(user.ID, user.Role, refreshTTL)
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

	claims, err := utils.ParseToken(req.RefreshToken)
	if err != nil {
		utils.Fail(c, 401, "Invalid or expired refresh token")
		return
	}

	accessTTL := time.Duration(config.GlobalConfig.JWTAccessExpire) * time.Second
	refreshTTL := time.Duration(config.GlobalConfig.JWTRefreshExpire) * time.Second
	nowUnix := time.Now().Unix()

	accessToken, err := utils.GenerateTokenWithTTL(claims.UserID, claims.Role, accessTTL)
	if err != nil {
		utils.Fail(c, 500, "Failed to generate access token")
		return
	}
	refreshToken, err := utils.GenerateTokenWithTTL(claims.UserID, claims.Role, refreshTTL)
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

	// 生成6位验证码
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06d", rnd.Intn(1000000))

	// 存储验证码到数据库，15分钟有效期（重置密码链接需要更长时间）
	expiresAt := time.Now().Add(15 * time.Minute)
	err = models.CreateVerificationCode(user.Email, code, "reset_password", expiresAt)
	if err != nil {
		fmt.Printf("[ERROR] Failed to save reset code: %v\n", err)
		// 即使失败也返回成功，避免邮箱枚举攻击
		utils.Success(c, gin.H{"message": "If the email exists, a reset code has been sent"})
		return
	}

	// 构造链接 (假设前端路由是 /login/reset-password-confirm)
	// 注意：ResetPwdConfirm.vue 从 route.query.token 获取 code
	// 所以我们发送的邮件里应该包含这个链接
	frontendURL := config.GlobalConfig.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // 默认开发地址
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
			fmt.Printf("[ERROR] Failed to send email: %v\n", err)
			utils.Fail(c, 500, "Failed to send email")
			return
		}
	} else {
		if config.GlobalConfig.AppMode != "production" {
			fmt.Printf("[DEV] Reset Password Link: %s\n", resetLink)
			fmt.Printf("[DEV] Reset Code: %s\n", code)
			utils.Fail(c, 500, "SMTP not configured (Check server logs for code)")
			return
		}
		// 生产模式下，如果没有配置SMTP，返回错误
		utils.Fail(c, 500, "SMTP service not configured")
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

	// 验证Code（从数据库获取）
	valid, codeID, err := models.VerifyCode(req.Email, req.Code, "reset_password")
	if err != nil || !valid {
		utils.Fail(c, 400, "Invalid or expired reset token")
		return
	}
	// 标记验证码为已使用
	_ = models.MarkVerificationCodeAsUsed(codeID)

	// 重置成功后：清理该邮箱所有重置密码验证码
	_ = models.DeleteVerificationCodesByEmail(req.Email, "reset_password")

	user, err := models.GetUserByEmail(req.Email)
	if err != nil {
		utils.Fail(c, 400, "User not found")
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
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
