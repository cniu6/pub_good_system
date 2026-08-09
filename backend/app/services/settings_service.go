package services

import (
	"encoding/json"
	"errors"
	"fst/backend/app/models"
	"fst/backend/pkg/config"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// SettingsService caches system settings from DB.
type SettingsService struct {
	cache     map[string]*models.SystemSetting
	cacheMu   sync.RWMutex
	cacheTime time.Time
	ttl       time.Duration
}

// GlobalSettingsService is the singleton settings service instance.
var GlobalSettingsService *SettingsService

// InitSettingsService initializes the global settings cache service.
func InitSettingsService() {
	GlobalSettingsService = NewSettingsService(5 * time.Minute)
	ReloadGlobalRuntimeConfig()
	// 注入在线心跳容忍窗口的动态实现，避免 models 包直接依赖 services（防止循环引用）。
	models.GetOnlineHeartbeatGraceSeconds = func() int64 {
		return GetGlobalOnlinePresenceRuntimeConfig().GraceSeconds
	}
	log.Println("[SettingsService] Initialized with cache TTL: 5m")
}

// NewSettingsService creates a settings service with the given cache TTL.
func NewSettingsService(ttl time.Duration) *SettingsService {
	return &SettingsService{
		cache: make(map[string]*models.SystemSetting),
		ttl:   ttl,
	}
}

// ensureFreshCache 在读取缓存前确保缓存尽量保持最新。
func (s *SettingsService) ensureFreshCache() {
	if s == nil {
		return
	}

	s.cacheMu.RLock()
	expired := s.cacheTime.IsZero() || time.Since(s.cacheTime) > s.ttl
	s.cacheMu.RUnlock()
	if !expired {
		return
	}

	if err := s.RefreshCache(); err != nil {
		log.Printf("[SettingsService] Refresh cache failed: %v", err)
	}
}

// ReloadGlobalRuntimeConfig 强制刷新全局配置缓存并重新应用运行时配置。
func ReloadGlobalRuntimeConfig() {
	if GlobalSettingsService != nil {
		if err := GlobalSettingsService.RefreshCache(); err != nil {
			log.Printf("[SettingsService] Refresh cache failed: %v", err)
		}
	}

	ApplyGlobalRuntimeConfig()
}

// RefreshCache refreshes all settings from DB.
func (s *SettingsService) RefreshCache() error {
	settings, err := models.GetAllSettings()
	if err != nil {
		return err
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.cache = make(map[string]*models.SystemSetting)
	for i := range settings {
		s.cache[settings[i].Key] = &settings[i]
	}
	s.cacheTime = time.Now()

	return nil
}

// GetSystemSetting 读取系统设置值（不存在返回空字符串和 false）
func GetSystemSetting(key string) (string, bool) {
	if GlobalSettingsService == nil {
		return "", false
	}
	return GlobalSettingsService.Get(key)
}

// SetSystemSetting 设置系统配置（不存在则创建，存在则更新），并刷新缓存
func SetSystemSetting(key, value string) error {
	existing, err := models.GetSettingByKey(key)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing == nil {
		setting := &models.SystemSetting{
			Key:   key,
			Value: value,
			Type:  "text",
		}
		if err := models.CreateSetting(setting); err != nil {
			return err
		}
	} else {
		existing.Value = value
		if err := models.UpdateSettingWithMeta(existing); err != nil {
			return err
		}
	}
	if GlobalSettingsService != nil {
		GlobalSettingsService.InvalidateCache()
	}
	return nil
}

// Get returns setting value by key.
func (s *SettingsService) Get(key string) (string, bool) {
	if s == nil {
		return "", false
	}
	s.ensureFreshCache()

	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	setting, ok := s.cache[key]
	if !ok {
		return "", false
	}
	return setting.Value, true
}

// GetWithDefault returns setting value or fallback if key does not exist.
func (s *SettingsService) GetWithDefault(key, defaultValue string) string {
	val, ok := s.Get(key)
	if !ok {
		return defaultValue
	}
	return val
}

// GetBool returns bool setting value. Missing keys return false.
func (s *SettingsService) GetBool(key string) bool {
	val, ok := s.Get(key)
	if !ok {
		return false
	}
	return val == "true" || val == "1"
}

// GetBoolWithDefault returns bool setting value or fallback.
func (s *SettingsService) GetBoolWithDefault(key string, defaultValue bool) bool {
	val, ok := s.Get(key)
	if !ok {
		return defaultValue
	}
	return val == "true" || val == "1"
}

// GetIntWithDefault returns int setting value or fallback.
func (s *SettingsService) GetIntWithDefault(key string, defaultValue int) int {
	val, ok := s.Get(key)
	if !ok {
		return defaultValue
	}
	var result int
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return defaultValue
	}
	return result
}

// getEffectiveGlobalConfig 返回全局配置快照（值拷贝），避免与运行时热更新并发读写。
func getEffectiveGlobalConfig() *config.Config {
	if cfg := config.CloneGlobalConfig(); cfg != nil {
		return cfg
	}
	return &config.Config{}
}

// GetSetting returns full setting model by key from cache.
func (s *SettingsService) GetSetting(key string) *models.SystemSetting {
	if s == nil {
		return nil
	}
	s.ensureFreshCache()

	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	return s.cache[key]
}

// GetAllFromCache returns a shallow copy of cache map.
func (s *SettingsService) GetAllFromCache() map[string]*models.SystemSetting {
	if s == nil {
		return map[string]*models.SystemSetting{}
	}
	s.ensureFreshCache()

	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	result := make(map[string]*models.SystemSetting)
	for k, v := range s.cache {
		result[k] = v
	}
	return result
}

// IsCacheExpired returns whether current cache is expired.
func (s *SettingsService) IsCacheExpired() bool {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	return time.Since(s.cacheTime) > s.ttl
}

// InvalidateCache marks cache as expired immediately.
func (s *SettingsService) InvalidateCache() {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.cacheTime = time.Time{}
}

// PublicAppConfig is the public-facing app config payload.
type PublicAppConfig struct {
	SiteName            string `json:"site_name"`
	SiteDesc            string `json:"site_desc"`
	SiteLogo            string `json:"site_logo"`
	Copyright           string `json:"copyright"`
	ICP                 string `json:"icp"`
	AllowRegister       bool   `json:"allow_register"`
	AnnouncementEnabled bool   `json:"announcement_enabled"`
	AllowDeleteAccount  bool   `json:"allow_delete_account"`
	DefaultLang         string `json:"default_lang"`
	Version             string `json:"version"`
	GeetestEnabled      bool   `json:"geetest_enabled"`
	GeetestCaptchaId    string `json:"geetest_captcha_id"`
	EmailVerifyEnabled  bool   `json:"email_verify_enabled"`
	SMSVerifyEnabled    bool   `json:"sms_verify_enabled"`
	// MobileCNOnly 为 true 时仅允许中国大陆手机号（+86）；false 时允许国际 E.164
	MobileCNOnly bool `json:"mobile_cn_only"`
	// MobileIPCountryDetect 国际号模式下是否按 IP/CDN 头预选国家区号
	MobileIPCountryDetect bool     `json:"mobile_ip_country_detect"`
	RealnameEnabled       bool     `json:"realname_enabled"`
	RealnameNotifyText    string   `json:"realname_notify_text"`
	WithdrawEnabled       bool     `json:"withdraw_enabled"`
	WithdrawMinAmount     float64  `json:"withdraw_min_amount"`
	WithdrawNotifyText    string   `json:"withdraw_notify_text"`
	WithdrawAccountTypes  []string `json:"withdraw_account_types"`
	// WithdrawRequireRealname 提现前是否必须已完成实名认证并通过审核，默认 false
	WithdrawRequireRealname bool `json:"withdraw_require_realname"`
	// AdminAPIPath 管理端 REST API 在 /api/v1 下的前缀（来自 env ADMIN_API_PATH，默认 /admin）
	AdminAPIPath string `json:"admin_api_path"`
	// PresenceEnabled 是否启用 Presence WebSocket / 在线心跳（默认 false）
	PresenceEnabled bool `json:"presence_enabled"`
	// OnlineReportIntervalSeconds 在线心跳上报周期（秒），前端 Presence 心跳按此间隔发送
	OnlineReportIntervalSeconds int `json:"online_report_interval_seconds"`
	// AllowUserLogin 是否允许普通用户密码登录（默认 true）；关闭后不发 JWT，管理员与 API Key 不受影响
	AllowUserLogin bool `json:"allow_user_login"`
	// UserAPILogVisible 用户中心是否展示「API 日志」（仅 API Key 调用）
	UserAPILogVisible bool `json:"user_api_log_visible"`
	// UserOperationLogVisible 用户中心是否展示「操作日志」
	UserOperationLogVisible bool `json:"user_operation_log_visible"`
}

// VerifyConfig 验证码功能开关运行时配置
type VerifyConfig struct {
	EmailEnabled bool
	SMSEnabled   bool
}

// SMSRuntimeConfig 短信服务运行时配置
type SMSRuntimeConfig struct {
	Provider       string
	AccessKey      string
	SecretKey      string
	SignName       string
	TemplateCode   string
	TemplateCodeEN string
	Region         string
	Endpoint       string
	SdkAppID       string
	BodyFormat     string
}

// APILogRuntimeConfig API访问日志运行时配置
type APILogRuntimeConfig struct {
	Enabled                bool
	QueryDays              int
	MaxCount               int
	CleanupIntervalSeconds int
	PerUserLimitEnabled    bool
	PerUserMaxCount        int
	UserVisible            bool // 用户中心是否可见本人 API Key 日志
}

const (
	defaultAPILogCleanupIntervalSeconds = 600
	minAPILogCleanupIntervalSeconds     = 60
	maxAPILogCleanupIntervalSeconds     = 86400
)

// normalizeAPILogCleanupIntervalSeconds 限制自动清理间隔，避免错误配置导致高频删除。
func normalizeAPILogCleanupIntervalSeconds(value int) int {
	if value < minAPILogCleanupIntervalSeconds || value > maxAPILogCleanupIntervalSeconds {
		return defaultAPILogCleanupIntervalSeconds
	}
	return value
}

// OperationLogRuntimeConfig 操作日志运行时配置
type OperationLogRuntimeConfig struct {
	QueryDays           int
	MaxCount            int
	PerUserLimitEnabled bool
	PerUserMaxCount     int
	UserVisible         bool // 用户中心是否可见本人操作日志
	UserShowBody        bool // 用户详情是否展示请求/响应体
}

// SMSLogRuntimeConfig 短信日志运行时配置
type SMSLogRuntimeConfig struct {
	MaxCount            int
	PerUserLimitEnabled bool
	PerUserMaxCount     int
}

// EmailLogRuntimeConfig 邮件日志运行时配置
type EmailLogRuntimeConfig struct {
	MaxCount            int
	PerUserLimitEnabled bool
	PerUserMaxCount     int
}

// RateLimitRuntimeConfig 全局/管理员限速运行时配置
type RateLimitRuntimeConfig struct {
	Enabled bool
	Rate    int
	Burst   int
}

// GeetestRuntimeConfig is the effective config used by backend validation.
type GeetestRuntimeConfig struct {
	Enabled    bool
	CaptchaID  string
	CaptchaKey string
}

// OnlinePresenceRuntimeConfig 在线状态运行时配置：管理端可配置「上报周期」，
// 容忍窗口（GraceSeconds）按上报周期的 3 倍换算，容许偶尔丢失 1~2 次心跳而不误判离线。
type OnlinePresenceRuntimeConfig struct {
	ReportIntervalSeconds int   // 客户端心跳上报周期（秒），默认 30
	GraceSeconds          int64 // 判定离线的容忍窗口（秒）
}

// RealnameAPIRuntimeConfig 实名认证API运行时配置
type RealnameAPIRuntimeConfig struct {
	Enabled   bool   // 是否启用自动实名认证验证
	Provider  string // 提供商: aliyun, tencent, baidu, custom
	AppKey    string // App Key / Access Key
	AppSecret string // App Secret / Secret Key
	Endpoint  string // 可选，自定义API地址
}

func parseBoolSettingWithFallback(val string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true":
		return true
	case "0", "false":
		return false
	case "":
		return fallback
	default:
		return fallback
	}
}

