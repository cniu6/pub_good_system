package services

import (
	"fst/backend/app/models"
	"fst/backend/pkg/config"
	"testing"
	"time"
)

func TestApplyGlobalRuntimeConfigUsesSQLAndEnvFallback(t *testing.T) {
	oldSettingsService := GlobalSettingsService
	oldConfig := config.GlobalConfig
	defer func() {
		GlobalSettingsService = oldSettingsService
		config.GlobalConfig = oldConfig
	}()

	config.GlobalConfig = &config.Config{
		AppName:                   "Env App",
		FrontendURL:               "https://env.example.com",
		BackendAPIURL:             "https://api-env.example.com",
		SMTPHost:                  "smtp.env.local",
		SMTPPort:                  "587",
		SMTPUser:                  "env-user@example.com",
		SMTPPass:                  "env-pass",
		SMTPSSL:                   false,
		SystemEmail:               "env-from@example.com",
		SystemEmailName:           "Env Sender",
		RegisterCodeExpireMinutes: 60,
		JWTAccessExpire:           7200,
		JWTRefreshExpire:          604800,
		LoginMaxFailureCount:      5,
		LoginLockDurationMinutes:  10,
		EmailVerifyEnabled:        true,
		SMSVerifyEnabled:          false,
		SMSProvider:               "console",
	}

	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"site_name":                    {Key: "site_name", Value: "DB App"},
			"frontend_url":                 {Key: "frontend_url", Value: "https://db.example.com"},
			"backend_api_url":              {Key: "backend_api_url", Value: "https://api-db.example.com"},
			"smtp_host":                    {Key: "smtp_host", Value: ""},
			"smtp_ssl":                     {Key: "smtp_ssl", Value: "true"},
			"system_email_address":         {Key: "system_email_address", Value: "db-from@example.com"},
			"system_email_name":            {Key: "system_email_name", Value: "DB Sender"},
			"register_code_expire_minutes": {Key: "register_code_expire_minutes", Value: "90"},
			"jwt_access_expire":            {Key: "jwt_access_expire", Value: "1800"},
			"jwt_refresh_expire":           {Key: "jwt_refresh_expire", Value: "3600"},
			"login_max_failure":            {Key: "login_max_failure", Value: "8"},
			"login_lock_duration":          {Key: "login_lock_duration", Value: "15"},
			"email_verify_enabled":         {Key: "email_verify_enabled", Value: "false"},
			"sms_verify_enabled":           {Key: "sms_verify_enabled", Value: "true"},
			"sms_provider":                 {Key: "sms_provider", Value: "aliyun"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	ApplyGlobalRuntimeConfig()

	if config.GlobalConfig.AppName != "DB App" {
		t.Fatalf("AppName = %q, want %q", config.GlobalConfig.AppName, "DB App")
	}
	if config.GlobalConfig.FrontendURL != "https://db.example.com" {
		t.Fatalf("FrontendURL = %q, want %q", config.GlobalConfig.FrontendURL, "https://db.example.com")
	}
	if config.GlobalConfig.BackendAPIURL != "https://api-db.example.com" {
		t.Fatalf("BackendAPIURL = %q, want %q", config.GlobalConfig.BackendAPIURL, "https://api-db.example.com")
	}
	if config.GlobalConfig.SMTPHost != "smtp.env.local" {
		t.Fatalf("SMTPHost = %q, want env fallback %q", config.GlobalConfig.SMTPHost, "smtp.env.local")
	}
	if !config.GlobalConfig.SMTPSSL {
		t.Fatalf("SMTPSSL = %v, want true", config.GlobalConfig.SMTPSSL)
	}
	if config.GlobalConfig.SystemEmail != "db-from@example.com" {
		t.Fatalf("SystemEmail = %q, want %q", config.GlobalConfig.SystemEmail, "db-from@example.com")
	}
	if config.GlobalConfig.SystemEmailName != "DB Sender" {
		t.Fatalf("SystemEmailName = %q, want %q", config.GlobalConfig.SystemEmailName, "DB Sender")
	}
	if config.GlobalConfig.RegisterCodeExpireMinutes != 90 {
		t.Fatalf("RegisterCodeExpireMinutes = %d, want 90", config.GlobalConfig.RegisterCodeExpireMinutes)
	}
	if config.GlobalConfig.JWTAccessExpire != 1800 {
		t.Fatalf("JWTAccessExpire = %d, want 1800", config.GlobalConfig.JWTAccessExpire)
	}
	if config.GlobalConfig.JWTRefreshExpire != 3600 {
		t.Fatalf("JWTRefreshExpire = %d, want 3600", config.GlobalConfig.JWTRefreshExpire)
	}
	if config.GlobalConfig.LoginMaxFailureCount != 8 {
		t.Fatalf("LoginMaxFailureCount = %d, want 8", config.GlobalConfig.LoginMaxFailureCount)
	}
	if config.GlobalConfig.LoginLockDurationMinutes != 15 {
		t.Fatalf("LoginLockDurationMinutes = %d, want 15", config.GlobalConfig.LoginLockDurationMinutes)
	}
	if config.GlobalConfig.EmailVerifyEnabled {
		t.Fatalf("EmailVerifyEnabled = %v, want false", config.GlobalConfig.EmailVerifyEnabled)
	}
	if !config.GlobalConfig.SMSVerifyEnabled {
		t.Fatalf("SMSVerifyEnabled = %v, want true", config.GlobalConfig.SMSVerifyEnabled)
	}
	if config.GlobalConfig.SMSProvider != "aliyun" {
		t.Fatalf("SMSProvider = %q, want %q", config.GlobalConfig.SMSProvider, "aliyun")
	}
}

func TestGetGlobalFrontendURLFallsBackWhenDBValueBlank(t *testing.T) {
	oldSettingsService := GlobalSettingsService
	oldConfig := config.GlobalConfig
	defer func() {
		GlobalSettingsService = oldSettingsService
		config.GlobalConfig = oldConfig
	}()

	config.GlobalConfig = &config.Config{FrontendURL: "https://env.example.com"}
	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"frontend_url": {Key: "frontend_url", Value: "   "},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	if got := GetGlobalFrontendURL(); got != "https://env.example.com" {
		t.Fatalf("GetGlobalFrontendURL() = %q, want %q", got, "https://env.example.com")
	}
}

func TestGetGlobalBackendAPIURLFallsBackWhenDBValueBlank(t *testing.T) {
	oldSettingsService := GlobalSettingsService
	oldConfig := config.GlobalConfig
	defer func() {
		GlobalSettingsService = oldSettingsService
		config.GlobalConfig = oldConfig
	}()

	config.GlobalConfig = &config.Config{BackendAPIURL: "https://api-env.example.com"}
	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"backend_api_url": {Key: "backend_api_url", Value: "   "},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	if got := GetGlobalBackendAPIURL(); got != "https://api-env.example.com" {
		t.Fatalf("GetGlobalBackendAPIURL() = %q, want %q", got, "https://api-env.example.com")
	}
}

func TestGetGlobalPaymentRuntimeHelpers(t *testing.T) {
	oldSettingsService := GlobalSettingsService
	defer func() {
		GlobalSettingsService = oldSettingsService
	}()

	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"payment_enabled":              {Key: "payment_enabled", Value: "true"},
			"payment_order_expire_minutes": {Key: "payment_order_expire_minutes", Value: "45"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	if !GetGlobalPaymentEnabled() {
		t.Fatalf("GetGlobalPaymentEnabled() = false, want true")
	}
	if got := GetGlobalPaymentOrderExpireMinutes(); got != 45 {
		t.Fatalf("GetGlobalPaymentOrderExpireMinutes() = %d, want %d", got, 45)
	}

	GlobalSettingsService.cache["payment_order_expire_minutes"] = &models.SystemSetting{Key: "payment_order_expire_minutes", Value: "0"}
	if got := GetGlobalPaymentOrderExpireMinutes(); got != 30 {
		t.Fatalf("GetGlobalPaymentOrderExpireMinutes() fallback = %d, want %d", got, 30)
	}
	GlobalSettingsService.cache["payment_enabled"] = &models.SystemSetting{Key: "payment_enabled", Value: "false"}
	if GetGlobalPaymentEnabled() {
		t.Fatalf("GetGlobalPaymentEnabled() = true, want false")
	}
}
