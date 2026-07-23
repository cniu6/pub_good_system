package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
)

const piiCipherPrefix = "enc:v1:"

var (
	piiKeyOnce sync.Once
	piiKey     []byte
	piiKeyErr  error
)

// piiMasterKey 优先读 PII_ENCRYPT_KEY；未配置时用 JWT_SECRET 派生（仅兼容开发，生产请单独配置）
func piiMasterKey() ([]byte, error) {
	piiKeyOnce.Do(func() {
		raw := strings.TrimSpace(os.Getenv("PII_ENCRYPT_KEY"))
		if raw == "" {
			raw = strings.TrimSpace(os.Getenv("JWT_SECRET"))
		}
		if raw == "" {
			piiKeyErr = errors.New("未配置 PII_ENCRYPT_KEY/JWT_SECRET，无法加密 PII")
			return
		}
		sum := sha256.Sum256([]byte(raw))
		piiKey = sum[:]
	})
	return piiKey, piiKeyErr
}

// EncryptPII 加密敏感字段；失败时返回原文并打日志由调用方决定是否继续。
func EncryptPII(plain string) (string, error) {
	plain = strings.TrimSpace(plain)
	if plain == "" || strings.HasPrefix(plain, piiCipherPrefix) {
		return plain, nil
	}
	key, err := piiMasterKey()
	if err != nil {
		return plain, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return plain, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return plain, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return plain, err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return piiCipherPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptPII 解密；非加密前缀原文原样返回（兼容存量明文）。
func DecryptPII(stored string) (string, error) {
	stored = strings.TrimSpace(stored)
	if stored == "" || !strings.HasPrefix(stored, piiCipherPrefix) {
		return stored, nil
	}
	key, err := piiMasterKey()
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, piiCipherPrefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("密文过短")
	}
	nonce, ciphertext := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
