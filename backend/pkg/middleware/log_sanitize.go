package middleware

import (
	"encoding/json"
	"net/url"
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
// 注意：使用“包含匹配”以覆盖各种命名风格（如 Authorization / X-Access-Token）。
var sensitiveLogHeaderFields = []string{
	"password", "passwd", "pwd",
	"token", "accesstoken", "refreshtoken", "apikey", "api_key", "authorization", "cookie", "session", "jwt",
	"secret",
}

// sensitiveValueMask 请求/响应体命中敏感字段后的统一替换值。
const sensitiveValueMask = "***"

// sensitiveBodyFieldNames 请求/响应体中需要整体脱敏的字段名（已归一化：小写 + 去掉下划线/连字符）。
// 仅做“精确字段名”匹配，不做模糊包含匹配：
//   - 避免把业务字段误伤成 ***（例如响应包裹字段 code 表示状态码，"status_code"/"error_code" 等业务码不应被遮盖）
//   - "code" 单独处理：仅在请求体（客户端提交的验证码）中脱敏，响应体统一保留（响应包裹的 code 是状态码，见下方 isSensitiveBodyField）
var sensitiveBodyFieldNames = map[string]struct{}{
	"password": {}, "passwd": {}, "pwd": {},
	"oldpassword": {}, "newpassword": {}, "confirmpassword": {}, "repassword": {},
	"token": {}, "accesstoken": {}, "refreshtoken": {}, "idtoken": {}, "sessiontoken": {},
	"apikey": {}, "apisecret": {}, "secretkey": {}, "clientsecret": {}, "secret": {},
	"sign": {}, "signature": {},
	"authorization": {},
	// 验证码类字段：这些字段名本身语义明确是验证码，请求/响应体中都脱敏
	"smscode": {}, "emailcode": {}, "verifycode": {}, "verificationcode": {}, "authcode": {}, "otp": {}, "captchacode": {}, "resetcode": {},
	// 证件号：与 utils.MaskCertificateNo 双重保护，日志层再兜底一次
	"certificateno": {}, "idcard": {},
}

// normalizeBodyFieldName 归一化字段名：小写 + 去掉下划线/连字符，兼容 snake_case / camelCase / kebab-case。
func normalizeBodyFieldName(key string) string {
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, "_", "")
	key = strings.ReplaceAll(key, "-", "")
	return key
}

// isSensitiveBodyField 判断请求/响应体字段名是否需要脱敏。
// isRequest=true 时才把裸字段名 "code" 视为敏感（用户提交的验证码）；
// 响应体的 "code" 是 utils.Response 状态码包裹字段，不能脱敏，否则所有响应都会被打成 ***。
func isSensitiveBodyField(key string, isRequest bool) bool {
	normalized := normalizeBodyFieldName(key)
	if normalized == "code" {
		return isRequest
	}
	_, ok := sensitiveBodyFieldNames[normalized]
	return ok
}

// maskSensitiveJSONValue 递归遍历 JSON 结构，命中敏感字段名则整体替换为 ***。
// changed 记录本次是否有字段被脱敏，避免对未命中任何敏感字段的内容做无意义的重新序列化。
func maskSensitiveJSONValue(v any, isRequest bool, changed *bool) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if isSensitiveBodyField(k, isRequest) {
				if s, ok := val.(string); !ok || s != sensitiveValueMask {
					*changed = true
				}
				t[k] = sensitiveValueMask
				continue
			}
			t[k] = maskSensitiveJSONValue(val, isRequest, changed)
		}
		return t
	case []any:
		for i, item := range t {
			t[i] = maskSensitiveJSONValue(item, isRequest, changed)
		}
		return t
	default:
		return v
	}
}

// maskSensitiveFieldsInText 对日志文本做字段级脱敏（而非仅长度截断）：
//  1. 优先按 JSON 解析（对象/数组），命中敏感字段名则整体替换为 ***后重新序列化；
//  2. 非 JSON 时尝试按 form-urlencoded 解析并脱敏（如 application/x-www-form-urlencoded 请求体）；
//  3. 两者都不是结构化数据（纯文本/HTML/SSE 流等）时原样返回，仅交由 truncateForLog 截断。
//
// 未命中任何敏感字段时原样返回原文，避免无意义的重新序列化改变原始格式（便于排查时比对）。
func maskSensitiveFieldsInText(raw string, isRequest bool) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw
	}

	if trimmed[0] == '{' || trimmed[0] == '[' {
		var parsed any
		if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
			changed := false
			masked := maskSensitiveJSONValue(parsed, isRequest, &changed)
			if changed {
				if data, err := json.Marshal(masked); err == nil {
					return string(data)
				}
			}
			return raw
		}
	}

	// form-urlencoded 兜底：仅当文本形如 key=value（且不含 <> 避免误判 HTML）时才尝试
	if strings.Contains(trimmed, "=") && !strings.ContainsAny(trimmed, "<>") {
		if values, err := url.ParseQuery(trimmed); err == nil && len(values) > 0 {
			changed := false
			for key := range values {
				if isSensitiveBodyField(key, isRequest) {
					values.Set(key, sensitiveValueMask)
					changed = true
				}
			}
			if changed {
				return values.Encode()
			}
		}
	}

	return raw
}

// sanitizeLogBody 对请求/响应体做字段级脱敏 + 长度截断。
// isRequest 区分请求体/响应体：请求体中的裸 "code" 字段（用户提交的验证码）会被脱敏，
// 响应体的 "code" 字段（utils.Response 状态码包裹）保留，避免所有响应日志都被打成 ***。
func sanitizeLogBody(raw string, limit int, isRequest bool) string {
	if raw == "" {
		return raw
	}
	masked := maskSensitiveFieldsInText(raw, isRequest)
	return truncateForLog(masked, limit)
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