func parsePositiveIntSettingWithFallback(val string, fallback int) int {
	val = strings.TrimSpace(val)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func getDirectSettingString(key, fallback string) string {
	setting, err := models.GetSettingByKey(key)
	if err != nil || setting == nil {
		return strings.TrimSpace(fallback)
	}
	val := strings.TrimSpace(setting.Value)
	if val != "" {
		return val
	}
	return strings.TrimSpace(fallback)
}

func getDirectSettingBool(key string, fallback bool) bool {
	setting, err := models.GetSettingByKey(key)
	if err != nil || setting == nil {
		return fallback
	}
	return parseBoolSettingWithFallback(setting.Value, fallback)
}

func getDirectSettingPositiveInt(key string, fallback int) int {
	setting, err := models.GetSettingByKey(key)
	if err != nil || setting == nil {
		return fallback
	}
	return parsePositiveIntSettingWithFallback(setting.Value, fallback)
}

func (s *SettingsService) getRuntimeString(key, fallback string) string {
	if s != nil {
		if val, ok := s.Get(key); ok {
			val = strings.TrimSpace(val)
			if val != "" {
				return val
			}
		}
	}
	return strings.TrimSpace(fallback)
}

func (s *SettingsService) getRuntimeBool(key string, fallback bool) bool {
	if s != nil {
		if val, ok := s.Get(key); ok {
			return parseBoolSettingWithFallback(val, fallback)
		}
	}
	return fallback
}

func (s *SettingsService) getRuntimePositiveInt(key string, fallback int) int {
	if s != nil {
		if val, ok := s.Get(key); ok {
			return parsePositiveIntSettingWithFallback(val, fallback)
		}
	}
	return fallback
}

// GetGeetestRuntimeConfig returns effective geetest config.
// Priority: database values -> environment fallback.
func (s *SettingsService) GetGeetestRuntimeConfig() GeetestRuntimeConfig {
	cfg := getEffectiveGlobalConfig()
	enabled := s.getRuntimeBool("geetest_enabled", cfg.GeetestEnabled)
	captchaID := s.getRuntimeString("geetest_captcha_id", cfg.GeetestID)
	captchaKey := s.getRuntimeString("geetest_captcha_key", cfg.GeetestKey)

	enabled = enabled && captchaID != "" && captchaKey != ""

	return GeetestRuntimeConfig{
		Enabled:    enabled,
		CaptchaID:  captchaID,
		CaptchaKey: captchaKey,
	}
}

// GetGlobalGeetestRuntimeConfig returns effective geetest config with global cache when available.
func GetGlobalGeetestRuntimeConfig() GeetestRuntimeConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetGeetestRuntimeConfig()
	}

	cfg := getEffectiveGlobalConfig()
	captchaID := strings.TrimSpace(cfg.GeetestID)
	captchaKey := strings.TrimSpace(cfg.GeetestKey)

	return GeetestRuntimeConfig{
		Enabled:    cfg.GeetestEnabled && captchaID != "" && captchaKey != "",
		CaptchaID:  captchaID,
		CaptchaKey: captchaKey,
	}
}

