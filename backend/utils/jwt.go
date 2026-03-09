package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"fst/backend/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint64 `json:"user_id"`
	Role   string `json:"role"`
	TokenType string `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

// RefreshClaims 用于RefreshToken的claims
type RefreshClaims struct {
	UserID uint64 `json:"user_id"`
	Role string `json:"role,omitempty"`
	TokenType string `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

const (
	accessTokenType = "access"
	refreshTokenType = "refresh"
)

func jwtSigningKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(config.GlobalConfig.JWTSecret), nil
}

func GenerateToken(userID uint64, role string) (string, error) {
	return GenerateTokenWithTTL(userID, role, 24*time.Hour)
}

func GenerateTokenWithTTL(userID uint64, role string, ttl time.Duration) (string, error) {
	expirationTime := time.Now().Add(ttl)
	claims := &Claims{
		UserID: userID,
		Role:   role,
		TokenType: accessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GlobalConfig.JWTSecret))
}

// ParseToken parses and validates a JWT token
func ParseToken(tokenString string) (*Claims, error) {
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

// GenerateRefreshToken 生成Refresh Token (7天有效期)
func GenerateRefreshToken(userID uint64, username string) (string, error) {
	return GenerateRefreshTokenWithTTL(userID, 7*24*time.Hour)
}

func GenerateRefreshTokenWithTTL(userID uint64, ttl time.Duration) (string, error) {
	expirationTime := time.Now().Add(ttl)
	claims := &RefreshClaims{
		UserID: userID,
		TokenType: refreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GlobalConfig.JWTSecret))
}

// ParseRefreshToken 解析Refresh Token
func ParseRefreshToken(tokenString string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, jwtSigningKey)

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

	return claims, nil
}
