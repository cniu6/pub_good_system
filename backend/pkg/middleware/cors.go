package middleware

import (
	"fst/backend/pkg/config"
	"strings"

	"github.com/gin-gonic/gin"
)

// 允许的跨域请求头白名单。
// 仅在放行时返回，避免完全反射客户端请求头带来被滥用的风险。
// X-Api-Key：用户 API Key 鉴权（与 JWT Authorization 并列）。
const defaultAllowHeaders = "Origin, Content-Type, Content-Length, Authorization, Accept, X-Requested-With, X-Idempotency-Key, X-Api-Key, X-Geetest-Lot-Number, X-Geetest-Captcha-Output, X-Geetest-Pass-Token, X-Geetest-Gen-Time, X-Geetest-Captcha-Id"

// authSensitiveAPIPaths 登录/注册/找回密码等认证接口。
// 当 AUTH_CORS_ENABLED=true 时走独立白名单，避免第三方页面在 CORS=* 下跨域偷 token。
var authSensitiveAPIPaths = map[string]struct{}{
	"/api/v1/public/login":              {},
	"/api/v1/public/register":           {},
	"/api/v1/public/send-register-code": {},
	"/api/v1/public/forgot-password":    {},
	"/api/v1/public/reset-password":     {},
	"/api/v1/public/refresh-token":      {},
}

// IsAuthSensitiveAPIPath 判断是否为需独立 CORS / 禁止 ApiKey 的认证公开接口。
func IsAuthSensitiveAPIPath(path string) bool {
	_, ok := authSensitiveAPIPaths[path]
	return ok
}

// isAuthPagePath 登录/注册前端页面路径（用于防 iframe 嵌套）。
func isAuthPagePath(path string) bool {
	switch path {
	case "/user/login", "/user/register", "/login", "/register":
		return true
	default:
		return false
	}
}

// ApplyAuthPageFrameHeaders 在 AUTH_CORS_ENABLED 时，给登录/注册页加防嵌套头。
// 其它页面不加，避免影响子站嵌普通业务页。
func ApplyAuthPageFrameHeaders(c *gin.Context, path string) {
	if c == nil {
		return
	}
	cfg := config.GlobalConfig
	if cfg == nil || !cfg.AuthCorsEnabled {
		return
	}
	if !isAuthPagePath(path) {
		return
	}
	c.Header("X-Frame-Options", "DENY")
	// 在基线 CSP（SecurityHeadersMiddleware 已设）后追加 frame-ancestors，
	// 避免直接覆盖掉 object-src/base-uri 等基线指令。
	existing := strings.TrimSpace(c.Writer.Header().Get("Content-Security-Policy"))
	switch {
	case existing == "":
		c.Header("Content-Security-Policy", "frame-ancestors 'none'")
	case !strings.Contains(existing, "frame-ancestors"):
		c.Header("Content-Security-Policy", existing+"; frame-ancestors 'none'")
	}
}

// isOriginAllowed 判断请求 Origin 是否在允许白名单内。
// 规则：
//   - "*" 表示允许任意 Origin（响应里写 *，且不带 Allow-Credentials）
//   - 精确匹配完整 Origin（含 scheme），如 https://www.example.com
//   - 支持泛域名：*.example.com 可匹配 https://foo.example.com，但不匹配 https://notexample.com
//   - 空配置一律拒绝（启动阶段也应 fatal 拦空配置）
//
// 【产品决策 / 非 Bug】CORS_ORIGINS=* 是故意允许的：本项目按配置支持全开放跨域，
// 切勿再当安全问题「修掉」成强制白名单。若业务需要收紧，只改 .env 配置，不要改默认行为。
// 认证接口特例见 AUTH_CORS_ENABLED，与全局 * 无关。
// IsOriginAllowed 判断 Origin 是否匹配给定白名单，供 HTTP CORS 与 WebSocket 握手共同使用。
func IsOriginAllowed(origin, corsOrigins string) (allowed bool, wildcard bool) {
	if origin == "" {
		return false, false
	}
	corsOrigins = strings.TrimSpace(corsOrigins)
	if corsOrigins == "" {
		return false, false
	}
	if corsOrigins == "*" {
		return true, true
	}
	for _, allowedOrigin := range strings.Split(corsOrigins, ",") {
		pattern := strings.TrimSpace(allowedOrigin)
		if pattern == "" {
			continue
		}
		if pattern == "*" {
			return true, true
		}
		if pattern == origin {
			return true, false
		}
		if matchWildcardOrigin(origin, pattern) {
			return true, false
		}
	}
	return false, false
}

