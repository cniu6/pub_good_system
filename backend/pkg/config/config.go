package config

import (
	crypto_rand "crypto/rand"
	"encoding/json"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName                   string
	AppTitle                  string
	AppMode                   string
	Environment               string
	Port                      string
	HTTPReadHeaderTimeoutSeconds int
	HTTPReadTimeoutSeconds       int
	HTTPWriteTimeoutSeconds      int
	HTTPIdleTimeoutSeconds       int
	HTTPShutdownTimeoutSeconds   int
	HTTPMaxHeaderBytes           int
	DBDriver                  string
	DBDSN                     string
	GeetestEnabled            bool
	GeetestID                 string
	GeetestKey                string
	JWTSecret                 string
	AdminJWTSecret            string
	AdminPath                 string // 管理后台前端页面入口（隐藏路径，如 /system-mgr）
	AdminAPIPath              string // 管理端 REST API 在 /api/v1 下的前缀（默认 /admin）
	CorsOrigins               string
	EnableSwagger             bool
	// EnableAdminDebugOps 是否开放管理端 debug/pprof/强制 GC 等高危运维接口。
	// 生产环境一律视为关闭；非生产默认 true，可用 ENABLE_ADMIN_DEBUG=false 关闭。
	EnableAdminDebugOps       bool
	FrontendURL               string
	BackendAPIURL             string
	SMTPHost                  string
	SMTPPort                  string
	SMTPUser                  string
	SMTPPass                  string
	SMTPSSL                   bool
	SystemEmail               string
	SystemEmailName           string
	RegisterCodeExpireMinutes int
	LoginMaxFailureCount      int    // 登录最大失败次数，超过此次数将锁定账户
	LoginLockDurationMinutes  int    // 账户锁定持续时间（分钟）
	JWTAccessExpire           int    // Access Token 过期时间（秒）
	JWTRefreshExpire          int    // Refresh Token 过期时间（秒）
	CleanupIntervalMinutes    int    // 验证码清理任务间隔（分钟）
	EmailVerifyEnabled        bool   // 邮箱验证码功能开关
	SMSVerifyEnabled          bool   // 短信验证码功能开关
	SMSProvider               string // 短信服务商: aliyun, tencent, console
	SMSAccessKey              string // 短信服务 AccessKey
	SMSSecretKey              string // 短信服务 SecretKey
	SMSSignName               string // 短信签名
	SMSTemplateCode           string // 短信验证码模板ID
	SMSTemplateCodeEN         string // 英文短信模板ID
	SMSRegion                 string // 短信服务区域
	SMSEndpoint               string // 自定义短信网关地址
	SMSSdkAppID               string // 腾讯云 SmsSdkAppId
	SMSBodyFormat             string // 自定义短信请求体格式
}

var (
	globalConfigMu sync.RWMutex
	GlobalConfig   *Config
)

const defaultJWTSecret = "secret"

func isProductionEnvMode(mode string) bool {
	mode = strings.ToLower(strings.TrimSpace(mode))
	return mode == "prod" || mode == "production" || mode == "release"
}

func resolveRuntimeEnv(candidates ...string) string {
	for _, candidate := range candidates {
		trimmed := strings.TrimSpace(candidate)
		if trimmed != "" {
			return trimmed
		}
	}
	return "development"
}

func generateDevelopmentSecret() string {
	buf := make([]byte, 32)
	if _, err := crypto_rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)
	}
	return "dev-fallback-secret-please-configure-jwt-secret"
}

func resolveJWTSecret(secret string) string {
	secret = strings.TrimSpace(secret)
	if secret != "" {
		return secret
	}
	generated := generateDevelopmentSecret()
	log.Println("[Security Warning] JWT_SECRET 未配置，已生成临时开发密钥")
	return generated
}

// isWeakJWTSecret 判断是否为明显不安全/占位密钥（仅开发环境 warning 用，生产不再因弱密钥 fatal）
func isWeakJWTSecret(secret string) bool {
	secret = strings.TrimSpace(secret)
	if secret == "" || secret == defaultJWTSecret {
		return true
	}
	// 常见占位串（与 .env.example 示例值对齐）
	weakSet := map[string]bool{
		"your_jwt_secret":                   true,
		"change_me_to_a_long_random_secret": true,
		"jwt_secret":                        true,
		"secret":                            true,
		"123456":                            true,
	}
	if weakSet[strings.ToLower(secret)] {
		return true
	}
	return len(secret) < 16
}

