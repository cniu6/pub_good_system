package middleware

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateForLogKeepsUTF8AndMarks(t *testing.T) {
	raw := strings.Repeat("测", 20) // 每字 3 字节
	out := truncateForLog(raw, 10)
	if !strings.HasSuffix(out, logTruncateMarker) {
		t.Fatalf("expected truncate marker, got %q", out)
	}
	body := strings.TrimSuffix(out, logTruncateMarker)
	if !utf8.ValidString(body) {
		t.Fatalf("truncated body is not valid UTF-8: %q", body)
	}
	if len(out) <= 10 {
		t.Fatalf("expected truncated content longer than limit due to marker, got len=%d", len(out))
	}
}

func TestTruncateForLogNoOpWhenShort(t *testing.T) {
	raw := "hello"
	if got := truncateForLog(raw, 100); got != raw {
		t.Fatalf("expected unchanged, got %q", got)
	}
}

func TestSanitizeLogBodyKeepsResponseEnvelopeCode(t *testing.T) {
	// 响应体的 code 是 utils.Response 状态码包裹字段（如 code:0 表示成功），
	// 不能脱敏，否则所有响应日志都会被打成 ***，业务字段（如 email）同样保留
	raw := `{"code":0,"data":{"email":"a@b.com","username":"zerohh"},"message":"OK"}`
	out := sanitizeLogBody(raw, maxLogStoredBodyBytes, false)
	if out != raw {
		t.Fatalf("expected response body unchanged, got %q", out)
	}
}

func TestSanitizeLogBodyMasksRequestVerificationCode(t *testing.T) {
	// 请求体中的裸 code 字段是用户提交的验证码，需要脱敏
	raw := `{"email":"a@b.com","code":"123456"}`
	out := sanitizeLogBody(raw, maxLogStoredBodyBytes, true)
	if strings.Contains(out, "123456") {
		t.Fatalf("expected verification code masked, got %q", out)
	}
	if !strings.Contains(out, `"code":"***"`) {
		t.Fatalf("expected code field replaced with ***, got %q", out)
	}
	if !strings.Contains(out, `"a@b.com"`) {
		t.Fatalf("expected business field (email) preserved, got %q", out)
	}
}

func TestSanitizeLogBodyMasksSensitiveFieldsRegardlessOfScope(t *testing.T) {
	// password/token/secret/apikey/sign/authorization 等字段无论请求体还是响应体都需脱敏
	raw := `{"username":"zerohh","password":"P@ssw0rd123","data":{"accessToken":"abc.def.ghi","refresh_token":"xyz"}}`
	out := sanitizeLogBody(raw, maxLogStoredBodyBytes, false)
	if strings.Contains(out, "P@ssw0rd123") || strings.Contains(out, "abc.def.ghi") || strings.Contains(out, "xyz") {
		t.Fatalf("expected sensitive fields masked, got %q", out)
	}
	if !strings.Contains(out, `"username":"zerohh"`) {
		t.Fatalf("expected business field (username) preserved, got %q", out)
	}
}

func TestSanitizeLogBodyMasksFormEncodedSecret(t *testing.T) {
	raw := "mobile=13800000000&password=abc123&remember=1"
	out := sanitizeLogBody(raw, maxLogStoredBodyBytes, true)
	if strings.Contains(out, "abc123") {
		t.Fatalf("expected form password masked, got %q", out)
	}
	if !strings.Contains(out, "mobile=13800000000") {
		t.Fatalf("expected business field preserved, got %q", out)
	}
}

func TestSanitizeLogBodyDoesNotOverMaskBusinessCodeFields(t *testing.T) {
	// status_code / error_code 等复合字段名是业务码，不应被当成验证码脱敏
	raw := `{"status_code":404,"errorCode":"NOT_FOUND"}`
	out := sanitizeLogBody(raw, maxLogStoredBodyBytes, true)
	if out != raw {
		t.Fatalf("expected business code fields unchanged, got %q", out)
	}
}

func TestSanitizeLogBodyTruncatesLongContent(t *testing.T) {
	raw := strings.Repeat("a", 100)
	out := sanitizeLogBody(raw, 20, false)
	if !strings.HasSuffix(out, logTruncateMarker) {
		t.Fatalf("expected truncate marker, got %q", out)
	}
	if strings.Contains(out, strings.Repeat("a", 50)) {
		t.Fatalf("expected truncated content, got len=%d", len(out))
	}
}

func TestIsSensitiveLogFieldForHeaders(t *testing.T) {
	if !isSensitiveLogField("Authorization") {
		t.Fatal("Authorization should be sensitive")
	}
	if !isSensitiveLogField("X-Access-Token") {
		t.Fatal("X-Access-Token should be sensitive")
	}
	// 业务字段不应被当成请求头敏感键
	if isSensitiveLogField("email") {
		t.Fatal("email should not be treated as header secret")
	}
	if isSensitiveLogField("code") {
		t.Fatal("code should not be treated as header secret")
	}
}