// GetRealnameAPIRuntimeConfig returns effective realname API config.
func (s *SettingsService) GetRealnameAPIRuntimeConfig() RealnameAPIRuntimeConfig {
	enabled := s.getRuntimeBool("realname_api_enabled", false)
	appKey := s.getRuntimeString("realname_api_app_key", "")
	appSecret := s.getRuntimeString("realname_api_app_secret", "")

	// 只有当启用且有密钥时才认为真正启用
	enabled = enabled && appKey != "" && appSecret != ""

	return RealnameAPIRuntimeConfig{
		Enabled:   enabled,
		Provider:  s.getRuntimeString("realname_api_provider", "aliyun"),
		AppKey:    appKey,
		AppSecret: appSecret,
		Endpoint:  s.getRuntimeString("realname_api_endpoint", ""),
	}
}

// GetGlobalRealnameAPIRuntimeConfig returns effective realname API config with global cache.
func GetGlobalRealnameAPIRuntimeConfig() RealnameAPIRuntimeConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetRealnameAPIRuntimeConfig()
	}
	return RealnameAPIRuntimeConfig{
		Enabled:  false,
		Provider: "aliyun",
	}
}

// GetVerifyConfig returns effective verify enable/disable config.
// Priority: database values -> environment fallback.
func (s *SettingsService) GetVerifyConfig() VerifyConfig {
	cfg := getEffectiveGlobalConfig()
	return VerifyConfig{
		EmailEnabled: s.getRuntimeBool("email_verify_enabled", cfg.EmailVerifyEnabled),
		SMSEnabled:   s.getRuntimeBool("sms_verify_enabled", cfg.SMSVerifyEnabled),
	}
}

