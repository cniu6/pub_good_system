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

// GetAdminAuditUser 从 gin.Context 中同时提取管理员用户ID与用户名，
// 用于写操作审计日志。
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

// GetUserID 取出 AuthMiddleware 写入的 userID（写入类型固定为 uint64，见 pkg/middleware/auth.go）。
// ok=false 表示未登录或中间件未生效。
func GetUserID(c *gin.Context) (uint64, bool) {
	id, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	uid, ok := id.(uint64)
	return uid, ok
}

// RequireUserID 取当前登录用户 ID；未登录/无效时已写好 401 响应，调用方直接 return。
// 统一用户侧 controller 重复的 GetUserID + Fail(401) 样板。
func RequireUserID(c *gin.Context) (uint64, bool) {
	uid, ok := GetUserID(c)
	if !ok || uid == 0 {
		Fail(c, 401, "User not logged in")
		return 0, false
	}
	return uid, true
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
