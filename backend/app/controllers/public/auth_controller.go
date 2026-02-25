package public

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/config"
	"fst/backend/internal/middleware"
	"fst/backend/utils"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthController 公共认证控制器（无需登录）
type AuthController struct {
	auth_svc  *services.AuthService
	email_svc *services.EmailService
}

// NewAuthController 创建认证控制器
func NewAuthController() *AuthController {
	return &AuthController{
		auth_svc:  services.NewAuthService(),
		email_svc: services.NewEmailService(),
	}
}

// ========================================
// 请求结构体
// ========================================

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

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// ========================================
// 辅助函数
// ========================================

// getLangFromRequest 从请求中获取语言参数
func getLangFromRequest(c *gin.Context, reqLang string) string {
	lang := reqLang
	if lang == "" {
		lang = c.GetHeader("Accept-Language")
	}
	if lang == "" {
		lang = "en-US"
	}
	if strings.Contains(lang, "zh") {
		lang = "zh-CN"
	} else {
		lang = "en-US"
	}
	return lang
}

// ========================================
// 控制器方法
// ========================================

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录并获取 Token
// @Tags Public-认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录信息"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/public/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入
	req.UserName = utils.Clean_XSS(req.UserName)
	req.Username = utils.Clean_XSS(req.Username)

	username := req.UserName
	if username == "" {
		username = req.Username
	}
	if username == "" {
		utils.Fail(c, 400, "username is required")
		return
	}

	// 极验验证
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

	// 调用服务层登录
	result, err := ctrl.auth_svc.Login(username, req.Password, clientIP)
	if err != nil {
		if config.GlobalConfig.AppMode == "dev" {
			fmt.Printf("[LOGIN-DEBUG] %v\n", err)
		}
		utils.Fail(c, err.Code, err.Message)
		return
	}

	utils.Success(c, result)
}

