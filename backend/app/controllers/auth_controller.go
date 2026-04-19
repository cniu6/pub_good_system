package controllers

import (
	"fmt"
	public "fst/backend/app/controllers/public"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	crypto_rand "crypto/rand"
	"fst/backend/utils"
	"math/big"
	"math/rand"
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
	return services.GetGlobalAllowRegister()
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
	public.NewAuthController().Register(c)
}

// SendRegisterCode 发送注册验证码
func (ctrl *AuthController) SendRegisterCode(c *gin.Context) {
	public.NewAuthController().SendRegisterCode(c)
}

// Login 用户登录 (已迁移到 public.AuthController)
// Deprecated: 请使用 /api/v1/public/login
func (ctrl *AuthController) Login(c *gin.Context) {
	public.NewAuthController().Login(c)
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (ctrl *AuthController) UpdateToken(c *gin.Context) {
	public.NewAuthController().UpdateToken(c)
}

func (ctrl *AuthController) GetUserRoutes(c *gin.Context) {
	utils.Success(c, []interface{}{})
}

// SendResetEmail 发送重置密码邮件
func (ctrl *AuthController) SendResetEmail(c *gin.Context) {
	public.NewAuthController().SendResetEmail(c)
}

// ResetPasswordConfirm 确认重置密码
func (ctrl *AuthController) ResetPasswordConfirm(c *gin.Context) {
	public.NewAuthController().ResetPasswordConfirm(c)
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

