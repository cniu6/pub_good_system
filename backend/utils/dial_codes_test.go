package utils

import "testing"

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
