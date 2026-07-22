package admin

import (
	"encoding/json"
	"fst/backend/app/models"
	sms_plugin "fst/backend/app/plugins/sms"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	"fst/backend/pkg/panicsafe"
	"log"
	"strconv"
	"strings"
)

const sensitiveSettingMaskedValue = "********"

var sensitiveSettingKeys = map[string]struct{}{
	"geetest_captcha_key": {},
	"smtp_password":       {},
	"smtp_proxy_password": {},
	"sms_access_key":      {},
	"sms_secret_key":      {},
}

// ========================================
// 辅助方法
// ========================================

// isValidType 验证配置类型是否有效
// password 语义等同 string，仅用于告知前端用密码输入框（带查看/隐藏眼睛）渲染，不做额外校验
func (ctrl *SettingsController) isValidType(t string) bool {
	validTypes := map[string]bool{
		"string":   true,
		"number":   true,
		"boolean":  true,
		"json":     true,
		"password": true,
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
	case "string", "password":
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
	case "smtp_proxy_enabled":
		if strings.TrimSpace(setting.Value) == "" {
			return cfg.SMTPProxyEnabled
		}
		return setting.GetTypedValue()
	case "smtp_proxy_type":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(cfg.SMTPProxyType)
		}
		if val == "" {
			val = "socks5"
		}
		return val
	case "smtp_proxy_host":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(cfg.SMTPProxyHost)
		}
		return val
	case "smtp_proxy_port":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			if port, err := strconv.Atoi(strings.TrimSpace(cfg.SMTPProxyPort)); err == nil && port > 0 {
				return port
			}
			return setting.GetTypedValue()
		}
		if port, err := strconv.Atoi(val); err == nil && port > 0 {
			return port
		}
		return setting.GetTypedValue()
	case "smtp_proxy_username":
		val := strings.TrimSpace(setting.Value)
		if val == "" {
			val = strings.TrimSpace(cfg.SMTPProxyUser)
		}
		return val
	case "smtp_proxy_password":
		return ctrl.maskSensitiveSettingValue(ctrl.resolveCurrentSensitiveSettingValue(setting))
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
	key = strings.TrimSpace(strings.ToLower(key))
	if key == "" {
		return false
	}
	if _, ok := sensitiveSettingKeys[key]; ok {
		return true
	}
	// 自定义配置名命中敏感关键词时，同样禁止公开，防止误标 is_public 外泄
	sensitiveFragments := []string{
		"password", "passwd", "secret", "private_key", "api_key", "apikey",
		"access_key", "token", "credential",
	}
	for _, frag := range sensitiveFragments {
		if strings.Contains(key, frag) {
			return true
		}
	}
	return false
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
	case "smtp_proxy_password":
		return cfg.SMTPProxyPass
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

	// 保存配置后立即触发各类日志保留清理（异步，互不影响）
	panicsafe.Go("Settings.postSaveCleanup", func() {
		apiLogConfig := services.GetGlobalAPILogRuntimeConfig()
		if apiLogConfig.MaxCount > 0 {
			if _, err := models.CleanExcessAPIAccessLogs(apiLogConfig.MaxCount); err != nil {
				log.Printf("[APIAccessLog] 应用运行时配置后清理超限日志失败: %v", err)
			} else {
				invalidateAPILogStatsCache()
			}
		}
		if apiLogConfig.PerUserLimitEnabled && apiLogConfig.PerUserMaxCount > 0 {
			if _, err := models.CleanExcessAPIAccessLogsPerUser(apiLogConfig.PerUserMaxCount); err != nil {
				log.Printf("[APIAccessLog] 应用运行时配置后按用户清理失败: %v", err)
			}
		}

		opCfg := services.GetGlobalOperationLogRuntimeConfig()
		if opCfg.MaxCount > 0 {
			if _, err := models.CleanExcessOperationLogs(opCfg.MaxCount); err != nil {
				log.Printf("[OperationLog] 应用运行时配置后清理超限日志失败: %v", err)
			}
		}
		if opCfg.PerUserLimitEnabled && opCfg.PerUserMaxCount > 0 {
			if _, err := models.CleanExcessOperationLogsPerUser(opCfg.PerUserMaxCount); err != nil {
				log.Printf("[OperationLog] 应用运行时配置后按用户清理失败: %v", err)
			}
		}

		smsLogCfg := services.GetGlobalSMSLogRuntimeConfig()
		if smsLogCfg.MaxCount > 0 {
			if _, err := models.CleanExcessSMSLogs(smsLogCfg.MaxCount); err != nil {
				log.Printf("[SMSLog] 应用运行时配置后清理超限日志失败: %v", err)
			}
		}
		if smsLogCfg.PerUserLimitEnabled && smsLogCfg.PerUserMaxCount > 0 {
			if _, err := models.CleanExcessSMSLogsPerRecipient(smsLogCfg.PerUserMaxCount); err != nil {
				log.Printf("[SMSLog] 应用运行时配置后按收件人清理失败: %v", err)
			}
		}

		emailLogCfg := services.GetGlobalEmailLogRuntimeConfig()
		if emailLogCfg.MaxCount > 0 {
			if _, err := models.CleanExcessEmailLogs(emailLogCfg.MaxCount); err != nil {
				log.Printf("[EmailLog] 应用运行时配置后清理超限日志失败: %v", err)
			}
		}
		if emailLogCfg.PerUserLimitEnabled && emailLogCfg.PerUserMaxCount > 0 {
			if _, err := models.CleanExcessEmailLogsPerRecipient(emailLogCfg.PerUserMaxCount); err != nil {
				log.Printf("[EmailLog] 应用运行时配置后按收件人清理失败: %v", err)
			}
		}
	})
}
