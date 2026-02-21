package utils

import (
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// Clean_XSS 清理用户输入，防止XSS攻击和SQL注入
func Clean_XSS(input string) string {
	if input == "" {
		return input
	}

	// 先去除首尾空白
	input = strings.TrimSpace(input)

	// HTML转义
	input = html.EscapeString(input)

	// SQL注入防护 - 移除或转义危险的SQL关键字和字符
	sqlDangerousPatterns := []struct {
		pattern string
		replace string
	}{
		// SQL注入关键字（不区分大小写）
		{`(?i)\b(union|select|insert|update|delete|drop|create|alter|exec|execute|sp_|xp_)\b`, "[FILTERED]"},
		// SQL注释
		{`--.*$`, ""},
		{`/\*.*?\*/`, ""},
		// 危险字符组合
		{`['";]`, ""},
		// 十六进制编码
		{`0x[0-9a-fA-F]+`, "[HEX]"},
	}

	for _, pattern := range sqlDangerousPatterns {
		re := regexp.MustCompile(pattern.pattern)
		input = re.ReplaceAllString(input, pattern.replace)
	}

	// XSS防护 - 移除危险的HTML/JS标签和属性
	xssDangerousPatterns := []struct {
		pattern string
		replace string
	}{
		// 脚本标签
		{`(?i)<script[^>]*>.*?</script>`, "[SCRIPT_REMOVED]"},
		{`(?i)<script[^>]*>`, "[SCRIPT_REMOVED]"},
		// 危险属性
		{`(?i)\s*on\w+\s*=`, " data-removed="},
		// JavaScript协议
		{`(?i)javascript:`, "data:"},
		{`(?i)vbscript:`, "data:"},
		// 数据协议中的危险内容
		{`(?i)data:.*?base64`, "data:text/plain"},
		// iframe和object标签
		{`(?i)<(iframe|object|embed|applet)[^>]*>`, "[OBJECT_REMOVED]"},
		// style标签中的expression
		{`(?i)expression\s*\(`, "[EXPRESSION_REMOVED]("},
	}

	for _, pattern := range xssDangerousPatterns {
		re := regexp.MustCompile(pattern.pattern)
		input = re.ReplaceAllString(input, pattern.replace)
	}

	// 移除控制字符（除了常见的空白字符）
	controlChars := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	input = controlChars.ReplaceAllString(input, "")

	// 最终清理多余空白
	input = strings.TrimSpace(input)

	return input
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(email)
}

// IsEmail 判断字符串是否为邮箱格式
func IsEmail(str string) bool {
	return ValidateEmail(str)
}

// IsDigit 判断字符串是否只包含数字
func IsDigit(str string) bool {
	if str == "" {
		return false
	}
	for _, c := range str {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// ValidatePort 检测端口合法性(返回true表示合法，false表示不合法)
func ValidatePort(port int) bool {
	return port >= 1 && port <= 65535
}

// SanitizeQueryParams 清理查询参数中的 "null"/"undefined" 字符串值
// 前端可能将 JavaScript 的 null/undefined 序列化为字符串 "null"/"undefined"
// 导致后端 ParseUint 等解析失败
func SanitizeQueryParams(ctx *gin.Context) {
	rawQuery := ctx.Request.URL.RawQuery
	if rawQuery == "" {
		return
	}

	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return
	}

	changed := false
	for key, vals := range values {
		for _, v := range vals {
			if v == "null" || v == "undefined" {
				values.Del(key)
				changed = true
				break
			}
		}
	}

	if changed {
		ctx.Request.URL.RawQuery = values.Encode()
	}
}

