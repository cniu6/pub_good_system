package services

import (
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/utils"
)

// createTestSessionForRefresh 模拟一次登录：生成 access/refresh token 并落一条会话记录，
// 供刷新流程测试直接复用（不经过 HTTP 层）。
func createTestSessionForRefresh(t *testing.T, userID uint64) (accessToken, refreshToken string) {
	t.Helper()
	ttl := time.Hour
	accessToken, err := utils.GenerateTokenForGuardWithTTL(userID, "user", "user", ttl)
	if err != nil {
		t.Fatalf("生成 access token 失败: %v", err)
	}
	refreshToken, err = utils.GenerateRefreshTokenForGuardWithTTL(userID, "user", ttl)
	if err != nil {
		t.Fatalf("生成 refresh token 失败: %v", err)
	}
	now := time.Now()
	if err := models.CreateUserSession(userID, "user", utils.HashToken(accessToken), utils.HashToken(refreshToken),
		"127.0.0.1", "test-agent", "pc", "web", "", now.Add(ttl).Unix(), now.Add(ttl).Unix()); err != nil {
		t.Fatalf("CreateUserSession 失败: %v", err)
	}
	return accessToken, refreshToken
}

// TestRefreshTokenNormalRotationSucceeds 验证正常刷新：合法 refresh token 能拿到新的一对令牌。
func TestRefreshTokenNormalRotationSucceeds(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "refresh_normal_user")
	_, refreshToken := createTestSessionForRefresh(t, u.ID)

	svc := NewAuthService()
	result, svcErr := svc.RefreshToken(refreshToken, "user", "127.0.0.1", "test-agent", "pc")
	if svcErr != nil {
		t.Fatalf("正常刷新应成功，实际报错: %v", svcErr.Message)
	}
	if result.RefreshToken == "" {
		t.Fatalf("刷新后 refresh token 不能为空")
	}
	// 同秒内 JWT 无 jti 时新旧 token 可能完全一致；只要会话仍按返回值可续期即可。
	active, err := models.IsRefreshSessionActive(u.ID, "user", utils.HashToken(result.RefreshToken))
	if err != nil {
		t.Fatalf("IsRefreshSessionActive: %v", err)
	}
	if !active {
		t.Fatalf("刷新后会话应仍活跃")
	}
	if result.AuthGuard != "user" {
		t.Fatalf("刷新响应 authGuard 应为 user，实际=%q", result.AuthGuard)
	}
}

// TestRefreshTokenReuseRevokesAllSessions 验证刷新令牌重放检测：
// 用已经被轮换过的旧 refresh token 再次刷新时，应被拒绝，且该用户名下全部会话被吊销
// （即便是刚刚轮换出来的新 refresh token，也应随之失效）。
func TestRefreshTokenReuseRevokesAllSessions(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "refresh_reuse_user")
	_, oldRefreshToken := createTestSessionForRefresh(t, u.ID)

	svc := NewAuthService()
	first, svcErr := svc.RefreshToken(oldRefreshToken, "user", "127.0.0.1", "test-agent", "pc")
	if svcErr != nil {
		t.Fatalf("首次刷新应成功: %v", svcErr.Message)
	}
	newRefreshToken := first.RefreshToken

	// 重放旧 refresh token（已失效）
	_, svcErr = svc.RefreshToken(oldRefreshToken, "user", "9.9.9.9", "attacker-agent", "pc")
	if svcErr == nil {
		t.Fatalf("重放旧 refresh token 应被拒绝")
	}
	if svcErr.Code != 401 {
		t.Fatalf("期望 401，实际=%d", svcErr.Code)
	}

	// 因重放检测触发了全量吊销，刚才轮换出来的新 refresh token 也应随之失效
	_, svcErr = svc.RefreshToken(newRefreshToken, "user", "127.0.0.1", "test-agent", "pc")
	if svcErr == nil {
		t.Fatalf("重放检测应吊销全部会话，新 refresh token 也应失效")
	}

	active, err := models.IsRefreshSessionActive(u.ID, "user", utils.HashToken(newRefreshToken))
	if err != nil {
		t.Fatalf("IsRefreshSessionActive: %v", err)
	}
	if active {
		t.Fatalf("重放检测后该用户所有会话应已被吊销")
	}
}
