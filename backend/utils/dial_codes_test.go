package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCountryFromLanguage(t *testing.T) {
	if got := CountryFromLanguage("zhCN"); got != "CN" {
		t.Fatalf("zhCN => %s", got)
	}
	if got := CountryFromLanguage("zh-CN"); got != "CN" {
		t.Fatalf("zh-CN => %s", got)
	}
	if got := CountryFromLanguage("enUS"); got != "US" {
		t.Fatalf("enUS => %s", got)
	}
	if got := CountryFromLanguage(""); got != "" {
		t.Fatalf("empty => %q want empty", got)
	}
}

func TestCountryFromCDNHeaders(t *testing.T) {
	h := map[string]string{"CF-IPCountry": "jp"}
	got := CountryFromCDNHeaders(func(k string) string { return h[k] })
	if got != "JP" {
		t.Fatalf("got %s", got)
	}
	h2 := map[string]string{"CF-IPCountry": "XX"}
	if got := CountryFromCDNHeaders(func(k string) string { return h2[k] }); got != "" {
		t.Fatalf("XX should ignore, got %s", got)
	}
}

func TestResolvePhoneCountry(t *testing.T) {
	c, src := ResolvePhoneCountry(ResolvePhoneCountryOptions{
		CNOnly: true,
		Lang:   "enUS",
		GetHeader: func(string) string { return "US" },
	})
	if c.Code != "CN" || src != "cn_only" {
		t.Fatalf("cn_only got %+v src=%s", c, src)
	}

	c, src = ResolvePhoneCountry(ResolvePhoneCountryOptions{
		CNOnly: false,
		Lang:   "zhCN",
		GetHeader: func(string) string { return "" },
	})
	if c.Code != "CN" || src != "lang" {
		t.Fatalf("lang got %+v src=%s", c, src)
	}

	c, src = ResolvePhoneCountry(ResolvePhoneCountryOptions{
		CNOnly: false,
		Lang:   "",
		GetHeader: func(string) string { return "" },
	})
	if c.Code != "US" || src != "default" {
		t.Fatalf("default got %+v src=%s", c, src)
	}
}

func TestComposeE164(t *testing.T) {
	if got := ComposeE164("86", "013800138000"); got != "+8613800138000" {
		t.Fatalf("got %s", got)
	}
	if got := ComposeE164("1", "4155552671"); got != "+14155552671" {
		t.Fatalf("got %s", got)
	}
}

func TestLookupIPCountryUsesProviderPool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"cc":"US"}`))
	}))
	defer srv.Close()

	savedPool := freeIPCountryProviders
	savedHook := GetCustomIPCountryLookup()
	defer func() { freeIPCountryProviders = savedPool; SetCustomIPCountryLookup(savedHook) }()

	SetCustomIPCountryLookup(nil)
	freeIPCountryProviders = []ipCountryProvider{{
		name:  "test",
		build: func(ip string) string { return srv.URL },
		parse: func(b []byte) string {
			var r struct {
				CC string `json:"cc"`
			}
			_ = json.Unmarshal(b, &r)
			return r.CC
		},
	}}

	if got := LookupIPCountry("8.8.8.8"); got != "US" {
		t.Fatalf("pool lookup => %q, want US", got)
	}
	// 私网/环回 IP 应直接返回空，不发请求
	if got := LookupIPCountry("192.168.1.1"); got != "" {
		t.Fatalf("private ip => %q, want empty", got)
	}
}

func TestLookupIPCountryCustomHookPriority(t *testing.T) {
	savedPool := freeIPCountryProviders
	savedHook := GetCustomIPCountryLookup()
	defer func() { freeIPCountryProviders = savedPool; SetCustomIPCountryLookup(savedHook) }()

	// 自定义 provider 命中 JP；若错误地回退到免费池会得到 US，用于验证优先级
	SetCustomIPCountryLookup(func(ip string) string { return "JP" })
	freeIPCountryProviders = []ipCountryProvider{{
		name:  "should-not-be-used",
		build: func(ip string) string { return "http://127.0.0.1:0" },
		parse: func(b []byte) string { return "US" },
	}}

	if got := LookupIPCountry("8.8.8.8"); got != "JP" {
		t.Fatalf("custom hook => %q, want JP", got)
	}
}

func TestBuildJSONCountryLookup(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"country_code":"US"}`))
	}))
	defer srv.Close()

	lookup := BuildJSONCountryLookup(srv.URL+"?ip={ip}", "country_code")
	if lookup == nil {
		t.Fatal("builder returned nil for valid args")
	}
	if got := lookup("8.8.8.8"); got != "US" {
		t.Fatalf("builder lookup => %q, want US", got)
	}
	// 非法参数应返回 nil
	if BuildJSONCountryLookup("no-placeholder", "cc") != nil {
		t.Fatal("expected nil when template lacks {ip}")
	}
	if BuildJSONCountryLookup("", "cc") != nil {
		t.Fatal("expected nil for empty template")
	}
}
