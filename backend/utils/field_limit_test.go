package utils

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestClampBytesKeepsUTF8(t *testing.T) {
	raw := strings.Repeat("测", 20) // 每字 3 字节
	out := ClampBytes(raw, 10)
	if len(out) > 10 {
		t.Fatalf("expected len<=10, got %d", len(out))
	}
	if !utf8.ValidString(out) {
		t.Fatalf("invalid utf8: %q", out)
	}
}

func TestClampBytesNoOpWhenShort(t *testing.T) {
	if got := ClampBytes("hello", 100); got != "hello" {
		t.Fatalf("expected unchanged, got %q", got)
	}
}

func TestValidateRuneLenRejectsOverLong(t *testing.T) {
	err := ValidateRuneLen(strings.Repeat("a", 65), "用户名", 64)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "不能超过64个字符") {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestValidateRuneLenAllowsExact(t *testing.T) {
	if err := ValidateRuneLen(strings.Repeat("测", 64), "昵称", 64); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
