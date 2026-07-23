package services

import (
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/utils"
)

// createLoginReadyUser 创建可密码登录的测试用户（含真实 bcrypt 哈希）。
func createLoginReadyUser(t *testing.T, username, role, password string) *models.User {
	t.Helper()
	hashed, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword 失败: %v", err)
	}
	u := &models.User{
		Username: username,
		Nickname: username,
		Email:    username + "@example.test",
		Password: hashed,
		Role:     role,
		Status:   1,
	}
	if err := models.CreateUser(u); err != nil {
		t.Fatalf("CreateUser(%s) 失败: %v", username, err)
	}
	return u
}

// TestLogin_ReturnsRequestedAuthGuard 登录响应必须显式带回本次签发的 authGuard，
// 前端以此存会话并刷新，不能再用 role 猜测。
func TestLogin_ReturnsRequestedAuthGuard(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	setAllowUserLoginForTest(t, true)

	const password = "TestPass123!"
	user := createLoginReadyUser(t, "login_guard_user", "user", password)
	admin := createLoginReadyUser(t, "login_guard_admin", "admin", password)

	svc := NewAuthService()

	userResult, svcErr := svc.Login(user.Username, password, "user", "127.0.0.1")
	if svcErr != nil {
		t.Fatalf("用户登录应成功: %v", svcErr.Message)
	}
	if userResult.AuthGuard != "user" {
		t.Fatalf("用户登录 authGuard 应为 user，实际=%q", userResult.AuthGuard)
	}

	adminResult, svcErr := svc.Login(admin.Username, password, "admin", "127.0.0.1")
	if svcErr != nil {
		t.Fatalf("管理员登录应成功: %v", svcErr.Message)
	}
	if adminResult.AuthGuard != "admin" {
		t.Fatalf("管理员登录 authGuard 应为 admin，实际=%q", adminResult.AuthGuard)
	}
}
