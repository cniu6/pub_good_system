package services

import (
	"testing"
)

// TestGenerateEpaySign 测试签名生成
func TestGenerateEpaySign(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]string
		key      string
		expected string
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
			sign := GenerateEpaySign(tt.params, tt.key)
			if sign == "" {
				t.Error("签名结果不应为空")
			}
			if len(sign) != 32 {
				t.Errorf("MD5签名长度应为32，实际为 %d", len(sign))
			}
			// 验证签名一致性：相同参数应产生相同签名
			sign2 := GenerateEpaySign(tt.params, tt.key)
			if sign != sign2 {
				t.Errorf("相同参数两次签名不一致: %s != %s", sign, sign2)
			}
		})
	}
}

// TestGenerateEpaySign_Deterministic 测试签名确定性（手动计算验证）
func TestGenerateEpaySign_Deterministic(t *testing.T) {
	// 手动构造一个简单的签名用例
	// 参数按 key 排序: money=1.00&name=test&out_trade_no=123&pid=1&type=alipay + key
	params := map[string]string{
		"pid":          "1",
		"type":         "alipay",
		"out_trade_no": "123",
		"name":         "test",
		"money":        "1.00",
	}
	key := "abc"

	sign1 := GenerateEpaySign(params, key)
	sign2 := GenerateEpaySign(params, key)

	if sign1 != sign2 {
		t.Fatalf("确定性签名失败: %s != %s", sign1, sign2)
	}

	// 改变一个参数，签名应不同
	params["money"] = "2.00"
	sign3 := GenerateEpaySign(params, key)
	if sign1 == sign3 {
		t.Error("不同参数产生了相同签名")
	}

	// 改变 key，签名应不同
	params["money"] = "1.00"
	sign4 := GenerateEpaySign(params, "different_key")
	if sign1 == sign4 {
		t.Error("不同密钥产生了相同签名")
	}
}

// TestVerifyEpaySign 测试签名验证
func TestVerifyEpaySign(t *testing.T) {
	key := "test_secret_key"

	// 构造参数并生成签名
	params := map[string]string{
		"pid":          "1001",
		"type":         "alipay",
		"out_trade_no": "P20240101120000123456",
		"name":         "余额充值",
		"money":        "10.00",
		"trade_no":     "2024010112345678",
		"trade_status": "TRADE_SUCCESS",
	}

	// 生成正确签名
	sign := GenerateEpayNotifySign(params, key)
	params["sign"] = sign
	params["sign_type"] = "MD5"

	t.Run("正确签名验证通过", func(t *testing.T) {
		if !VerifyEpaySign(params, key) {
			t.Error("正确签名应验证通过")
		}
	})

	t.Run("兼容包含type的回调签名", func(t *testing.T) {
		genericParams := make(map[string]string)
		for k, v := range params {
			genericParams[k] = v
		}
		genericParams["sign"] = GenerateEpaySign(genericParams, key)
		if !VerifyEpaySign(genericParams, key) {
			t.Error("包含type的回调签名也应验证通过")
		}
	})

	t.Run("错误签名验证失败", func(t *testing.T) {
		badParams := make(map[string]string)
		for k, v := range params {
			badParams[k] = v
		}
		badParams["sign"] = "0000000000000000000000000000000"
		if VerifyEpaySign(badParams, key) {
			t.Error("错误签名不应验证通过")
		}
	})

	t.Run("篡改金额后验证失败", func(t *testing.T) {
		tamperedParams := make(map[string]string)
		for k, v := range params {
			tamperedParams[k] = v
		}
		tamperedParams["money"] = "99999.00" // 篡改金额
		if VerifyEpaySign(tamperedParams, key) {
			t.Error("篡改金额后签名不应验证通过")
		}
	})

	t.Run("篡改订单号后验证失败", func(t *testing.T) {
		tamperedParams := make(map[string]string)
		for k, v := range params {
			tamperedParams[k] = v
		}
		tamperedParams["out_trade_no"] = "FAKE_ORDER" // 篡改订单号
		if VerifyEpaySign(tamperedParams, key) {
			t.Error("篡改订单号后签名不应验证通过")
		}
	})

	t.Run("错误密钥验证失败", func(t *testing.T) {
		if VerifyEpaySign(params, "wrong_key") {
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
		if VerifyEpaySign(noSignParams, key) {
			t.Error("缺少sign字段不应验证通过")
		}
	})

	t.Run("空sign字段验证失败", func(t *testing.T) {
		emptySignParams := make(map[string]string)
		for k, v := range params {
			emptySignParams[k] = v
		}
		emptySignParams["sign"] = ""
		if VerifyEpaySign(emptySignParams, key) {
			t.Error("空sign字段不应验证通过")
		}
	})
}

// TestGenerateEpaySign_EmptyParams 测试空参数处理
func TestGenerateEpaySign_EmptyParams(t *testing.T) {
	t.Run("空参数map", func(t *testing.T) {
		sign := GenerateEpaySign(map[string]string{}, "key")
		if sign == "" {
			t.Error("空参数也应生成签名（仅含key的MD5）")
		}
	})

	t.Run("所有值为空的参数", func(t *testing.T) {
		params := map[string]string{
			"pid":  "",
			"type": "",
		}
		sign := GenerateEpaySign(params, "key")
		if sign == "" {
			t.Error("所有值为空时也应生成签名")
		}
	})
}

// TestValidatePaymentType 测试支付方式验证
func TestValidatePaymentType(t *testing.T) {
	config := &EpayConfig{
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
		{"ALIPAY", false}, // 大小写敏感
	}

	for _, tt := range tests {
		t.Run(tt.paymentType, func(t *testing.T) {
			result := ValidatePaymentType(config, tt.paymentType)
			if result != tt.expected {
				t.Errorf("ValidatePaymentType(%q) = %v, want %v", tt.paymentType, result, tt.expected)
			}
		})
	}
}

// TestSignFilterFields 测试签名时过滤 sign/sign_type/空值
func TestSignFilterFields(t *testing.T) {
	key := "mykey"

	// 不含 sign/sign_type 的参数
	baseParams := map[string]string{
		"pid":   "1",
		"money": "10.00",
	}
	baseSig := GenerateEpaySign(baseParams, key)

	// 含 sign/sign_type 的参数（应被过滤，签名一致）
	withSignParams := map[string]string{
		"pid":       "1",
		"money":     "10.00",
		"sign":      "whatever",
		"sign_type": "MD5",
	}
	withSignSig := GenerateEpaySign(withSignParams, key)

	if baseSig != withSignSig {
		t.Errorf("sign/sign_type 未被正确过滤: base=%s, withSign=%s", baseSig, withSignSig)
	}

	// 含空值的参数（空值应被过滤，签名一致）
	withEmptyParams := map[string]string{
		"pid":     "1",
		"money":   "10.00",
		"name":    "",
		"type":    "",
		"garbage": "",
	}
	withEmptySig := GenerateEpaySign(withEmptyParams, key)

	if baseSig != withEmptySig {
		t.Errorf("空值字段未被正确过滤: base=%s, withEmpty=%s", baseSig, withEmptySig)
	}
}

// BenchmarkGenerateEpaySign 签名性能基准测试
func BenchmarkGenerateEpaySign(b *testing.B) {
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
		GenerateEpaySign(params, key)
	}
}

// BenchmarkVerifyEpaySign 验签性能基准测试
func BenchmarkVerifyEpaySign(b *testing.B) {
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
	params["sign"] = GenerateEpayNotifySign(params, key)
	params["sign_type"] = "MD5"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyEpaySign(params, key)
	}
}
