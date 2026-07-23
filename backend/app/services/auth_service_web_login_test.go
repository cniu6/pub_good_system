package services

import (
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
)

func setAllowUserLoginForTest(t *testing.T, allow bool) {
	t.Helper()
	old := GlobalSettingsService
	t.Cleanup(func() { GlobalSettingsService = old })
	value := "false"
	if allow {
		value = "true"
	}
	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"allow_user_login": {Key: "allow_user_login", Value: value},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}
}

// TestLogin_BlocksUserWhenLoginDisabled 关闭允许登录后，普通用户登录直接 403（查库前拦截）。
func TestLogin_BlocksUserWhenLoginDisabled(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	setAllowUserLoginForTest(t, false)

	svc := NewAuthService()
	_, svcErr := svc.Login("no-such-user", "whatever", "user", "127.0.0.1")
	if svcErr == nil {
		t.Fatal("禁止用户登录时应拒绝")
	}
	if svcErr.Code != 403 {
		t.Fatalf("期望 403，实际=%d，message=%s", svcErr.Code, svcErr.Message)
	}
}

// TestLogin_AdminNotAffectedByUserLoginDisabled 开关只拦 authGuard=user。
func TestLogin_AdminNotAffectedByUserLoginDisabled(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	setAllowUserLoginForTest(t, false)

	svc := NewAuthService()
	_, svcErr := svc.Login("no-such-admin", "whatever", "admin", "127.0.0.1")
	if svcErr == nil {
		t.Fatal("用户不存在应报错")
	}
	if svcErr.Code == 403 {
		t.Fatalf("管理员登录不应受影响，message=%s", svcErr.Message)
	}
}

// TestLogin_UserLoginAllowedByDefault 默认允许登录。
func TestLogin_UserLoginAllowedByDefault(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	setAllowUserLoginForTest(t, true)

	svc := NewAuthService()
	_, svcErr := svc.Login("no-such-user", "whatever", "user", "127.0.0.1")
	if svcErr == nil {
		t.Fatal("用户不存在应报错")
	}
	if svcErr.Code == 403 {
		t.Fatalf("允许登录时不应 403，message=%s", svcErr.Message)
	}
}
