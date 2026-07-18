package admin

import (
	"encoding/json"
	"fst/backend/app/models"
	sms_plugin "fst/backend/app/plugins/sms"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	"log"
	"strconv"
	"strings"
)

const sensitiveSettingMaskedValue = "********"

var sensitiveSettingKeys = map[string]struct{}{
	"geetest_captcha_key": {},
	"smtp_password":       {},
	"sms_access_key":      {},
	"sms_secret_key":      {},
}

// ========================================
// 辅助方法
// ========================================

// isValidType 验证配置类型是否有效
func (ctrl *SettingsController) isValidType(t string) bool {
	validTypes := map[string]bool{
		"string":  true,
		"number":  true,
		"boolean": true,
		"json":    true,
	}
	return validTypes[t]
}

// isValidCategory 验证分类是否有效
func (ctrl *SettingsController) isValidCategory(cat string) bool {
	validCategories := map[string]bool{
		"basic":    true,
		"security": true,
		"email":    true,
		"payment":  true,
		"sms":      true,
		"custom":   true,
	}
	return validCategories[cat]
}

// validateSettingValue 根据类型验证值
func (ctrl *SettingsController) validateSettingValue(value, typ string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return true
	}
	switch typ {
	case "number":
		_, err := strconv.ParseFloat(trimmed, 64)
		return err == nil
	case "boolean":
		lower := strings.ToLower(trimmed)
		return lower == "true" || lower == "false" || lower == "1" || lower == "0"
	case "json":
		var payload interface{}
		return json.Unmarshal([]byte(trimmed), &payload) == nil
	case "string":
		return true
	default:
		return true
	}
}

func (ctrl *SettingsController) resolveSettingValueForAdmin(setting models.SystemSetting) interface{} {
	cfg := currentGlobalConfig()
	switch setting.Key {
	case "geetest_enabled":
		return services.GetGlobalGeetestRuntimeConfig().Enabled
	case "geetest_captcha_id":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(cfg.GeetestID)
		}
		return val
	case "geetest_captcha_key":
		return ctrl.maskSensitiveSettingValue(ctrl.resolveCurrentSensitiveSettingValue(setting))
	case "smtp_host":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(cfg.SMTPHost)
		}
		return val
	case "smtp_port":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			if port, err := strconv.Atoi(strings.TrimSpace(cfg.SMTPPort)); err == nil && port > 0 {
				return port
			}
			return setting.GetTypedValue()
		}
		if port, err := strconv.Atoi(val); err == nil && port > 0 {
			return port
		}
		return setting.GetTypedValue()
	case "smtp_username":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(cfg.SMTPUser)
		}
		return val
	case "smtp_password":
		return ctrl.maskSensitiveSettingValue(ctrl.resolveCurrentSensitiveSettingValue(setting))
	case "smtp_ssl":
		if strings.TrimSpace(setting.Value) == "" {
			return cfg.SMTPSSL
		}
		return setting.GetTypedValue()
	case "system_email_address":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(cfg.SystemEmail)
		}
		return val
	case "system_email_name":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(cfg.SystemEmailName)
		}
		return val
	case "frontend_url":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(cfg.FrontendURL)
		}
		return val
	case "register_code_expire_minutes":
		if strings.TrimSpace(setting.Value) == "" {
			return services.GetGlobalRegisterCodeExpireMinutes()
		}
		return setting.GetTypedValue()
	case "jwt_access_expire":
		if strings.TrimSpace(setting.Value) == "" {
			if cfg.JWTAccessExpire > 0 {
				return cfg.JWTAccessExpire
			}
		}
		return setting.GetTypedValue()
	case "jwt_refresh_expire":
		if strings.TrimSpace(setting.Value) == "" {
			if cfg.JWTRefreshExpire > 0 {
				return cfg.JWTRefreshExpire
			}
		}
		return setting.GetTypedValue()
	case "login_max_failure":
		if strings.TrimSpace(setting.Value) == "" {
			if cfg.LoginMaxFailureCount > 0 {
				return cfg.LoginMaxFailureCount
			}
		}
		return setting.GetTypedValue()
	case "login_lock_duration":
		if strings.TrimSpace(setting.Value) == "" {
			if cfg.LoginLockDurationMinutes > 0 {
				return cfg.LoginLockDurationMinutes
			}
		}
		return setting.GetTypedValue()
	case "email_verify_enabled":
		return services.GetGlobalVerifyConfig().EmailEnabled
	case "sms_verify_enabled":
		return services.GetGlobalVerifyConfig().SMSEnabled
	case "sms_provider":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = cfg.SMSProvider
		}
		if val == "" {
			val = "console"
		}
		return val
	case "sms_access_key":
		return ctrl.maskSensitiveSettingValue(ctrl.resolveCurrentSensitiveSettingValue(setting))
	case "sms_secret_key":
		return ctrl.maskSensitiveSettingValue(ctrl.resolveCurrentSensitiveSettingValue(setting))
	case "sms_sign_name":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = cfg.SMSSignName
		}
		return val
	case "sms_template_code":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = cfg.SMSTemplateCode
		}
		return val
	case "sms_template_code_en":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = cfg.SMSTemplateCodeEN
		}
		return val
	case "sms_region":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = cfg.SMSRegion
		}
		return val
	case "sms_sdk_app_id":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = cfg.SMSSdkAppID
		}
		return val
	case "sms_endpoint":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = cfg.SMSEndpoint
		}
		return val
	case "sms_body_format":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = cfg.SMSBodyFormat
		}
		if val == "" {
			val = "json"
		}
		return val
	default:
		return setting.GetTypedValue()
	}
}

