package middleware

import (
	"encoding/json"
	"strings"
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

func truncateForLog(raw string, limit int) string {
	if limit <= 0 || len(raw) <= limit {
		return raw
	}
	return raw[:limit] + "...(truncated)"
}
