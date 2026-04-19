package services

import (
	"encoding/json"
	"fst/backend/app/models"
	"fst/backend/pkg/config"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
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

// GetInt returns int setting value. Missing keys return 0.
func (s *SettingsService) GetInt(key string) int {
	val, ok := s.Get(key)
	if !ok {
		return 0
	}
	var result int
	_ = json.Unmarshal([]byte(val), &result)
	return result
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
	SiteName           string `json:"site_name"`
	SiteDesc           string `json:"site_desc"`
	SiteLogo           string `json:"site_logo"`
	Copyright          string `json:"copyright"`
	ICP                string `json:"icp"`
	AllowRegister      bool   `json:"allow_register"`
	AllowDeleteAccount bool   `json:"allow_delete_account"`
	DefaultLang        string `json:"default_lang"`
	Version            string `json:"version"`
	GeetestEnabled     bool   `json:"geetest_enabled"`
	GeetestCaptchaId   string `json:"geetest_captcha_id"`
	EmailVerifyEnabled bool   `json:"email_verify_enabled"`
	SMSVerifyEnabled   bool   `json:"sms_verify_enabled"`
	RealnameEnabled    bool   `json:"realname_enabled"`
	RealnameNotifyText string `json:"realname_notify_text"`
	WithdrawEnabled    bool     `json:"withdraw_enabled"`
	WithdrawMinAmount  float64  `json:"withdraw_min_amount"`
	WithdrawNotifyText string   `json:"withdraw_notify_text"`
	WithdrawAccountTypes []string `json:"withdraw_account_types"`
}

// VerifyConfig 验证码功能开关运行时配置
type VerifyConfig struct {
	EmailEnabled bool
	SMSEnabled   bool
}

// SMSRuntimeConfig 短信服务运行时配置
type SMSRuntimeConfig struct {
	Provider     string
	AccessKey    string
	SecretKey    string
	SignName     string
	TemplateCode string
	TemplateCodeEN string
	Region       string
	Endpoint     string
	SdkAppID     string
	BodyFormat   string
}

// APILogRuntimeConfig API访问日志运行时配置
type APILogRuntimeConfig struct {
	Enabled   bool
	QueryDays int
	MaxCount  int
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

// RealnameAPIRuntimeConfig 实名认证API运行时配置
type RealnameAPIRuntimeConfig struct {
	Enabled   bool   // 是否启用自动实名认证验证
	Provider  string // 提供商: aliyun, tencent, baidu, custom
	AppKey    string // App Key / Access Key
	AppSecret string // App Secret / Secret Key
	Endpoint  string // 可选，自定义API地址
}

func parseBoolSetting(val string) bool {
	return parseBoolSettingWithFallback(val, false)
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
	enabled := s.getRuntimeBool("geetest_enabled", config.GlobalConfig.GeetestEnabled)
	captchaID := s.getRuntimeString("geetest_captcha_id", config.GlobalConfig.GeetestID)
	captchaKey := s.getRuntimeString("geetest_captcha_key", config.GlobalConfig.GeetestKey)

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

	captchaID := strings.TrimSpace(config.GlobalConfig.GeetestID)
	captchaKey := strings.TrimSpace(config.GlobalConfig.GeetestKey)

	return GeetestRuntimeConfig{
		Enabled:    config.GlobalConfig.GeetestEnabled && captchaID != "" && captchaKey != "",
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
	return VerifyConfig{
		EmailEnabled: s.getRuntimeBool("email_verify_enabled", config.GlobalConfig.EmailVerifyEnabled),
		SMSEnabled:   s.getRuntimeBool("sms_verify_enabled", config.GlobalConfig.SMSVerifyEnabled),
	}
}

// GetSMSRuntimeConfig returns effective SMS provider config.
func (s *SettingsService) GetSMSRuntimeConfig() SMSRuntimeConfig {
	return SMSRuntimeConfig{
		Provider:       s.getRuntimeString("sms_provider", config.GlobalConfig.SMSProvider),
		AccessKey:      s.getRuntimeString("sms_access_key", config.GlobalConfig.SMSAccessKey),
		SecretKey:      s.getRuntimeString("sms_secret_key", config.GlobalConfig.SMSSecretKey),
		SignName:       s.getRuntimeString("sms_sign_name", config.GlobalConfig.SMSSignName),
		TemplateCode:   s.getRuntimeString("sms_template_code", config.GlobalConfig.SMSTemplateCode),
		TemplateCodeEN: s.getRuntimeString("sms_template_code_en", config.GlobalConfig.SMSTemplateCodeEN),
		Region:         s.getRuntimeString("sms_region", config.GlobalConfig.SMSRegion),
		Endpoint:       s.getRuntimeString("sms_endpoint", config.GlobalConfig.SMSEndpoint),
		SdkAppID:       s.getRuntimeString("sms_sdk_app_id", config.GlobalConfig.SMSSdkAppID),
		BodyFormat:     s.getRuntimeString("sms_body_format", config.GlobalConfig.SMSBodyFormat),
	}
}

// GetGlobalVerifyConfig returns effective verify config with global cache.
func GetGlobalVerifyConfig() VerifyConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetVerifyConfig()
	}
	return VerifyConfig{
		EmailEnabled: config.GlobalConfig.EmailVerifyEnabled,
		SMSEnabled:   config.GlobalConfig.SMSVerifyEnabled,
	}
}

// GetGlobalSMSRuntimeConfig returns effective SMS config with global cache.
func GetGlobalSMSRuntimeConfig() SMSRuntimeConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetSMSRuntimeConfig()
	}
	return SMSRuntimeConfig{
		Provider:       config.GlobalConfig.SMSProvider,
		AccessKey:      config.GlobalConfig.SMSAccessKey,
		SecretKey:      config.GlobalConfig.SMSSecretKey,
		SignName:       config.GlobalConfig.SMSSignName,
		TemplateCode:   config.GlobalConfig.SMSTemplateCode,
		TemplateCodeEN: config.GlobalConfig.SMSTemplateCodeEN,
		Region:         config.GlobalConfig.SMSRegion,
		Endpoint:       config.GlobalConfig.SMSEndpoint,
		SdkAppID:       config.GlobalConfig.SMSSdkAppID,
		BodyFormat:     config.GlobalConfig.SMSBodyFormat,
	}
}

