package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fst/backend/pkg/config"
	"fst/backend/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func TestResolveWSCorsAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		CorsOrigins:   "*",
		WSCorsEnabled: true,
		WSCorsOrigins: "https://app.example.com,https://*.trusted.example",
		FrontendURL:   "https://frontend.example.com",
	}
	config.SetGlobalConfig(cfg)
	t.Cleanup(func() { config.SetGlobalConfig(nil) })

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/ws/presence", nil)
	allowlist := middleware.ResolveWSCorsAllowlist(c)
	if allowed, wildcard := middleware.IsOriginAllowed("https://app.example.com", allowlist); !allowed || wildcard {
		t.Fatalf("精确 WS Origin 应被放行：allowed=%v wildcard=%v", allowed, wildcard)
	}
	if allowed, _ := middleware.IsOriginAllowed("https://foo.trusted.example", allowlist); !allowed {
		t.Fatal("泛域名 WS Origin 应被放行")
	}
	if allowed, _ := middleware.IsOriginAllowed("https://evil.example.com", allowlist); allowed {
		t.Fatal("不在 WS 白名单的 Origin 不应被放行")
	}

	cfg.WSCorsOrigins = ""
	if got := middleware.ResolveWSCorsAllowlist(c); got != "https://frontend.example.com" {
		t.Fatalf("WS 独立白名单为空时应回退 FRONTEND_URL，实际 %q", got)
	}
	cfg.WSCorsEnabled = false
	if got := middleware.ResolveWSCorsAllowlist(c); got != "*" {
		t.Fatalf("WS 独立 CORS 关闭时应沿用全局 CORS，实际 %q", got)
	}
}
