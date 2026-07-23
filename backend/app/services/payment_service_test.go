package services

import (
	"fst/backend/app/models"
	"testing"
)

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

// TestValidateCallbackMoney 对齐生产逻辑：按「分」整数精确比较（Round），不再使用 float 容差。
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

	t.Run("金额完全一致应通过", func(t *testing.T) {
		if err := validateCallbackMoney(10, "10"); err != nil {
			t.Fatalf("expected exact amount to pass, got %v", err)
		}
	})

	// 亚分噪声经 Round 仍落到同一分，应通过
	t.Run("亚分噪声四舍五入到同一分应通过", func(t *testing.T) {
		for _, money := range []string{"10.0005", "10.0009", "10.001", "10.004"} {
			if err := validateCallbackMoney(10, money); err != nil {
				t.Fatalf("money=%s should pass (same fen), got %v", money, err)
			}
		}
	})

	// 差至少 1 分必须拒绝
	t.Run("差1分及以上应拒绝", func(t *testing.T) {
		for _, money := range []string{"10.01", "10.02", "10.005", "9.99"} {
			err := validateCallbackMoney(10, money)
			if err == nil || err.Error() != "回调金额与订单金额不一致" {
				t.Fatalf("money=%s should reject, got %v", money, err)
			}
		}
	})
}

func TestValidatePaymentOrderDeletion(t *testing.T) {
	// 物理删除已禁用；保留状态校验函数供历史兼容/测试
	allowed := []int{models.PaymentStatusCanceled, models.PaymentStatusFailed}
	for _, status := range allowed {
		if err := validatePaymentOrderDeletion(status); err != nil {
			t.Fatalf("status %d should be deletable by validator, got %v", status, err)
		}
	}

	denied := []int{models.PaymentStatusPending, models.PaymentStatusPaid, models.PaymentStatusRefunded}
	for _, status := range denied {
		if err := validatePaymentOrderDeletion(status); err == nil {
			t.Fatalf("status %d should not be deletable", status)
		}
	}

	if err := AdminDeleteOrder(1); err == nil || !IsClientError(err) {
		t.Fatalf("AdminDeleteOrder should be disabled, got %v", err)
	}
}
