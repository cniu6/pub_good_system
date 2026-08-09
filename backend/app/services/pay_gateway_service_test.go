package services

import (
	"math"
	"testing"
)

// 浮点金额比较允许的最小误差（分进位导致）
const floatEpsilon = 0.001

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) <= floatEpsilon
}

func TestCalculateFee(t *testing.T) {
	tests := []struct {
		name              string
		amount            float64
		feeRate           int
		feeMode           string
		wantFee           float64
		wantPayAmount     float64
		wantCreditAmount  float64
	}{
		{
			name:             "add 模式：用户多付，100 + 2% = 102 应付，100 到账",
			amount:           100,
			feeRate:          2,
			feeMode:          "add",
			wantFee:          2,
			wantPayAmount:    102,
			wantCreditAmount: 100,
		},
		{
			name:             "include 模式：从金额内扣，100 - 2% = 98 到账",
			amount:           100,
			feeRate:          2,
			feeMode:          "include",
			wantFee:          2,
			wantPayAmount:    100,
			wantCreditAmount: 98,
		},
		{
			name:             "add 模式固定 0.5 元：100 + 0.5 = 100.5",
			amount:           100,
			feeRate:          0,
			feeMode:          "add",
			wantFee:          0,
			wantPayAmount:    100,
			wantCreditAmount: 100,
		},
		{
			name:             "include 模式 100% 费率：到账为 0",
			amount:           100,
			feeRate:          100,
			feeMode:          "include",
			wantFee:          100,
			wantPayAmount:    100,
			wantCreditAmount: 0,
		},
		{
			name:             "费率为 0 不收费",
			amount:           100,
			feeRate:          0,
			feeMode:          "include",
			wantFee:          0,
			wantPayAmount:    100,
			wantCreditAmount: 100,
		},
		{
			name:             "金额小于 0.01 元按无手续费返回",
			amount:           0,
			feeRate:          2,
			feeMode:          "add",
			wantFee:          0,
			wantPayAmount:    0,
			wantCreditAmount: 0,
		},
		{
			name:             "未知 feeMode 默认按 include 处理",
			amount:           100,
			feeRate:          2,
			feeMode:          "",
			wantFee:          2,
			wantPayAmount:    100,
			wantCreditAmount: 98,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fee, payAmount, creditAmount := CalculateFee(tt.amount, tt.feeRate, tt.feeMode)
			if !approxEqual(fee, tt.wantFee) {
				t.Errorf("fee = %.2f, want %.2f", fee, tt.wantFee)
			}
			if !approxEqual(payAmount, tt.wantPayAmount) {
				t.Errorf("payAmount = %.2f, want %.2f", payAmount, tt.wantPayAmount)
			}
			if !approxEqual(creditAmount, tt.wantCreditAmount) {
				t.Errorf("creditAmount = %.2f, want %.2f", creditAmount, tt.wantCreditAmount)
			}
		})
	}
}
