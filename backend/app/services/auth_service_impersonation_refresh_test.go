package services

import (
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/utils"
)

// createSessionForRefreshTest 模拟登录后写入一条 user_sessions 会话记录，返回 accessToken/refreshToken，
// 供下面两个测试直接调用 RefreshToken 验证续期后的 TTL。
func createSessionForRefreshTest(t *testing.T, userID uint64, device string, accessTTL, refreshTTL time.Duration) (accessToken, refreshToken string) {
	t.Helper()

	var err error
	accessToken, err = utils.GenerateTokenForGuardWithTTL(userID, "user", "user", accessTTL)
	if err != nil {
		t.Fatalf("生成 access token 失败: %v", err)
	}
	refreshToken, err = utils.GenerateRefreshTokenForGuardWithTTL(userID, "user", refreshTTL)
	if err != nil {
		t.Fatalf("生成 refresh token 失败: %v", err)
	}

	now := time.Now().Unix()
	expiresAt := now + int64(accessTTL.Seconds())
	refreshExpiresAt := now + int64(refreshTTL.Seconds())
	if err := models.CreateUserSession(
		userID, "user", utils.HashToken(accessToken), utils.HashToken(refreshToken),
		"127.0.0.1", "test-agent", device, "web", "", expiresAt, refreshExpiresAt,
	); err != nil {
		t.Fatalf("创建测试会话失败: %v", err)
	}
	return accessToken, refreshToken
}

// TestRefreshToken_ImpersonationSessionKeepsShortTTL 回归测试：管理员代登录(login-as)会话
// 执行 refresh 续期后，TTL 仍被限制在短时上限内（access<=15min, refresh<=30min），不会被拉回
// 运行时默认的长 TTL（默认 access 2h / refresh 7d）。修复前这里会直接签发默认长 TTL，代登的
// 短时限制形同虚设（见 CLAUDE.md 已知问题修复记录）。
// 同时验证：客户端传入的 device 字段不能覆盖代登标记，第二次 refresh 仍保持短时上限。
func TestRefreshToken_ImpersonationSessionKeepsShortTTL(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	user := testutil.CreateTestUser(t, "impersonated-user")

	// 模拟 LoginToUser 签发的短时代登会话（15分钟/30分钟）
	_, refreshToken := createSessionForRefreshTest(t, user.ID, models.ImpersonationDeviceLabel, 15*time.Minute, 30*time.Minute)

	svc := NewAuthService()
	// 故意传入普通 device，模拟前端 refresh 时覆盖 device；服务端必须保留代登标记
	result, svcErr := svc.RefreshToken(refreshToken, "user", "127.0.0.1", "test-agent", "Chrome on Windows")
	if svcErr != nil {
		t.Fatalf("刷新代登会话失败: code=%d message=%s", svcErr.Code, svcErr.Message)
	}

	assertImpersonationShortTTL(t, result)

	// 第二次 refresh：若 device 标记在第一次轮换时被覆盖，此处会退回正常长 TTL
	result2, svcErr := svc.RefreshToken(result.RefreshToken, "user", "127.0.0.1", "test-agent", "Another Device")
	if svcErr != nil {
		t.Fatalf("第二次刷新代登会话失败: code=%d message=%s", svcErr.Code, svcErr.Message)
	}
	assertImpersonationShortTTL(t, result2)
}

func assertImpersonationShortTTL(t *testing.T, result *LoginResult) {
	t.Helper()
	now := time.Now().Unix()
	gotAccessTTL := time.Duration(result.ExpiresAt-now) * time.Second
	gotRefreshTTL := time.Duration(result.RefreshExpiresAt-now) * time.Second

	// 允许 1 分钟误差（测试执行耗时）
	if gotAccessTTL > ImpersonationAccessTTLCap+time.Minute {
		t.Fatalf("代登会话 refresh 后 access TTL 超出短时上限: got=%s cap=%s", gotAccessTTL, ImpersonationAccessTTLCap)
	}
	if gotRefreshTTL > ImpersonationRefreshTTLCap+time.Minute {
		t.Fatalf("代登会话 refresh 后 refresh TTL 超出短时上限: got=%s cap=%s", gotRefreshTTL, ImpersonationRefreshTTLCap)
	}
}

// TestRefreshToken_NormalSessionUsesRuntimeDefaultTTL 对照组：普通登录会话（非代登）refresh
// 续期后应使用运行时默认的长 TTL，确认修复没有误伤正常会话的续期逻辑。
func TestRefreshToken_NormalSessionUsesRuntimeDefaultTTL(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	user := testutil.CreateTestUser(t, "normal-user")

	// 模拟一次正常登录写入的会话（非代登，device 为普通设备信息）
	_, refreshToken := createSessionForRefreshTest(t, user.ID, "Chrome on Windows", 2*time.Hour, 7*24*time.Hour)

	svc := NewAuthService()
	result, svcErr := svc.RefreshToken(refreshToken, "user", "127.0.0.1", "test-agent", "Chrome on Windows")
	if svcErr != nil {
		t.Fatalf("刷新普通会话失败: code=%d message=%s", svcErr.Code, svcErr.Message)
	}

	now := time.Now().Unix()
	gotAccessTTL := time.Duration(result.ExpiresAt-now) * time.Second
	gotRefreshTTL := time.Duration(result.RefreshExpiresAt-now) * time.Second

	// 运行时默认 TTL（测试环境未配置 JWT_ACCESS_EXPIRE/JWT_REFRESH_EXPIRE，走默认值 7200s/604800s）
	if gotAccessTTL <= ImpersonationAccessTTLCap {
		t.Fatalf("普通会话 refresh 后 access TTL 不应被代登上限误伤: got=%s", gotAccessTTL)
	}
	if gotRefreshTTL <= ImpersonationRefreshTTLCap {
		t.Fatalf("普通会话 refresh 后 refresh TTL 不应被代登上限误伤: got=%s", gotRefreshTTL)
	}
}
