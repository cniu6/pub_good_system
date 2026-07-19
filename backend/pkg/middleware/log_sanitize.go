package middleware

import (
	"strings"
	"unicode/utf8"
)

// MySQL 文本类型容量（字节）：
//   TEXT       ≈ 65,535
//   MEDIUMTEXT ≈ 16,777,215（约 16MB）
//   LONGTEXT   ≈ 4GB
//
// 本项目 operation_logs / api_access_logs 的 request_body、response_body 均为 MEDIUMTEXT。
// 存库前统一截断到远小于上限的安全值，避免撑爆数据库与拖慢详情接口。
const (
	// mysqlMediumTextMaxBytes MySQL MEDIUMTEXT 理论上限（字节）
	mysqlMediumTextMaxBytes = 16_777_215
	// maxLogStoredBodyBytes 实际写入 request_body / response_body 的上限（64KB）
	// 约为 MEDIUMTEXT 的 0.4%，留足余量（含 utf8mb4 与截断标记）
	maxLogStoredBodyBytes = 64 * 1024
	// maxLogReadableBodyBytes 读取请求体时的上限，超过则不读入内存
	maxLogReadableBodyBytes = 64 * 1024
	// maxLogStoredHeadersLength 请求头 JSON 存库上限
	maxLogStoredHeadersLength = 8 * 1024
	// maxLogHeaderValueLength 单个请求头值截断长度
	maxLogHeaderValueLength = 500
	// logTruncateMarker 截断标记（追加在截断内容末尾）
	logTruncateMarker = "...(已截断)"
)

// sensitiveLogHeaderFields 仅用于请求头脱敏（Authorization / Cookie / Token 等）。
// 请求体/响应体不做字段脱敏：操作日志与 API 日志仅管理员（及本人）可查看，
// 过度脱敏会把业务字段（如响应 code、email）打成 ***，影响排查。
// 注意：使用“包含匹配”以覆盖各种命名风格（如 Authorization / X-Access-Token）。
var sensitiveLogHeaderFields = []string{
	"password", "passwd", "pwd",
	"token", "accesstoken", "refreshtoken", "apikey", "api_key", "authorization", "cookie", "session", "jwt",
	"secret",
}

// sanitizeLogBody 对请求/响应体只做长度截断，不做字段脱敏。
func sanitizeLogBody(raw string, limit int) string {
	if raw == "" {
		return raw
	}
	return truncateForLog(raw, limit)
}

// isSensitiveLogField 判断请求头字段名是否涉及敏感凭证（仅用于请求头脱敏）。
func isSensitiveLogField(key string) bool {
	normalized := strings.ToLower(key)
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "_", "")
	for _, keyword := range sensitiveLogHeaderFields {
		clean := strings.ReplaceAll(keyword, "_", "")
		if clean == "" {
			continue
		}
		if strings.Contains(normalized, clean) {
			return true
		}
	}
	return false
}

// truncateForLog 按字节上限截断，避免切断 UTF-8 多字节字符，并追加「已截断」标记。
func truncateForLog(raw string, limit int) string {
	if limit <= 0 || len(raw) <= limit {
		return raw
	}
	// 为截断标记预留空间，确保最终长度不超过 limit + 标记
	cut := limit
	if cut > len(raw) {
		cut = len(raw)
	}
	// 回退到完整 rune 边界，避免乱码
	for cut > 0 && !utf8.RuneStart(raw[cut]) {
		cut--
	}
	if cut <= 0 {
		return logTruncateMarker
	}
	return raw[:cut] + logTruncateMarker
}