func (s *SettingsService) GetAPILogRuntimeConfig() APILogRuntimeConfig {
	return APILogRuntimeConfig{
		Enabled:   s.getRuntimeBool("api_access_log_enabled", true),
		QueryDays: s.getRuntimePositiveInt("api_log_query_days", 7),
		MaxCount:  s.getRuntimePositiveInt("api_log_max_count", 1000),
	}
}

func GetGlobalAPILogRuntimeConfig() APILogRuntimeConfig {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.GetAPILogRuntimeConfig()
	}
	return APILogRuntimeConfig{
		Enabled:   getDirectSettingBool("api_access_log_enabled", true),
		QueryDays: getDirectSettingPositiveInt("api_log_query_days", 7),
		MaxCount:  getDirectSettingPositiveInt("api_log_max_count", 1000),
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

func GetGlobalAllowRegister() bool {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimeBool("allow_register", true)
	}
	return getDirectSettingBool("allow_register", true)
}

func GetGlobalAllowDeleteAccount() bool {
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimeBool("allow_delete_account", false)
	}
	return getDirectSettingBool("allow_delete_account", false)
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

func GetGlobalFrontendURL() string {
	fallback := ""
	if config.GlobalConfig != nil {
		fallback = config.GlobalConfig.FrontendURL
	}
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimeString("frontend_url", fallback)
	}
	return getDirectSettingString("frontend_url", fallback)
}

func GetGlobalBackendAPIURL() string {
	fallback := ""
	if config.GlobalConfig != nil {
		fallback = config.GlobalConfig.BackendAPIURL
	}
	if GlobalSettingsService != nil {
		return GlobalSettingsService.getRuntimeString("backend_api_url", fallback)
	}
	return getDirectSettingString("backend_api_url", fallback)
}