// GetSMSRuntimeConfig returns effective SMS provider config.
func (s *SettingsService) GetSMSRuntimeConfig() SMSRuntimeConfig {
	cfg := getEffectiveGlobalConfig()
	return SMSRuntimeConfig{
		Provider:       s.getRuntimeString("sms_provider", cfg.SMSProvider),
		AccessKey:      s.getRuntimeString("sms_access_key", cfg.SMSAccessKey),
		SecretKey:      s.getRuntimeString("sms_secret_key", cfg.SMSSecretKey),
		SignName:       s.getRuntimeString("sms_sign_name", cfg.SMSSignName),
		TemplateCode:   s.getRuntimeString("sms_template_code", cfg.SMSTemplateCode),
		TemplateCodeEN: s.getRuntimeString("sms_template_code_en", cfg.SMSTemplateCodeEN),
		Region:         s.getRuntimeString("sms_region", cfg.SMSRegion),
		Endpoint:       s.getRuntimeString("sms_endpoint", cfg.SMSEndpoint),
		SdkAppID:       s.getRuntimeString("sms_sdk_app_id", cfg.SMSSdkAppID),
		BodyFormat:     s.getRuntimeString("sms_body_format", cfg.SMSBodyFormat),
	}
}

// GetGlobalVerifyConfig returns effective verify config with global cache.
func GetGlobalVerifyConfig() VerifyConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetVerifyConfig()
	}
	cfg := getEffectiveGlobalConfig()
	return VerifyConfig{
		EmailEnabled: cfg.EmailVerifyEnabled,
		SMSEnabled:   cfg.SMSVerifyEnabled,
	}
}

// GetGlobalMobileCNOnly 是否仅允许中国大陆手机号（默认 true）。
// GetOnlinePresenceRuntimeConfig 计算在线心跳的上报周期与离线容忍窗口。
func (s *SettingsService) GetOnlinePresenceRuntimeConfig() OnlinePresenceRuntimeConfig {
	interval := s.getRuntimePositiveInt("online_report_interval_seconds", 30)
	if interval < 10 {
		interval = 10
	}
	if interval > 300 {
		interval = 300
	}
	grace := int64(interval) * 3
	if grace < 60 {
		grace = 60
	}
	return OnlinePresenceRuntimeConfig{ReportIntervalSeconds: interval, GraceSeconds: grace}
}

// GetGlobalOnlinePresenceRuntimeConfig 全局访问入口，服务未初始化时回退默认值（30s 上报 / 90s 容忍）。
func GetGlobalOnlinePresenceRuntimeConfig() OnlinePresenceRuntimeConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetOnlinePresenceRuntimeConfig()
	}
	return OnlinePresenceRuntimeConfig{ReportIntervalSeconds: 30, GraceSeconds: 90}
}

// GetGlobalPresenceEnabled 是否启用 Presence（默认关闭；服务未就绪时也按关闭处理）。
func GetGlobalPresenceEnabled() bool {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetBoolWithDefault("presence_enabled", false)
	}
	return false
}

func GetGlobalMobileCNOnly() bool {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetBoolWithDefault("mobile_cn_only", true)
	}
	return true
}

// GetGlobalMobileIPCountryDetect 国际号模式下是否按 IP 自动匹配国家（默认 false）。
func GetGlobalMobileIPCountryDetect() bool {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetBoolWithDefault("mobile_ip_country_detect", false)
	}
	return false
}

