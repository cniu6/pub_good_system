package payment

import "testing"

// 回调日志只允许保留对账字段，签名类凭据必须打码后再落盘。
func TestRedactCallbackParamsMasksSignature(t *testing.T) {
	params := map[string]string{
		"out_trade_no": "ORDER123",
		"trade_no":     "T123",
		"money":        "9.90",
		"sign":         "abc123signature",
		"sign_type":    "MD5",
	}

	safe := redactCallbackParams(params)

	if safe["sign"] != "***" {
		t.Fatalf("sign 未打码: %q", safe["sign"])
	}
	for k, want := range map[string]string{
		"out_trade_no": "ORDER123",
		"trade_no":     "T123",
		"money":        "9.90",
		"sign_type":    "MD5",
	} {
		if safe[k] != want {
			t.Errorf("%s = %q, want %q", k, safe[k], want)
		}
	}
	if params["sign"] != "abc123signature" {
		t.Error("redactCallbackParams 不应修改入参")
	}
}
