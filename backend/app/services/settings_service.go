package services

import (
	"encoding/json"
	"fst/backend/app/models"
	"fst/backend/internal/config"
	"log"
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
	if err := GlobalSettingsService.RefreshCache(); err != nil {
		log.Printf("[SettingsService] Refresh cache failed: %v", err)
	}
	ApplyGlobalRuntimeConfig()
	log.Println("[SettingsService] Initialized with cache TTL: 5m")
}

// NewSettingsService creates a settings service with the given cache TTL.
func NewSettingsService(ttl time.Duration) *SettingsService {
	return &SettingsService{
		cache: make(map[string]*models.SystemSetting),
		ttl:   ttl,
	}
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
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	if time.Since(s.cacheTime) > s.ttl {
		s.cacheMu.RUnlock()
		_ = s.RefreshCache()
		s.cacheMu.RLock()
	}

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
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	if time.Since(s.cacheTime) > s.ttl {
		s.cacheMu.RUnlock()
		_ = s.RefreshCache()
		s.cacheMu.RLock()
	}

	return s.cache[key]
}

// GetAllFromCache returns a shallow copy of cache map.
func (s *SettingsService) GetAllFromCache() map[string]*models.SystemSetting {
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
	return val == "true" || val == "1" || strings.EqualFold(val, "true")
}

// GetGeetestRuntimeConfig returns effective geetest config.
// Priority: database values -> environment fallback.
func (s *SettingsService) GetGeetestRuntimeConfig() GeetestRuntimeConfig {
	enabled := config.GlobalConfig.GeetestEnabled
	if val, ok := s.Get("geetest_enabled"); ok {
		enabled = parseBoolSetting(strings.TrimSpace(val))
	}

	captchaID := strings.TrimSpace(config.GlobalConfig.GeetestID)
	if val, ok := s.Get("geetest_captcha_id"); ok {
		val = strings.TrimSpace(val)
		if val != "" {
			captchaID = val
		}
	}

	captchaKey := strings.TrimSpace(config.GlobalConfig.GeetestKey)
	if val, ok := s.Get("geetest_captcha_key"); ok {
		val = strings.TrimSpace(val)
		if val != "" {
			captchaKey = val
		}
	}

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
	get := func(dbKey, envFallback string) string {
		if val, ok := s.Get(dbKey); ok && strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
		return envFallback
	}

	enabled := parseBoolSetting(get("realname_api_enabled", "false"))
	appKey := get("realname_api_app_key", "")
	appSecret := get("realname_api_app_secret", "")

	// 只有当启用且有密钥时才认为真正启用
	enabled = enabled && appKey != "" && appSecret != ""

	return RealnameAPIRuntimeConfig{
		Enabled:   enabled,
		Provider:  get("realname_api_provider", "aliyun"),
		AppKey:    appKey,
		AppSecret: appSecret,
		Endpoint:  get("realname_api_endpoint", ""),
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
	emailEnabled := config.GlobalConfig.EmailVerifyEnabled
	if val, ok := s.Get("email_verify_enabled"); ok {
		emailEnabled = parseBoolSetting(strings.TrimSpace(val))
	}

	smsEnabled := config.GlobalConfig.SMSVerifyEnabled
	if val, ok := s.Get("sms_verify_enabled"); ok {
		smsEnabled = parseBoolSetting(strings.TrimSpace(val))
	}

	return VerifyConfig{
		EmailEnabled: emailEnabled,
		SMSEnabled:   smsEnabled,
	}
}

// GetSMSRuntimeConfig returns effective SMS provider config.
func (s *SettingsService) GetSMSRuntimeConfig() SMSRuntimeConfig {
	get := func(dbKey, envFallback string) string {
		if val, ok := s.Get(dbKey); ok && strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
		return envFallback
	}

	return SMSRuntimeConfig{
		Provider:       get("sms_provider", config.GlobalConfig.SMSProvider),
		AccessKey:      get("sms_access_key", config.GlobalConfig.SMSAccessKey),
		SecretKey:      get("sms_secret_key", config.GlobalConfig.SMSSecretKey),
		SignName:       get("sms_sign_name", config.GlobalConfig.SMSSignName),
		TemplateCode:   get("sms_template_code", config.GlobalConfig.SMSTemplateCode),
		TemplateCodeEN: get("sms_template_code_en", config.GlobalConfig.SMSTemplateCodeEN),
		Region:         get("sms_region", config.GlobalConfig.SMSRegion),
		Endpoint:       get("sms_endpoint", config.GlobalConfig.SMSEndpoint),
		SdkAppID:       get("sms_sdk_app_id", config.GlobalConfig.SMSSdkAppID),
		BodyFormat:     get("sms_body_format", config.GlobalConfig.SMSBodyFormat),
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

func ApplyGlobalRuntimeConfig() {
	geetestConfig := GetGlobalGeetestRuntimeConfig()
	config.GlobalConfig.GeetestEnabled = geetestConfig.Enabled
	config.GlobalConfig.GeetestID = geetestConfig.CaptchaID
	config.GlobalConfig.GeetestKey = geetestConfig.CaptchaKey

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
		RealnameNotifyText: s.GetWithDefault("realname_notify_text", "完成实名认证后可享受更多服务"),
		WithdrawEnabled:    s.GetBoolWithDefault("withdraw_enabled", true),
		WithdrawMinAmount:  parseJSONFloatWithDefault(s.GetWithDefault("withdraw_min_amount", "10"), 10),
		WithdrawNotifyText: s.GetWithDefault("withdraw_notify_text", "提现申请提交后需管理员审核，通过后人工打款。"),
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

// UpdateSettingsWithCache updates settings in DB and invalidates cache.
func (s *SettingsService) UpdateSettingsWithCache(settings map[string]string) error {
	err := models.BatchUpdateSettings(settings)
	if err != nil {
		return err
	}
	s.InvalidateCache()
	return nil
}

// UpdateSingleSettingWithCache updates one setting in DB and invalidates cache.
func (s *SettingsService) UpdateSingleSettingWithCache(key, value string) error {
	err := models.UpdateSetting(key, value)
	if err != nil {
		return err
	}
	s.InvalidateCache()
	return nil
}