// NormalizeAdminAPIPath 规范化管理 API 前缀：保证以 / 开头、去掉尾斜杠；空则默认 /admin
func NormalizeAdminAPIPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/admin"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(path, "/")
}

func validateCriticalSecurityConfig(cfg *Config) {
	if cfg == nil {
		return
	}

	secret := strings.TrimSpace(cfg.JWTSecret)
	prod := isProductionEnvMode(cfg.Environment) || isProductionEnvMode(cfg.AppMode)
	cors := strings.TrimSpace(cfg.CorsOrigins)

	// CORS：任意环境都必须配置；允许 * 与泛域名（如 *.example.com）
	if cors == "" {
		log.Fatal("[Security] CORS_ORIGINS 不能为空，请在根目录 .env 配置（可用 * 或具体域名/泛域名）")
	}

	if prod {
		// 生产：JWT 仅空才 fatal；env 填了默认值即可启动，不再因占位/过短拦死
		if secret == "" {
			log.Fatal("[Security] 生产环境拒绝启动：JWT_SECRET 为空，请在根目录 .env 中配置")
		}
		return
	}

	// 开发：弱密钥仅 warning
	if secret == "" || secret == defaultJWTSecret || isWeakJWTSecret(secret) {
		log.Println("[Security Warning] JWT_SECRET 仍是弱/占位值，仅开发可用；生产建议换成强随机密钥")
	}
}

func IsProductionMode() bool {
	cfg := GetGlobalConfig()
	if cfg == nil {
		return false
	}
	return isProductionEnvMode(cfg.Environment) || isProductionEnvMode(cfg.AppMode)
}

// IsAdminDebugOpsEnabled 管理端 debug/pprof/重启类高危能力是否可用。
// 规则：生产永远 false；非生产看 EnableAdminDebugOps（默认 true）。
func IsAdminDebugOpsEnabled() bool {
	if IsProductionMode() {
		return false
	}
	cfg := GetGlobalConfig()
	if cfg == nil {
		// 无配置时按开发态默认允许，避免本地启动被静默关掉
		return true
	}
	return cfg.EnableAdminDebugOps
}

// GetGlobalConfig 返回当前全局配置指针（只读场景请优先 CloneGlobalConfig）。
// 指针本身在锁内读取；调用方不得直接写字段，写入请用 UpdateGlobalConfig。
func GetGlobalConfig() *Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	return GlobalConfig
}

// SetGlobalConfig 整体替换全局配置（启动加载 / 测试注入）。
func SetGlobalConfig(cfg *Config) {
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	GlobalConfig = cfg
}

// UpdateGlobalConfig 在写锁内原地更新配置字段，避免并发写竞态。
func UpdateGlobalConfig(updater func(cfg *Config)) {
	if updater == nil {
		return
	}
	globalConfigMu.Lock()
	defer globalConfigMu.Unlock()
	if GlobalConfig == nil {
		GlobalConfig = &Config{}
	}
	updater(GlobalConfig)
}

// CloneGlobalConfig 返回全局配置的值拷贝（快照），适合跨请求只读使用。
func CloneGlobalConfig() *Config {
	globalConfigMu.RLock()
	defer globalConfigMu.RUnlock()
	if GlobalConfig == nil {
		return nil
	}
	clone := *GlobalConfig
	return &clone
}

// isFrontendDotEnv 判断 .env 是否为纯前端配置（仅含 VITE_ 变量）
func isFrontendDotEnv(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	lines := strings.Split(string(data), "\n")
	nonEmptyLines := 0
	viteOnly := true
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		nonEmptyLines++
		if !strings.HasPrefix(line, "VITE_") {
			viteOnly = false
			break
		}
	}
	return nonEmptyLines > 0 && viteOnly
}

