package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fst/backend/internal/testutil"
	"fst/backend/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func TestCorsAndAuthMiddleware(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	t.Run("CORS放行预检", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.CorsMiddleware())
		r.GET("/x", func(c *gin.Context) { c.String(200, "ok") })
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/x", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("Access-Control-Request-Method", "GET")
		r.ServeHTTP(w, req)
		if w.Code >= 500 {
			t.Fatalf("CORS 预检失败: %d", w.Code)
		}
	})

	t.Run("无token鉴权失败", func(t *testing.T) {
		r := gin.New()
		r.GET("/need", middleware.AuthMiddleware(), func(c *gin.Context) {
			c.JSON(200, gin.H{"code": 0})
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/need", nil)
		r.ServeHTTP(w, req)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["code"] == float64(0) {
			t.Fatalf("无 token 不应业务成功: %s", w.Body.String())
		}
		if resp["code"] != float64(401) {
			t.Fatalf("期望 401，实际 code=%v body=%s", resp["code"], w.Body.String())
		}
	})

	t.Run("AdminOnly无角色拒绝", func(t *testing.T) {
		r := gin.New()
		r.GET("/adm", middleware.AdminOnly(), func(c *gin.Context) {
			c.JSON(200, gin.H{"code": 0})
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/adm", nil)
		r.ServeHTTP(w, req)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp["code"] == float64(0) {
			t.Fatalf("无管理员上下文不应成功: %s", w.Body.String())
		}
		if resp["code"] != float64(403) {
			t.Fatalf("期望 403，实际 code=%v body=%s", resp["code"], w.Body.String())
		}
	})
}
