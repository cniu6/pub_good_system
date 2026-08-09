package models

import (
	"database/sql"
	"errors"
	"fmt"
	"fst/backend/pkg/db"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ExchangeRate 全局汇率表（固定/动态汇率统一落库）
type ExchangeRate struct {
	ID           uint64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	FromCurrency string  `gorm:"column:from_currency;size:10;not null;index:idx_er_from_to,unique,priority:1" json:"from_currency"`
	ToCurrency   string  `gorm:"column:to_currency;size:10;not null;index:idx_er_from_to,unique,priority:2" json:"to_currency"`
	Rate         float64 `gorm:"column:rate;type:decimal(18,8);not null;default:0" json:"rate"`
	FixedAmount  float64 `gorm:"column:fixed_amount;type:decimal(18,8);not null;default:0" json:"fixed_amount"`
	RateType     string  `gorm:"column:rate_type;size:20;not null;default:'fixed'" json:"rate_type"`
	Source       string  `gorm:"column:source;size:255;not null;default:''" json:"source"`
	CreateTime   int64   `gorm:"column:create_time;not null;default:0" json:"create_time"`
	UpdateTime   int64   `gorm:"column:update_time;not null;default:0" json:"update_time"`
}

// TableName 表名
func (ExchangeRate) TableName() string { return "exchange_rates" }

// NormalizeExchangeRate 规范化货币代码
func NormalizeExchangeRate(from, to string) (string, string) {
	return strings.ToUpper(strings.TrimSpace(from)), strings.ToUpper(strings.TrimSpace(to))
}

// CreateExchangeRate 创建汇率
func CreateExchangeRate(rate *ExchangeRate) error {
	rate.FromCurrency, rate.ToCurrency = NormalizeExchangeRate(rate.FromCurrency, rate.ToCurrency)
	now := time.Now().Unix()
	rate.CreateTime = now
	rate.UpdateTime = now
	return db.DB.Create(rate).Error
}

// UpdateExchangeRate 更新汇率
func UpdateExchangeRate(rate *ExchangeRate) error {
	rate.FromCurrency, rate.ToCurrency = NormalizeExchangeRate(rate.FromCurrency, rate.ToCurrency)
	rate.UpdateTime = time.Now().Unix()
	return db.DB.Save(rate).Error
}

// DeleteExchangeRate 删除汇率
func DeleteExchangeRate(id uint64) error {
	return db.DB.Delete(&ExchangeRate{}, id).Error
}

// GetExchangeRateByID 按 ID 取汇率
func GetExchangeRateByID(id uint64) (*ExchangeRate, error) {
	var rate ExchangeRate
	err := db.DB.Where("id = ?", id).First(&rate).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

// GetExchangeRateByPair 按货币对取汇率
func GetExchangeRateByPair(from, to string) (*ExchangeRate, error) {
	from, to = NormalizeExchangeRate(from, to)
	var rate ExchangeRate
	err := db.DB.Where("from_currency = ? AND to_currency = ?", from, to).First(&rate).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

// ListExchangeRates 列出全部或按货币筛选
func ListExchangeRates(from, to string) ([]ExchangeRate, error) {
	q := db.DB.Model(&ExchangeRate{})
	if from != "" {
		q = q.Where("from_currency = ?", strings.ToUpper(strings.TrimSpace(from)))
	}
	if to != "" {
		q = q.Where("to_currency = ?", strings.ToUpper(strings.TrimSpace(to)))
	}
	var rates []ExchangeRate
	err := q.Order("from_currency, to_currency").Find(&rates).Error
	return rates, err
}

// SeedExchangeRates 内置常用汇率种子（启动时无则插入，有则不动）
func SeedExchangeRates() {
	now := time.Now().Unix()
	rates := []ExchangeRate{
		{FromCurrency: "USD", ToCurrency: "CNY", Rate: 7.2500, RateType: "fixed", Source: "builtin-seed", CreateTime: now, UpdateTime: now},
		{FromCurrency: "CNY", ToCurrency: "USD", Rate: 0.1379, RateType: "fixed", Source: "builtin-seed", CreateTime: now, UpdateTime: now},
		{FromCurrency: "USD", ToCurrency: "EUR", Rate: 0.9200, RateType: "fixed", Source: "builtin-seed", CreateTime: now, UpdateTime: now},
		{FromCurrency: "EUR", ToCurrency: "USD", Rate: 1.0870, RateType: "fixed", Source: "builtin-seed", CreateTime: now, UpdateTime: now},
		{FromCurrency: "USD", ToCurrency: "JPY", Rate: 150.0000, RateType: "fixed", Source: "builtin-seed", CreateTime: now, UpdateTime: now},
		{FromCurrency: "JPY", ToCurrency: "USD", Rate: 0.0067, RateType: "fixed", Source: "builtin-seed", CreateTime: now, UpdateTime: now},
		{FromCurrency: "USD", ToCurrency: "GBP", Rate: 0.7900, RateType: "fixed", Source: "builtin-seed", CreateTime: now, UpdateTime: now},
		{FromCurrency: "GBP", ToCurrency: "USD", Rate: 1.2658, RateType: "fixed", Source: "builtin-seed", CreateTime: now, UpdateTime: now},
		{FromCurrency: "USD", ToCurrency: "HKD", Rate: 7.8000, RateType: "fixed", Source: "builtin-seed", CreateTime: now, UpdateTime: now},
		{FromCurrency: "HKD", ToCurrency: "USD", Rate: 0.1282, RateType: "fixed", Source: "builtin-seed", CreateTime: now, UpdateTime: now},
		{FromCurrency: "USD", ToCurrency: "TWD", Rate: 32.0000, RateType: "fixed", Source: "builtin-seed", CreateTime: now, UpdateTime: now},
		{FromCurrency: "TWD", ToCurrency: "USD", Rate: 0.0313, RateType: "fixed", Source: "builtin-seed", CreateTime: now, UpdateTime: now},
	}

	for _, r := range rates {
		_, err := GetExchangeRateByPair(r.FromCurrency, r.ToCurrency)
		if err == nil {
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("[SeedExchangeRates] query %s -> %s failed: %v", r.FromCurrency, r.ToCurrency, err)
			continue
		}
		if err := CreateExchangeRate(&r); err != nil {
			log.Printf("[SeedExchangeRates] insert %s -> %s failed: %v", r.FromCurrency, r.ToCurrency, err)
		}
	}
}

// GetExchangeRate 取最近一条可用汇率：优先 direct，否则尝试反向
// 返回 (rate, fixed_amount, error)：fixed_amount 是汇率自身的固定加额/差价
func GetExchangeRate(from, to string) (float64, float64, error) {
	from, to = NormalizeExchangeRate(from, to)
	if from == to {
		return 1, 0, nil
	}

	// 1. 直接汇率
	rate, err := GetExchangeRateByPair(from, to)
	if err == nil && rate.Rate > 0 {
		return rate.Rate, rate.FixedAmount, nil
	}

	// 2. 反向汇率取倒数；固定加额不反向折算，避免货币单位混乱
	rate, err = GetExchangeRateByPair(to, from)
	if err == nil && rate.Rate > 0 {
		return 1.0 / rate.Rate, 0, nil
	}

	return 0, 0, fmt.Errorf("exchange rate not found for %s -> %s", from, to)
}