// GetGlobalSMSRuntimeConfig returns effective SMS config with global cache.
func GetGlobalSMSRuntimeConfig() SMSRuntimeConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetSMSRuntimeConfig()
	}
	cfg := getEffectiveGlobalConfig()
	return SMSRuntimeConfig{
		Provider:       cfg.SMSProvider,
		AccessKey:      cfg.SMSAccessKey,
		SecretKey:      cfg.SMSSecretKey,
		SignName:       cfg.SMSSignName,
		TemplateCode:   cfg.SMSTemplateCode,
		TemplateCodeEN: cfg.SMSTemplateCodeEN,
		Region:         cfg.SMSRegion,
		Endpoint:       cfg.SMSEndpoint,
		SdkAppID:       cfg.SMSSdkAppID,
		BodyFormat:     cfg.SMSBodyFormat,
	}
}

func (s *SettingsService) GetAPILogRuntimeConfig() APILogRuntimeConfig {
	return APILogRuntimeConfig{
		Enabled:                s.getRuntimeBool("api_access_log_enabled", true),
		QueryDays:              s.getRuntimePositiveInt("api_log_query_days", 7),
		MaxCount:               s.getRuntimePositiveInt("api_log_max_count", 5000),
		CleanupIntervalSeconds: normalizeAPILogCleanupIntervalSeconds(s.getRuntimePositiveInt("api_log_cleanup_interval_seconds", defaultAPILogCleanupIntervalSeconds)),
		PerUserLimitEnabled:    s.getRuntimeBool("api_log_per_user_limit_enabled", false),
		PerUserMaxCount:        s.getRuntimePositiveInt("api_log_per_user_max_count", 1000),
		UserVisible:            s.getRuntimeBool("user_api_log_visible", true),
	}
}

func GetGlobalAPILogRuntimeConfig() APILogRuntimeConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetAPILogRuntimeConfig()
	}
	return APILogRuntimeConfig{
		Enabled:                getDirectSettingBool("api_access_log_enabled", true),
		QueryDays:              getDirectSettingPositiveInt("api_log_query_days", 7),
		MaxCount:               getDirectSettingPositiveInt("api_log_max_count", 5000),
		CleanupIntervalSeconds: normalizeAPILogCleanupIntervalSeconds(getDirectSettingPositiveInt("api_log_cleanup_interval_seconds", defaultAPILogCleanupIntervalSeconds)),
		PerUserLimitEnabled:    getDirectSettingBool("api_log_per_user_limit_enabled", false),
		PerUserMaxCount:        getDirectSettingPositiveInt("api_log_per_user_max_count", 1000),
		UserVisible:            getDirectSettingBool("user_api_log_visible", true),
	}
}

func (s *SettingsService) GetOperationLogRuntimeConfig() OperationLogRuntimeConfig {
	return OperationLogRuntimeConfig{
		QueryDays:           s.getRuntimePositiveInt("operation_log_query_days", 30),
		MaxCount:            s.getRuntimePositiveInt("operation_log_max_count", 1000),
		PerUserLimitEnabled: s.getRuntimeBool("operation_log_per_user_limit_enabled", false),
		PerUserMaxCount:     s.getRuntimePositiveInt("operation_log_per_user_max_count", 1000),
		UserVisible:         s.getRuntimeBool("user_operation_log_visible", true),
		UserShowBody:        s.getRuntimeBool("user_operation_log_show_body", false),
	}
}

func GetGlobalOperationLogRuntimeConfig() OperationLogRuntimeConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetOperationLogRuntimeConfig()
	}
	return OperationLogRuntimeConfig{
		QueryDays:           getDirectSettingPositiveInt("operation_log_query_days", 30),
		MaxCount:            getDirectSettingPositiveInt("operation_log_max_count", 1000),
		PerUserLimitEnabled: getDirectSettingBool("operation_log_per_user_limit_enabled", false),
		PerUserMaxCount:     getDirectSettingPositiveInt("operation_log_per_user_max_count", 1000),
		UserVisible:         getDirectSettingBool("user_operation_log_visible", true),
		UserShowBody:        getDirectSettingBool("user_operation_log_show_body", false),
	}
}

func (s *SettingsService) GetSMSLogRuntimeConfig() SMSLogRuntimeConfig {
	return SMSLogRuntimeConfig{
		MaxCount:            s.getRuntimePositiveInt("sms_log_max_count", 1000),
		PerUserLimitEnabled: s.getRuntimeBool("sms_log_per_user_limit_enabled", false),
		PerUserMaxCount:     s.getRuntimePositiveInt("sms_log_per_user_max_count", 1000),
	}
}

func GetGlobalSMSLogRuntimeConfig() SMSLogRuntimeConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetSMSLogRuntimeConfig()
	}
	return SMSLogRuntimeConfig{
		MaxCount:            getDirectSettingPositiveInt("sms_log_max_count", 1000),
		PerUserLimitEnabled: getDirectSettingBool("sms_log_per_user_limit_enabled", false),
		PerUserMaxCount:     getDirectSettingPositiveInt("sms_log_per_user_max_count", 1000),
	}
}

