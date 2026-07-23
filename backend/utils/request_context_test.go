package utils

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResolveRequestLang(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	ctx.Request = req

	if got := ResolveRequestLang(ctx, "", "zh-CN"); got != "en-US" {
		t.Fatalf("ResolveRequestLang() = %q, want %q", got, "en-US")
	}
	if got := ResolveRequestLang(ctx, "zh-Hans-CN", "en-US"); got != "zh-CN" {
		t.Fatalf("ResolveRequestLang(reqLang) = %q, want %q", got, "zh-CN")
	}
}

func TestGetClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.0.2.10:12345"
	ctx.Request = req

	if got := GetClientIP(ctx); got != "192.0.2.10" {
		t.Fatalf("GetClientIP() = %q, want %q", got, "192.0.2.10")
	}
}

func TestClampStoredIP(t *testing.T) {
	if got := ClampStoredIP("  1.2.3.4  "); got != "1.2.3.4" {
		t.Fatalf("ClampStoredIP trim = %q", got)
	}
	long := strings.Repeat("a", MaxStoredIPLength+10)
	if got := ClampStoredIP(long); len(got) != MaxStoredIPLength {
		t.Fatalf("ClampStoredIP len = %d, want %d", len(got), MaxStoredIPLength)
	}
}

func TestExtractBearerToken(t *testing.T) {
	if got := ExtractBearerToken("Bearer abc.def.ghi"); got != "abc.def.ghi" {
		t.Fatalf("ExtractBearerToken() = %q, want %q", got, "abc.def.ghi")
	}
	if got := ExtractBearerToken("abc.def.ghi"); got != "abc.def.ghi" {
		t.Fatalf("ExtractBearerToken(raw) = %q, want %q", got, "abc.def.ghi")
	}
}

func TestParseDeviceFromUserAgent(t *testing.T) {
	if got := ParseDeviceFromUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64)"); got != "Windows PC" {
		t.Fatalf("ParseDeviceFromUserAgent(windows) = %q, want %q", got, "Windows PC")
	}
	if got := ParseDeviceFromUserAgent("Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)"); got != "iPhone" {
		t.Fatalf("ParseDeviceFromUserAgent(iphone) = %q, want %q", got, "iPhone")
	}
}
