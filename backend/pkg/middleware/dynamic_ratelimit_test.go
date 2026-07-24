package middleware

import "testing"

func TestIsGlobalRateLimitExemptPath_PaymentCallback(t *testing.T) {
	cases := []struct {
		path   string
		exempt bool
	}{
		{"/api/v1/public/payment/notify", true},
		{"/api/v1/public/payment/notify/", true},
		{"/api/v1/public/payment/return", true},
		{"/api/v1/public/payment/return/extra", true},
		{"/api/v1/public/payment/orders", false}, // 其它支付公开路径仍走限流
		{"/api/v1/public/login", false},
		{"/api/v1/user/profile", false},
		{"/health", true}, // 非 /api/ 一律豁免（与原先行为一致）
		{"", false},
	}
	for _, tc := range cases {
		got := isGlobalRateLimitExemptPath(tc.path)
		if got != tc.exempt {
			t.Fatalf("path=%q exempt=%v, want %v", tc.path, got, tc.exempt)
		}
	}
}
