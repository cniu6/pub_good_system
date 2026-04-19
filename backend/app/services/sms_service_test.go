package services

import (
	"errors"
	"testing"
)

type stubSMSProvider struct {
	name         string
	configured   bool
	sendErr      error
	sendCodeErr  error
	sendCalls    int
	sendCodeCalls int
}

func (p *stubSMSProvider) Name() string { return p.name }
func (p *stubSMSProvider) IsConfigured() bool { return p.configured }
func (p *stubSMSProvider) Send(phone, content string) error {
	p.sendCalls++
	return p.sendErr
}
func (p *stubSMSProvider) SendCode(phone, code string, expireMinutes int, templateParams map[string]string, lang string) error {
	p.sendCodeCalls++
	return p.sendCodeErr
}

func TestSMSServiceUseProviderReplacesNonConsoleProviders(t *testing.T) {
	service := &SMSService{
		providers: map[string]SMSProvider{
			"console": &stubSMSProvider{name: "console", configured: true},
			"old":     &stubSMSProvider{name: "old", configured: true},
		},
	}

	newProvider := &stubSMSProvider{name: "tencent", configured: true}
	service.UseProvider("tencent", newProvider)

	if len(service.providers) != 2 {
		t.Fatalf("expected 2 providers after replacement, got %d", len(service.providers))
	}
	if _, ok := service.providers["old"]; ok {
		t.Fatal("expected old provider to be removed")
	}
	if got := service.providers["tencent"]; got != newProvider {
		t.Fatal("expected new provider to be registered")
	}
}

func TestSMSServiceSendCodeFallsBackToConsole(t *testing.T) {
	primary := &stubSMSProvider{name: "aliyun", configured: true, sendCodeErr: errors.New("primary failed")}
	console := &stubSMSProvider{name: "console", configured: true}

	service := &SMSService{
		providers: map[string]SMSProvider{
			"console": console,
			"aliyun":  primary,
		},
		config: SMSConfig{Provider: "aliyun"},
	}

	if err := service.SendCode("13800138000", "123456", 10, map[string]string{"__template_name": "bind_phone"}, "zh-CN"); err != nil {
		t.Fatalf("expected console fallback to succeed, got error: %v", err)
	}
	if primary.sendCodeCalls != 1 {
		t.Fatalf("expected primary provider to be called once, got %d", primary.sendCodeCalls)
	}
	if console.sendCodeCalls != 1 {
		t.Fatalf("expected console provider to be called once, got %d", console.sendCodeCalls)
	}
}

func TestSMSServiceIsConfiguredAndGetProviderName(t *testing.T) {
	service := &SMSService{
		providers: map[string]SMSProvider{
			"console": &stubSMSProvider{name: "console", configured: true},
			"custom":  &stubSMSProvider{name: "custom", configured: true},
		},
		config: SMSConfig{Provider: "custom"},
	}

	if !service.IsConfigured() {
		t.Fatal("expected service to be configured")
	}
	if got := service.GetProviderName(); got != "custom" {
		t.Fatalf("expected provider name custom, got %s", got)
	}
}

