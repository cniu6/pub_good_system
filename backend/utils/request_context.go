package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ResolveRequestLang 统一解析请求语言，避免各控制器重复维护相同逻辑。
func ResolveRequestLang(c *gin.Context, reqLang, defaultLang string) string {
	lang := strings.TrimSpace(reqLang)
	if lang == "" && c != nil {
		lang = strings.TrimSpace(c.GetHeader("Accept-Language"))
	}

	defaultLang = strings.TrimSpace(defaultLang)
	if defaultLang == "" {
		defaultLang = "en-US"
	}
	if lang == "" {
		lang = defaultLang
	}

	if strings.Contains(strings.ToLower(lang), "zh") {
		return "zh-CN"
	}
	return "en-US"
}

// GetUserID 从 gin.Context 里取出 AuthMiddleware 写入的 userID。
// 中间件里写入的是 uint64（见 pkg/middleware/auth.go），这里额外兼容 int64/float64/int 只是防御性写法
// （历史上多个 controller 各自写了一份几乎一样的 type switch，这里统一收敛成一个函数）。
// exists=false 表示 context 里没有 userID 或类型对不上（未登录 / 中间件未生效）。
// GetAdminAuditUser 从 gin.Context 中同时提取管理员用户ID与用户名，
// 用于写操作审计日志。与 GetUserID 配套，避免各管理端 controller 重复实现。
func GetAdminAuditUser(c *gin.Context) (uint64, string) {
	uid, _ := GetUserID(c)
	var name string
	if v, ok := c.Get("username"); ok {
		if s, ok2 := v.(string); ok2 {
			name = s
		}
	}
	return uid, name
}

func GetUserID(c *gin.Context) (uint64, bool) {
	v, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	switch id := v.(type) {
	case uint64:
		return id, true
	case int64:
		return uint64(id), true
	case float64:
		return uint64(id), true
	case int:
		return uint64(id), true
	default:
		return 0, false
	}
}

// GetClientIP 统一获取客户端 IP。
// 当前项目已显式关闭 TrustedProxies，因此这里不再手动信任可伪造的转发头。
func GetClientIP(c *gin.Context) string {
	if c == nil {
		return "unknown"
	}

	clientIP := strings.TrimSpace(c.ClientIP())
	if clientIP == "" {
		return "unknown"
	}
	return ClampStoredIP(clientIP)
}

// ExtractBearerToken 从 Authorization 头中提取 Bearer Token。
func ExtractBearerToken(authHeader string) string {
	authHeader = strings.TrimSpace(authHeader)
	if len(authHeader) >= 7 && strings.EqualFold(authHeader[:7], "Bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return authHeader
}

// ParseDeviceFromUserAgent 根据 User-Agent 粗略识别设备类型。
func ParseDeviceFromUserAgent(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "iphone"):
		return "iPhone"
	case strings.Contains(ua, "ipad"):
		return "iPad"
	case strings.Contains(ua, "android") && strings.Contains(ua, "mobile"):
		return "Android Phone"
	case strings.Contains(ua, "android"):
		return "Android Tablet"
	case strings.Contains(ua, "macintosh"):
		return "Mac"
	case strings.Contains(ua, "windows"):
		return "Windows PC"
	case strings.Contains(ua, "linux"):
		return "Linux PC"
	default:
		return "Unknown"
	}
}
