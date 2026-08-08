package utils

import (
	"strconv"
	"unicode/utf8"
)

// InterfaceToString 把常见的 JSON 解出来的动态类型转成字符串，用于模板变量替换等场景
// （邮件模板预览、短信模板预览等原来各自写了一份几乎一样的 toString，统一到这里）。
// 只覆盖 JSON 常见的 string/float64/int，其它类型返回空串（模板占位符替换失败时留空比崩溃更安全）。
func InterfaceToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	default:
		return ""
	}
}

// TruncateString 按 rune 数量截断字符串，避免按 byte 截断导致中文乱码。
// 返回前 n 个 rune，超过则在末尾追加 "..." 提示被截断。
func TruncateString(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n]) + "..."
}
