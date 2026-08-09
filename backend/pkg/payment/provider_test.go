package payment

import (
	"testing"
)

func TestParseMoneyMinor(t *testing.T) {
	cases := []struct {
		in     string
		want   int64
		hasErr bool
	}{
		{"1.23", 123, false},
		{"100", 10000, false},
		{"0.01", 1, false},
		{"35.90", 3590, false},
		{"35.9", 3590, false},
		{"", 0, true},
		{"abc", 0, true},
		{"1.234", 0, true},
	}
	for _, c := range cases {
		got, err := ParseMoneyMinor(c.in)
		if c.hasErr {
			if err == nil {
				t.Errorf("ParseMoneyMinor(%q) expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseMoneyMinor(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseMoneyMinor(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestFormatMoneyYuan(t *testing.T) {
	if FormatMoneyYuan(123) != "1.23" {
		t.Errorf("FormatMoneyYuan(123) = %s", FormatMoneyYuan(123))
	}
}

func TestNormalizeTradeStatus(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"1", "TRADE_SUCCESS"},
		{"TRADE_SUCCESS", "TRADE_SUCCESS"},
		{"pending", "PENDING"},
		{"WAIT", "PENDING"},
		{"unknown", "UNKNOWN"},
	}
	for _, c := range cases {
		if got := NormalizeTradeStatus(c.in); got != c.want {
			t.Errorf("NormalizeTradeStatus(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseExtConfig(t *testing.T) {
	m := ParseExtConfig(`{"merchant_key":"abc","version":"v2"}`)
	if m["merchant_key"] != "abc" || m["version"] != "v2" {
		t.Errorf("ParseExtConfig unexpected result: %v", m)
	}
	empty := ParseExtConfig("")
	if empty == nil {
		t.Error("ParseExtConfig(\"\") should return non-nil map")
	}
}

func TestSignWithMD5(t *testing.T) {
	params := map[string]string{
		"pid":  "1001",
		"type": "alipay",
	}
	key := "secret"
	sign := SignWithMD5(params, key)
	if sign == "" {
		t.Fatal("SignWithMD5 returned empty string")
	}
	// 确定性
	if SignWithMD5(params, key) != sign {
		t.Error("SignWithMD5 is not deterministic")
	}
	// 修改参数签名应变
	params2 := map[string]string{
		"pid":  "1001",
		"type": "wxpay",
	}
	if SignWithMD5(params2, key) == sign {
		t.Error("SignWithMD5 should differ with different params")
	}
}

func TestNormalizeTradeNo(t *testing.T) {
	if got := NormalizeTradeNo("  abC123  "); got != "abC123" {
		t.Errorf("NormalizeTradeNo got %q", got)
	}
	if got := NormalizeTradeNo("NULL"); got != "" {
		t.Errorf("NormalizeTradeNo(NULL) = %q", got)
	}
}