func (s *SettingsService) GetEmailLogRuntimeConfig() EmailLogRuntimeConfig {
	return EmailLogRuntimeConfig{
		MaxCount:            s.getRuntimePositiveInt("email_log_max_count", 1000),
		PerUserLimitEnabled: s.getRuntimeBool("email_log_per_user_limit_enabled", false),
		PerUserMaxCount:     s.getRuntimePositiveInt("email_log_per_user_max_count", 1000),
	}
}

func GetGlobalEmailLogRuntimeConfig() EmailLogRuntimeConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetEmailLogRuntimeConfig()
	}
	return EmailLogRuntimeConfig{
		MaxCount:            getDirectSettingPositiveInt("email_log_max_count", 1000),
		PerUserLimitEnabled: getDirectSettingBool("email_log_per_user_limit_enabled", false),
		PerUserMaxCount:     getDirectSettingPositiveInt("email_log_per_user_max_count", 1000),
	}
}

func (s *SettingsService) GetAPIRateLimitRuntimeConfig() RateLimitRuntimeConfig {
	return RateLimitRuntimeConfig{
		Enabled: s.getRuntimeBool("api_rate_limit_enabled", false),
		Rate:    s.getRuntimePositiveInt("api_rate_limit_rate", 120),
		Burst:   s.getRuntimePositiveInt("api_rate_limit_burst", 240),
	}
}

func GetGlobalAPIRateLimitRuntimeConfig() RateLimitRuntimeConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetAPIRateLimitRuntimeConfig()
	}
	return RateLimitRuntimeConfig{
		Enabled: getDirectSettingBool("api_rate_limit_enabled", false),
		Rate:    getDirectSettingPositiveInt("api_rate_limit_rate", 120),
		Burst:   getDirectSettingPositiveInt("api_rate_limit_burst", 240),
	}
}

func (s *SettingsService) GetAdminRateLimitRuntimeConfig() RateLimitRuntimeConfig {
	return RateLimitRuntimeConfig{
		Enabled: s.getRuntimeBool("admin_rate_limit_enabled", false),
		Rate:    s.getRuntimePositiveInt("admin_rate_limit_rate", 60),
		Burst:   s.getRuntimePositiveInt("admin_rate_limit_burst", 120),
	}
}

func GetGlobalAdminRateLimitRuntimeConfig() RateLimitRuntimeConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetAdminRateLimitRuntimeConfig()
	}
	return RateLimitRuntimeConfig{
		Enabled: getDirectSettingBool("admin_rate_limit_enabled", false),
		Rate:    getDirectSettingPositiveInt("admin_rate_limit_rate", 60),
		Burst:   getDirectSettingPositiveInt("admin_rate_limit_burst", 120),
	}
}

// GetGlobalAPIKeyAuthEnabled 是否允许 X-Api-Key 方式鉴权（默认 false，需管理员在后台主动开启）。
// 走与其他运行时配置一致的内存缓存（GlobalSettingsService），单机场景下管理员后台保存后立即生效，
// 不需要重启进程；命中的是进程内存里的 map，不会对每次请求都查一次数据库，性能上是安全的。
func GetGlobalAPIKeyAuthEnabled() bool {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimeBool("api_key_auth_enabled", false)
	}
	return getDirectSettingBool("api_key_auth_enabled", false)
}

func GetGlobalAllowRegister() bool {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimeBool("allow_register", false)
	}
	return getDirectSettingBool("allow_register", false)
}

func GetGlobalAllowDeleteAccount() bool {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimeBool("allow_delete_account", false)
	}
	return getDirectSettingBool("allow_delete_account", false)
}

// GetGlobalAllowUserLogin 是否允许普通用户密码登录拿 JWT（默认 false）。关闭不影响管理员登录与 API Key。
func GetGlobalAllowUserLogin() bool {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimeBool("allow_user_login", false)
	}
	return getDirectSettingBool("allow_user_login", false)
}

func GetGlobalPaymentEnabled() bool {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimeBool("payment_enabled", false)
	}
	return getDirectSettingBool("payment_enabled", false)
}

func GetGlobalPaymentOrderExpireMinutes() int {
	const fallback = 30
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimePositiveInt("payment_order_expire_minutes", fallback)
	}
	return getDirectSettingPositiveInt("payment_order_expire_minutes", fallback)
}

func GetGlobalRegisterCodeExpireMinutes() int {
	const fallback = 60
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimePositiveInt("register_code_expire_minutes", fallback)
	}
	return getDirectSettingPositiveInt("register_code_expire_minutes", fallback)
}

// GetGlobalRegisterDefaultLevel 自助注册时写入的 users.level，默认 1。
func GetGlobalRegisterDefaultLevel() uint64 {
	const fallback = 1
	var level int
	if GlobalSettingsService != nil {
		level = GlobalSettingsService.getRuntimePositiveInt("register_default_level", fallback)
	} else {
		level = getDirectSettingPositiveInt("register_default_level", fallback)
	}
	if level < 1 {
		return fallback
	}
	// 防止误配超大值；等级能力表一般只维护少量档位
	if level > 9999 {
		return fallback
	}
	return uint64(level)
}

