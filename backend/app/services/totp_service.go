package services

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/utils"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TotpSetupResult 绑定前返回给前端的密钥与 otpauth URL
type TotpSetupResult struct {
	Secret     string `json:"secret"`
	OtpauthURL string `json:"otpauth_url"`
}

// SetupAdminTOTP 为当前管理员生成（或重置未启用的）TOTP 密钥，写入但暂不启用
func SetupAdminTOTP(userID uint64, accountName string) (*TotpSetupResult, error) {
	user, err := models.GetUserByID(userID)
	if err != nil || user == nil {
		return nil, NewClientError("用户不存在")
	}
	if user.Role != "admin" {
		return nil, NewClientError("仅管理员可配置 TOTP")
	}
	if user.TotpEnabled {
		return nil, NewClientError("已启用 TOTP，请先禁用再重新绑定")
	}

	issuer := "FST-Admin"
	if setting, serr := models.GetSettingByKey("site_name"); serr == nil && setting != nil {
		name := strings.TrimSpace(setting.Value)
		if name != "" {
			issuer = name + "-Admin"
		}
	}
	if accountName == "" {
		accountName = user.Username
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
		SecretSize:  20,
	})
	if err != nil {
		return nil, fmt.Errorf("generate totp: %w", err)
	}

	secret := key.Secret()
	if err := models.UpdateUserTOTPSecret(userID, secret, false); err != nil {
		return nil, fmt.Errorf("save totp secret: %w", err)
	}

	return &TotpSetupResult{
		Secret:     secret,
		OtpauthURL: key.URL(),
	}, nil
}

// EnableAdminTOTP 校验动态码后启用 TOTP
func EnableAdminTOTP(userID uint64, code string) error {
	user, err := models.GetUserByID(userID)
	if err != nil || user == nil {
		return NewClientError("用户不存在")
	}
	if user.TotpSecret == nil || strings.TrimSpace(*user.TotpSecret) == "" {
		return NewClientError("请先调用 setup 获取密钥")
	}
	if !utils.ValidateTOTP(*user.TotpSecret, code) {
		return NewClientError("验证码错误")
	}
	return models.UpdateUserTOTPSecret(userID, *user.TotpSecret, true)
}

// DisableAdminTOTP 禁用并清空密钥（需正确动态码）
func DisableAdminTOTP(userID uint64, code string) error {
	user, err := models.GetUserByID(userID)
	if err != nil || user == nil {
		return NewClientError("用户不存在")
	}
	if !user.TotpEnabled {
		return NewClientError("未启用 TOTP")
	}
	if user.TotpSecret == nil || !utils.ValidateTOTP(*user.TotpSecret, code) {
		return NewClientError("验证码错误")
	}
	return models.ClearUserTOTP(userID)
}

// CompleteAdminLoginWithTOTP 用临时令牌 + 动态码完成管理端登录
func CompleteAdminLoginWithTOTP(tempToken, code, clientIP string) (*LoginResult, *ServiceError) {
	claims, err := utils.ParseTOTPPendingToken(tempToken)
	if err != nil {
		return nil, NewServiceError(401, "Invalid or expired temp token")
	}
	user, err := models.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		return nil, NewServiceError(401, "User not found")
	}
	if accessErr := validateUserAccessForGuard(user, utils.AdminAuthGuard); accessErr != nil {
		return nil, accessErr
	}
	if !user.TotpEnabled || user.TotpSecret == nil {
		return nil, NewServiceError(400, "TOTP not enabled")
	}
	if !utils.ValidateTOTP(*user.TotpSecret, code) {
		return nil, NewServiceError(401, "Invalid TOTP code")
	}

	_ = models.UpdateLoginInfo(user.ID, clientIP)

	_, _, accessExpireSeconds, refreshExpireSeconds := getAuthRuntimeDefaults()
	accessTTL := time.Duration(accessExpireSeconds) * time.Second
	refreshTTL := time.Duration(refreshExpireSeconds) * time.Second

	accessToken, err := utils.GenerateTokenForGuardWithTTL(user.ID, user.Role, utils.AdminAuthGuard, accessTTL)
	if err != nil {
		return nil, NewServiceError(500, "Failed to generate access token")
	}
	refreshToken, err := utils.GenerateRefreshTokenForGuardWithTTL(user.ID, utils.AdminAuthGuard, refreshTTL)
	if err != nil {
		return nil, NewServiceError(500, "Failed to generate refresh token")
	}

	return &LoginResult{
		ID:               user.ID,
		UserName:         user.Username,
		Email:            user.Email,
		Role:             []string{user.Role},
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresAt:        time.Now().Unix() + int64(accessTTL.Seconds()),
		RefreshExpiresAt: time.Now().Unix() + int64(refreshTTL.Seconds()),
		Realname:         buildLoginRealnameSummary(user.ID),
	}, nil
}
