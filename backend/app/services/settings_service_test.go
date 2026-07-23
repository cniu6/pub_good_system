package services

import (
	"fst/backend/app/models"
	"fst/backend/pkg/config"
	"testing"
	"time"
)

func TestApplyGlobalRuntimeConfigUsesSQLAndEnvFallback(t *testing.T) {
	oldSettingsService := GlobalSettingsService
	oldConfig := config.CloneGlobalConfig()
	defer func() {
		GlobalSettingsService = oldSettingsService
		config.SetGlobalConfig(oldConfig)
	}()

	config.SetGlobalConfig(&config.Config{
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
	})

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

	got := config.CloneGlobalConfig()
	if got == nil {
		t.Fatal("GlobalConfig is nil after ApplyGlobalRuntimeConfig")
	}
	if got.AppName != "DB App" {
		t.Fatalf("AppName = %q, want %q", got.AppName, "DB App")
	}
	if got.FrontendURL != "https://db.example.com" {
		t.Fatalf("FrontendURL = %q, want %q", got.FrontendURL, "https://db.example.com")
	}
	if got.BackendAPIURL != "https://api-db.example.com" {
		t.Fatalf("BackendAPIURL = %q, want %q", got.BackendAPIURL, "https://api-db.example.com")
	}
	if got.SMTPHost != "smtp.env.local" {
		t.Fatalf("SMTPHost = %q, want env fallback %q", got.SMTPHost, "smtp.env.local")
	}
	if !got.SMTPSSL {
		t.Fatalf("SMTPSSL = %v, want true", got.SMTPSSL)
	}
	if got.SystemEmail != "db-from@example.com" {
		t.Fatalf("SystemEmail = %q, want %q", got.SystemEmail, "db-from@example.com")
	}
	if got.SystemEmailName != "DB Sender" {
		t.Fatalf("SystemEmailName = %q, want %q", got.SystemEmailName, "DB Sender")
	}
	if got.RegisterCodeExpireMinutes != 90 {
		t.Fatalf("RegisterCodeExpireMinutes = %d, want 90", got.RegisterCodeExpireMinutes)
	}
	if got.JWTAccessExpire != 1800 {
		t.Fatalf("JWTAccessExpire = %d, want 1800", got.JWTAccessExpire)
	}
	if got.JWTRefreshExpire != 3600 {
		t.Fatalf("JWTRefreshExpire = %d, want 3600", got.JWTRefreshExpire)
	}
	if got.LoginMaxFailureCount != 8 {
		t.Fatalf("LoginMaxFailureCount = %d, want 8", got.LoginMaxFailureCount)
	}
	if got.LoginLockDurationMinutes != 15 {
		t.Fatalf("LoginLockDurationMinutes = %d, want 15", got.LoginLockDurationMinutes)
	}
	if got.EmailVerifyEnabled {
		t.Fatalf("EmailVerifyEnabled = %v, want false", got.EmailVerifyEnabled)
	}
	if !got.SMSVerifyEnabled {
		t.Fatalf("SMSVerifyEnabled = %v, want true", got.SMSVerifyEnabled)
	}
	if got.SMSProvider != "aliyun" {
		t.Fatalf("SMSProvider = %q, want %q", got.SMSProvider, "aliyun")
	}
}

func TestGetGlobalFrontendURLFallsBackWhenDBValueBlank(t *testing.T) {
	oldSettingsService := GlobalSettingsService
	oldConfig := config.CloneGlobalConfig()
	defer func() {
		GlobalSettingsService = oldSettingsService
		config.SetGlobalConfig(oldConfig)
	}()

	config.SetGlobalConfig(&config.Config{FrontendURL: "https://env.example.com"})
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
	oldConfig := config.CloneGlobalConfig()
	defer func() {
		GlobalSettingsService = oldSettingsService
		config.SetGlobalConfig(oldConfig)
	}()

	config.SetGlobalConfig(&config.Config{BackendAPIURL: "https://api-env.example.com"})
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

func TestGetGlobalAPILogRuntimeConfigNormalizesCleanupInterval(t *testing.T) {
	oldSettingsService := GlobalSettingsService
	defer func() {
		GlobalSettingsService = oldSettingsService
	}()

	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"api_log_max_count":                {Key: "api_log_max_count", Value: "5000"},
			"api_log_cleanup_interval_seconds": {Key: "api_log_cleanup_interval_seconds", Value: "5"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	if got := GetGlobalAPILogRuntimeConfig().CleanupIntervalSeconds; got != 600 {
		t.Fatalf("CleanupIntervalSeconds = %d, want fallback 600", got)
	}

	GlobalSettingsService.cache["api_log_cleanup_interval_seconds"] = &models.SystemSetting{
		Key: "api_log_cleanup_interval_seconds", Value: "1200",
	}
	if got := GetGlobalAPILogRuntimeConfig().CleanupIntervalSeconds; got != 1200 {
		t.Fatalf("CleanupIntervalSeconds = %d, want 1200", got)
	}
}
