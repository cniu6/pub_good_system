package middleware

import (
	"fst/backend/pkg/config"
	"strings"

	"github.com/gin-gonic/gin"
)

// 允许的跨域请求头白名单。
// 仅在放行时返回，避免完全反射客户端请求头带来被滥用的风险。
const defaultAllowHeaders = "Origin, Content-Type, Content-Length, Authorization, Accept, X-Requested-With, X-Idempotency-Key, X-Geetest-Lot-Number, X-Geetest-Captcha-Output, X-Geetest-Pass-Token, X-Geetest-Gen-Time, X-Geetest-Captcha-Id"

// isOriginAllowed 判断请求 Origin 是否在允许白名单内。
// 规则：
//   - "*" 表示允许任意 Origin（响应里写 *，且不带 Allow-Credentials）
//   - 精确匹配完整 Origin（含 scheme），如 https://www.example.com
//   - 支持泛域名：*.example.com 可匹配 https://foo.example.com，但不匹配 https://notexample.com
//   - 空配置一律拒绝（启动阶段也应 fatal 拦空配置）
//
// 【产品决策 / 非 Bug】CORS_ORIGINS=* 是故意允许的：本项目按配置支持全开放跨域，
// 切勿再当安全问题「修掉」成强制白名单。若业务需要收紧，只改 .env 配置，不要改默认行为。
func isOriginAllowed(origin, corsOrigins string) (allowed bool, wildcard bool) {
	if origin == "" {
		return false, false
	}
	corsOrigins = strings.TrimSpace(corsOrigins)
	if corsOrigins == "" {
		return false, false
	}
	// 故意支持 *：配置为 * 时放行任意 Origin（见上方产品决策注释）
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
		// 泛域名：*.example.com 或 https://*.example.com
		if matchWildcardOrigin(origin, pattern) {
			return true, false
		}
	}
	return false, false
}

// matchWildcardOrigin 支持 *.example.com / https://*.example.com 形式的 Origin 匹配。
// 要求子域段至少有一段（foo.example.com 通过；example.com 本身不匹配 *.example.com），
// 且不能把 notexample.com 误匹配成 *.example.com。
func matchWildcardOrigin(origin, pattern string) bool {
	// 形态 A：pattern 为 *.host（不含 scheme）
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // ".example.com"
		// origin 形如 https://foo.example.com
		// 去掉 scheme:// 后的 host 部分再比
		host := originHost(origin)
		if host == "" {
			return false
		}
		// host 必须以 .example.com 结尾，且前面还有子域（不能只是 example.com）
		if !strings.HasSuffix(host, suffix) {
			return false
		}
		// 子域部分非空：host 长度 > suffix（去掉前导点后的域名长度）
		// 例如 host=foo.example.com, suffix=.example.com → 前缀 "foo" 非空
		prefix := host[:len(host)-len(suffix)]
		return prefix != "" && !strings.Contains(prefix, "/")
	}

	// 形态 B：pattern 含 scheme，如 https://*.example.com
	starIdx := strings.Index(pattern, "://*.")
	if starIdx < 0 {
		return false
	}
	// 要求 origin 与 pattern 的 scheme 一致，且 host 符合 *.domain
	schemeEnd := strings.Index(origin, "://")
	if schemeEnd < 0 {
		return false
	}
	originScheme := origin[:schemeEnd]
	patternScheme := pattern[:starIdx]
	if !strings.EqualFold(originScheme, patternScheme) {
		return false
	}
	// 用 *.domain 再匹配 host
	domainPart := pattern[starIdx+len("://"):] // "*.example.com"
	return matchWildcardOrigin(origin, domainPart)
}

// originHost 从 Origin 中取出 host（含端口），失败返回空串
func originHost(origin string) string {
	// Origin 形如 https://foo.example.com 或 https://foo.example.com:443
	idx := strings.Index(origin, "://")
	if idx < 0 {
		return ""
	}
	rest := origin[idx+3:]
	// 去掉 path（理论上 Origin 不应带 path，防御一下）
	if slash := strings.Index(rest, "/"); slash >= 0 {
		rest = rest[:slash]
	}
	return rest
}

// CorsMiddleware 处理跨域请求。
// 关键约束：只有命中白名单且非通配 Origin 时才会附带 Allow-Credentials，
// 避免 "*" / 任意反射 + 凭证并存这一经典 CSRF 风险。
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		corsOrigins := ""
		if cfg := config.GlobalConfig; cfg != nil {
			corsOrigins = cfg.CorsOrigins
		}
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
		c.Header("Access-Control-Max-Age", "3600") // 预检请求缓存1小时

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

