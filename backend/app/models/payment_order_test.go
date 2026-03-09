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

// TestPaymentOrderStruct 测试订单结构体字段
func TestPaymentOrderStruct(t *testing.T) {
	order := PaymentOrder{
		OrderNo:        "P20240101120000123456",
		UserID:         1,
		TradeNo:        "EP123456",
		PaymentChannel: "epay",
		PaymentType:    "alipay",
		Amount:         10.00,
		Subject:        "余额充值",
		Status:         PaymentStatusPending,
		ClientIP:       "127.0.0.1",
	}

	if order.OrderNo == "" {
		t.Error("OrderNo 不应为空")
	}
	if order.UserID != 1 {
		t.Errorf("UserID = %d, want 1", order.UserID)
	}
	if order.Amount != 10.00 {
		t.Errorf("Amount = %f, want 10.00", order.Amount)
	}
	if order.Status != PaymentStatusPending {
		t.Errorf("Status = %d, want %d (Pending)", order.Status, PaymentStatusPending)
	}
}

// TestPaymentStatusConstants 测试订单状态常量
func TestPaymentStatusConstants(t *testing.T) {
	if PaymentStatusPending != 0 {
		t.Errorf("PaymentStatusPending = %d, want 0", PaymentStatusPending)
	}
	if PaymentStatusPaid != 1 {
		t.Errorf("PaymentStatusPaid = %d, want 1", PaymentStatusPaid)
	}
	if PaymentStatusCanceled != 2 {
		t.Errorf("PaymentStatusCanceled = %d, want 2", PaymentStatusCanceled)
	}
	if PaymentStatusRefunded != 3 {
		t.Errorf("PaymentStatusRefunded = %d, want 3", PaymentStatusRefunded)
	}
	if PaymentStatusFailed != 4 {
		t.Errorf("PaymentStatusFailed = %d, want 4", PaymentStatusFailed)
	}

	// 确保状态值互不相同
	statuses := []int{
		PaymentStatusPending,
		PaymentStatusPaid,
		PaymentStatusCanceled,
		PaymentStatusRefunded,
		PaymentStatusFailed,
	}
	seen := make(map[int]bool)
	for _, s := range statuses {
		if seen[s] {
			t.Errorf("状态值 %d 重复", s)
		}
		seen[s] = true
	}
}

// TestPaymentStatsStruct 测试统计结构体
func TestPaymentStatsStruct(t *testing.T) {
	stats := PaymentStats{
		TotalOrders:   100,
		PaidOrders:    80,
		TotalAmount:   5000.00,
		TodayOrders:   5,
		TodayAmount:   300.00,
		PendingOrders: 3,
	}

	if stats.TotalOrders != 100 {
		t.Errorf("TotalOrders = %d, want 100", stats.TotalOrders)
	}
	if stats.PaidOrders != 80 {
		t.Errorf("PaidOrders = %d, want 80", stats.PaidOrders)
	}
	if stats.TotalAmount != 5000.00 {
		t.Errorf("TotalAmount = %f, want 5000.00", stats.TotalAmount)
	}
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

// BenchmarkGenerateOrderNo 订单号生成性能基准
func BenchmarkGenerateOrderNo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateOrderNo()
	}
}
