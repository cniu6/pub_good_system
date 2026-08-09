package epay

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"fst/backend/app/models"
)

func TestGenerateSign(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
		key    string
	}{
		{
			name: "标准参数签名",
			params: map[string]string{
				"pid":          "1001",
				"type":         "alipay",
				"out_trade_no": "P20240101120000123456",
				"notify_url":   "https://example.com/api/v1/public/payment/notify",
				"return_url":   "https://example.com/api/v1/public/payment/return",
				"name":         "余额充值",
				"money":        "10.00",
			},
			key: "testkey123",
		},
		{
			name: "空值字段应被过滤",
			params: map[string]string{
				"pid":          "1001",
				"type":         "wxpay",
				"out_trade_no": "P20240101120000999999",
				"notify_url":   "https://example.com/notify",
				"return_url":   "",
				"name":         "充值",
				"money":        "50.00",
			},
			key: "mykey456",
		},
		{
			name: "sign和sign_type应被过滤",
			params: map[string]string{
				"pid":          "1001",
				"type":         "alipay",
				"out_trade_no": "P20240101120000111111",
				"name":         "测试",
				"money":        "1.00",
				"sign":         "should_be_ignored",
				"sign_type":    "MD5",
			},
			key: "key789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sign := GenerateSign(tt.params, tt.key)
			if sign == "" {
				t.Error("签名结果不应为空")
			}
			if len(sign) != 32 {
				t.Errorf("MD5签名长度应为32，实际为 %d", len(sign))
			}
			sign2 := GenerateSign(tt.params, tt.key)
			if sign != sign2 {
				t.Errorf("相同参数两次签名不一致: %s != %s", sign, sign2)
			}
		})
	}
}

func TestGenerateSign_Deterministic(t *testing.T) {
	params := map[string]string{
		"pid":          "1",
		"type":         "alipay",
		"out_trade_no": "123",
		"name":         "test",
		"money":        "1.00",
	}
	key := "abc"

	sign1 := GenerateSign(params, key)
	sign2 := GenerateSign(params, key)

	if sign1 != sign2 {
		t.Fatalf("确定性签名失败: %s != %s", sign1, sign2)
	}

	params["money"] = "2.00"
	sign3 := GenerateSign(params, key)
	if sign1 == sign3 {
		t.Error("不同参数产生了相同签名")
	}

	params["money"] = "1.00"
	sign4 := GenerateSign(params, "different_key")
	if sign1 == sign4 {
		t.Error("不同密钥产生了相同签名")
	}
}

func TestVerifySign(t *testing.T) {
	key := "test_secret_key"

	params := map[string]string{
		"pid":          "1001",
		"type":         "alipay",
		"out_trade_no": "P20240101120000123456",
		"name":         "余额充值",
		"money":        "10.00",
		"trade_no":     "2024010112345678",
		"trade_status": "TRADE_SUCCESS",
	}

	sign := GenerateNotifySign(params, key)
	params["sign"] = sign
	params["sign_type"] = "MD5"

	t.Run("正确签名验证通过", func(t *testing.T) {
		if !VerifySign(params, key) {
			t.Error("正确签名应验证通过")
		}
	})

	t.Run("兼容包含type的回调签名", func(t *testing.T) {
		genericParams := make(map[string]string)
		for k, v := range params {
			genericParams[k] = v
		}
		genericParams["sign"] = GenerateSign(genericParams, key)
		if !VerifySign(genericParams, key) {
			t.Error("包含type的回调签名也应验证通过")
		}
	})

	t.Run("错误签名验证失败", func(t *testing.T) {
		badParams := make(map[string]string)
		for k, v := range params {
			badParams[k] = v
		}
		badParams["sign"] = "0000000000000000000000000000000"
		if VerifySign(badParams, key) {
			t.Error("错误签名不应验证通过")
		}
	})

	t.Run("篡改金额后验证失败", func(t *testing.T) {
		tamperedParams := make(map[string]string)
		for k, v := range params {
			tamperedParams[k] = v
		}
		tamperedParams["money"] = "99999.00"
		if VerifySign(tamperedParams, key) {
			t.Error("篡改金额后签名不应验证通过")
		}
	})

	t.Run("篡改订单号后验证失败", func(t *testing.T) {
		tamperedParams := make(map[string]string)
		for k, v := range params {
			tamperedParams[k] = v
		}
		tamperedParams["out_trade_no"] = "FAKE_ORDER"
		if VerifySign(tamperedParams, key) {
			t.Error("篡改订单号后签名不应验证通过")
		}
	})

	t.Run("错误密钥验证失败", func(t *testing.T) {
		if VerifySign(params, "wrong_key") {
			t.Error("错误密钥不应验证通过")
		}
	})

	t.Run("缺少sign字段验证失败", func(t *testing.T) {
		noSignParams := make(map[string]string)
		for k, v := range params {
			if k != "sign" {
				noSignParams[k] = v
			}
		}
		if VerifySign(noSignParams, key) {
			t.Error("缺少sign字段不应验证通过")
		}
	})

	t.Run("空sign字段验证失败", func(t *testing.T) {
		emptySignParams := make(map[string]string)
		for k, v := range params {
			emptySignParams[k] = v
		}
		emptySignParams["sign"] = ""
		if VerifySign(emptySignParams, key) {
			t.Error("空sign字段不应验证通过")
		}
	})
}

