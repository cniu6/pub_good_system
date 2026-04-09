package sms_plugin

import (
	"encoding/json"
	"fst/backend/app/services"
	"fst/backend/internal/config"
	"net/url"
	"testing"
)

func TestBuildSMSConfigMapsAllFields(t *testing.T) {
	cfg := &config.Config{
		SMSProvider:       "custom",
		SMSAccessKey:      "ak",
		SMSSecretKey:      "sk",
		SMSSignName:       "sign",
		SMSTemplateCode:   "tpl",
		SMSTemplateCodeEN: "tpl_en",
		SMSRegion:         "ap-guangzhou",
		SMSEndpoint:       "https://example.com/sms",
		SMSSdkAppID:       "1400123456",
		SMSBodyFormat:     "form",
	}

	got := buildSMSConfig(cfg)
	if got.Provider != "custom" || got.AccessKey != "ak" || got.SecretKey != "sk" ||
		got.SignName != "sign" || got.TemplateCode != "tpl" || got.TemplateCodeEN != "tpl_en" || got.Region != "ap-guangzhou" ||
		got.Endpoint != "https://example.com/sms" || got.SdkAppID != "1400123456" || got.BodyFormat != "form" {
		t.Fatalf("buildSMSConfig returned unexpected result: %+v", got)
	}
}

func TestNewProviderSelection(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantNil  bool
	}{
		{name: "aliyun", provider: "aliyun"},
		{name: "tencent", provider: "tencent"},
		{name: "custom", provider: "custom"},
		{name: "unknown", provider: "unknown", wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newProvider(services.SMSConfig{Provider: tt.provider})
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil provider for %s", tt.provider)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected provider instance for %s", tt.provider)
			}
			if got.Name() != tt.provider {
				t.Fatalf("expected provider name %s, got %s", tt.provider, got.Name())
			}
		})
	}
}

func TestApplyRuntimeProviderUpdatesServiceProvider(t *testing.T) {
	services.InitSMSService()

	ApplyRuntimeProvider(services.SMSConfig{
		Provider:   "custom",
		Endpoint:   "https://example.com/sms",
		BodyFormat: "json",
	})

	if got := services.GlobalSMSService.GetProviderName(); got != "custom" {
		t.Fatalf("expected runtime provider custom, got %s", got)
	}
	if !services.GlobalSMSService.IsConfigured() {
		t.Fatal("expected runtime provider to be configured")
	}
}

func TestNormalizeTemplateParams(t *testing.T) {
	name, payload, order := normalizeTemplateParams("123456", 10, map[string]string{
		"__template_name":  "bind_phone",
		"__template_order": "code,expire,app_name",
		"app_name":         "F.st",
	})

	if name != "bind_phone" {
		t.Fatalf("expected template name bind_phone, got %s", name)
	}
	if payload["code"] != "123456" || payload["expire"] != "10" || payload["app_name"] != "F.st" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if len(order) != 3 || order[0] != "code" || order[1] != "expire" || order[2] != "app_name" {
		t.Fatalf("unexpected order: %+v", order)
	}
}

func TestCustomProviderBuildBodyJSON(t *testing.T) {
	provider := NewCustomProvider(services.SMSConfig{
		Provider:   "custom",
		SignName:   "F.st",
		BodyFormat: "json",
	})

	body, err := provider.buildBody("13800138000", "test-content", map[string]string{
		"code":   "123456",
		"expire": "10",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("expected valid json body, got %v", err)
	}
	if payload["phone"] != "13800138000" || payload["content"] != "test-content" || payload["sign"] != "F.st" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	params, ok := payload["params"].(map[string]any)
	if !ok || params["code"] != "123456" || params["expire"] != "10" {
		t.Fatalf("unexpected params payload: %+v", payload["params"])
	}
}

func TestCustomProviderBuildBodyForm(t *testing.T) {
	provider := NewCustomProvider(services.SMSConfig{
		Provider:   "custom",
		SignName:   "F.st",
		BodyFormat: "form",
	})

	body, err := provider.buildBody("13800138000", "验证码 123456", map[string]string{
		"code":   "123456",
		"expire": "10",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	values, err := url.ParseQuery(body)
	if err != nil {
		t.Fatalf("expected valid form body, got %v", err)
	}
	if values.Get("phone") != "13800138000" || values.Get("content") != "验证码 123456" || values.Get("sign") != "F.st" {
		t.Fatalf("unexpected form values: %+v", values)
	}
	if values.Get("code") != "123456" || values.Get("expire") != "10" {
		t.Fatalf("unexpected form params: %+v", values)
	}
}

func TestNormalizeSMSLang(t *testing.T) {
	tests := map[string]string{
		"":      "zh-CN",
		"zh":    "zh-CN",
		"zh-CN": "zh-CN",
		"en":    "en-US",
		"en-US": "en-US",
	}

	for input, want := range tests {
		if got := normalizeSMSLang(input); got != want {
			t.Fatalf("normalizeSMSLang(%q) expected %q, got %q", input, want, got)
		}
	}
}

func TestAliyunProviderTemplateCodeForLang(t *testing.T) {
	provider := NewAliyunProvider(services.SMSConfig{
		TemplateCode:   "SMS_ZH",
		TemplateCodeEN: "SMS_EN",
	})

	if got := provider.templateCodeForLang("en-US"); got != "SMS_EN" {
		t.Fatalf("expected english template code, got %s", got)
	}
	if got := provider.templateCodeForLang("zh-CN"); got != "SMS_ZH" {
		t.Fatalf("expected chinese template code, got %s", got)
	}
}

func TestTencentProviderTemplateCodeForLang(t *testing.T) {
	provider := NewTencentProvider(services.SMSConfig{
		TemplateCode:   "1860001",
		TemplateCodeEN: "1860002",
	})

	if got := provider.templateCodeForLang("en"); got != "1860002" {
		t.Fatalf("expected english template id, got %s", got)
	}
	if got := provider.templateCodeForLang(""); got != "1860001" {
		t.Fatalf("expected default template id, got %s", got)
	}
}

func TestCustomProviderIsSuccess(t *testing.T) {
	provider := NewCustomProvider(services.SMSConfig{Provider: "custom"})

	successCases := []string{"success", "ok", "", `{"code":0}`, `{"errcode":"ok"}`}
	for _, value := range successCases {
		if !provider.isSuccess(value) {
			t.Fatalf("expected %q to be recognized as success", value)
		}
	}

	if provider.isSuccess(`{"code":500,"message":"fail"}`) {
		t.Fatal("expected failure payload to be recognized as failure")
	}
}
