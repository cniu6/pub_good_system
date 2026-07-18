package utils

import (
	"html"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// Clean_XSS 清理用户输入，防 XSS / 控制字符；不做 SQL 关键词黑名单替换。
//
// 设计说明：
//   - SQL 注入应由参数化查询兜底（本项目已普遍使用 ? 占位），禁止用黑名单词替换
//     冒充防 SQL（会误伤备注里的 "select"、合法 URL 里的引号等）。
//   - 输入层：长度上限、去掉危险标签/事件属性/危险协议、剔除控制字符。
//   - 输出层：展示时由前端/模板再转义；此处 EscapeString 作为额外保险。
//
// 注意：对「需要原样存 URL」的字段，调用方应在 Clean_XSS 之后用 ValidateURL 校验，
// 或对 URL 字段单独处理（本函数会 EscapeString，& 会变成 &amp;）。
func Clean_XSS(input string) string {
	if input == "" {
		return input
	}

	input = strings.TrimSpace(input)

	// 长度上限，避免超大 payload
	const maxLen = 10000
	if utf8.RuneCountInString(input) > maxLen {
		runes := []rune(input)
		input = string(runes[:maxLen])
	}

	// 去掉控制字符（保留常见空白：\t \n \r）
	input = controlCharsRe.ReplaceAllString(input, "")

	// 去掉危险 HTML/脚本片段（在 Escape 之前做，便于匹配原始标签）
	input = stripDangerousMarkup(input)

	// HTML 实体转义，降低反射型 XSS 风险
	input = html.EscapeString(input)

	return strings.TrimSpace(input)
}

var (
	controlCharsRe = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	scriptTagRe    = regexp.MustCompile(`(?is)<\s*script[^>]*>.*?<\s*/\s*script\s*>`)
	scriptOpenRe   = regexp.MustCompile(`(?is)<\s*/?\s*script[^>]*>`)
	eventAttrRe    = regexp.MustCompile(`(?i)\s+on[a-z]+\s*=`)
	jsProtocolRe   = regexp.MustCompile(`(?i)javascript\s*:`)
	vbProtocolRe   = regexp.MustCompile(`(?i)vbscript\s*:`)
	objectTagRe    = regexp.MustCompile(`(?is)<\s*(iframe|object|embed|applet|link|meta|base)[^>]*>`)
	expressionRe   = regexp.MustCompile(`(?i)expression\s*\(`)
	dataBase64Re   = regexp.MustCompile(`(?i)data\s*:\s*[^,]*;?\s*base64`)
)

func stripDangerousMarkup(input string) string {
	input = scriptTagRe.ReplaceAllString(input, "")
	input = scriptOpenRe.ReplaceAllString(input, "")
	input = objectTagRe.ReplaceAllString(input, "")
	input = eventAttrRe.ReplaceAllString(input, " ")
	input = jsProtocolRe.ReplaceAllString(input, "")
	input = vbProtocolRe.ReplaceAllString(input, "")
	input = expressionRe.ReplaceAllString(input, "")
	input = dataBase64Re.ReplaceAllString(input, "data:text/plain")
	return input
}

// ValidateURL 校验用户提供的URL是否为安全的 http/https 地址
// 返回 true 表示合法，false 表示非法或危险
func ValidateURL(rawURL string) bool {
	if rawURL == "" {
		return true // 空值允许（代表清除）
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
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
