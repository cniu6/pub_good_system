package routes_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/routes"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

func TestAdminRoutes_RequireAuthViaSetupRoutes(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	routes.SetupRoutes(r)

	paths := []string{
		"/api/v1/admin/users",
		"/api/v1/admin/dashboard",
		"/api/v1/admin/settings",
		"/api/v1/admin/logs",
		"/api/v1/admin/api-logs",
		"/api/v1/admin/payments",
		"/api/v1/admin/withdraws",
	}
	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			r.ServeHTTP(w, req)
			var resp map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if w.Code == 200 && resp["code"] == float64(0) {
				t.Fatalf("无 token 访问 %s 不应成功: %s", path, w.Body.String())
			}
		})
	}
}

func TestAdminRoutes_UserTokenRejected(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	u := testutil.CreateTestUser(t, "admin_route_regular_user")
	accessToken, err := utils.GenerateTokenForGuardWithTTL(u.ID, "user", utils.UserAuthGuard, time.Minute)
	if err != nil {
		t.Fatalf("生成用户 token 失败: %v", err)
	}
	if err := models.CreateUserSession(u.ID, utils.UserAuthGuard, utils.HashToken(accessToken), utils.HashToken("refresh-token"), "127.0.0.1", "test", "pc", "web", "", time.Now().Add(time.Hour).Unix(), time.Now().Add(2*time.Hour).Unix()); err != nil {
		t.Fatalf("创建用户会话失败: %v", err)
	}

	r := gin.New()
	routes.SetupRoutes(r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
		t.Fatalf("普通用户 token 访问 admin 路由应为 401/403，实际 HTTP %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminRoutes_AdminTokenAccepted(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	admin := testutil.CreateTestAdmin(t, "admin_route_admin")
	accessToken, err := utils.GenerateTokenForGuardWithTTL(admin.ID, "admin", utils.AdminAuthGuard, time.Minute)
	if err != nil {
		t.Fatalf("生成管理员 token 失败: %v", err)
	}
	if err := models.CreateUserSession(admin.ID, utils.AdminAuthGuard, utils.HashToken(accessToken), utils.HashToken("refresh-token"), "127.0.0.1", "test", "pc", "web", "", time.Now().Add(time.Hour).Unix(), time.Now().Add(2*time.Hour).Unix()); err != nil {
		t.Fatalf("创建管理员会话失败: %v", err)
	}

	r := gin.New()
	routes.SetupRoutes(r)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	r.ServeHTTP(w, req)

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if w.Code != http.StatusOK || resp["code"] != float64(http.StatusOK) {
		t.Fatalf("管理员 token 应可访问 admin 路由，HTTP %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminRoutes_BatchUserActions(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	admin := testutil.CreateTestAdmin(t, "admin_batch_admin")
	user1 := testutil.CreateTestUser(t, "admin_batch_user_1")
	user2 := testutil.CreateTestUser(t, "admin_batch_user_2")

	accessToken, err := utils.GenerateTokenForGuardWithTTL(admin.ID, "admin", utils.AdminAuthGuard, time.Minute)
	if err != nil {
		t.Fatalf("生成管理员 token 失败: %v", err)
	}
	if err := models.CreateUserSession(admin.ID, utils.AdminAuthGuard, utils.HashToken(accessToken), utils.HashToken("refresh-token"), "127.0.0.1", "test", "pc", "web", "", time.Now().Add(time.Hour).Unix(), time.Now().Add(2*time.Hour).Unix()); err != nil {
		t.Fatalf("创建管理员会话失败: %v", err)
	}

	r := gin.New()
	routes.SetupRoutes(r)

	statusBody := fmt.Sprintf(`{"ids":[%d,%d],"status":0}`, user1.ID, user2.ID)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/batch-status", strings.NewReader(statusBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	r.ServeHTTP(w, req)

	var statusResp struct {
		Code int `json:"code"`
		Data struct {
			Succeeded int `json:"succeeded"`
			Failed    int `json:"failed"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("批量状态响应不是 JSON: %v", err)
	}
	if w.Code != http.StatusOK || statusResp.Code != http.StatusOK || statusResp.Data.Succeeded != 2 || statusResp.Data.Failed != 0 {
		t.Fatalf("批量禁用失败，HTTP %d: %s", w.Code, w.Body.String())
	}
	for _, id := range []uint64{user1.ID, user2.ID} {
		u, err := models.GetUserByID(id)
		if err != nil || u.Status != 0 {
			t.Fatalf("用户 %d 应已被禁用: status=%v err=%v", id, u, err)
		}
	}

	deleteBody := fmt.Sprintf(`{"ids":[%d]}`, user1.ID)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/batch-delete", strings.NewReader(deleteBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	r.ServeHTTP(w, req)

	var deleteResp struct {
		Code int `json:"code"`
		Data struct {
			Succeeded int `json:"succeeded"`
			Failed    int `json:"failed"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &deleteResp); err != nil {
		t.Fatalf("批量删除响应不是 JSON: %v", err)
	}
	if w.Code != http.StatusOK || deleteResp.Code != http.StatusOK || deleteResp.Data.Succeeded != 1 || deleteResp.Data.Failed != 0 {
		t.Fatalf("批量删除失败，HTTP %d: %s", w.Code, w.Body.String())
	}
}
