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

func TestSanitizeLogBodyMasksSensitive(t *testing.T) {
	raw := `{"password":"secret123","name":"tom"}`
	out := sanitizeLogBody(raw, maxLogStoredBodyBytes)
	if strings.Contains(out, "secret123") {
		t.Fatalf("password should be masked: %s", out)
	}
	if !strings.Contains(out, "***") {
		t.Fatalf("expected mask marker: %s", out)
	}
}
