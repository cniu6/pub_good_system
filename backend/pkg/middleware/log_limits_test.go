package middleware

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateForLogKeepsUTF8AndMarks(t *testing.T) {
	raw := strings.Repeat("测", 20) // 每字 3 字节
	out := truncateForLog(raw, 40)
	if !strings.HasSuffix(out, logTruncateMarker) {
		t.Fatalf("expected truncate marker, got %q", out)
	}
	if len(out) > 40 {
		t.Fatalf("expected len<=limit, got len=%d", len(out))
	}
	body := strings.TrimSuffix(out, logTruncateMarker)
	if !utf8.ValidString(body) {
		t.Fatalf("truncated body is not valid UTF-8: %q", body)
	}
}

func TestTruncateForLogRespectsLimitWithMarker(t *testing.T) {
	raw := strings.Repeat("a", 100)
	out := truncateForLog(raw, 20)
	if len(out) > 20 {
		t.Fatalf("expected len<=20, got %d (%q)", len(out), out)
	}
	if !strings.HasSuffix(out, logTruncateMarker) {
		t.Fatalf("expected truncate marker, got %q", out)
	}
}

func TestTruncateForLogNoOpWhenShort(t *testing.T) {
	raw := "hello"
	if got := truncateForLog(raw, 100); got != raw {
		t.Fatalf("expected unchanged, got %q", got)
	}
}

func TestCaptureRequestHeadersKeepsPlaintext(t *testing.T) {
	// 访问日志按产品要求明文落库（含 Authorization / Cookie），仅长度截断
	raw := captureRequestHeaders(map[string][]string{
		"Authorization": {"Bearer super-secret-token"},
		"X-Api-Key":     {"plain-api-key"},
		"Cookie":        {"session=abc"},
		"Content-Type":  {"application/json"},
		"X-Custom":      {"keep-me"},
	})
	if raw == nil {
		t.Fatal("expected headers json")
	}
	out := *raw
	if !strings.Contains(out, "super-secret-token") || !strings.Contains(out, "plain-api-key") || !strings.Contains(out, "session=abc") {
		t.Fatalf("expected secrets kept as plaintext, got %s", out)
	}
	if !strings.Contains(out, "keep-me") || !strings.Contains(out, "application/json") {
		t.Fatalf("expected non-secret headers preserved, got %s", out)
	}
}

func TestFormatQueryStringKeepsToken(t *testing.T) {
	out := formatQueryString("page=1&token=abc.def&keyword=hello")
	if !strings.Contains(out, "abc.def") {
		t.Fatalf("expected token kept as plaintext, got %s", out)
	}
	if !strings.Contains(out, "hello") {
		t.Fatalf("expected keyword preserved, got %s", out)
	}
}
