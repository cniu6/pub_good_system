package env

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// EnvConfig 草稿栈环境配置
type EnvConfig struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	APIPort        int
	JWTSecret      string // 固定默认字符串即可；正式环境用 env（与现网 config 对齐思路）
	JWTExpireHours int
}

var envConfig *EnvConfig

// InitEnv 加载 .env，缺失则用默认值
func InitEnv() *EnvConfig {
	_ = godotenv.Load()

	envConfig = &EnvConfig{
		DBHost:         GetEnvString("DB_HOST", "127.0.0.1"),
		DBPort:         getEnvInt("DB_PORT", 3306),
		DBUser:         GetEnvString("DB_USER", "root"),
		DBPassword:     GetEnvString("DB_PASSWORD", "root"),
		DBName:         GetEnvString("DB_NAME", "fst"),
		APIPort:        getEnvInt("API_PORT", 8080),
		// JWT 默认固定密钥（不随机）；生产请设置 JWT_SECRET
		JWTSecret:      GetEnvString("JWT_SECRET", "fst-secret-key-change-in-production"),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
	}

	log.Println("环境变量加载完成")
	return envConfig
}

// GetEnv 获取配置单例
func GetEnv() *EnvConfig {
	if envConfig == nil {
		return InitEnv()
	}
	return envConfig
}

// GetEnvString 读字符串环境变量
func GetEnvString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("环境变量 %s 解析失败，使用默认值: %d", key, defaultValue)
		return defaultValue
	}
	return intValue
}
