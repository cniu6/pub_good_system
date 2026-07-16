package appinit

import (
	"fst/backend/pkg/config"
	"log"
)

// ensureJWTConfig 保证现网 utils.JWT 使用的 GlobalConfig 已初始化
// 草稿栈不随机生成密钥：使用 env 中的固定/配置字符串
func ensureJWTConfig(jwtSecret string) {
	if jwtSecret == "" {
		jwtSecret = "fst-secret-key-change-in-production"
	}
	if config.GetGlobalConfig() == nil {
		// 最小配置：仅 JWT 相关，足够 GenerateToken/ParseToken
		config.SetGlobalConfig(&config.Config{
			JWTSecret:      jwtSecret,
			AdminJWTSecret: jwtSecret,
			AppMode:        "debug",
			Environment:    "development",
		})
		log.Println("  - 已初始化 JWT GlobalConfig（草稿栈）")
		return
	}
	// 若现网 config 已存在但密钥为空，则补齐（写锁内更新）
	config.UpdateGlobalConfig(func(cfg *config.Config) {
		if cfg.JWTSecret == "" {
			cfg.JWTSecret = jwtSecret
		}
		if cfg.AdminJWTSecret == "" {
			cfg.AdminJWTSecret = cfg.JWTSecret
		}
	})
}
