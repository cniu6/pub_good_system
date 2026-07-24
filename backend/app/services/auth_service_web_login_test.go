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

// TestLogin_AdminNotAffectedByUserLoginDisabled 关闭用户登录时，管理员 guard 仍可认证；
// 同一账号若尝试 user guard，必须仍被用户登录开关拒绝。
func TestLogin_AdminNotAffectedByUserLoginDisabled(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	setAllowUserLoginForTest(t, false)

	const password = "TestPass123!"
	admin := createLoginReadyUser(t, "admin_login_guard", "admin", password)
	svc := NewAuthService()
	result, svcErr := svc.Login(admin.Username, password, "admin", "127.0.0.1")
	if svcErr != nil {
		t.Fatalf("管理员 guard 不应受用户登录开关影响: %v", svcErr.Message)
	}
	if result.AuthGuard != "admin" {
		t.Fatalf("管理员登录 authGuard=%q，期望 admin", result.AuthGuard)
	}

	_, svcErr = svc.Login(admin.Username, password, "user", "127.0.0.1")
	if svcErr == nil || svcErr.Code != 403 {
		t.Fatalf("用户 guard 在开关关闭时应返回 403，实际=%v", svcErr)
	}
}

// TestLogin_UserLoginAllowedWhenEnabled 显式开启 allow_user_login 后，普通用户登录不再因开关被拦。
func TestLogin_UserLoginAllowedWhenEnabled(t *testing.T) {
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

// TestLogin_UserLoginBlockedByDefault 缺配置时默认关闭用户登录。
func TestLogin_UserLoginBlockedByDefault(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	old := GlobalSettingsService
	t.Cleanup(func() { GlobalSettingsService = old })
	GlobalSettingsService = &SettingsService{
		cache:     map[string]*models.SystemSetting{},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	svc := NewAuthService()
	_, svcErr := svc.Login("no-such-user", "whatever", "user", "127.0.0.1")
	if svcErr == nil {
		t.Fatal("默认禁止用户登录时应拒绝")
	}
	if svcErr.Code != 403 {
		t.Fatalf("期望 403，实际=%d，message=%s", svcErr.Code, svcErr.Message)
	}
}