// Register 注册新用户
// @Summary 用户注册
// @Description 注册一个新用户
// @Tags Public-认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /api/v1/public/register [post]
func (ctrl *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入
	req.Username = utils.Clean_XSS(req.Username)
	req.Email = utils.Clean_XSS(req.Email)
	req.Code = utils.Clean_XSS(req.Code)

	// 验证验证码
	valid, codeID, err := models.VerifyCode(req.Email, req.Code, "register")
	if err != nil || !valid {
		utils.Fail(c, 400, "Invalid or expired verification code")
		return
	}
	_ = models.MarkVerificationCodeAsUsed(codeID)
	_ = models.DeleteVerificationCodesByEmail(req.Email, "register")

	// 验证用户名格式
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)
	if !usernameRegex.MatchString(req.Username) {
		utils.Fail(c, 400, "Username must be 3-50 characters long and contain only letters, numbers, and underscores")
		return
	}

	// 检查用户名是否存在
	if _, err := models.GetUserByUsername(req.Username); err == nil {
		utils.Fail(c, 400, "Username already exists")
		return
	}

	// 检查邮箱是否存在
	if _, err := models.GetUserByEmail(req.Email); err == nil {
		utils.Fail(c, 400, "Email already exists")
		return
	}

	// 极验验证
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

	// 创建用户
	user := &models.User{
		Username: req.Username,
		Password: req.Password, // 服务层会进行哈希
		Email:    req.Email,
		Role:     "user",
		Status:   1,
	}

	if err := ctrl.auth_svc.Register(user); err != nil {
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
// @Summary 发送注册验证码
// @Description 发送注册验证码到邮箱
// @Tags Public-认证
// @Accept json
// @Produce json
// @Param request body SendCodeRequest true "邮箱信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/public/send-register-code [post]
func (ctrl *AuthController) SendRegisterCode(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入
	req.Email = utils.Clean_XSS(req.Email)
	req.Lang = utils.Clean_XSS(req.Lang)

	// 生成验证码
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06d", rnd.Intn(1000000))

	// 存储验证码
	expireMinutes := config.GlobalConfig.RegisterCodeExpireMinutes
	expiresAt := time.Now().Add(time.Duration(expireMinutes) * time.Minute)
	if err := models.CreateVerificationCode(req.Email, code, "register", expiresAt); err != nil {
		utils.Fail(c, 500, "Failed to generate verification code")
		return
	}

	// 获取语言
	lang := getLangFromRequest(c, req.Lang)

	// 检查邮件服务是否可用
	if !ctrl.email_svc.IsEmailConfigured() {
		if config.GlobalConfig.AppMode != "production" {
			fmt.Printf("[DEV] SMTP not configured. Code: %s\n", code)
			utils.Fail(c, 500, "SMTP not configured (Check server logs for code)")
			return
		}
		utils.Fail(c, 500, "SMTP service not configured")
		return
	}

	// 发送验证码邮件
	vars := map[string]string{
		"code":            code,
		"expire_minutes":  fmt.Sprintf("%d", expireMinutes),
	}

	if err := ctrl.email_svc.SendTemplateEmail(req.Email, "register_code", lang, vars); err != nil {
		if config.GlobalConfig.AppMode == "dev" {
			fmt.Printf("[DEV] Email send failed. Code: %s, Error: %v\n", code, err)
		}
		utils.Fail(c, 500, "Failed to send email")
		return
	}

	utils.Success(c, gin.H{"message": "Verification code sent"})
}

// SendResetEmail 发送重置密码邮件
// @Summary 发送重置密码邮件
// @Description 发送重置密码验证码到邮箱
// @Tags Public-认证
// @Accept json
// @Produce json
// @Param request body ResetEmailRequest true "邮箱信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/public/forgot-password [post]
func (ctrl *AuthController) SendResetEmail(c *gin.Context) {
	var req ResetEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入
	req.Email = utils.Clean_XSS(req.Email)
	req.Lang = utils.Clean_XSS(req.Lang)

	// 检查邮箱是否存在
	user, err := models.GetUserByUsernameOrEmail(req.Email)
	if err != nil || user == nil {
		// 安全考虑：即使邮箱不存在也返回成功
		utils.Success(c, gin.H{"message": "If the email exists, a reset code has been sent"})
		return
	}

	// 生成验证码
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06d", rnd.Intn(1000000))

	// 存储验证码
	expiresAt := time.Now().Add(15 * time.Minute)
	if err := models.CreateVerificationCode(user.Email, code, "reset_password", expiresAt); err != nil {
		utils.Success(c, gin.H{"message": "If the email exists, a reset code has been sent"})
		return
	}

	// 构造重置链接
	frontendURL := config.GlobalConfig.FrontendURL
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	resetLink := fmt.Sprintf("%s/#/login/reset-password-confirm?email=%s&token=%s", frontendURL, user.Email, code)

	// 获取语言
	lang := getLangFromRequest(c, req.Lang)

	// 检查邮件服务
	if !ctrl.email_svc.IsEmailConfigured() {
		if config.GlobalConfig.AppMode != "production" {
			fmt.Printf("[DEV] Reset Link: %s\n", resetLink)
			fmt.Printf("[DEV] Reset Code: %s\n", code)
			utils.Fail(c, 500, "SMTP not configured (Check server logs for code)")
			return
		}
		utils.Fail(c, 500, "SMTP service not configured")
		return
	}

	// 发送邮件
	vars := map[string]string{
		"code": code,
		"link": resetLink,
	}

	if err := ctrl.email_svc.SendTemplateEmail(user.Email, "reset_password", lang, vars); err != nil {
		utils.Fail(c, 500, "Failed to send email")
		return
	}

	utils.Success(c, gin.H{"message": "Reset email sent"})
}

// ResetPasswordConfirm 确认重置密码
// @Summary 确认重置密码
// @Description 使用验证码重置密码
// @Tags Public-认证
// @Accept json
// @Produce json
// @Param request body ResetPasswordConfirmRequest true "重置信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/public/reset-password [post]
func (ctrl *AuthController) ResetPasswordConfirm(c *gin.Context) {
	var req ResetPasswordConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	// 过滤用户输入
	req.Email = utils.Clean_XSS(req.Email)
	req.Code = utils.Clean_XSS(req.Code)

	// 验证验证码
	valid, codeID, err := models.VerifyCode(req.Email, req.Code, "reset_password")
	if err != nil || !valid {
		utils.Fail(c, 400, "Invalid or expired reset token")
		return
	}
	_ = models.MarkVerificationCodeAsUsed(codeID)
	_ = models.DeleteVerificationCodesByEmail(req.Email, "reset_password")

	// 获取用户
	user, err := models.GetUserByEmail(req.Email)
	if err != nil {
		utils.Fail(c, 400, "User not found")
		return
	}

	// 更新密码
	if err := ctrl.auth_svc.UpdatePassword(user.ID, req.NewPassword); err != nil {
		utils.Fail(c, 500, "Failed to update password")
		return
	}

	utils.Success(c, gin.H{"message": "Password reset successfully"})
}

// UpdateToken 刷新Token
// @Summary 刷新Token
// @Description 使用refresh token获取新的access token
// @Tags Public-认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌"
// @Success 200 {object} utils.Response
// @Router /api/v1/public/refresh-token [post]
func (ctrl *AuthController) UpdateToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	result, err := ctrl.auth_svc.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.Fail(c, err.Code, err.Message)
		return
	}

	utils.Success(c, result)
}

// RegisterRoutes 注册公共认证路由
func (ctrl *AuthController) RegisterRoutes(group *gin.RouterGroup) {
	// 应用严格限流
	authGroup := group.Group("")
	authGroup.Use(middleware.StrictRateLimitMiddleware())
	{
		authGroup.POST("/login", ctrl.Login)
		authGroup.POST("/register", ctrl.Register)
		authGroup.POST("/send-register-code", ctrl.SendRegisterCode)
		authGroup.POST("/forgot-password", ctrl.SendResetEmail)
		authGroup.POST("/reset-password", ctrl.ResetPasswordConfirm)
		authGroup.POST("/refresh-token", ctrl.UpdateToken)
	}
}