func TestGenerateSign_EmptyParams(t *testing.T) {
	t.Run("空参数map", func(t *testing.T) {
		sign := GenerateSign(map[string]string{}, "key")
		if sign == "" {
			t.Error("空参数也应生成签名（仅含key的MD5）")
		}
	})

	t.Run("所有值为空的参数", func(t *testing.T) {
		params := map[string]string{
			"pid":  "",
			"type": "",
		}
		sign := GenerateSign(params, "key")
		if sign == "" {
			t.Error("所有值为空时也应生成签名")
		}
	})
}

func TestValidatePayType(t *testing.T) {
	config := &Config{
		PaymentTypes: []string{"alipay", "wxpay", "qqpay"},
	}

	tests := []struct {
		paymentType string
		expected    bool
	}{
		{"alipay", true},
		{"wxpay", true},
		{"qqpay", true},
		{"bankcard", false},
		{"", false},
		{"ALIPAY", false},
	}

	for _, tt := range tests {
		t.Run(tt.paymentType, func(t *testing.T) {
			result := ValidatePayType(config, tt.paymentType)
			if result != tt.expected {
				t.Errorf("ValidatePayType(%q) = %v, want %v", tt.paymentType, result, tt.expected)
			}
		})
	}
}

func TestChannelValidatePayType(t *testing.T) {
	ch := NewChannel()
	gw := &models.PayGateway{PayType: "alipay"}
	if !ch.ValidatePayType(gw, "alipay") {
		t.Fatal("expected alipay to be valid")
	}
	if ch.ValidatePayType(gw, "wxpay") {
		t.Fatal("expected wxpay to be invalid for alipay-only gateway")
	}
}

func TestSignFilterFields(t *testing.T) {
	key := "mykey"

	baseParams := map[string]string{
		"pid":   "1",
		"money": "10.00",
	}
	baseSig := GenerateSign(baseParams, key)

	withSignParams := map[string]string{
		"pid":       "1",
		"money":     "10.00",
		"sign":      "whatever",
		"sign_type": "MD5",
	}
	withSignSig := GenerateSign(withSignParams, key)

	if baseSig != withSignSig {
		t.Errorf("sign/sign_type 未被正确过滤: base=%s, withSign=%s", baseSig, withSignSig)
	}

	withEmptyParams := map[string]string{
		"pid":     "1",
		"money":   "10.00",
		"name":    "",
		"type":    "",
		"garbage": "",
	}
	withEmptySig := GenerateSign(withEmptyParams, key)

	if baseSig != withEmptySig {
		t.Errorf("空值字段未被正确过滤: base=%s, withEmpty=%s", baseSig, withEmptySig)
	}
}

func TestQueryOrder_FallbackIdentifiers(t *testing.T) {
	queries := make([]url.Values, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query())
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("out_trade_no") != "" {
			_, _ = w.Write([]byte(`{"code":-3,"msg":"order not found"}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":1,"msg":"success","trade_no":"TN123","out_trade_no":"P123","type":false,"money":"10.00","status":"1"}`))
	}))
	defer server.Close()

	result, err := QueryOrder(&Config{
		ApiURL: server.URL,
		PID:    "1001",
		Key:    "secret-key",
	}, "P123", "TN123")
	if err != nil {
		t.Fatalf("QueryOrder returned error: %v", err)
	}
	if result == nil {
		t.Fatal("QueryOrder returned nil result")
	}
	if len(queries) != 2 {
		t.Fatalf("expected 2 query attempts, got %d", len(queries))
	}
	if got := queries[0].Get("out_trade_no"); got != "P123" {
		t.Fatalf("expected first query to use out_trade_no, got %q", got)
	}
	if got := queries[0].Get("key"); got != "secret-key" {
		t.Fatalf("expected first query to include key, got %q", got)
	}
	if got := queries[0].Get("sign"); got != "" {
		t.Fatalf("expected first query not to include sign, got %q", got)
	}
	if got := queries[1].Get("trade_no"); got != "TN123" {
		t.Fatalf("expected second query to use trade_no, got %q", got)
	}
	if result.Code != 1 || result.OutTradeNo != "P123" || result.TradeNo != "TN123" {
		t.Fatalf("unexpected query result: %+v", result)
	}
	if result.TradeStatus != "TRADE_SUCCESS" {
		t.Fatalf("expected normalized trade status to be TRADE_SUCCESS, got %q", result.TradeStatus)
	}
	if result.Type != "" {
		t.Fatalf("expected false type to normalize to empty string, got %q", result.Type)
	}
}

