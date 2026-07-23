package sms_plugin

import (
	"net"
	"strings"
	"testing"
)

func TestValidateOutboundURL_RejectNonHTTPSchemes(t *testing.T) {
	t.Parallel()
	cases := []string{
		"file:///etc/passwd",
		"gopher://127.0.0.1:70/",
		"ftp://example.com/a",
		"dict://127.0.0.1:11211/",
	}
	for _, raw := range cases {
		if err := ValidateOutboundURL(raw); err == nil {
			t.Fatalf("expected reject for scheme in %q", raw)
		}
	}
}

func TestValidateOutboundURL_RejectBlockedIPs(t *testing.T) {
	t.Parallel()
	cases := []string{
		"http://127.0.0.1/sms",
		"https://127.0.0.1:8080/x",
		"http://localhost/sms",
		"http://10.0.0.1/hook",
		"http://192.168.1.1/hook",
		"http://172.16.5.5/hook",
		"http://169.254.169.254/latest/meta-data/",
		"http://[::1]:8080/x",
		"http://0.0.0.0/x",
	}
	for _, raw := range cases {
		if err := ValidateOutboundURL(raw); err == nil {
			t.Fatalf("expected reject for blocked URL %q", raw)
		}
	}
}

func TestValidateOutboundURL_RejectEmptyAndInvalid(t *testing.T) {
	t.Parallel()
	if err := ValidateOutboundURL(""); err == nil {
		t.Fatal("expected reject empty URL")
	}
	if err := ValidateOutboundURL("http:///no-host"); err == nil {
		t.Fatal("expected reject empty host")
	}
}

func TestValidateOutboundURL_AllowPublicIPLiteral(t *testing.T) {
	t.Parallel()
	// 字面量公网 IP：不依赖 DNS，测试更稳定
	if err := ValidateOutboundURL("https://8.8.8.8/sms"); err != nil {
		t.Fatalf("expected allow public IP, got: %v", err)
	}
	if err := ValidateOutboundURL("http://1.1.1.1:443/hook"); err != nil {
		t.Fatalf("expected allow public IP, got: %v", err)
	}
}

func TestIsBlockedOutboundIP(t *testing.T) {
	t.Parallel()
	blocked := []string{"127.0.0.1", "::1", "10.1.2.3", "192.168.0.1", "172.31.255.1", "169.254.169.254", "0.0.0.0"}
	for _, s := range blocked {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("bad IP %q", s)
		}
		if !isBlockedOutboundIP(ip) {
			t.Fatalf("expected %s blocked", s)
		}
	}
	allowed := []string{"8.8.8.8", "1.1.1.1", "93.184.216.34"}
	for _, s := range allowed {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("bad IP %q", s)
		}
		if isBlockedOutboundIP(ip) {
			t.Fatalf("expected %s allowed", s)
		}
	}
}

func TestValidateOutboundURL_ErrorMentionsBlock(t *testing.T) {
	t.Parallel()
	err := ValidateOutboundURL("http://169.254.169.254/meta")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("expected blocked message, got: %v", err)
	}
}
