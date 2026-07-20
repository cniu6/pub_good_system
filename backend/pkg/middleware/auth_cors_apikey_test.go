package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/config"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

func TestAuthCorsAndAPIKey(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	t.Run("AUTH_CORS关闭时登录接口仍随全局星号放行", func(t *testing.T) {
		cfg := config.GlobalConfig
		cfg.CorsOrigins = "*"
		cfg.AuthCorsEnabled = false
		cfg.AuthCorsOrigins = ""

		r := gin.New()
		r.Use(middleware.CorsMiddleware())
		r.POST("/api/v1/public/login", func(c *gin.Context) { c.String(200, "ok") })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/public/login", nil)
		req.Header.Set("Origin", "https://evil.example.com")
		r.ServeHTTP(w, req)
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Fatalf("期望 ACAO=*，实际 %q", got)
		}
	})

	t.Run("AUTH_CORS开启时登录接口拒绝第三方Origin", func(t *testing.T) {
		cfg := config.GlobalConfig
		cfg.CorsOrigins = "*"
		cfg.AuthCorsEnabled = true
		cfg.AuthCorsOrigins = "https://app.example.com"

		r := gin.New()
		r.Use(middleware.CorsMiddleware())
		r.POST("/api/v1/public/login", func(c *gin.Context) { c.String(200, "ok") })
		r.GET("/api/v1/public/app-config", func(c *gin.Context) { c.String(200, "ok") })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/public/login", nil)
		req.Header.Set("Origin", "https://evil.example.com")
		r.ServeHTTP(w, req)
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("第三方 Origin 不应有 ACAO，实际 %q", got)
		}

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/public/login", nil)
		req2.Header.Set("Origin", "https://app.example.com")
		r.ServeHTTP(w2, req2)
		if got := w2.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
			t.Fatalf("白名单 Origin 期望回显，实际 %q", got)
		}

		// 非认证路径仍走全局 *
		w3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodGet, "/api/v1/public/app-config", nil)
		req3.Header.Set("Origin", "https://evil.example.com")
		r.ServeHTTP(w3, req3)
		if got := w3.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Fatalf("非认证路径应仍为 *，实际 %q", got)
		}
	})

	t.Run("总开关默认关闭时ApiKey鉴权被拒绝", func(t *testing.T) {
		u := testutil.CreateTestUser(t, "apikey_user_default_off")
		key, err := models.ResetUserApiKey(u.ID)
		if err != nil {
			t.Fatalf("ResetUserApiKey: %v", err)
		}

		r := gin.New()
		r.GET("/need", middleware.AuthMiddleware(), func(c *gin.Context) {
			c.JSON(200, gin.H{"code": 0})
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/need", nil)
		req.Header.Set("X-Api-Key", key)
		r.ServeHTTP(w, req)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["code"] != float64(403) {
			t.Fatalf("默认关闭时期望 403，实际 %v body=%s", resp["code"], w.Body.String())
		}
	})

	t.Run("X-Api-Key鉴权成功并写入authMethod", func(t *testing.T) {
		u := testutil.CreateTestUser(t, "apikey_user_1")
		key, err := models.ResetUserApiKey(u.ID)
		if err != nil {
			t.Fatalf("ResetUserApiKey: %v", err)
		}
		// 主动开启总开关后才能走 X-Api-Key 鉴权
		if err := models.UpdateSetting("api_key_auth_enabled", "true"); err != nil {
			t.Fatalf("开启api_key_auth_enabled失败: %v", err)
		}

		r := gin.New()
		r.GET("/need", middleware.AuthMiddlewareForGuard(utils.UserAuthGuard, utils.AdminAuthGuard), func(c *gin.Context) {
			am, _ := c.Get("authMethod")
			uid, _ := c.Get("userID")
			c.JSON(200, gin.H{"code": 0, "authMethod": am, "userID": uid})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/need", nil)
		req.Header.Set("X-Api-Key", key)
		r.ServeHTTP(w, req)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["code"] != float64(0) {
			t.Fatalf("ApiKey 鉴权应成功: %s", w.Body.String())
		}
		if resp["authMethod"] != middleware.AuthMethodAPIKey {
			t.Fatalf("authMethod 期望 apikey，实际 %v", resp["authMethod"])
		}
	})

	t.Run("错误ApiKey拒绝", func(t *testing.T) {
		r := gin.New()
		r.GET("/need", middleware.AuthMiddleware(), func(c *gin.Context) {
			c.JSON(200, gin.H{"code": 0})
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/need", nil)
		req.Header.Set("X-Api-Key", "not-a-real-key")
		r.ServeHTTP(w, req)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["code"] != float64(401) {
			t.Fatalf("期望 401，实际 %v body=%s", resp["code"], w.Body.String())
		}
	})

	t.Run("认证路径拒绝携带ApiKey", func(t *testing.T) {
		r := gin.New()
		r.POST("/api/v1/public/login", middleware.RejectApiKeyOnAuthPaths(), func(c *gin.Context) {
			c.JSON(200, gin.H{"code": 0})
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/public/login", nil)
		req.Header.Set("X-Api-Key", "whatever")
		r.ServeHTTP(w, req)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["code"] != float64(403) {
			t.Fatalf("期望 403，实际 %v body=%s", resp["code"], w.Body.String())
		}
	})
}
