package middleware

import (
	"net/url"
	"strings"
	"testing"
)

func TestIsPaymentCallbackPath(t *testing.T) {
	for _, path := range []string{
		"/api/v1/public/payment/notify",
		"/api/v1/public/payment/notify/epay",
		"/api/v1/public/payment/return",
	} {
		if !isPaymentCallbackPath(path) {
			t.Errorf("%s 应判定为支付回调", path)
		}
	}
	for _, path := range []string{
		"/api/v1/public/payment",
		"/api/v1/public/settings",
		"/api/v1/user/payment/orders",
	} {
		if isPaymentCallbackPath(path) {
			t.Errorf("%s 不应判定为支付回调", path)
		}
	}
}

// 回调 query 里的签名必须打码，订单号/金额/状态这类对账字段要保持原样。
func TestRedactPaymentQueryMasksSignature(t *testing.T) {
	raw := "out_trade_no=ORDER123&trade_no=T1&money=9.90&sign=deadbeef&sign_type=MD5&trade_status=TRADE_SUCCESS"

	got := redactPaymentQuery(raw)
	if strings.Contains(got, "deadbeef") {
		t.Fatalf("签名未打码: %s", got)
	}

	values, err := url.ParseQuery(got)
	if err != nil {
		t.Fatalf("打码后 query 不可解析: %v", err)
	}
	if values.Get("sign") != "***" {
		t.Errorf("sign = %q, want ***", values.Get("sign"))
	}
	if values.Get("out_trade_no") != "ORDER123" || values.Get("money") != "9.90" || values.Get("sign_type") != "MD5" {
		t.Errorf("对账字段被误改: %s", got)
	}
}