func GetGlobalFrontendURL() string {
	fallback := ""
	if cfg := config.CloneGlobalConfig(); cfg != nil {
		fallback = cfg.FrontendURL
	}
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimeString("frontend_url", fallback)
	}
	return getDirectSettingString("frontend_url", fallback)
}

func GetGlobalBackendAPIURL() string {
	fallback := ""
	if cfg := config.CloneGlobalConfig(); cfg != nil {
		fallback = cfg.BackendAPIURL
	}
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimeString("backend_api_url", fallback)
	}
	return getDirectSettingString("backend_api_url", fallback)
}

// ApplyGlobalRuntimeConfig 将 DB/缓存中的运行时配置合并进全局配置。
// 在锁外准备好全部新值，再通过 UpdateGlobalConfig 一次性写入，避免并发字段竞态。
func ApplyGlobalRuntimeConfig() {
	base := config.CloneGlobalConfig()
	if base == nil {
		return
	}

	geetestConfig := GetGlobalGeetestRuntimeConfig()
	verifyConfig := GetGlobalVerifyConfig()
	smsConfig := GetGlobalSMSRuntimeConfig()

	// 基于当前快照合并 DB 运行时配置（settings 读取在锁外完成）
	next := *base
	if GlobalSettingsService != nil {
		s := GlobalSettingsService
		next.AppName = s.getRuntimeString("site_name", base.AppName)
		next.FrontendURL = s.getRuntimeString("frontend_url", base.FrontendURL)
		next.BackendAPIURL = s.getRuntimeString("backend_api_url", base.BackendAPIURL)
		next.SMTPHost = s.getRuntimeString("smtp_host", base.SMTPHost)
		next.SMTPPort = s.getRuntimeString("smtp_port", base.SMTPPort)
		next.SMTPUser = s.getRuntimeString("smtp_username", base.SMTPUser)
		next.SMTPPass = s.getRuntimeString("smtp_password", base.SMTPPass)
		next.SMTPSSL = s.getRuntimeBool("smtp_ssl", base.SMTPSSL)
		next.SMTPProxyEnabled = s.getRuntimeBool("smtp_proxy_enabled", base.SMTPProxyEnabled)
		next.SMTPProxyType = s.getRuntimeString("smtp_proxy_type", base.SMTPProxyType)
		next.SMTPProxyHost = s.getRuntimeString("smtp_proxy_host", base.SMTPProxyHost)
		next.SMTPProxyPort = s.getRuntimeString("smtp_proxy_port", base.SMTPProxyPort)
		next.SMTPProxyUser = s.getRuntimeString("smtp_proxy_username", base.SMTPProxyUser)
		next.SMTPProxyPass = s.getRuntimeString("smtp_proxy_password", base.SMTPProxyPass)
		next.SystemEmail = s.getRuntimeString("system_email_address", base.SystemEmail)
		next.SystemEmailName = s.getRuntimeString("system_email_name", base.SystemEmailName)
		next.RegisterCodeExpireMinutes = s.getRuntimePositiveInt("register_code_expire_minutes", base.RegisterCodeExpireMinutes)
		next.JWTAccessExpire = s.getRuntimePositiveInt("jwt_access_expire", base.JWTAccessExpire)
		next.JWTRefreshExpire = s.getRuntimePositiveInt("jwt_refresh_expire", base.JWTRefreshExpire)
		next.LoginMaxFailureCount = s.getRuntimePositiveInt("login_max_failure", base.LoginMaxFailureCount)
		next.LoginLockDurationMinutes = s.getRuntimePositiveInt("login_lock_duration", base.LoginLockDurationMinutes)
	}

	next.GeetestEnabled = geetestConfig.Enabled
	next.GeetestID = geetestConfig.CaptchaID
	next.GeetestKey = geetestConfig.CaptchaKey
	next.EmailVerifyEnabled = verifyConfig.EmailEnabled
	next.SMSVerifyEnabled = verifyConfig.SMSEnabled
	next.SMSProvider = smsConfig.Provider
	next.SMSAccessKey = smsConfig.AccessKey
	next.SMSSecretKey = smsConfig.SecretKey
	next.SMSSignName = smsConfig.SignName
	next.SMSTemplateCode = smsConfig.TemplateCode
	next.SMSTemplateCodeEN = smsConfig.TemplateCodeEN
	next.SMSRegion = smsConfig.Region
	next.SMSEndpoint = smsConfig.Endpoint
	next.SMSSdkAppID = smsConfig.SdkAppID
	next.SMSBodyFormat = smsConfig.BodyFormat

	// 写锁内整体替换字段（保持指针稳定，兼容仍直接读 GlobalConfig 的旧代码）
	config.UpdateGlobalConfig(func(cfg *config.Config) {
		*cfg = next
	})

	warnIfProductionRateLimitDisabled()
}

// warnIfProductionRateLimitDisabled 生产环境限流软提醒。
// 仅打印非致命 Warning，不阻断启动、不影响业务；不涉及 JWT 校验逻辑
// （JWT 相关的强校验见 pkg/config.validateCriticalSecurityConfig，本函数不与之交叉）。
func warnIfProductionRateLimitDisabled() {
	if !config.IsProductionMode() {
		return
	}
	if GetGlobalAPIRateLimitRuntimeConfig().Enabled {
		return
	}
	log.Println("[Security Warning] 当前判定为生产环境（APP_ENV/GO_ENV/GIN_MODE 或 APP_MODE），但系统设置里全局 API 限流 api_rate_limit_enabled=false；建议在管理后台「系统设置-安全」中开启，避免接口被刷/被爬")
}

