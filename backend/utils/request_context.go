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
	return clientIP
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
