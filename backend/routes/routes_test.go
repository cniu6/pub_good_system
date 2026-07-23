package routes_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fst/backend/internal/testutil"
	"fst/backend/pkg/config"
	"fst/backend/routes"

	"github.com/gin-gonic/gin"
)

func TestSetupRoutes_HealthAndPublic(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	routes.SetupRoutes(r)

	t.Run("health", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json: %v", err)
		}
		if resp["code"] != float64(0) {
			t.Fatalf("code=%v", resp["code"])
		}
	})

	t.Run("ready", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		req.Header.Set("X-Request-Id", "test-ready-rid")
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		var resp map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("json: %v", err)
		}
		if resp["code"] != float64(200) {
			t.Fatalf("code=%v body=%s", resp["code"], w.Body.String())
		}
		data, _ := resp["data"].(map[string]any)
		if data["status"] != "ready" {
			t.Fatalf("data.status=%v", data["status"])
		}
	})

	t.Run("metrics", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
		body := w.Body.String()
		if !strings.Contains(body, "fst_uptime_seconds") {
			t.Fatalf("metrics missing uptime: %s", body)
		}
	})

	t.Run("公开app-config", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/public/app-config", nil)
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("登录缺参应失败", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/public/login", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code == 200 {
			// 业务可能仍 200 + code!=0
			var resp map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["code"] == float64(0) {
				t.Fatalf("空登录不应成功: %s", w.Body.String())
			}
		}
	})

	t.Run("用户接口无token应拒绝", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
		r.ServeHTTP(w, req)
		if w.Code == 200 {
			var resp map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["code"] == float64(0) {
				t.Fatalf("无 token 不应成功: %s", w.Body.String())
			}
		}
	})

	t.Run("管理端无token应拒绝", func(t *testing.T) {
		adminPath := "/admin"
		if config.GlobalConfig != nil && config.GlobalConfig.AdminAPIPath != "" {
			adminPath = config.GlobalConfig.AdminAPIPath
			if !strings.HasPrefix(adminPath, "/") {
				adminPath = "/" + adminPath
			}
		}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1"+adminPath+"/users", nil)
		r.ServeHTTP(w, req)
		if w.Code == 200 {
			var resp map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["code"] == float64(0) {
				t.Fatalf("管理端无 token 不应成功: %s", w.Body.String())
			}
		}
	})
}
