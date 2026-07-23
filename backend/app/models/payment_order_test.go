package models

import (
	"strings"
	"testing"
	"time"
)

// TestGenerateOrderNo 测试订单号生成
func TestGenerateOrderNo(t *testing.T) {
	t.Run("订单号格式正确", func(t *testing.T) {
		orderNo := GenerateOrderNo()

		// 以 P 开头
		if !strings.HasPrefix(orderNo, "P") {
			t.Errorf("订单号应以 P 开头: %s", orderNo)
		}

		// 长度: P + 14位时间 + 4位序列 + 4位随机 = 23
		if len(orderNo) != 23 {
			t.Errorf("订单号长度应为23，实际为 %d: %s", len(orderNo), orderNo)
		}
	})

	t.Run("订单号唯一性", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			orderNo := GenerateOrderNo()
			if seen[orderNo] {
				t.Errorf("发现重复订单号: %s (在 %d 次迭代中)", orderNo, i)
			}
			seen[orderNo] = true
		}
	})

	t.Run("订单号包含时间信息", func(t *testing.T) {
		now := time.Now()
		orderNo := GenerateOrderNo()

		// 提取日期部分 (P + YYYYMMDD)
		dateStr := orderNo[1:9]
		expectedDate := now.Format("20060102")

		if dateStr != expectedDate {
			t.Errorf("订单号日期部分不正确: got %s, want %s", dateStr, expectedDate)
		}
	})
}

func TestNormalizeTradeNo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "空字符串", input: "", expected: ""},
		{name: "纯空白", input: "  ", expected: ""},
		{name: "标准交易号", input: "2026030823244167397", expected: "2026030823244167397"},
		{name: "保留前后空格的真实交易号", input: "  EP123456789  ", expected: "EP123456789"},
		{name: "占位符TRADE_NO", input: "TRADE_NO", expected: ""},
		{name: "带前缀符号的占位符", input: "/TRADE_NO", expected: ""},
		{name: "占位符OUT_TRADE_NO", input: "OUT_TRADE_NO", expected: ""},
		{name: "NULL占位符", input: " null ", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeTradeNo(tt.input); got != tt.expected {
				t.Fatalf("NormalizeTradeNo(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCanTransitionPaymentStatus(t *testing.T) {
	tests := []struct {
		name     string
		from     int
		to       int
		expected bool
	}{
		{name: "待支付可变已支付", from: PaymentStatusPending, to: PaymentStatusPaid, expected: true},
		{name: "待支付可变已取消", from: PaymentStatusPending, to: PaymentStatusCanceled, expected: true},
		{name: "待支付可变失败", from: PaymentStatusPending, to: PaymentStatusFailed, expected: true},
		{name: "待支付保持待支付允许幂等", from: PaymentStatusPending, to: PaymentStatusPending, expected: true},
		{name: "已支付可变已退款", from: PaymentStatusPaid, to: PaymentStatusRefunded, expected: true},
		{name: "已支付不能改已取消", from: PaymentStatusPaid, to: PaymentStatusCanceled, expected: false},
		{name: "已支付不能改失败", from: PaymentStatusPaid, to: PaymentStatusFailed, expected: false},
		{name: "已取消不能回待支付", from: PaymentStatusCanceled, to: PaymentStatusPending, expected: false},
		{name: "已取消可迟到恢复已支付", from: PaymentStatusCanceled, to: PaymentStatusPaid, expected: true},
		{name: "已失败不能回待支付", from: PaymentStatusFailed, to: PaymentStatusPending, expected: false},
		{name: "已失败可迟到恢复已支付", from: PaymentStatusFailed, to: PaymentStatusPaid, expected: true},
		{name: "已支付不能回待支付", from: PaymentStatusPaid, to: PaymentStatusPending, expected: false},
		{name: "已退款不能回已支付", from: PaymentStatusRefunded, to: PaymentStatusPaid, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canTransitionPaymentStatus(tt.from, tt.to); got != tt.expected {
				t.Fatalf("canTransitionPaymentStatus(%d, %d) = %v, want %v", tt.from, tt.to, got, tt.expected)
			}
		})
	}
}

// BenchmarkGenerateOrderNo 订单号生成性能基准
func BenchmarkGenerateOrderNo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateOrderNo()
	}
}

