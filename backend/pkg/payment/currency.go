package payment

import (
	"encoding/json"
	"fmt"
	"fst/backend/utils"
	"math"
	"strconv"
	"strings"
)

// ExchangeRateMode 汇率模式
const (
	ExchangeRateModeSystem  = "system"  // 使用系统汇率
	ExchangeRateModeFixed   = "fixed"   // 使用固定倍率
	ExchangeRateModeDynamic = "dynamic" // 使用实时动态汇率
)

// FeeMode 手续费模式
const (
	FeeModeAdd     = "add"     // 用户多付（手续费加到支付金额上）
	FeeModeInclude = "include" // 内扣（手续费从目标金额中扣除）
)

// ExchangeRate 汇率记录（内存/缓存/DB 通用结构）
type ExchangeRate struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Rate         float64 `json:"rate"`
	RateType     string  `json:"rate_type"`
	Source       string  `json:"source"`
	UpdatedAt    int64   `json:"updated_at"`
}

// RateKey 返回货币对唯一 key，如 CNY:USD
func (r *ExchangeRate) RateKey() string {
	return RateKey(r.FromCurrency, r.ToCurrency)
}

// RateKey 拼接货币对 key
func RateKey(from, to string) string {
	return fmt.Sprintf("%s:%s", strings.ToUpper(strings.TrimSpace(from)), strings.ToUpper(strings.TrimSpace(to)))
}

// NormalizeCurrency 规范化货币代码
func NormalizeCurrency(c string) string {
	return strings.ToUpper(strings.TrimSpace(c))
}

// ExchangeRates 内存汇率表（从 DB 或 settings 加载后缓存）
type ExchangeRates struct {
	BaseCurrency string                  `json:"base_currency"`
	Rates        map[string]ExchangeRate `json:"rates"`
	UpdatedAt    int64                   `json:"updated_at"`
}

// MarshalRates JSON 序列化
func (er *ExchangeRates) Marshal() string {
	b, _ := json.Marshal(er)
	return string(b)
}

// UnmarshalRates 从 JSON 反序列化
func UnmarshalRates(s string) (*ExchangeRates, error) {
	var er ExchangeRates
	if s == "" {
		er.Rates = make(map[string]ExchangeRate)
		return &er, nil
	}
	if err := json.Unmarshal([]byte(s), &er); err != nil {
		return nil, err
	}
	if er.Rates == nil {
		er.Rates = make(map[string]ExchangeRate)
	}
	return &er, nil
}

// TargetMoneyResult 转换 + 目标手续费计算结果
type TargetMoneyResult struct {
	SourceAmount    float64 // 原币种金额
	SourceCurrency  string  // 原币种
	TargetAmount    float64 // 转换后目标币种本金
	TargetCurrency  string  // 目标币种
	ExchangeRate    float64 // 实际使用汇率
	ExchangeFixed   float64 // 固定加额
	TargetFee       float64 // 目标通道手续费
	TargetPayAmount float64 // 用户实际支付金额（目标币种）
	TargetCredit    float64 // 到账/商品金额（目标币种）
	TargetFeeRate   int     // 费率（百分之 x）
	TargetFeeFixed  float64 // 固定手续费
	TargetFeeMode   string  // add / include
}

// ConvertToTarget 把源币种金额转换为目标币种，并计算目标通道手续费
// sourceAmount: 源币种金额
// sourceCurrency: 源币种（如 CNY）
// targetCurrency: 目标币种（如 USD）
// exchangeRate: 汇率倍率（>0），0 表示不转换
// exchangeFixed: 转换后固定加额
// feeRate: 目标手续费率（百分之 x，如 200 = 2%）
// feeFixed: 目标固定手续费
// feeMode: add / include
func ConvertToTarget(sourceAmount float64, sourceCurrency, targetCurrency, exchangeRateMode string, exchangeRate, exchangeFixed float64, feeRate int, feeFixed float64, feeMode string) (*TargetMoneyResult, error) {
	sourceCurrency = NormalizeCurrency(sourceCurrency)
	targetCurrency = NormalizeCurrency(targetCurrency)

	if sourceAmount <= 0 {
		return nil, fmt.Errorf("source amount must be greater than 0")
	}

	// 目标币种未设置或与源币种相同，无需转换
	actualRate := 1.0
	targetAmount := sourceAmount
	if targetCurrency != "" && targetCurrency != sourceCurrency {
		if exchangeRate <= 0 {
			return nil, fmt.Errorf("exchange rate not configured for %s -> %s", sourceCurrency, targetCurrency)
		}
		actualRate = exchangeRate
		targetAmount = sourceAmount*exchangeRate + exchangeFixed
	}

	// 使用分精度避免浮点误差
	targetFen, err := utils.YuanToFen(targetAmount)
	if err != nil {
		return nil, err
	}

	feeFen := int64(math.Round(float64(targetFen) * float64(feeRate) / 100.0))
	fixedFeeFen, err := utils.YuanToFen(feeFixed)
	if err != nil {
		fixedFeeFen = 0
	}
	feeFen += fixedFeeFen
	if feeFen < 0 {
		feeFen = 0
	}

	var payFen, creditFen int64
	mode := strings.ToLower(strings.TrimSpace(feeMode))
	if mode == FeeModeAdd {
		payFen = targetFen + feeFen
		creditFen = targetFen
	} else {
		payFen = targetFen
		creditFen = targetFen - feeFen
		if creditFen < 0 {
			creditFen = 0
		}
	}

	return &TargetMoneyResult{
		SourceAmount:    sourceAmount,
		SourceCurrency:  sourceCurrency,
		TargetAmount:    utils.FenToYuan(targetFen),
		TargetCurrency:  targetCurrency,
		ExchangeRate:    actualRate,
		ExchangeFixed:   exchangeFixed,
		TargetFee:       utils.FenToYuan(feeFen),
		TargetPayAmount: utils.FenToYuan(payFen),
		TargetCredit:    utils.FenToYuan(creditFen),
		TargetFeeRate:   feeRate,
		TargetFeeFixed:  feeFixed,
		TargetFeeMode:   mode,
	}, nil
}

// yuanToFen 元转分
func yuanToFen(yuan float64) (int64, error) {
	s := fmt.Sprintf("%.4f", yuan)
	d, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(d * 100)), nil
}

// yuanToFenFast 元转分（快速，忽略错误，用于固定手续费）
func yuanToFenFast(yuan float64) int64 {
	n, _ := yuanToFen(yuan)
	return n
}

// fenToYuan 分转元
func fenToYuan(fen int64) float64 {
	return float64(fen) / 100.0
}

// DefaultBaseCurrency 默认本位币
const DefaultBaseCurrency = "CNY"

// ExchangeRateSources 内置动态汇率源别名
var ExchangeRateSources = map[string]string{
	"exchangerate-api": "https://api.exchangerate-api.com/v4/latest/%s",
}
