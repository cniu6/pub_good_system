package utils

import (
	"testing"
	"time"

	"fst/backend/pkg/config"
	"github.com/pquerna/otp/totp"
)

func TestValidateTOTP_RoundTrip(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP" // 标准测试密钥
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("GenerateCode: %v", err)
	}
	if !ValidateTOTP(secret, code) {
		t.Fatal("当前码应校验通过")
	}
	if ValidateTOTP(secret, "000000") {
		t.Fatal("错误码不应通过")
	}
	if ValidateTOTP("", code) {
		t.Fatal("空密钥不应通过")
	}
}

func TestTOTPPendingToken_RoundTrip(t *testing.T) {
	old := config.CloneGlobalConfig()
	config.SetGlobalConfig(&config.Config{
		JWTSecret:      "unit-test-secret",
		AdminJWTSecret: "unit-admin-secret",
	})
	defer config.SetGlobalConfig(old)

	tok, err := GenerateTOTPPendingToken(99, AdminAuthGuard, time.Minute)
	if err != nil {
		t.Fatalf("GenerateTOTPPendingToken: %v", err)
	}
	claims, err := ParseTOTPPendingToken(tok)
	if err != nil {
		t.Fatalf("ParseTOTPPendingToken: %v", err)
	}
	if claims.UserID != 99 || claims.TokenType != totpPendingType {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	access, err := GenerateTokenForGuardWithTTL(99, AdminAuthGuard, AdminAuthGuard, time.Minute)
	if err != nil {
		t.Fatalf("GenerateTokenForGuardWithTTL: %v", err)
	}
	if _, err := ParseTOTPPendingToken(access); err == nil {
		t.Fatal("access token 不应被当作 totp_pending")
	}
}
