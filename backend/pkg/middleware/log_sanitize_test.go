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

func TestSanitizeLogBodyKeepsBusinessFields(t *testing.T) {
	// 管理端审计日志需要完整业务字段（如 code / email），不做字段脱敏
	raw := `{"code":0,"data":{"email":"a@b.com","username":"zerohh"},"message":"OK"}`
	out := sanitizeLogBody(raw, maxLogStoredBodyBytes)
	if out != raw {
		t.Fatalf("expected body unchanged, got %q", out)
	}
}

func TestSanitizeLogBodyTruncatesLongContent(t *testing.T) {
	raw := strings.Repeat("a", 100)
	out := sanitizeLogBody(raw, 20)
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
