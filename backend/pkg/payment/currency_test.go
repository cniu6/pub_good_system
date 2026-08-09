package payment

import (
	"math"
	"testing"
)

const floatEpsilon = 0.001

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) <= floatEpsilon
}

func TestConvertToTarget(t *testing.T) {
	tests := []struct {
		name                 string
		sourceAmount         float64
		sourceCurrency       string
		targetCurrency       string
		exchangeRate         float64
		exchangeFixed        float64
		feeRate              int
		feeFixed             float64
		feeMode              string
		wantTargetAmount     float64
		wantTargetFee        float64
		wantTargetPayAmount  float64
		wantTargetCredit     float64
	}{
		{
			name:                "add 模式：100 CNY -> 13.72 USD，用户应付 13.99，商户到账 13.72",
			sourceAmount:        100,
			sourceCurrency:      "CNY",
			targetCurrency:      "USD",
			exchangeRate:        0.1372,
			exchangeFixed:       0,
			feeRate:             2,
			feeFixed:            0,
			feeMode:             "add",
			wantTargetAmount:    13.72,
			wantTargetFee:       0.27,
			wantTargetPayAmount: 13.99,
			wantTargetCredit:    13.72,
		},
		{
			name:                "include 模式：98 CNY -> 13.45 USD，到账 13.18",
			sourceAmount:        98,
			sourceCurrency:      "CNY",
			targetCurrency:      "USD",
			exchangeRate:        0.1372,
			exchangeFixed:       0,
			feeRate:             2,
			feeFixed:            0,
			feeMode:             "include",
			wantTargetAmount:    13.45,
			wantTargetFee:       0.27,
			wantTargetPayAmount: 13.45,
			wantTargetCredit:    13.18,
		},
		{
			name:                "目标币种为空或同币种时不转换，直接按目标费率计算",
			sourceAmount:        100,
			sourceCurrency:      "CNY",
			targetCurrency:      "",
			exchangeRate:        0,
			exchangeFixed:       0,
			feeRate:             2,
			feeFixed:            0,
			feeMode:             "add",
			wantTargetAmount:    100,
			wantTargetFee:       2,
			wantTargetPayAmount: 102,
			wantTargetCredit:    100,
		},
		{
			name:                "固定加额 + 固定手续费：100 * 0.1372 + 0.05 = 13.77，再加 0.1 固定费，add 应付 13.87",
			sourceAmount:        100,
			sourceCurrency:      "CNY",
			targetCurrency:      "USD",
			exchangeRate:        0.1372,
			exchangeFixed:       0.05,
			feeRate:             0,
			feeFixed:            0.10,
			feeMode:             "add",
			wantTargetAmount:    13.77,
			wantTargetFee:       0.10,
			wantTargetPayAmount: 13.87,
			wantTargetCredit:    13.77,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ConvertToTarget(
				tt.sourceAmount,
				tt.sourceCurrency,
				tt.targetCurrency,
				ExchangeRateModeSystem,
				tt.exchangeRate,
				tt.exchangeFixed,
				tt.feeRate,
				tt.feeFixed,
				tt.feeMode,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !approxEqual(res.TargetAmount, tt.wantTargetAmount) {
				t.Errorf("TargetAmount = %.2f, want %.2f", res.TargetAmount, tt.wantTargetAmount)
			}
			if !approxEqual(res.TargetFee, tt.wantTargetFee) {
				t.Errorf("TargetFee = %.2f, want %.2f", res.TargetFee, tt.wantTargetFee)
			}
			if !approxEqual(res.TargetPayAmount, tt.wantTargetPayAmount) {
				t.Errorf("TargetPayAmount = %.2f, want %.2f", res.TargetPayAmount, tt.wantTargetPayAmount)
			}
			if !approxEqual(res.TargetCredit, tt.wantTargetCredit) {
				t.Errorf("TargetCredit = %.2f, want %.2f", res.TargetCredit, tt.wantTargetCredit)
			}
		})
	}
}
