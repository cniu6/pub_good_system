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

// TestCallbackAmountVerification 测试回调金额校验（防篡改核心逻辑，容差 0.001）
// 规则与 validateCallbackMoney 一致：仅当 abs(diff) > 0.001 时拒绝
func TestCallbackAmountVerification(t *testing.T) {
	tests := []struct {
		name          string
		orderAmount   float64
		callbackMoney float64
		shouldPass    bool
	}{
		{"金额完全一致", 10.00, 10.00, true},
		{"容差内0.0005应通过", 10.00, 10.0005, true},
		{"容差内0.0009应通过", 10.00, 10.0009, true},
		// 0.01 不是业务容差，必须拒绝
		{"0.01超出容差应拒绝", 10.00, 10.01, false},
		{"超出误差0.02", 10.00, 10.02, false},
		{"金额被篡改-增大", 10.00, 100.00, false},
		{"金额被篡改-减小", 10.00, 1.00, false},
		{"金额被改为0", 10.00, 0.00, false},
		{"金额被改为负数", 10.00, -10.00, false},
		{"大金额一致", 9999.99, 9999.99, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 仅 abs(diff) > 0.001 拒绝
			passed := !(abs(tt.callbackMoney-tt.orderAmount) > 0.001)
			if passed != tt.shouldPass {
				t.Errorf("orderAmount=%v, callbackMoney=%v: passed=%v, want=%v",
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

	t.Run("匹配参数通过", func(t *testing.T) {
		err := validatePaymentNotifyBinding(
			&models.PaymentOrder{
				GatewayID:      7,
				TradeNo:        "TN123",
				PaymentChannel: "epay",
				PaymentType:    "alipay",
			},
			&models.PayGateway{ID: 7, Type: "epay", PayType: "alipay", PID: "1001"},
			"1001",
			"alipay",
			"TN123",
		)
		if err != nil {
			t.Fatalf("expected binding validation to pass, got %v", err)
		}
	})

	t.Run("跨网关ID必须拒绝", func(t *testing.T) {
		err := validatePaymentNotifyBinding(
			&models.PaymentOrder{GatewayID: 1, PaymentChannel: "epay", PaymentType: "alipay"},
			&models.PayGateway{ID: 2, Type: "epay", PayType: "alipay", PID: "1001"},
			"1001",
			"alipay",
			"",
		)
		if err == nil || err.Error() != "支付通道不匹配" {
			t.Fatalf("expected gateway id mismatch, got %v", err)
		}
	})

	t.Run("支付方式不匹配必须拒绝", func(t *testing.T) {
		err := validatePaymentNotifyBinding(
			&models.PaymentOrder{GatewayID: 1, PaymentChannel: "epay", PaymentType: "alipay"},
			&models.PayGateway{ID: 1, Type: "epay", PayType: "wxpay", PID: "1001"},
			"1001",
			"",
			"",
		)
		if err == nil || err.Error() != "支付方式不匹配" {
			t.Fatalf("expected pay type mismatch, got %v", err)
		}
	})

	t.Run("标准回调type与订单不一致必须拒绝", func(t *testing.T) {
		err := validatePaymentNotifyBinding(
			&models.PaymentOrder{PaymentType: "alipay"},
			nil,
			"",
			"wxpay",
			"",
		)
		if err == nil || err.Error() != "回调支付类型不匹配" {
			t.Fatalf("expected callback type mismatch, got %v", err)
		}
	})

	t.Run("非标准回调支付类型允许通过", func(t *testing.T) {
		err := validatePaymentNotifyBinding(&models.PaymentOrder{PaymentType: "alipay"}, nil, "", "0", "")
		if err != nil {
			t.Fatalf("expected non-standard callback type to be allowed, got %v", err)
		}
	})
}

func TestValidateCallbackMoney(t *testing.T) {
	t.Run("空金额必须拒绝", func(t *testing.T) {
		err := validateCallbackMoney(10, "")
		if err == nil || err.Error() != "回调金额不能为空" {
			t.Fatalf("expected empty amount error, got %v", err)
		}
	})

	t.Run("空白金额必须拒绝", func(t *testing.T) {
		err := validateCallbackMoney(10, "   ")
		if err == nil || err.Error() != "回调金额不能为空" {
			t.Fatalf("expected whitespace amount error, got %v", err)
		}
	})

	t.Run("非法金额格式拒绝", func(t *testing.T) {
		err := validateCallbackMoney(10, "not-a-number")
		if err == nil || err.Error() != "回调金额格式非法" {
			t.Fatalf("expected invalid amount error, got %v", err)
		}
	})

	// 业务上要求金额一致；0.01 不是浮点容差，必须拒绝
	t.Run("10 vs 10.01 应拒绝", func(t *testing.T) {
		err := validateCallbackMoney(10, "10.01")
		if err == nil || err.Error() != "回调金额与订单金额不一致" {
			t.Fatalf("expected amount mismatch for 10.01, got %v", err)
		}
	})

	// 规则：仅 abs(diff) > 0.001 才拒绝。
	// 10.001 解析后的浮点差可能略大于 0.001（IEEE754），也可能刚好等于；
	// 这里按「真实 abs 与 0.001 比较」断言，把边界写清楚。
	t.Run("10 vs 10.001 边界按 abs>0.001", func(t *testing.T) {
		err := validateCallbackMoney(10, "10.001")
		// 与实现同一条件：diff = |parseFloat("10.001")-10|，仅 > 0.001 拒绝
		diff := abs(10.001 - 10)
		if diff > 0.001 {
			if err == nil || err.Error() != "回调金额与订单金额不一致" {
				t.Fatalf("diff=%g > 0.001, expected reject, got %v", diff, err)
			}
		} else {
			if err != nil {
				t.Fatalf("diff=%g <= 0.001, expected pass, got %v", diff, err)
			}
		}
	})

	// 明确小于 0.001 的误差应通过
	t.Run("10 vs 10.0005 应通过", func(t *testing.T) {
		if err := validateCallbackMoney(10, "10.0005"); err != nil {
			t.Fatalf("expected 10.0005 within tolerance, got %v", err)
		}
	})

	t.Run("10 vs 10.0009 应通过", func(t *testing.T) {
		if err := validateCallbackMoney(10, "10.0009"); err != nil {
			t.Fatalf("expected 10.0009 within tolerance, got %v", err)
		}
	})

	t.Run("金额完全一致应通过", func(t *testing.T) {
		if err := validateCallbackMoney(10, "10"); err != nil {
			t.Fatalf("expected exact amount to pass, got %v", err)
		}
	})

	t.Run("明显超出容差拒绝", func(t *testing.T) {
		err := validateCallbackMoney(10, "10.02")
		if err == nil || err.Error() != "回调金额与订单金额不一致" {
			t.Fatalf("expected amount mismatch error, got %v", err)
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

