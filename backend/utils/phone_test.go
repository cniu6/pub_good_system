package utils

import "testing"

func TestNormalizeAndValidateMobile_CNOnly(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"13800138000", "13800138000", true},
		{"+86 138-0013-8000", "13800138000", true},
		{"8613800138000", "13800138000", true},
		{"008613800138000", "13800138000", true},
		{"12800138000", "", false}, // 第二位非法
		{"+14155552671", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, err := NormalizeAndValidateMobile(c.in, true)
		if c.ok {
			if err != nil || got != c.want {
				t.Fatalf("cnOnly in=%q got=%q err=%v want=%q", c.in, got, err, c.want)
			}
		} else if err == nil {
			t.Fatalf("cnOnly in=%q expect error, got %q", c.in, got)
		}
	}
}

func TestNormalizeAndValidateMobile_International(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"13800138000", "+8613800138000", true},
		{"+86 13800138000", "+8613800138000", true},
		{"+14155552671", "+14155552671", true},
		// 无 + 的北美号与大陆 11 位规则歧义，必须写 +1...
		{"14155552671", "+8614155552671", true}, // 按大陆号便捷规则（与 1[3-9] 重合）
		{"442071838750", "", false},             // 英国号无 + → 拒绝，需 +44...
		{"+442071838750", "+442071838750", true},
		{"+0123", "", false},
		{"123", "", false},
	}
	for _, c := range cases {
		got, err := NormalizeAndValidateMobile(c.in, false)
		if c.ok {
			if err != nil || got != c.want {
				t.Fatalf("intl in=%q got=%q err=%v want=%q", c.in, got, err, c.want)
			}
		} else if err == nil {
			t.Fatalf("intl in=%q expect error, got %q", c.in, got)
		}
	}
}

func TestMobileLookupVariants(t *testing.T) {
	v1 := MobileLookupVariants("13800138000")
	if len(v1) != 2 || v1[0] != "13800138000" || v1[1] != "+8613800138000" {
		t.Fatalf("variants cn = %#v", v1)
	}
	v2 := MobileLookupVariants("+8613800138000")
	if len(v2) != 2 || v2[0] != "+8613800138000" || v2[1] != "13800138000" {
		t.Fatalf("variants e164 = %#v", v2)
	}
	v3 := MobileLookupVariants("+14155552671")
	if len(v3) != 1 || v3[0] != "+14155552671" {
		t.Fatalf("variants us = %#v", v3)
	}
}

func TestFormatPhoneForSMS(t *testing.T) {
	if got := FormatPhoneForSMS("+8613800138000"); got != "8613800138000" {
		t.Fatalf("got %q", got)
	}
	if got := FormatPhoneForSMS("13800138000"); got != "13800138000" {
		t.Fatalf("got %q", got)
	}
}