func isSensitiveSettingKey(key string) bool {
	_, ok := sensitiveSettingKeys[key]
	return ok
}

func (ctrl *SettingsController) maskSensitiveSettingValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return sensitiveSettingMaskedValue
}

func (ctrl *SettingsController) resolveCurrentSensitiveSettingValue(setting models.SystemSetting) string {
	if strings.TrimSpace(setting.Value) != "" {
		return setting.Value
	}
	cfg := currentGlobalConfig()

	switch setting.Key {
	case "geetest_captcha_key":
		return strings.TrimSpace(cfg.GeetestKey)
	case "smtp_password":
		return cfg.SMTPPass
	case "sms_access_key":
		return strings.TrimSpace(cfg.SMSAccessKey)
	case "sms_secret_key":
		return strings.TrimSpace(cfg.SMSSecretKey)
	default:
		return setting.Value
	}
}

func (ctrl *SettingsController) normalizeSettingValueForWrite(setting models.SystemSetting, value string) string {
	if isSensitiveSettingKey(setting.Key) && value == sensitiveSettingMaskedValue {
		return ctrl.resolveCurrentSensitiveSettingValue(setting)
	}
	return value
}

// currentGlobalConfig 读取全局配置快照，避免直接持有可变指针带来的竞态风险。
func currentGlobalConfig() *config.Config {
	if cfg := config.CloneGlobalConfig(); cfg != nil {
		return cfg
	}
	return &config.Config{}
}

func (ctrl *SettingsController) refreshRuntimeConfig() {
	services.ReloadGlobalRuntimeConfig()

	// 同步短信服务配置并更新 SMS Provider
	smsConfig := services.GetGlobalSMSRuntimeConfig()

	if services.GlobalSMSService != nil {
		sms_plugin.ApplyRuntimeProvider(services.SMSConfig(smsConfig))
	}

	apiLogConfig := services.GetGlobalAPILogRuntimeConfig()
	if apiLogConfig.MaxCount > 0 {
		go func(maxCount int) {
			if _, err := models.CleanExcessAPIAccessLogs(maxCount); err != nil {
				log.Printf("[APIAccessLog] 应用运行时配置后清理超限日志失败: %v", err)
				return
			}
			invalidateAPILogStatsCache()
		}(apiLogConfig.MaxCount)
	}
}
