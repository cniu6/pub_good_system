package middleware

import (
	"encoding/json"
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

// sensitiveLogFields 审计日志中需要脱敏的字段。
// 注意：这里使用“包含匹配”以覆盖各种命名风格（如 new_password / oldPassword）。
var sensitiveLogFields = []string{
	"password", "passwd", "pwd",
	"token", "accesstoken", "refreshtoken", "apikey", "api_key", "authorization", "cookie", "session", "jwt",
	"secret", "sign", "signature",
	"captcha", "code", "verification",
	"mobile", "phone", "email",
	"certificate_no", "certificateno", "id_card", "idcard",
	"account_no", "accountno", "card_no", "cardno",
}

// sanitizeLogBody 对请求/响应体做关键字段脱敏，并控制长度。
// 优先尝试 JSON 脱敏；如果不是 JSON，则退化为截断。
func sanitizeLogBody(raw string, limit int) string {
	if raw == "" {
		return raw
	}

	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var payload any
		if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
			masked := maskSensitiveValues(payload)
			if data, err := json.Marshal(masked); err == nil {
				return truncateForLog(string(data), limit)
			}
		}
	}

	return truncateForLog(raw, limit)
}

// maskSensitiveValues 递归遍历 JSON 结构，对敏感键值做遮蔽。
func maskSensitiveValues(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, inner := range v {
			if isSensitiveLogField(key) {
				out[key] = "***"
				continue
			}
			out[key] = maskSensitiveValues(inner)
		}
		return out
	case []any:
		for i := range v {
			v[i] = maskSensitiveValues(v[i])
		}
		return v
	default:
		return v
	}
}

// isSensitiveLogField 判断字段名是否涉及敏感数据。
func isSensitiveLogField(key string) bool {
	normalized := strings.ToLower(key)
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, "_", "")
	for _, keyword := range sensitiveLogFields {
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
