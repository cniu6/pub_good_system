package models

import (
	"strings"
	"testing"
)

func TestClampBytesKeepsUTF8(t *testing.T) {
	raw := strings.Repeat("测", 20)
	out := clampBytes(raw, 10)
	if len(out) > 10 {
		t.Fatalf("expected len<=10, got %d", len(out))
	}
}

func TestClampStoredIP(t *testing.T) {
	got := clampStoredIP(strings.Repeat("a", 100))
	if len(got) != storedIPMaxLen {
		t.Fatalf("expected %d, got %d", storedIPMaxLen, len(got))
	}
}

func TestClampBrowserIDAndDevice(t *testing.T) {
	if len(clampBrowserID(strings.Repeat("b", 100))) != storedBrowserIDLen {
		t.Fatal("browser_id clamp failed")
	}
	if len(clampDevice(strings.Repeat("d", 200))) != storedDeviceLen {
		t.Fatal("device clamp failed")
	}
}

func TestNormalizeTradeNoClampsLength(t *testing.T) {
	long := strings.Repeat("T", 100)
	got := NormalizeTradeNo(long)
	if len(got) > 64 {
		t.Fatalf("expected <=64, got %d", len(got))
	}
	if got == "" {
		t.Fatal("expected non-empty trade no")
	}
}

func TestResolveAPIAccessLogAggregateRouteClamps(t *testing.T) {
	got := resolveAPIAccessLogAggregateRoute(strings.Repeat("/p", 200), "")
	if len(got) > storedPathLen {
		t.Fatalf("expected <=%d, got %d", storedPathLen, len(got))
	}
}
