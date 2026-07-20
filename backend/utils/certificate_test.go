package utils

import "testing"

func TestMaskCertificateNo(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"110101199001011234", "110101********1234"}, // 18位身份证
		{"1234", "***4"},                              // 4位：仅留最后1位
		{"A", "A"},
		{"12345678", "1******8"},   // 8位：首尾各1位
		{"E12345678", "E*******8"}, // 9位护照号：落入5~10位区间，首尾各1位
	}

	for _, c := range cases {
		got := MaskCertificateNo(c.in)
		if got != c.want {
			t.Errorf("MaskCertificateNo(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
