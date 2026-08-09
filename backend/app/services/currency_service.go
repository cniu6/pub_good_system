package services

import (
	"context"
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/payment"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CurrencySettingKey 系统设置里全局汇率配置 key
const (
	CurrencySettingBaseCurrency     = "base_currency"
	CurrencySettingDynamicSource    = "currency_dynamic_source"
	CurrencySettingDynamicSourceURL = "currency_dynamic_source_url"
)

// GetBaseCurrency 获取系统本位币
func GetBaseCurrency() string {
	v, _ := GetSystemSetting(CurrencySettingBaseCurrency)
	if v = strings.ToUpper(strings.TrimSpace(v)); v != "" {
		return v
	}
	return payment.DefaultBaseCurrency
}

// SetBaseCurrency 设置系统本位币
func SetBaseCurrency(currency string) error {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return fmt.Errorf("currency cannot be empty")
	}
	return SetSystemSetting(CurrencySettingBaseCurrency, currency)
}

// GetExchangeRate 获取系统汇率：直接/反向/本位币兜底
// 若通道自己设置了 fixed 或 dynamic，调用方先用通道的；这里只负责系统汇率
func GetExchangeRate(from, to string) (float64, error) {
	from, to = payment.NormalizeCurrency(from), payment.NormalizeCurrency(to)
	if from == to {
		return 1, nil
	}
	return models.GetExchangeRate(from, to)
}

// ResolveExchangeRate 解析通道应使用的汇率
// mode: system / fixed / dynamic
// 返回实际 rate 和固定加额（fixedAmount 取自网关）
func ResolveExchangeRate(sourceCurrency, targetCurrency, mode string, fixedRate float64) (float64, error) {
	sourceCurrency = payment.NormalizeCurrency(sourceCurrency)
	targetCurrency = payment.NormalizeCurrency(targetCurrency)

	if sourceCurrency == targetCurrency || targetCurrency == "" {
		return 1, nil
	}

	switch strings.ToLower(strings.TrimSpace(mode)) {
	case payment.ExchangeRateModeFixed:
		if fixedRate <= 0 {
			return 0, fmt.Errorf("fixed exchange rate missing for %s -> %s", sourceCurrency, targetCurrency)
		}
		return fixedRate, nil
	case payment.ExchangeRateModeDynamic:
		// 动态汇率先尝试系统表，系统表没有再实时拉取
		rate, err := GetExchangeRate(sourceCurrency, targetCurrency)
		if err == nil && rate > 0 {
			return rate, nil
		}
		return FetchDynamicExchangeRate(sourceCurrency, targetCurrency)
	default: // system
		return GetExchangeRate(sourceCurrency, targetCurrency)
	}
}

// FetchDynamicExchangeRate 从外部 API 实时拉取汇率
func FetchDynamicExchangeRate(from, to string) (float64, error) {
	from, to = payment.NormalizeCurrency(from), payment.NormalizeCurrency(to)

	sourceURL := GetSystemSettingOrDefault(CurrencySettingDynamicSourceURL, "")
	source := GetSystemSettingOrDefault(CurrencySettingDynamicSource, "")
	if sourceURL == "" {
		if u, ok := payment.ExchangeRateSources[source]; ok {
			sourceURL = fmt.Sprintf(u, from)
		} else {
			sourceURL = fmt.Sprintf(payment.ExchangeRateSources["exchangerate-api"], from)
		}
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(sourceURL)
	if err != nil {
		return 0, fmt.Errorf("fetch dynamic rate failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// 兼容 exchangerate-api 返回格式：{ "rates": { "CNY": 7.2, ... } }
	var payload struct {
		Rates map[string]interface{} `json:"rates"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, fmt.Errorf("parse dynamic rate failed: %w", err)
	}

	raw, ok := payload.Rates[to]
	if !ok {
		return 0, fmt.Errorf("rate for %s not found in dynamic response", to)
	}
	rate, err := strconv.ParseFloat(fmt.Sprintf("%v", raw), 64)
	if err != nil || rate <= 0 {
		return 0, fmt.Errorf("invalid dynamic rate value for %s: %v", to, raw)
	}

	// 落库缓存，便于后续查
	if err := UpsertExchangeRate(&models.ExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         rate,
		RateType:     payment.ExchangeRateModeDynamic,
		Source:       sourceURL,
	}); err != nil {
		log.Printf("[Currency] cache dynamic rate failed: %v", err)
	}

	return rate, nil
}

// UpsertExchangeRate 创建或更新汇率
func UpsertExchangeRate(rate *models.ExchangeRate) error {
	existing, err := models.GetExchangeRateByPair(rate.FromCurrency, rate.ToCurrency)
	if err != nil && err != fmt.Errorf("sql: no rows in result set") { // sql.ErrNoRows 比较
		return err
	}
	if existing != nil {
		existing.Rate = rate.Rate
		existing.RateType = rate.RateType
		existing.Source = rate.Source
		return models.UpdateExchangeRate(existing)
	}
	return models.CreateExchangeRate(rate)
}

// ConvertOrderAmountToTarget 统一转换入口
// 输入：用户订单金额（orderCurrency）、网关配置
// 输出：目标币种金额 + 目标手续费计算结果
func ConvertOrderAmountToTarget(orderAmount float64, orderCurrency string, gateway *models.PayGateway) (*payment.TargetMoneyResult, error) {
	targetCurrency := payment.NormalizeCurrency(gateway.TargetCurrency)
	if targetCurrency == "" {
		targetCurrency = payment.NormalizeCurrency(gateway.Currency)
	}

	exchangeRate, err := ResolveExchangeRate(orderCurrency, targetCurrency, gateway.ExchangeRateMode, gateway.ExchangeRate)
	if err != nil {
		return nil, err
	}

	return payment.ConvertToTarget(
		orderAmount,
		orderCurrency,
		targetCurrency,
		gateway.ExchangeRateMode,
		exchangeRate,
		gateway.ExchangeFixedAmount,
		gateway.TargetFeeRate,
		gateway.TargetFeeFixed,
		gateway.TargetFeeMode,
	)
}

// GetSystemSettingOrDefault 读系统设置，带默认值
func GetSystemSettingOrDefault(key, defaultValue string) string {
	v, _ := GetSystemSetting(key)
	if v == "" {
		return defaultValue
	}
	return v
}

// RefreshDynamicRates 刷新全部动态汇率（后台手动/定时任务入口）
func RefreshDynamicRates(ctx context.Context) (map[string]float64, error) {
	rates, err := models.ListExchangeRates("", "")
	if err != nil {
		return nil, err
	}
	result := make(map[string]float64)
	for _, r := range rates {
		if r.RateType != payment.ExchangeRateModeDynamic {
			continue
		}
		rate, err := FetchDynamicExchangeRate(r.FromCurrency, r.ToCurrency)
		if err != nil {
			log.Printf("[Currency] refresh %s -> %s failed: %v", r.FromCurrency, r.ToCurrency, err)
			continue
		}
		result[payment.RateKey(r.FromCurrency, r.ToCurrency)] = rate
	}
	return result, nil
}
