package utils

import (
	"time"

	"github.com/pquerna/otp/totp"
)

// ValidateTOTP 校验 TOTP 动态码（允许 ±1 步长窗口）
func ValidateTOTP(secret, code string) bool {
	if secret == "" || code == "" {
		return false
	}
	return totp.Validate(code, secret)
}

// GenerateTOTPCode 生成当前时刻 TOTP（仅测试用）
func GenerateTOTPCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}
