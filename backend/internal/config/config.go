package config

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName                   string
	AppTitle                  string
	AppMode                   string
	Port                      string
	DBDriver                  string
	DBDSN                     string
	GeetestEnabled            bool
	GeetestID                 string
	GeetestKey                string
	JWTSecret                 string
	AdminPath                 string
	CorsOrigins               string
	EnableSwagger             bool
	FrontendURL               string
	SMTPHost                  string
	SMTPPort                  string
	SMTPUser                  string
	SMTPPass                  string
	SMTPSSL                   bool
	SystemEmail               string
	SystemEmailName           string
	RegisterCodeExpireMinutes int
	LoginMaxFailureCount      int // 登录最大失败次数，超过此次数将锁定账户
	LoginLockDurationMinutes  int // 账户锁定持续时间（分钟）
	JWTAccessExpire           int // Access Token 过期时间（秒）
	JWTRefreshExpire          int // Refresh Token 过期时间（秒）
	CleanupIntervalMinutes    int // 验证码清理任务间隔（分钟）
}

var GlobalConfig *Config

func InitConfig() {
	if cfg, ok := loadJSONDotEnv(".env"); ok {
		GlobalConfig = cfg
		return
	}

	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file, using default environment variables")
	}

	geetestEnabled, _ := strconv.ParseBool(strings.TrimSpace(getEnv("GEETEST_ENABLED", "false")))
	if !geetestEnabled {
		geetestEnabled, _ = strconv.ParseBool(strings.TrimSpace(getEnv("GEETEST_ENABLE", "false")))
	}

	enableSwagger, _ := strconv.ParseBool(strings.TrimSpace(getEnv("ENABLE_SWAGGER", "false")))

	geetestID := getEnv("GEETEST_ID", "")
	if geetestID == "" {
		geetestID = getEnv("GEETEST_CAPTCHA_ID", "")
	}
	geetestKey := getEnv("GEETEST_KEY", "")
	if geetestKey == "" {
		geetestKey = getEnv("GEETEST_CAPTCHA_KEY", "")
	}

	GlobalConfig = &Config{
		AppName:         getEnv("APP_NAME", "F.st"),
		AppTitle:        getEnv("APP_TITLE", "F.st - Think Fast,Run F.st"),
		AppMode:         getEnv("APP_MODE", "separate"),
		Port:            getEnv("PORT", "8080"),
		DBDriver:        getEnv("DB_DRIVER", "mysql"),
		DBDSN:           buildDSN(),
		GeetestEnabled:  geetestEnabled && geetestID != "" && geetestKey != "",
		GeetestID:       geetestID,
		GeetestKey:      geetestKey,
		JWTSecret:       getEnv("JWT_SECRET", "secret"),
		AdminPath:       getEnv("ADMIN_PATH", "/system-mgr"),
		CorsOrigins:     getEnv("CORS_ORIGINS", ""),
		EnableSwagger:   enableSwagger,
		FrontendURL:     getEnv("FRONTEND_URL", ""),
		SMTPHost:        getEnv("SMTP_HOST", ""),
		SMTPPort:        getEnv("SMTP_PORT", ""),
		SMTPUser:        getEnv("SMTP_USERNAME", ""),
		SMTPPass:        getEnv("SMTP_PASSWORD", ""),
		SMTPSSL:         getEnv("SMTP_SSL_TYPE", "") == "ssl",
		SystemEmail:     getEnv("SYSTEM_EMAIL_ADDRESS", ""),
		SystemEmailName: getEnv("SYSTEM_EMAIL_NAME", ""),
		RegisterCodeExpireMinutes: func() int {
			v, err := strconv.Atoi(strings.TrimSpace(getEnv("REGISTER_CODE_EXPIRE_MINUTES", "60")))
			if err != nil || v <= 0 {
				return 60
			}
			return v
		}(),
		LoginMaxFailureCount: func() int {
			v, err := strconv.Atoi(strings.TrimSpace(getEnv("LOGIN_MAX_FAILURE_COUNT", "5")))
			if err != nil || v <= 0 {
				return 5
			}
			return v
		}(),
		LoginLockDurationMinutes: func() int {
			v, err := strconv.Atoi(strings.TrimSpace(getEnv("LOGIN_LOCK_DURATION_MINUTES", "10")))
			if err != nil || v <= 0 {
				return 10
			}
			return v
		}(),
		JWTAccessExpire: func() int {
			v, err := strconv.Atoi(strings.TrimSpace(getEnv("JWT_ACCESS_EXPIRE", "7200")))
			if err != nil || v <= 0 {
				return 7200
			}
			return v
		}(),
		JWTRefreshExpire: func() int {
			v, err := strconv.Atoi(strings.TrimSpace(getEnv("JWT_REFRESH_EXPIRE", "604800")))
			if err != nil || v <= 0 {
				return 604800
			}
			return v
		}(),
		CleanupIntervalMinutes: func() int {
			v, err := strconv.Atoi(strings.TrimSpace(getEnv("CLEANUP_INTERVAL_MINUTES", "10")))
			if err != nil || v <= 0 {
				return 10
			}
			return v
		}(),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func buildDSN() string {
	user := getEnv("DB_USER", "root")
	pass := getEnv("DB_PASSWORD", "")
	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "3306")
	name := getEnv("DB_NAME", "fst_platform")

	return user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + name + "?charset=utf8mb4&parseTime=True&loc=Local"
}

