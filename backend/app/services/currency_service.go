package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
// 返回 (rate, fixed_amount, error)
func GetExchangeRate(from, to string) (float64, float64, error) {
	from, to = payment.NormalizeCurrency(from), payment.NormalizeCurrency(to)
	if from == to {
		return 1, 0, nil
	}
	return models.GetExchangeRate(from, to)
}

// ResolveExchangeRate 解析通道应使用的汇率
// mode: system / fixed / dynamic
// 返回 (rate, fixed_amount, error)：fixed_amount 是系统汇率记录的固定加额/差价
func ResolveExchangeRate(sourceCurrency, targetCurrency, mode string, fixedRate float64) (float64, float64, error) {
	sourceCurrency = payment.NormalizeCurrency(sourceCurrency)
	targetCurrency = payment.NormalizeCurrency(targetCurrency)

	if sourceCurrency == targetCurrency || targetCurrency == "" {
		return 1, 0, nil
	}

	switch strings.ToLower(strings.TrimSpace(mode)) {
	case payment.ExchangeRateModeFixed:
		if fixedRate <= 0 {
			return 0, 0, fmt.Errorf("fixed exchange rate missing for %s -> %s", sourceCurrency, targetCurrency)
		}
		return fixedRate, 0, nil
	case payment.ExchangeRateModeDynamic:
		// 动态汇率先尝试系统表，系统表没有再实时拉取
		rate, fixed, err := GetExchangeRate(sourceCurrency, targetCurrency)
		if err == nil && rate > 0 {
			return rate, fixed, nil
		}
		rate, err = FetchDynamicExchangeRate(sourceCurrency, targetCurrency)
		if err != nil {
			return 0, 0, err
		}
		// 动态拉取成功后，再从库里读该记录的 fixed_amount
		_, fixed, _ = GetExchangeRate(sourceCurrency, targetCurrency)
		return rate, fixed, nil
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

	// 落库缓存，便于后续查；保留已有的 fixed_amount
	existing, _ := models.GetExchangeRateByPair(from, to)
	upsertRate := &models.ExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         rate,
		RateType:     payment.ExchangeRateModeDynamic,
		Source:       sourceURL,
	}
	if existing != nil {
		upsertRate.FixedAmount = existing.FixedAmount
	}
	if err := UpsertExchangeRate(upsertRate); err != nil {
		log.Printf("[Currency] cache dynamic rate failed: %v", err)
	}

	return rate, nil
}

// UpsertExchangeRate 创建或更新汇率
func UpsertExchangeRate(rate *models.ExchangeRate) error {
	existing, err := models.GetExchangeRateByPair(rate.FromCurrency, rate.ToCurrency)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
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
	targetCurrency := payment.NormalizeCurrency(gateway.GetTargetCurrency())
	if targetCurrency == "" {
		targetCurrency = payment.NormalizeCurrency(gateway.Currency)
	}

	exchangeRate, exchangeFixed, err := ResolveExchangeRate(orderCurrency, targetCurrency, gateway.GetExchangeRateMode(), gateway.GetExchangeRate())
	if err != nil {
		return nil, err
	}

	// 系统/动态汇率的 fixed_amount 与网关自身的 exchange_fixed_amount 累加
	exchangeFixed += gateway.GetExchangeFixedAmount()

	return payment.ConvertToTarget(
		orderAmount,
		orderCurrency,
		targetCurrency,
		gateway.GetExchangeRateMode(),
		exchangeRate,
		exchangeFixed,
		gateway.GetTargetFeeRate(),
		gateway.GetTargetFeeFixed(),
		gateway.GetTargetFeeMode(),
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

// FetchLatestRate 只从上游 API 拉取最新汇率倍率，不落库，也不强制改 rate_type
func FetchLatestRate(from, to string) (float64, string, error) {
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
		return 0, sourceURL, fmt.Errorf("fetch latest rate failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, sourceURL, err
	}

	// 兼容 exchangerate-api 返回格式：{ "rates": { "CNY": 7.2, ... } }
	var payload struct {
		Rates map[string]interface{} `json:"rates"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, sourceURL, fmt.Errorf("parse latest rate failed: %w", err)
	}

	raw, ok := payload.Rates[to]
	if !ok {
		return 0, sourceURL, fmt.Errorf("rate for %s not found in dynamic response", to)
	}
	rate, err := strconv.ParseFloat(fmt.Sprintf("%v", raw), 64)
	if err != nil || rate <= 0 {
		return 0, sourceURL, fmt.Errorf("invalid latest rate value for %s: %v", to, raw)
	}

	return rate, sourceURL, nil
}

// RefreshExchangeRateByID 按 ID 从上游拉取最新汇率并更新，保留原有 rate_type 与 fixed_amount
func RefreshExchangeRateByID(id uint64) (*models.ExchangeRate, error) {
	rate, err := models.GetExchangeRateByID(id)
	if err != nil {
		return nil, err
	}

	newRate, sourceURL, err := FetchLatestRate(rate.FromCurrency, rate.ToCurrency)
	if err != nil {
		return nil, err
	}

	rate.Rate = newRate
	rate.Source = sourceURL
	rate.UpdateTime = time.Now().Unix()

	if err := models.UpdateExchangeRate(rate); err != nil {
		return nil, err
	}

	// 返回最新的 fixed_amount 以保持一致性
	return rate, nil
}

// BatchRefreshPreview 批量获取最新汇率预览（不更新数据库）
type BatchRefreshPreviewItem struct {
	ID           uint64  `json:"id"`
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	OldRate      float64 `json:"old_rate"`
	NewRate      float64 `json:"new_rate"`
	Source       string  `json:"source"`
	Error        string  `json:"error,omitempty"`
}

func BatchRefreshPreview(ids []uint64) ([]BatchRefreshPreviewItem, error) {
	items := make([]BatchRefreshPreviewItem, 0, len(ids))
	for _, id := range ids {
		rate, err := models.GetExchangeRateByID(id)
		if err != nil {
			items = append(items, BatchRefreshPreviewItem{ID: id, Error: "rate not found"})
			continue
		}
		newRate, sourceURL, err := FetchLatestRate(rate.FromCurrency, rate.ToCurrency)
		item := BatchRefreshPreviewItem{
			ID:           id,
			FromCurrency: rate.FromCurrency,
			ToCurrency:   rate.ToCurrency,
			OldRate:      rate.Rate,
			Source:       sourceURL,
		}
		if err != nil {
			item.Error = err.Error()
		} else {
			item.NewRate = newRate
		}
		items = append(items, item)
	}
	return items, nil
}

// BatchRefreshExchangeRates 批量确认后更新数据库
func BatchRefreshExchangeRates(ids []uint64) ([]BatchRefreshPreviewItem, error) {
	items, err := BatchRefreshPreview(ids)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].Error != "" || items[i].NewRate <= 0 {
			continue
		}
		rate, err := models.GetExchangeRateByID(items[i].ID)
		if err != nil {
			items[i].Error = err.Error()
			continue
		}
		rate.Rate = items[i].NewRate
		rate.Source = items[i].Source
		rate.UpdateTime = time.Now().Unix()
		if err := models.UpdateExchangeRate(rate); err != nil {
			items[i].Error = err.Error()
		}
	}
	return items, nil
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