func TestSignVerifyRoundTrip(t *testing.T) {
	key := "merchant_secret_key_2024"

	submitParams := map[string]string{
		"pid":          "10086",
		"type":         "alipay",
		"out_trade_no": "P20240615143000123456",
		"notify_url":   "https://mysite.com/api/v1/public/payment/notify",
		"return_url":   "https://mysite.com/api/v1/public/payment/return",
		"name":         "余额充值",
		"money":        "50.00",
	}

	submitSign := GenerateSign(submitParams, key)
	if submitSign == "" {
		t.Fatal("生成提交签名失败")
	}

	callbackParams := map[string]string{
		"pid":          "10086",
		"trade_no":     "EP2024061500001234",
		"out_trade_no": "P20240615143000123456",
		"type":         "alipay",
		"name":         "余额充值",
		"money":        "50.00",
		"trade_status": "TRADE_SUCCESS",
	}

	callbackSign := GenerateNotifySign(callbackParams, key)
	callbackParams["sign"] = callbackSign
	callbackParams["sign_type"] = "MD5"

	if !VerifySign(callbackParams, key) {
		t.Fatal("回调验签失败")
	}

	attackParams := make(map[string]string)
	for k, v := range callbackParams {
		attackParams[k] = v
	}
	attackParams["money"] = "0.01"
	if VerifySign(attackParams, key) {
		t.Fatal("篡改金额后验签不应通过")
	}

	attack2 := make(map[string]string)
	for k, v := range callbackParams {
		attack2[k] = v
	}
	attack2["out_trade_no"] = "ATTACKER_ORDER"
	if VerifySign(attack2, key) {
		t.Fatal("篡改订单号后验签不应通过")
	}

	fakeSign := GenerateNotifySign(callbackParams, "fake_key")
	fakeParams := make(map[string]string)
	for k, v := range callbackParams {
		fakeParams[k] = v
	}
	fakeParams["sign"] = fakeSign
	if VerifySign(fakeParams, key) {
		t.Fatal("伪造密钥签名不应通过验签")
	}
}

func TestChannelType(t *testing.T) {
	ch := NewChannel()
	if ch.Type() != ChannelType || ch.Type() != "epay" {
		t.Fatalf("Type() = %q, want epay", ch.Type())
	}
}

func TestConfigFromGateway(t *testing.T) {
	cfg := ConfigFromGateway(&models.PayGateway{
		ApiURL:    "https://pay.example.com/",
		PID:       "1001",
		ExtConfig: `{"key":"secret"}`,
		PayType:   "alipay",
	})
	if cfg.ApiURL != "https://pay.example.com" {
		t.Fatalf("ApiURL should trim trailing slash, got %q", cfg.ApiURL)
	}
	if cfg.PID != "1001" || cfg.Key != "secret" || len(cfg.PaymentTypes) != 1 || cfg.PaymentTypes[0] != "alipay" {
		t.Fatalf("unexpected config: %+v", cfg)
	}

	empty := ConfigFromGateway(nil)
	if empty == nil || empty.ApiURL != "" {
		t.Fatal("nil gateway should return empty config")
	}
}