func findDotEnvPath() (string, bool) {
	wd, err := os.Getwd()
	if err != nil {
		return "", false
	}

	dir := wd
	for {
		candidate := filepath.Join(dir, ".env")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			// 跳过纯前端 .env（仅含 VITE_ 变量），继续向上查找后端 .env
			if !isFrontendDotEnv(candidate) {
				return candidate, true
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", false
}

func InitConfig() {
	dotEnvPath, hasDotEnv := findDotEnvPath()
	if hasDotEnv {
		if cfg, ok := loadJSONDotEnv(dotEnvPath); ok {
			SetGlobalConfig(cfg)
			log.Printf("[Config] Loaded JSON .env from %s", dotEnvPath)
			validateCriticalSecurityConfig(cfg)
			return
		}

		if err := godotenv.Load(dotEnvPath); err != nil {
			log.Printf("Error loading .env file %s, using default environment variables", dotEnvPath)
		} else {
			log.Printf("[Config] Loaded .env from %s", dotEnvPath)
		}
	} else {
		log.Println("Error loading .env file, using default environment variables")
	}

	geetestEnabled, _ := strconv.ParseBool(strings.TrimSpace(getEnv("GEETEST_ENABLED", "false")))
	if !geetestEnabled {
		geetestEnabled, _ = strconv.ParseBool(strings.TrimSpace(getEnv("GEETEST_ENABLE", "false")))
	}

	enableSwagger, _ := strconv.ParseBool(strings.TrimSpace(getEnv("ENABLE_SWAGGER", "false")))
	// 非生产默认开启 debug 运维接口；显式 false/0 可关闭
	enableAdminDebugOps := true
	if v := strings.TrimSpace(getEnv("ENABLE_ADMIN_DEBUG", "true")); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			enableAdminDebugOps = parsed
		} else if v == "0" {
			enableAdminDebugOps = false
		}
	}

	geetestID := getEnv("GEETEST_ID", "")
	if geetestID == "" {
		geetestID = getEnv("GEETEST_CAPTCHA_ID", "")
	}
	geetestKey := getEnv("GEETEST_KEY", "")
	if geetestKey == "" {
		geetestKey = getEnv("GEETEST_CAPTCHA_KEY", "")
	}
	runtimeEnv := resolveRuntimeEnv(getEnv("GO_ENV", ""), getEnv("APP_ENV", ""), getEnv("GIN_MODE", ""))
	jwtSecret := resolveJWTSecret(getEnv("JWT_SECRET", ""))
	adminJWTSecret := strings.TrimSpace(getEnv("JWT_ADMIN_SECRET", ""))
	if adminJWTSecret == "" {
		adminJWTSecret = jwtSecret
	}

	cfg := &Config{
		AppName:         getEnv("APP_NAME", "F.st"),
		AppTitle:        getEnv("APP_TITLE", "F.st - Think Fast,Run F.st"),
		AppMode:         getEnv("APP_MODE", "separate"),
		Environment:     runtimeEnv,
		Port:            getEnv("PORT", "8080"),
		HTTPReadHeaderTimeoutSeconds: getEnvAsPositiveInt("HTTP_READ_HEADER_TIMEOUT_SECONDS", 5),
		HTTPReadTimeoutSeconds:       getEnvAsPositiveInt("HTTP_READ_TIMEOUT_SECONDS", 5),
		HTTPWriteTimeoutSeconds:      getEnvAsPositiveInt("HTTP_WRITE_TIMEOUT_SECONDS", 60),
		HTTPIdleTimeoutSeconds:       getEnvAsPositiveInt("HTTP_IDLE_TIMEOUT_SECONDS", 120),
		HTTPShutdownTimeoutSeconds:   getEnvAsPositiveInt("HTTP_SHUTDOWN_TIMEOUT_SECONDS", 10),
		HTTPMaxHeaderBytes:           getEnvAsPositiveInt("HTTP_MAX_HEADER_BYTES", 1<<20),
		DBDriver:        getEnv("DB_DRIVER", "mysql"),
		DBDSN:           buildDSN(),
		GeetestEnabled:  geetestEnabled && geetestID != "" && geetestKey != "",
		GeetestID:       geetestID,
		GeetestKey:      geetestKey,
		JWTSecret:       jwtSecret,
		AdminJWTSecret:  adminJWTSecret,
		AdminPath:       getEnv("ADMIN_PATH", "/system-mgr"),
		AdminAPIPath:    NormalizeAdminAPIPath(getEnv("ADMIN_API_PATH", "/admin")),
		CorsOrigins:           getEnv("CORS_ORIGINS", ""),
		EnableSwagger:         enableSwagger,
		EnableAdminDebugOps:   enableAdminDebugOps,
		FrontendURL:           getEnv("FRONTEND_URL", ""),
		BackendAPIURL:         getEnv("BACKEND_API_URL", ""),
		SMTPHost:              getEnv("SMTP_HOST", ""),
		SMTPPort:              getEnv("SMTP_PORT", ""),
		SMTPUser:              getEnv("SMTP_USERNAME", ""),
		SMTPPass:              getEnv("SMTP_PASSWORD", ""),
		SMTPSSL:               getEnv("SMTP_SSL_TYPE", "") == "ssl",
		SystemEmail:           getEnv("SYSTEM_EMAIL_ADDRESS", ""),
		SystemEmailName:       getEnv("SYSTEM_EMAIL_NAME", ""),
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
		EmailVerifyEnabled: func() bool {
			v, _ := strconv.ParseBool(strings.TrimSpace(getEnv("EMAIL_VERIFY_ENABLED", "true")))
			return v
		}(),
		SMSVerifyEnabled: func() bool {
			v, _ := strconv.ParseBool(strings.TrimSpace(getEnv("SMS_VERIFY_ENABLED", "false")))
			return v
		}(),
		SMSProvider:     getEnv("SMS_PROVIDER", "console"),
		SMSAccessKey:    getEnv("SMS_ACCESS_KEY", ""),
		SMSSecretKey:    getEnv("SMS_SECRET_KEY", ""),
		SMSSignName:       getEnv("SMS_SIGN_NAME", ""),
		SMSTemplateCode:   getEnv("SMS_TEMPLATE_CODE", ""),
		SMSTemplateCodeEN: getEnv("SMS_TEMPLATE_CODE_EN", ""),
		SMSRegion:         getEnv("SMS_REGION", ""),
		SMSEndpoint:     getEnv("SMS_ENDPOINT", ""),
		SMSSdkAppID:     getEnv("SMS_SDK_APP_ID", ""),
		SMSBodyFormat:   getEnv("SMS_BODY_FORMAT", "json"),
	}

	SetGlobalConfig(cfg)
	validateCriticalSecurityConfig(cfg)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsPositiveInt(key string, fallback int) int {
	return parsePositiveIntValue(getEnv(key, strconv.Itoa(fallback)), fallback)
}

func parsePositiveIntValue(raw string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v <= 0 {
		return fallback
	}
	return v
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
	HTTPReadHeaderTimeoutSeconds string `json:"http_read_header_timeout_seconds"`
	HTTPReadTimeoutSeconds       string `json:"http_read_timeout_seconds"`
	HTTPWriteTimeoutSeconds      string `json:"http_write_timeout_seconds"`
	HTTPIdleTimeoutSeconds       string `json:"http_idle_timeout_seconds"`
	HTTPShutdownTimeoutSeconds   string `json:"http_shutdown_timeout_seconds"`
	HTTPMaxHeaderBytes           string `json:"http_max_header_bytes"`
	CorsOrigins               string `json:"cors_origins"`
	JWTSecret                 string `json:"jwt_secret"`
	JWTAdminSecret            string `json:"jwt_admin_secret"`
	AdminPath                 string `json:"admin_path"`
	AdminAPIPath              string `json:"admin_api_path"`
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
	BackendAPIURL             string `json:"backend_api_url"`
	LoginMaxFailureCount      string `json:"login_max_failure_count"`
	LoginLockDurationMinutes  string `json:"login_lock_duration_minutes"`
	JWTAccessExpire           string `json:"jwt_access_expire"`
	JWTRefreshExpire          string `json:"jwt_refresh_expire"`
	CleanupIntervalMinutes    string `json:"cleanup_interval_minutes"`
	EmailVerifyEnabled        string `json:"email_verify_enabled"`
	SMSVerifyEnabled          string `json:"sms_verify_enabled"`
	SMSProvider               string `json:"sms_provider"`
	SMSAccessKey              string `json:"sms_access_key"`
	SMSSecretKey              string `json:"sms_secret_key"`
	SMSSignName               string `json:"sms_sign_name"`
	SMSTemplateCode           string `json:"sms_template_code"`
	SMSTemplateCodeEN         string `json:"sms_template_code_en"`
	SMSRegion                 string `json:"sms_region"`
	SMSEndpoint               string `json:"sms_endpoint"`
	SMSSdkAppID               string `json:"sms_sdk_app_id"`
	SMSBodyFormat             string `json:"sms_body_format"`
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

	runtimeEnv := resolveRuntimeEnv(func() string {
		debug := strings.TrimSpace(raw.Debug)
		if strings.EqualFold(debug, "false") || debug == "0" {
			return "production"
		}
		return "development"
	}())

	jwtSecret := resolveJWTSecret(raw.JWTSecret)

	jwtAdminSecret := raw.JWTAdminSecret
	if jwtAdminSecret == "" {
		jwtAdminSecret = jwtSecret
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
		Environment:     runtimeEnv,
		Port:            port,
		HTTPReadHeaderTimeoutSeconds: parsePositiveIntValue(raw.HTTPReadHeaderTimeoutSeconds, 5),
		HTTPReadTimeoutSeconds:       parsePositiveIntValue(raw.HTTPReadTimeoutSeconds, 5),
		HTTPWriteTimeoutSeconds:      parsePositiveIntValue(raw.HTTPWriteTimeoutSeconds, 60),
		HTTPIdleTimeoutSeconds:       parsePositiveIntValue(raw.HTTPIdleTimeoutSeconds, 120),
		HTTPShutdownTimeoutSeconds:   parsePositiveIntValue(raw.HTTPShutdownTimeoutSeconds, 10),
		HTTPMaxHeaderBytes:           parsePositiveIntValue(raw.HTTPMaxHeaderBytes, 1<<20),
		DBDriver:        "mysql",
		DBDSN:           dsn,
		GeetestEnabled:  geetestEnabled && raw.GeetestCaptchaID != "" && raw.GeetestCaptchaKey != "",
		GeetestID:       raw.GeetestCaptchaID,
		GeetestKey:      raw.GeetestCaptchaKey,
		JWTSecret:       jwtSecret,
		AdminJWTSecret:  jwtAdminSecret,
		AdminPath: func() string {
			if strings.TrimSpace(raw.AdminPath) == "" {
				return "/system-mgr"
			}
			return strings.TrimSpace(raw.AdminPath)
		}(),
		AdminAPIPath: NormalizeAdminAPIPath(raw.AdminAPIPath),
		CorsOrigins:         raw.CorsOrigins,
		EnableSwagger:       enableSwagger,
		// JSON .env 无独立开关时默认开启；生产仍会被 IsAdminDebugOpsEnabled 拦截
		EnableAdminDebugOps: true,
		FrontendURL:         raw.FrontendURL,
		BackendAPIURL:       raw.BackendAPIURL,
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
		EmailVerifyEnabled: func() bool {
			if raw.EmailVerifyEnabled == "" {
				return true
			}
			v, _ := strconv.ParseBool(strings.TrimSpace(raw.EmailVerifyEnabled))
			return v
		}(),
		SMSVerifyEnabled: func() bool {
			if raw.SMSVerifyEnabled == "" {
				return false
			}
			v, _ := strconv.ParseBool(strings.TrimSpace(raw.SMSVerifyEnabled))
			return v
		}(),
		SMSProvider:     raw.SMSProvider,
		SMSAccessKey:    raw.SMSAccessKey,
		SMSSecretKey:    raw.SMSSecretKey,
		SMSSignName:       raw.SMSSignName,
		SMSTemplateCode:   raw.SMSTemplateCode,
		SMSTemplateCodeEN: raw.SMSTemplateCodeEN,
		SMSRegion:         raw.SMSRegion,
		SMSEndpoint:       raw.SMSEndpoint,
		SMSSdkAppID:       raw.SMSSdkAppID,
		SMSBodyFormat:     raw.SMSBodyFormat,
	}

	log.Printf("[Config] RegisterCodeExpireMinutes: %d\n", cfg.RegisterCodeExpireMinutes)
	log.Printf("[Config] LoginMaxFailureCount: %d\n", cfg.LoginMaxFailureCount)
	log.Printf("[Config] LoginLockDurationMinutes: %d\n", cfg.LoginLockDurationMinutes)
	return cfg, true
}

