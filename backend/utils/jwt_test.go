package utils

import (
	"fst/backend/internal/config"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func useTestJWTConfig() func() {
	old := config.GlobalConfig
	config.GlobalConfig = &config.Config{JWTSecret: "unit-test-secret"}
	return func() {
		config.GlobalConfig = old
	}
}

func TestTokenTypeSeparationCurrentTokens(t *testing.T) {
	restore := useTestJWTConfig()
	defer restore()

	accessToken, err := GenerateTokenWithTTL(1, "user", time.Minute)
	if err != nil {
		t.Fatalf("GenerateTokenWithTTL returned error: %v", err)
	}
	refreshToken, err := GenerateRefreshTokenWithTTL(1, time.Minute)
	if err != nil {
		t.Fatalf("GenerateRefreshTokenWithTTL returned error: %v", err)
	}

	if _, err := ParseToken(accessToken); err != nil {
		t.Fatalf("ParseToken should accept access token: %v", err)
	}
	if _, err := ParseRefreshToken(refreshToken); err != nil {
		t.Fatalf("ParseRefreshToken should accept refresh token: %v", err)
	}
	if _, err := ParseRefreshToken(accessToken); err == nil {
		t.Fatal("ParseRefreshToken should reject access token")
	}
	if _, err := ParseToken(refreshToken); err == nil {
		t.Fatal("ParseToken should reject refresh token")
	}
}

func TestTokenTypeSeparationLegacyTokens(t *testing.T) {
	restore := useTestJWTConfig()
	defer restore()

	legacyAccessClaims := &Claims{
		UserID: 2,
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}
	legacyAccessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, legacyAccessClaims).SignedString([]byte(config.GlobalConfig.JWTSecret))
	if err != nil {
		t.Fatalf("failed to sign legacy access token: %v", err)
	}

	legacyRefreshClaims := &RefreshClaims{
		UserID: 2,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}
	legacyRefreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, legacyRefreshClaims).SignedString([]byte(config.GlobalConfig.JWTSecret))
	if err != nil {
		t.Fatalf("failed to sign legacy refresh token: %v", err)
	}

	if _, err := ParseToken(legacyAccessToken); err != nil {
		t.Fatalf("ParseToken should accept legacy access token: %v", err)
	}
	if _, err := ParseRefreshToken(legacyRefreshToken); err != nil {
		t.Fatalf("ParseRefreshToken should accept legacy refresh token: %v", err)
	}
	if _, err := ParseRefreshToken(legacyAccessToken); err == nil {
		t.Fatal("ParseRefreshToken should reject legacy access token")
	}
	if _, err := ParseToken(legacyRefreshToken); err == nil {
		t.Fatal("ParseToken should reject legacy refresh token")
	}
}

func TestAuthGuardSeparationAccessToken(t *testing.T) {
	restore := useTestJWTConfig()
	defer restore()

	adminToken, err := GenerateTokenForGuardWithTTL(3, "admin", AdminAuthGuard, time.Minute)
	if err != nil {
		t.Fatalf("GenerateTokenForGuardWithTTL(admin) returned error: %v", err)
	}
	userToken, err := GenerateTokenForGuardWithTTL(3, "user", UserAuthGuard, time.Minute)
	if err != nil {
		t.Fatalf("GenerateTokenForGuardWithTTL(user) returned error: %v", err)
	}

	if _, err := ParseTokenForGuard(adminToken, AdminAuthGuard); err != nil {
		t.Fatalf("ParseTokenForGuard should accept admin token: %v", err)
	}
	if _, err := ParseTokenForGuard(userToken, UserAuthGuard); err != nil {
		t.Fatalf("ParseTokenForGuard should accept user token: %v", err)
	}
	if _, err := ParseTokenForGuard(adminToken, UserAuthGuard); err == nil {
		t.Fatal("ParseTokenForGuard should reject admin token for user guard")
	}
	if _, err := ParseTokenForGuard(userToken, AdminAuthGuard); err == nil {
		t.Fatal("ParseTokenForGuard should reject user token for admin guard")
	}
}

func TestAuthGuardSeparationRefreshToken(t *testing.T) {
	restore := useTestJWTConfig()
	defer restore()

	adminRefreshToken, err := GenerateRefreshTokenForGuardWithTTL(7, AdminAuthGuard, time.Minute)
	if err != nil {
		t.Fatalf("GenerateRefreshTokenForGuardWithTTL(admin) returned error: %v", err)
	}
	userRefreshToken, err := GenerateRefreshTokenForGuardWithTTL(7, UserAuthGuard, time.Minute)
	if err != nil {
		t.Fatalf("GenerateRefreshTokenForGuardWithTTL(user) returned error: %v", err)
	}

	if _, err := ParseRefreshTokenForGuard(adminRefreshToken, AdminAuthGuard); err != nil {
		t.Fatalf("ParseRefreshTokenForGuard should accept admin refresh token: %v", err)
	}
	if _, err := ParseRefreshTokenForGuard(userRefreshToken, UserAuthGuard); err != nil {
		t.Fatalf("ParseRefreshTokenForGuard should accept user refresh token: %v", err)
	}
	if _, err := ParseRefreshTokenForGuard(adminRefreshToken, UserAuthGuard); err == nil {
		t.Fatal("ParseRefreshTokenForGuard should reject admin refresh token for user guard")
	}
	if _, err := ParseRefreshTokenForGuard(userRefreshToken, AdminAuthGuard); err == nil {
		t.Fatal("ParseRefreshTokenForGuard should reject user refresh token for admin guard")
	}
}