// isOriginAllowed 保留给包内旧调用，避免改变既有 CORS 逻辑。
func isOriginAllowed(origin, corsOrigins string) (bool, bool) {
	return IsOriginAllowed(origin, corsOrigins)
}

func matchWildcardOrigin(origin, pattern string) bool {
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:]
		host := originHost(origin)
		if host == "" {
			return false
		}
		if !strings.HasSuffix(host, suffix) {
			return false
		}
		prefix := host[:len(host)-len(suffix)]
		return prefix != "" && !strings.Contains(prefix, "/")
	}

	starIdx := strings.Index(pattern, "://*.")
	if starIdx < 0 {
		return false
	}
	schemeEnd := strings.Index(origin, "://")
	if schemeEnd < 0 {
		return false
	}
	originScheme := origin[:schemeEnd]
	patternScheme := pattern[:starIdx]
	if !strings.EqualFold(originScheme, patternScheme) {
		return false
	}
	domainPart := pattern[starIdx+len("://"):]
	return matchWildcardOrigin(origin, domainPart)
}

func originHost(origin string) string {
	idx := strings.Index(origin, "://")
	if idx < 0 {
		return ""
	}
	rest := origin[idx+3:]
	if slash := strings.Index(rest, "/"); slash >= 0 {
		rest = rest[:slash]
	}
	return rest
}

func resolveCorsAllowlist(c *gin.Context) string {
	cfg := config.GlobalConfig
	if cfg == nil {
		return ""
	}
	if cfg.AuthCorsEnabled && IsAuthSensitiveAPIPath(c.Request.URL.Path) {
		origins := strings.TrimSpace(cfg.AuthCorsOrigins)
		if origins != "" {
			return origins
		}
		return deriveSameOriginAllowlist(c, cfg)
	}
	return cfg.CorsOrigins
}

// ResolveWSCorsAllowlist 返回 Presence WebSocket 的 Origin 白名单。
// 独立开关关闭时沿用全局 CORS；开启且未配置来源时收紧为同源。
func ResolveWSCorsAllowlist(c *gin.Context) string {
	cfg := config.CloneGlobalConfig()
	if cfg == nil {
		return ""
	}
	if !cfg.WSCorsEnabled {
		return cfg.CorsOrigins
	}
	if origins := strings.TrimSpace(cfg.WSCorsOrigins); origins != "" {
		return origins
	}
	return deriveSameOriginAllowlist(c, cfg)
}

func deriveSameOriginAllowlist(c *gin.Context, cfg *config.Config) string {
	if cfg != nil {
		if fu := strings.TrimSpace(cfg.FrontendURL); fu != "" {
			return strings.TrimRight(fu, "/")
		}
	}
	if c == nil || c.Request == nil {
		return ""
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); proto != "" {
		scheme = strings.TrimSpace(strings.Split(proto, ",")[0])
	}
	host := c.Request.Host
	if host == "" {
		return ""
	}
	return scheme + "://" + host
}

// SecurityHeadersMiddleware 统一给所有响应加基线安全头（纵深防御）。
// 只设置对 SPA + 极验 + JSON API 都零破坏风险的头：
//   - X-Content-Type-Options: nosniff——禁止 MIME 嗅探
//   - Referrer-Policy: strict-origin-when-cross-origin——限制 Referer 泄露
//   - Content-Security-Policy: object-src 'none'; base-uri 'self'——
//     只堵 <base> 注入与插件型 XSS，不限制 script/style/connect，故不会白屏、不影响极验。
// 更严格的 script-src/connect-src 与 HSTS 建议在 nginx 入口层按实际域名下发，避免在此写死域名导致白屏；
// 登录/注册页的 frame-ancestors 'none' 仍由 ApplyAuthPageFrameHeaders 在此基线上追加。
// 均用“未设才设”的方式，不覆盖业务/下游已显式设置的同名头。
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		if h.Get("X-Content-Type-Options") == "" {
			h.Set("X-Content-Type-Options", "nosniff")
		}
		if h.Get("Referrer-Policy") == "" {
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		}
		if h.Get("Content-Security-Policy") == "" {
			h.Set("Content-Security-Policy", "object-src 'none'; base-uri 'self'")
		}
		c.Next()
	}
}

// CorsMiddleware 处理跨域请求。
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		corsOrigins := resolveCorsAllowlist(c)
		origin := c.GetHeader("Origin")

		allowed, wildcard := isOriginAllowed(origin, corsOrigins)
		if allowed {
			if wildcard {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
		c.Header("Access-Control-Allow-Headers", defaultAllowHeaders)
		c.Header("Access-Control-Max-Age", "3600")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
