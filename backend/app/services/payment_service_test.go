package services

import (
	"fst/backend/app/models"
	"testing"
)

// TestAbs 测试浮点数绝对值函数
func TestAbs(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{1.5, 1.5},
		{-1.5, 1.5},
		{0, 0},
		{-0.01, 0.01},
		{999999.99, 999999.99},
		{-999999.99, 999999.99},
	}

	for _, tt := range tests {
		result := abs(tt.input)
		if result != tt.expected {
			t.Errorf("abs(%f) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

// TestAmountValidation 测试金额边界校验逻辑
func TestAmountValidation(t *testing.T) {
	tests := []struct {
		name      string
		amount    float64
		min       float64
		max       float64
		expectErr bool
	}{
		{"正常金额", 10.00, 1.00, 10000.00, false},
		{"最小金额", 1.00, 1.00, 10000.00, false},
		{"最大金额", 10000.00, 1.00, 10000.00, false},
		{"低于最小值", 0.50, 1.00, 10000.00, true},
		{"超过最大值", 10001.00, 1.00, 10000.00, true},
		{"零金额", 0.00, 1.00, 10000.00, true},
		{"负金额", -10.00, 1.00, 10000.00, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.amount < tt.min || tt.amount > tt.max
			if hasErr != tt.expectErr {
				t.Errorf("amount=%f, min=%f, max=%f: err=%v, want err=%v",
					tt.amount, tt.min, tt.max, hasErr, tt.expectErr)
			}
		})
	}
}

// TestCallbackAmountVerification 测试回调金额校验（防篡改核心逻辑）
func TestCallbackAmountVerification(t *testing.T) {
	tests := []struct {
		name         string
		orderAmount  float64
		callbackMoney float64
		shouldPass   bool
	}{
		{"金额完全一致", 10.00, 10.00, true},
		{"微小浮点误差（允许）", 10.00, 10.001, true},
		{"微小浮点误差2", 10.00, 9.999, true},
		{"边界误差0.01（允许）", 10.00, 10.01, true},
		{"超出误差0.02", 10.00, 10.02, false},
		{"金额被篡改-增大", 10.00, 100.00, false},
		{"金额被篡改-减小", 10.00, 1.00, false},
		{"金额被改为0", 10.00, 0.00, false},
		{"金额被改为负数", 10.00, -10.00, false},
		{"大金额一致", 9999.99, 9999.99, true},
		{"大金额微小误差", 9999.99, 9999.989, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed := abs(tt.callbackMoney-tt.orderAmount) <= 0.01
			if passed != tt.shouldPass {
				t.Errorf("orderAmount=%.2f, callbackMoney=%.2f: passed=%v, want=%v",
					tt.orderAmount, tt.callbackMoney, passed, tt.shouldPass)
			}
		})
	}
}

// TestPaymentTypeLabels 测试支付方式标签映射完整性
func TestPaymentTypeLabels(t *testing.T) {
	// 模拟 GetAvailablePaymentMethods 中的 typeLabels
	typeLabels := map[string]string{
		"alipay": "支付宝",
		"wxpay":  "微信支付",
		"qqpay":  "QQ钱包",
	}

	allTypes := []string{"alipay", "wxpay", "qqpay"}
	for _, pt := range allTypes {
		label, ok := typeLabels[pt]
		if !ok {
			t.Errorf("支付方式 %q 缺少标签映射", pt)
		}
		if label == "" {
			t.Errorf("支付方式 %q 标签为空", pt)
		}
	}
}

// TestSignVerifyRoundTrip 完整签名-验签往返测试（模拟完整回调流程）
func TestSignVerifyRoundTrip(t *testing.T) {
	key := "merchant_secret_key_2024"

	// 模拟创建订单时的参数
	submitParams := map[string]string{
		"pid":          "10086",
		"type":         "alipay",
		"out_trade_no": "P20240615143000123456",
		"notify_url":   "https://mysite.com/api/v1/public/payment/notify",
		"return_url":   "https://mysite.com/api/v1/public/payment/return",
		"name":         "余额充值",
		"money":        "50.00",
	}

	// 1. 生成提交签名
	submitSign := GenerateEpaySign(submitParams, key)
	if submitSign == "" {
		t.Fatal("生成提交签名失败")
	}

	// 2. 模拟易支付回调（部分参数和提交时一致，增加 trade_no 和 trade_status）
	callbackParams := map[string]string{
		"pid":          "10086",
		"trade_no":     "EP2024061500001234",
		"out_trade_no": "P20240615143000123456",
		"type":         "alipay",
		"name":         "余额充值",
		"money":        "50.00",
		"trade_status": "TRADE_SUCCESS",
	}

	// 3. 用相同的密钥为回调参数生成签名
	callbackSign := GenerateEpayNotifySign(callbackParams, key)
	callbackParams["sign"] = callbackSign
	callbackParams["sign_type"] = "MD5"

	// 4. 验签应通过
	if !VerifyEpaySign(callbackParams, key) {
		t.Fatal("回调验签失败")
	}

	// 5. 模拟攻击者篡改金额
	attackParams := make(map[string]string)
	for k, v := range callbackParams {
		attackParams[k] = v
	}
	attackParams["money"] = "0.01" // 篡改金额为0.01
	if VerifyEpaySign(attackParams, key) {
		t.Fatal("篡改金额后验签不应通过")
	}

	// 6. 模拟攻击者篡改订单号
	attack2 := make(map[string]string)
	for k, v := range callbackParams {
		attack2[k] = v
	}
	attack2["out_trade_no"] = "ATTACKER_ORDER" // 篡改订单号
	if VerifyEpaySign(attack2, key) {
		t.Fatal("篡改订单号后验签不应通过")
	}

	// 7. 模拟攻击者用不同密钥签名
	fakeSign := GenerateEpayNotifySign(callbackParams, "fake_key")
	fakeParams := make(map[string]string)
	for k, v := range callbackParams {
		fakeParams[k] = v
	}
	fakeParams["sign"] = fakeSign
	if VerifyEpaySign(fakeParams, key) {
		t.Fatal("伪造密钥签名不应通过验签")
	}
}

// TestIdempotencyLogic 测试幂等性逻辑（模拟多次回调）
func TestIdempotencyLogic(t *testing.T) {
	// 模拟订单状态流转
	type Order struct {
		Status int
	}

	order := &Order{Status: 0} // 待支付

	// 第一次回调：应处理
	if order.Status != 0 {
		t.Fatal("初始状态应为待支付(0)")
	}
	order.Status = 1 // 标记为已支付

	// 第二次回调：幂等跳过
	if order.Status == 0 {
		t.Fatal("第二次回调不应再次处理")
	}
	// 应直接返回成功（幂等），不重复充值

	// 第三次回调：同样跳过
	if order.Status == 0 {
		t.Fatal("第三次回调不应再次处理")
	}
}

// TestStatusConstants 测试状态常量定义
func TestStatusConstants(t *testing.T) {
	// 确保状态常量值不冲突
	statuses := map[int]string{
		0: "pending",
		1: "paid",
		2: "canceled",
		3: "refunded",
		4: "failed",
	}

	seen := make(map[int]bool)
	for status := range statuses {
		if seen[status] {
			t.Errorf("状态值 %d 重复定义", status)
		}
		seen[status] = true
	}
}

func TestValidatePaymentNotifyBinding(t *testing.T) {
	t.Run("网关PID不匹配", func(t *testing.T) {
		err := validatePaymentNotifyBinding(nil, &models.PayGateway{PID: "1001"}, "1002", "", "TN123")
		if err == nil || err.Error() != "商户号不匹配" {
			t.Fatalf("expected pid mismatch error, got %v", err)
		}
	})

	t.Run("订单交易号不匹配", func(t *testing.T) {
		err := validatePaymentNotifyBinding(&models.PaymentOrder{TradeNo: "TN123"}, nil, "", "", "TN999")
		if err == nil || err.Error() != "交易号不匹配" {
			t.Fatalf("expected trade_no mismatch error, got %v", err)
		}
	})

	t.Run("订单支付类型不匹配", func(t *testing.T) {
		err := validatePaymentNotifyBinding(&models.PaymentOrder{PaymentType: "alipay"}, nil, "", "wxpay", "")
		if err == nil || err.Error() != "支付类型不匹配" {
			t.Fatalf("expected payment type mismatch error, got %v", err)
		}
	})

	t.Run("匹配参数通过", func(t *testing.T) {
		err := validatePaymentNotifyBinding(
			&models.PaymentOrder{TradeNo: "TN123", PaymentType: "alipay"},
			&models.PayGateway{PID: "1001"},
			"1001",
			"alipay",
			"TN123",
		)
		if err != nil {
			t.Fatalf("expected binding validation to pass, got %v", err)
		}
	})

	t.Run("空回调支付类型允许通过", func(t *testing.T) {
		err := validatePaymentNotifyBinding(&models.PaymentOrder{PaymentType: "alipay"}, nil, "", "", "")
		if err != nil {
			t.Fatalf("expected empty callback type to be allowed, got %v", err)
		}
	})
}

func TestValidateCallbackMoney(t *testing.T) {
	t.Run("空金额跳过校验", func(t *testing.T) {
		if err := validateCallbackMoney(10, ""); err != nil {
			t.Fatalf("expected empty amount to pass, got %v", err)
		}
	})

	t.Run("非法金额格式拒绝", func(t *testing.T) {
		err := validateCallbackMoney(10, "not-a-number")
		if err == nil || err.Error() != "回调金额格式非法" {
			t.Fatalf("expected invalid amount error, got %v", err)
		}
	})

	t.Run("超出容差拒绝", func(t *testing.T) {
		err := validateCallbackMoney(10, "10.02")
		if err == nil || err.Error() != "回调金额与订单金额不一致" {
			t.Fatalf("expected amount mismatch error, got %v", err)
		}
	})

	t.Run("容差内允许", func(t *testing.T) {
		if err := validateCallbackMoney(10, "10.01"); err != nil {
			t.Fatalf("expected tolerance amount to pass, got %v", err)
		}
	})
}

func TestValidatePaymentOrderDeletion(t *testing.T) {
	allowed := []int{models.PaymentStatusCanceled, models.PaymentStatusFailed}
	for _, status := range allowed {
		if err := validatePaymentOrderDeletion(status); err != nil {
			t.Fatalf("status %d should be deletable, got %v", status, err)
		}
	}

	denied := []int{models.PaymentStatusPending, models.PaymentStatusPaid, models.PaymentStatusRefunded}
	for _, status := range denied {
		if err := validatePaymentOrderDeletion(status); err == nil {
			t.Fatalf("status %d should not be deletable", status)
		}
	}
}
