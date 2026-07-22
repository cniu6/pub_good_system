package services

import (
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
)

// setDisableWebLoginForTest 注入一份仅用于测试的设置缓存，控制「禁止网页端登录」开关，
// 不触发真实 DB 刷新（cacheTime 设为当前时间 + 足够长 ttl，避免 ensureFreshCache 覆盖手工注入的值）。
func setDisableWebLoginForTest(t *testing.T, disabled bool) {
	t.Helper()
	old := GlobalSettingsService
	t.Cleanup(func() { GlobalSettingsService = old })
	value := "false"
	if disabled {
		value = "true"
	}
	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"disable_web_login": {Key: "disable_web_login", Value: value},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}
}

// TestLogin_BlocksUserWebLoginWhenDisabled 开关开启后，普通用户走默认 web 客户端类型登录应被直接拦截
// （在查用户之前就拒绝，不泄露账号是否存在）。
func TestLogin_BlocksUserWebLoginWhenDisabled(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	setDisableWebLoginForTest(t, true)

	svc := NewAuthService()
	_, svcErr := svc.Login("no-such-user", "whatever", "user", "", "127.0.0.1")
	if svcErr == nil {
		t.Fatal("网页端登录已禁止时应拒绝登录")
	}
	if svcErr.Code != 403 {
		t.Fatalf("期望 403，实际=%d，message=%s", svcErr.Code, svcErr.Message)
	}
}

// TestLogin_AllowsAppClientTypeWhenWebLoginDisabled 开关开启后，登录请求带 clientType=app 应绕过网页端限制，
// 正常走到用户查找逻辑（因用户不存在最终返回 401，而不是被网页端开关拦截的 403）。
func TestLogin_AllowsAppClientTypeWhenWebLoginDisabled(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	setDisableWebLoginForTest(t, true)

	svc := NewAuthService()
	_, svcErr := svc.Login("no-such-user", "whatever", "user", "app", "127.0.0.1")
	if svcErr == nil {
		t.Fatal("用户不存在应报错")
	}
	if svcErr.Code == 403 {
		t.Fatalf("clientType=app 应绕过网页端登录限制，不应命中 403，message=%s", svcErr.Message)
	}
}

// TestLogin_AdminNotAffectedByWebLoginDisabled 开关只拦截 authGuard=user，管理员登录不受影响。
func TestLogin_AdminNotAffectedByWebLoginDisabled(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	setDisableWebLoginForTest(t, true)

	svc := NewAuthService()
	_, svcErr := svc.Login("no-such-admin", "whatever", "admin", "", "127.0.0.1")
	if svcErr == nil {
		t.Fatal("用户不存在应报错")
	}
	if svcErr.Code == 403 {
		t.Fatalf("管理员登录不应受网页端登录开关影响，message=%s", svcErr.Message)
	}
}

// TestLogin_WebLoginAllowedByDefault 开关默认关闭（false）时，网页端登录不受限制。
func TestLogin_WebLoginAllowedByDefault(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	setDisableWebLoginForTest(t, false)

	svc := NewAuthService()
	_, svcErr := svc.Login("no-such-user", "whatever", "user", "", "127.0.0.1")
	if svcErr == nil {
		t.Fatal("用户不存在应报错")
	}
	if svcErr.Code == 403 {
		t.Fatalf("开关默认关闭时不应拦截网页端登录，message=%s", svcErr.Message)
	}
}