func ApplyGlobalRuntimeConfig() {
	if config.GlobalConfig == nil {
		return
	}

	geetestConfig := GetGlobalGeetestRuntimeConfig()
	config.GlobalConfig.GeetestEnabled = geetestConfig.Enabled
	config.GlobalConfig.GeetestID = geetestConfig.CaptchaID
	config.GlobalConfig.GeetestKey = geetestConfig.CaptchaKey

	if GlobalSettingsService != nil {
		config.GlobalConfig.AppName = GlobalSettingsService.getRuntimeString("site_name", config.GlobalConfig.AppName)
		config.GlobalConfig.FrontendURL = GlobalSettingsService.getRuntimeString("frontend_url", config.GlobalConfig.FrontendURL)
		config.GlobalConfig.BackendAPIURL = GlobalSettingsService.getRuntimeString("backend_api_url", config.GlobalConfig.BackendAPIURL)
		config.GlobalConfig.SMTPHost = GlobalSettingsService.getRuntimeString("smtp_host", config.GlobalConfig.SMTPHost)
		config.GlobalConfig.SMTPPort = GlobalSettingsService.getRuntimeString("smtp_port", config.GlobalConfig.SMTPPort)
		config.GlobalConfig.SMTPUser = GlobalSettingsService.getRuntimeString("smtp_username", config.GlobalConfig.SMTPUser)
		config.GlobalConfig.SMTPPass = GlobalSettingsService.getRuntimeString("smtp_password", config.GlobalConfig.SMTPPass)
		config.GlobalConfig.SMTPSSL = GlobalSettingsService.getRuntimeBool("smtp_ssl", config.GlobalConfig.SMTPSSL)
		config.GlobalConfig.SystemEmail = GlobalSettingsService.getRuntimeString("system_email_address", config.GlobalConfig.SystemEmail)
		config.GlobalConfig.SystemEmailName = GlobalSettingsService.getRuntimeString("system_email_name", config.GlobalConfig.SystemEmailName)
		config.GlobalConfig.RegisterCodeExpireMinutes = GlobalSettingsService.getRuntimePositiveInt("register_code_expire_minutes", config.GlobalConfig.RegisterCodeExpireMinutes)
		config.GlobalConfig.JWTAccessExpire = GlobalSettingsService.getRuntimePositiveInt("jwt_access_expire", config.GlobalConfig.JWTAccessExpire)
		config.GlobalConfig.JWTRefreshExpire = GlobalSettingsService.getRuntimePositiveInt("jwt_refresh_expire", config.GlobalConfig.JWTRefreshExpire)
		config.GlobalConfig.LoginMaxFailureCount = GlobalSettingsService.getRuntimePositiveInt("login_max_failure", config.GlobalConfig.LoginMaxFailureCount)
		config.GlobalConfig.LoginLockDurationMinutes = GlobalSettingsService.getRuntimePositiveInt("login_lock_duration", config.GlobalConfig.LoginLockDurationMinutes)
	}

	verifyConfig := GetGlobalVerifyConfig()
	config.GlobalConfig.EmailVerifyEnabled = verifyConfig.EmailEnabled
	config.GlobalConfig.SMSVerifyEnabled = verifyConfig.SMSEnabled

	smsConfig := GetGlobalSMSRuntimeConfig()
	config.GlobalConfig.SMSProvider = smsConfig.Provider
	config.GlobalConfig.SMSAccessKey = smsConfig.AccessKey
	config.GlobalConfig.SMSSecretKey = smsConfig.SecretKey
	config.GlobalConfig.SMSSignName = smsConfig.SignName
	config.GlobalConfig.SMSTemplateCode = smsConfig.TemplateCode
	config.GlobalConfig.SMSTemplateCodeEN = smsConfig.TemplateCodeEN
	config.GlobalConfig.SMSRegion = smsConfig.Region
	config.GlobalConfig.SMSEndpoint = smsConfig.Endpoint
	config.GlobalConfig.SMSSdkAppID = smsConfig.SdkAppID
	config.GlobalConfig.SMSBodyFormat = smsConfig.BodyFormat
}

// GetPublicAppConfig returns public app config consumed by frontend bootstrap.
func (s *SettingsService) GetPublicAppConfig() *PublicAppConfig {
	geetestConfig := s.GetGeetestRuntimeConfig()
	verifyConfig := s.GetVerifyConfig()

	return &PublicAppConfig{
		SiteName:           s.GetWithDefault("site_name", "F.st"),
		SiteDesc:           s.GetWithDefault("site_desc", "Full-stack admin template based on Go + Vue 3"),
		SiteLogo:           s.GetWithDefault("site_logo", ""),
		Copyright:          s.GetWithDefault("copyright", "(c) 2024 F.st"),
		ICP:                s.GetWithDefault("icp", ""),
		AllowRegister:      s.GetBoolWithDefault("allow_register", true),
		AllowDeleteAccount: s.GetBool("allow_delete_account"),
		DefaultLang:        s.GetWithDefault("default_lang", "zhCN"),
		Version:            s.GetWithDefault("version", "1.0.0"),
		GeetestEnabled:     geetestConfig.Enabled,
		GeetestCaptchaId:   geetestConfig.CaptchaID,
		EmailVerifyEnabled: verifyConfig.EmailEnabled,
		SMSVerifyEnabled:   verifyConfig.SMSEnabled,
		RealnameEnabled:    s.GetBoolWithDefault("realname_enabled", true),
		RealnameNotifyText: s.GetWithDefault("realname_notify_text", ""),
		WithdrawEnabled:    s.GetBoolWithDefault("withdraw_enabled", true),
		WithdrawMinAmount:  parseJSONFloatWithDefault(s.GetWithDefault("withdraw_min_amount", "10"), 10),
		WithdrawNotifyText: s.GetWithDefault("withdraw_notify_text", ""),
		WithdrawAccountTypes: parseJSONStringArrayWithDefault(s.GetWithDefault("withdraw_account_types", "[\"bank\",\"alipay\",\"wechat\",\"usdt\"]"), []string{"bank", "alipay", "wechat", "usdt"}),
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
