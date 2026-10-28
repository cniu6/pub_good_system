package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fst/backend/pkg/config"
	"fst/backend/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// TestCorsCredentialsNotLeakedToUntrustedOrigins 覆盖恶意 Origin 与凭证组合，
// 防止“反射任意 Origin + Access-Control-Allow-Credentials:true”导致跨站读取。
func TestCorsCredentialsNotLeakedToUntrustedOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	old := config.CloneGlobalConfig()
	t.Cleanup(func() { config.SetGlobalConfig(old) })

	run := func(origins, origin string) *httptest.ResponseRecorder {
		config.SetGlobalConfig(&config.Config{CorsOrigins: origins})
		r := gin.New()
		r.Use(middleware.CorsMiddleware())
		r.GET("/x", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Cookie", "session=secret")
		r.ServeHTTP(w, req)
		return w
	}

	t.Run("exact allowlist echoes origin with credentials", func(t *testing.T) {
		w := run("https://app.example.com", "https://app.example.com")
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
			t.Fatalf("白名单 Origin 应回显，实际 %q", got)
		}
		if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
			t.Fatalf("白名单 Origin 应允许凭证，实际 %q", got)
		}
	})

	t.Run("evil origin gets no CORS grant", func(t *testing.T) {
		w := run("https://app.example.com", "https://evil.example")
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("恶意 Origin 不应获得 ACAO，实际 %q", got)
		}
		if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
			t.Fatalf("恶意 Origin 不应获得凭证许可，实际 %q", got)
		}
	})

	t.Run("wildcard never carries credentials", func(t *testing.T) {
		w := run("*", "https://evil.example")
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Fatalf("通配配置应返回 ACAO=*，实际 %q", got)
		}
		if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
			t.Fatalf("ACAO=* 时不得同时允许凭证，实际 %q", got)
		}
	})

	t.Run("wildcard subdomain rejects suffix trick", func(t *testing.T) {
		w := run("*.example.com", "https://notexample.com")
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("notexample.com 不应匹配 *.example.com，实际 %q", got)
		}
	})
}