// GetPublicAppConfig returns public app config consumed by frontend bootstrap.
func (s *SettingsService) GetPublicAppConfig() *PublicAppConfig {
	geetestConfig := s.GetGeetestRuntimeConfig()
	verifyConfig := s.GetVerifyConfig()

	adminAPIPath := "/admin"
	if cfg := config.CloneGlobalConfig(); cfg != nil {
		adminAPIPath = config.NormalizeAdminAPIPath(cfg.AdminAPIPath)
	}

	return &PublicAppConfig{
		SiteName:                    s.GetWithDefault("site_name", "F.st"),
		SiteDesc:                    s.GetWithDefault("site_desc", "Full-stack admin template based on Go + Vue 3"),
		SiteLogo:                    s.GetWithDefault("site_logo", ""),
		Copyright:                   s.GetWithDefault("copyright", "(c) 2024 F.st"),
		ICP:                         s.GetWithDefault("icp", ""),
		AllowRegister:               s.GetBoolWithDefault("allow_register", false),
		AllowUserLogin:              s.GetBoolWithDefault("allow_user_login", false),
		AnnouncementEnabled:         s.GetBoolWithDefault("announcement_enabled", true),
		AllowDeleteAccount:          s.GetBool("allow_delete_account"),
		DefaultLang:                 s.GetWithDefault("default_lang", "zhCN"),
		Version:                     s.GetWithDefault("version", "1.0.0"),
		GeetestEnabled:              geetestConfig.Enabled,
		GeetestCaptchaId:            geetestConfig.CaptchaID,
		EmailVerifyEnabled:          verifyConfig.EmailEnabled,
		SMSVerifyEnabled:            verifyConfig.SMSEnabled,
		MobileCNOnly:                s.GetBoolWithDefault("mobile_cn_only", true),
		MobileIPCountryDetect:       s.GetBoolWithDefault("mobile_ip_country_detect", false),
		RealnameEnabled:             s.GetBoolWithDefault("realname_enabled", true),
		RealnameNotifyText:          s.GetWithDefault("realname_notify_text", ""),
		WithdrawEnabled:             s.GetBoolWithDefault("withdraw_enabled", true),
		WithdrawMinAmount:           parseJSONFloatWithDefault(s.GetWithDefault("withdraw_min_amount", "10"), 10),
		WithdrawNotifyText:          s.GetWithDefault("withdraw_notify_text", ""),
		WithdrawAccountTypes:        parseJSONStringArrayWithDefault(s.GetWithDefault("withdraw_account_types", "[\"bank\",\"alipay\",\"wechat\",\"usdt\"]"), []string{"bank", "alipay", "wechat", "usdt"}),
		WithdrawRequireRealname:     s.GetBoolWithDefault("withdraw_require_realname", false),
		AdminAPIPath:                adminAPIPath,
		PresenceEnabled:             s.GetBoolWithDefault("presence_enabled", false),
		OnlineReportIntervalSeconds: s.GetOnlinePresenceRuntimeConfig().ReportIntervalSeconds,
		UserAPILogVisible:           s.GetBoolWithDefault("user_api_log_visible", true),
		UserOperationLogVisible:     s.GetBoolWithDefault("user_operation_log_visible", true),
	}
}

func parseJSONFloatWithDefault(val string, fallback float64) float64 {
	var result float64
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return fallback
	}
	return result
}

func parseJSONStringArrayWithDefault(val string, fallback []string) []string {
	var result []string
	if err := json.Unmarshal([]byte(val), &result); err != nil || len(result) == 0 {
		return fallback
	}
	return result
}

func ParseJSONFloatForPublic(val string, fallback float64) float64 {
	return parseJSONFloatWithDefault(val, fallback)
}

func ParseJSONStringArrayForPublic(val string, fallback []string) []string {
	return parseJSONStringArrayWithDefault(val, fallback)
}

// refreshAfterWrite 在写入配置后刷新缓存，并在需要时同步运行时配置。
func (s *SettingsService) refreshAfterWrite() error {
	if s == nil {
		ApplyGlobalRuntimeConfig()
		return nil
	}
	if err := s.RefreshCache(); err != nil {
		return err
	}
	if s == GlobalSettingsService {
		ApplyGlobalRuntimeConfig()
	}
	return nil
}

// UpdateSettingsWithCache 批量更新配置并刷新缓存。
func (s *SettingsService) UpdateSettingsWithCache(settings map[string]string) error {
	err := models.BatchUpdateSettings(settings)
	if err != nil {
		return err
	}
	return s.refreshAfterWrite()
}

// UpdateSingleSettingWithCache 更新单项配置并刷新缓存。
func (s *SettingsService) UpdateSingleSettingWithCache(key, value string) error {
	err := models.UpdateSetting(key, value)
	if err != nil {
		return err
	}
	return s.refreshAfterWrite()
}