func TestCreatePay_APIPaySuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mapi.php" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		_ = r.ParseForm()
		if r.Form.Get("pid") != "1001" || r.Form.Get("out_trade_no") != "P100" {
			t.Errorf("unexpected form: %v", r.Form)
		}
		if r.Form.Get("sign") == "" || r.Form.Get("sign_type") != "MD5" {
			t.Error("missing sign fields")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"msg":"success","payurl":"https://cashier.example/pay","trade_no":"TN999"}`))
	}))
	defer server.Close()

	ch := NewChannel()
	gw := &models.PayGateway{
		ApiURL:    server.URL,
		PID:       "1001",
		ExtConfig: `{"key":"secret-key"}`,
		PayType:   "alipay",
	}
	order := &models.PaymentOrder{
		OrderNo:     "P100",
		PaymentType: "alipay",
		Subject:     "余额充值",
		PayAmount:   10.5,
		ClientIP:    "1.2.3.4",
	}

	payURL, tradeNo, err := ch.CreatePay(gw, order, "https://n.example/notify", "https://r.example/return")
	if err != nil {
		t.Fatalf("CreatePay error: %v", err)
	}
	if payURL != "https://cashier.example/pay" {
		t.Fatalf("payURL = %q", payURL)
	}
	if tradeNo != "TN999" {
		t.Fatalf("tradeNo = %q", tradeNo)
	}
}

func TestCreatePay_FallbackToSubmit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/mapi.php" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`oops`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	ch := NewChannel()
	gw := &models.PayGateway{
		ApiURL:    server.URL,
		PID:       "1001",
		ExtConfig: `{"key":"secret-key"}`,
		PayType:   "wxpay",
	}
	order := &models.PaymentOrder{
		OrderNo:     "P200",
		PaymentType: "wxpay",
		Subject:     "充值",
		PayAmount:   20,
	}

	payURL, tradeNo, err := ch.CreatePay(gw, order, "https://n.example/notify", "https://r.example/return")
	if err != nil {
		t.Fatalf("CreatePay fallback error: %v", err)
	}
	if tradeNo != "" {
		t.Fatalf("fallback submit should not return trade_no, got %q", tradeNo)
	}
	if !strings.Contains(payURL, "/submit.php") {
		t.Fatalf("expected submit.php URL, got %q", payURL)
	}
	if !strings.Contains(payURL, "out_trade_no=P200") {
		t.Fatalf("submit URL missing order no: %q", payURL)
	}
	if !strings.Contains(payURL, "sign=") {
		t.Fatalf("submit URL missing sign: %q", payURL)
	}
}

func TestCreatePay_ConfigIncomplete(t *testing.T) {
	ch := NewChannel()
	_, _, err := ch.CreatePay(&models.PayGateway{}, &models.PaymentOrder{OrderNo: "P1"}, "", "")
	if err == nil {
		t.Fatal("expected error for incomplete config")
	}
}

func TestChannelQueryOrderAndVerifyNotify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":1,"msg":"ok","trade_no":"TN1","out_trade_no":"P1","type":"alipay","money":"9.90","trade_status":"TRADE_SUCCESS"}`))
	}))
	defer server.Close()

	ch := NewChannel()
	gw := &models.PayGateway{ApiURL: server.URL, PID: "1", ExtConfig: `{"key":"k"}`, PayType: "alipay"}

	result, err := ch.QueryOrder(gw, "P1", "")
	if err != nil {
		t.Fatalf("QueryOrder error: %v", err)
	}
	if result == nil || result.Code != 1 || result.TradeNo != "TN1" || result.Money != "9.90" {
		t.Fatalf("unexpected result: %+v", result)
	}

	params := map[string]string{
		"pid":          "1",
		"out_trade_no": "P1",
		"money":        "9.90",
		"trade_status": "TRADE_SUCCESS",
	}
	params["sign"] = GenerateNotifySign(params, "k")
	if !ch.VerifyNotify(params, "k") {
		t.Fatal("VerifyNotify should pass")
	}
	if ch.VerifyNotify(params, "wrong") {
		t.Fatal("VerifyNotify should fail with wrong key")
	}
}

func TestBuildSubmitURL(t *testing.T) {
	url, err := BuildSubmitURL(&Config{
		ApiURL: "https://pay.example.com",
		PID:    "1001",
		Key:    "key",
	}, &models.PaymentOrder{
		OrderNo:     "P9",
		PaymentType: "alipay",
		Subject:     "余额充值",
		PayAmount:   1.5,
	}, "https://n", "https://r")
	if err != nil {
		t.Fatalf("BuildSubmitURL error: %v", err)
	}
	if !strings.HasPrefix(url, "https://pay.example.com/submit.php?") {
		t.Fatalf("unexpected url prefix: %q", url)
	}
	if !strings.Contains(url, "money=1.50") || !strings.Contains(url, "pid=1001") {
		t.Fatalf("url missing fields: %q", url)
	}
}

func BenchmarkGenerateSign(b *testing.B) {
	params := map[string]string{
		"pid":          "1001",
		"type":         "alipay",
		"out_trade_no": "P20240101120000123456",
		"notify_url":   "https://example.com/api/v1/public/payment/notify",
		"return_url":   "https://example.com/api/v1/public/payment/return",
		"name":         "余额充值",
		"money":        "10.00",
	}
	key := "benchmark_key_12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateSign(params, key)
	}
}

func BenchmarkVerifySign(b *testing.B) {
	key := "benchmark_key_12345"
	params := map[string]string{
		"pid":          "1001",
		"type":         "alipay",
		"out_trade_no": "P20240101120000123456",
		"name":         "余额充值",
		"money":        "10.00",
		"trade_no":     "2024010112345678",
		"trade_status": "TRADE_SUCCESS",
	}
	params["sign"] = GenerateNotifySign(params, key)
	params["sign_type"] = "MD5"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifySign(params, key)
	}
}