type jsonDotEnv struct {
	DBHost                    string `json:"db_host"`
	DBPort                    string `json:"db_port"`
	DBUser                    string `json:"db_user"`
	DBPassword                string `json:"db_password"`
	DBName                    string `json:"db_name"`
	Port                      string `json:"port"`
	CorsOrigins               string `json:"cors_origins"`
	JWTSecret                 string `json:"jwt_secret"`
	JWTExpireHours            string `json:"jwt_expire_hours"`
	Debug                     string `json:"debug"`
	GeetestEnabled            string `json:"geetest_enabled"`
	GeetestCaptchaID          string `json:"geetest_captcha_id"`
	GeetestCaptchaKey         string `json:"geetest_captcha_key"`
	EnableSwagger             string `json:"enable_swagger"`
	SMTPHost                  string `json:"smtp_host"`
	SMTPPort                  string `json:"smtp_port"`
	SMTPUser                  string `json:"smtp_username"`
	SMTPPass                  string `json:"smtp_password"`
	SMTPSSL                   string `json:"smtp_ssl_type"`
	SystemEmail               string `json:"system_email_address"`
	SystemEmailName           string `json:"system_email_name"`
	RegisterCodeExpireMinutes string `json:"register_code_expire_minutes"`
	FrontendURL               string `json:"frontend_url"`
	LoginMaxFailureCount      string `json:"login_max_failure_count"`
	LoginLockDurationMinutes  string `json:"login_lock_duration_minutes"`
	JWTAccessExpire           string `json:"jwt_access_expire"`
	JWTRefreshExpire          string `json:"jwt_refresh_expire"`
	CleanupIntervalMinutes    string `json:"cleanup_interval_minutes"`
}

func loadJSONDotEnv(path string) (*Config, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	s := strings.TrimSpace(string(b))
	if !strings.HasPrefix(s, "{") {
		return nil, false
	}

	var raw jsonDotEnv
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return nil, false
	}

	port := raw.Port
	if port == "" {
		port = "8080"
	}

	jwtSecret := raw.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "secret"
	}

	geetestEnabled, _ := strconv.ParseBool(strings.TrimSpace(raw.GeetestEnabled))
	enableSwagger, _ := strconv.ParseBool(strings.TrimSpace(raw.EnableSwagger))

	user := raw.DBUser
	if user == "" {
		user = "root"
	}
	pass := raw.DBPassword
	host := raw.DBHost
	if host == "" {
		host = "127.0.0.1"
	}
	dbPort := raw.DBPort
	if dbPort == "" {
		dbPort = "3306"
	}
	name := raw.DBName
	if name == "" {
		name = "fst_platform"
	}
	dsn := user + ":" + pass + "@tcp(" + host + ":" + dbPort + ")/" + name + "?charset=utf8mb4&parseTime=True&loc=Local"

	cfg := &Config{
		AppName:         "F.st",
		AppTitle:        "F.st - Think Fast,Run F.st",
		AppMode:         "separate",
		Port:            port,
		DBDriver:        "mysql",
		DBDSN:           dsn,
		GeetestEnabled:  geetestEnabled && raw.GeetestCaptchaID != "" && raw.GeetestCaptchaKey != "",
		GeetestID:       raw.GeetestCaptchaID,
		GeetestKey:      raw.GeetestCaptchaKey,
		JWTSecret:       jwtSecret,
		AdminPath:       "/admin",
		CorsOrigins:     raw.CorsOrigins,
		EnableSwagger:   enableSwagger,
		FrontendURL:     raw.FrontendURL,
		SMTPHost:        raw.SMTPHost,
		SMTPPort:        raw.SMTPPort,
		SMTPUser:        raw.SMTPUser,
		SMTPPass:        raw.SMTPPass,
		SMTPSSL:         raw.SMTPSSL == "ssl",
		SystemEmail:     raw.SystemEmail,
		SystemEmailName: raw.SystemEmailName,
		RegisterCodeExpireMinutes: func() int {
			if raw.RegisterCodeExpireMinutes == "" {
				return 60
			}
			v, err := strconv.Atoi(strings.TrimSpace(raw.RegisterCodeExpireMinutes))
			if err != nil || v <= 0 {
				return 60
			}
			return v
		}(),
		LoginMaxFailureCount: func() int {
			if raw.LoginMaxFailureCount == "" {
				return 5
			}
			v, err := strconv.Atoi(strings.TrimSpace(raw.LoginMaxFailureCount))
			if err != nil || v <= 0 {
				return 5
			}
			return v
		}(),
		LoginLockDurationMinutes: func() int {
			if raw.LoginLockDurationMinutes == "" {
				return 10
			}
			v, err := strconv.Atoi(strings.TrimSpace(raw.LoginLockDurationMinutes))
			if err != nil || v <= 0 {
				return 10
			}
			return v
		}(),
		JWTAccessExpire: func() int {
			if raw.JWTAccessExpire == "" {
				return 7200
			}
			v, err := strconv.Atoi(strings.TrimSpace(raw.JWTAccessExpire))
			if err != nil || v <= 0 {
				return 7200
			}
			return v
		}(),
		JWTRefreshExpire: func() int {
			if raw.JWTRefreshExpire == "" {
				return 604800
			}
			v, err := strconv.Atoi(strings.TrimSpace(raw.JWTRefreshExpire))
			if err != nil || v <= 0 {
				return 604800
			}
			return v
		}(),
		CleanupIntervalMinutes: func() int {
			if raw.CleanupIntervalMinutes == "" {
				return 10
			}
			v, err := strconv.Atoi(strings.TrimSpace(raw.CleanupIntervalMinutes))
			if err != nil || v <= 0 {
				return 10
			}
			return v
		}(),
	}
	log.Printf("[Config] RegisterCodeExpireMinutes: %d\n", cfg.RegisterCodeExpireMinutes)
	log.Printf("[Config] LoginMaxFailureCount: %d\n", cfg.LoginMaxFailureCount)
	log.Printf("[Config] LoginLockDurationMinutes: %d\n", cfg.LoginLockDurationMinutes)
	return cfg, true
}
