package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"fst/backend/pkg/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    uint64 `json:"user_id"`
	Role      string `json:"role"`
	AuthGuard string `json:"auth_guard,omitempty"`
	TokenType string `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

// RefreshClaims 用于RefreshToken的claims
type RefreshClaims struct {
	UserID    uint64 `json:"user_id"`
	Role      string `json:"role,omitempty"`
	AuthGuard string `json:"auth_guard,omitempty"`
	TokenType string `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

const (
	UserAuthGuard    = "user"
	AdminAuthGuard   = "admin"
	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

func getJWTSecretByGuard(authGuard string) (string, error) {
	// 使用快照读取，避免热更新配置时与写路径竞态
	cfg := config.CloneGlobalConfig()
	if cfg == nil {
		return "", fmt.Errorf("JWT config not initialized")
	}

	secret := cfg.JWTSecret
	if authGuard == AdminAuthGuard && cfg.AdminJWTSecret != "" {
		secret = cfg.AdminJWTSecret
	}
	if secret == "" {
		return "", fmt.Errorf("JWT secret not configured")
	}
	return secret, nil
}

func jwtSigningKeyByGuard(authGuard string) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		secret, err := getJWTSecretByGuard(authGuard)
		if err != nil {
			return nil, err
		}
		return []byte(secret), nil
	}
}

func jwtSigningKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	secret, err := getJWTSecretByGuard(UserAuthGuard)
	if err != nil {
		return nil, err
	}
	return []byte(secret), nil
}

func GenerateTokenWithTTL(userID uint64, role string, ttl time.Duration) (string, error) {
	return GenerateTokenForGuardWithTTL(userID, role, UserAuthGuard, ttl)
}

func GenerateTokenForGuardWithTTL(userID uint64, role, authGuard string, ttl time.Duration) (string, error) {
	if authGuard == "" {
		authGuard = UserAuthGuard
	}
	expirationTime := time.Now().Add(ttl)
	claims := &Claims{
		UserID:    userID,
		Role:      role,
		AuthGuard: authGuard,
		TokenType: accessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret, err := getJWTSecretByGuard(authGuard)
	if err != nil {
		return "", err
	}
	return token.SignedString([]byte(secret))
}

// ParseToken parses and validates a JWT token
func ParseToken(tokenString string) (*Claims, error) {
	return ParseTokenForGuard(tokenString, UserAuthGuard)
}

func ParseTokenForGuard(tokenString, expectedGuard string) (*Claims, error) {
	claims := &Claims{}
	if expectedGuard == "" {
		expectedGuard = UserAuthGuard
	}
	token, err := jwt.ParseWithClaims(tokenString, claims, jwtSigningKeyByGuard(expectedGuard))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	if claims.TokenType != "" {
		if claims.TokenType != accessTokenType {
			return nil, fmt.Errorf("unexpected token type: %s", claims.TokenType)
		}
	} else if claims.Role == "" {
		return nil, fmt.Errorf("unexpected token type")
	}

	authGuard := claims.AuthGuard
	if authGuard == "" {
		authGuard = UserAuthGuard
	}
	if authGuard != expectedGuard {
		return nil, fmt.Errorf("unexpected auth guard: %s", authGuard)
	}
	if expectedGuard == AdminAuthGuard && claims.Role != AdminAuthGuard {
		return nil, fmt.Errorf("admin token requires admin role")
	}
	claims.AuthGuard = authGuard
	return claims, nil
}

// ParseTokenForGuardIgnoreExpiry 与 ParseTokenForGuard 类似，但不校验过期时间（exp）。
// 仅用于"token 已过期后，前端仍希望通知后端撤销该会话记录"这种场景：
// 签名/guard/角色等其它校验都不放宽，只是允许 exp 已过去，避免过期 token 无法定位到具体会话。
// 绝不能用它来放行任何需要鉴权的正常业务接口。
func ParseTokenForGuardIgnoreExpiry(tokenString, expectedGuard string) (*Claims, error) {
	claims := &Claims{}
	if expectedGuard == "" {
		expectedGuard = UserAuthGuard
	}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.ParseWithClaims(tokenString, claims, jwtSigningKeyByGuard(expectedGuard))
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	if claims.TokenType != "" {
		if claims.TokenType != accessTokenType {
			return nil, fmt.Errorf("unexpected token type: %s", claims.TokenType)
		}
	} else if claims.Role == "" {
		return nil, fmt.Errorf("unexpected token type")
	}

	authGuard := claims.AuthGuard
	if authGuard == "" {
		authGuard = UserAuthGuard
	}
	if authGuard != expectedGuard {
		return nil, fmt.Errorf("unexpected auth guard: %s", authGuard)
	}
	if expectedGuard == AdminAuthGuard && claims.Role != AdminAuthGuard {
		return nil, fmt.Errorf("admin token requires admin role")
	}
	claims.AuthGuard = authGuard
	return claims, nil
}

// ParseTokenLegacy keeps compatibility for older callers.
func ParseTokenLegacy(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, jwtSigningKey)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	if claims.TokenType != "" {
		if claims.TokenType != accessTokenType {
			return nil, fmt.Errorf("unexpected token type: %s", claims.TokenType)
		}
	} else if claims.Role == "" {
		return nil, fmt.Errorf("unexpected token type")
	}

	return claims, nil
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func GenerateRefreshTokenWithTTL(userID uint64, ttl time.Duration) (string, error) {
	return GenerateRefreshTokenForGuardWithTTL(userID, UserAuthGuard, ttl)
}

func GenerateRefreshTokenForGuardWithTTL(userID uint64, authGuard string, ttl time.Duration) (string, error) {
	if authGuard == "" {
		authGuard = UserAuthGuard
	}
	expirationTime := time.Now().Add(ttl)
	claims := &RefreshClaims{
		UserID:    userID,
		AuthGuard: authGuard,
		TokenType: refreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret, err := getJWTSecretByGuard(authGuard)
	if err != nil {
		return "", err
	}
	return token.SignedString([]byte(secret))
}

// ParseRefreshToken 解析Refresh Token
func ParseRefreshToken(tokenString string) (*RefreshClaims, error) {
	return ParseRefreshTokenForGuard(tokenString, UserAuthGuard)
}

func ParseRefreshTokenForGuard(tokenString, expectedGuard string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	if expectedGuard == "" {
		expectedGuard = UserAuthGuard
	}
	token, err := jwt.ParseWithClaims(tokenString, claims, jwtSigningKeyByGuard(expectedGuard))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	if claims.TokenType != "" {
		if claims.TokenType != refreshTokenType {
			return nil, fmt.Errorf("unexpected token type: %s", claims.TokenType)
		}
	} else if claims.Role != "" {
		return nil, fmt.Errorf("unexpected token type")
	}

	authGuard := claims.AuthGuard
	if authGuard == "" {
		authGuard = UserAuthGuard
	}
	if authGuard != expectedGuard {
		return nil, fmt.Errorf("unexpected auth guard: %s", authGuard)
	}
	claims.AuthGuard = authGuard
	return claims, nil
}

