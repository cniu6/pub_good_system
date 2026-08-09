package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"fst/backend/pkg/db"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 支付通道状态常量
const (
	PayGatewayStatusDisabled = 0 // 禁用
	PayGatewayStatusEnabled  = 1 // 启用
)

// 手续费计算方式
const (
	FeeModAdd     = "add"     // 在充值金额基础上加收手续费（用户多付）
	FeeModInclude = "include" // 手续费包含在充值金额中（到账金额减少）
)

// 汇率模式
const (
	ExchangeRateModeSystem  = "system"  // 使用系统全局汇率
	ExchangeRateModeFixed   = "fixed"   // 使用通道固定汇率
	ExchangeRateModeDynamic = "dynamic" // 使用实时动态汇率
)

// PayGateway 支付通道模型
type PayGateway struct {
	ID            uint64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name          string  `gorm:"column:name;size:100;not null;default:''" json:"name"`
	Type          string  `gorm:"column:type;size:50;not null;default:'epay'" json:"type"`
	PayType       string  `gorm:"column:pay_type;size:50;not null;default:''" json:"pay_type"`
	Version       string  `gorm:"column:version;size:50;not null;default:''" json:"version"`
	Device        string  `gorm:"column:device;size:50;not null;default:'pc'" json:"device"`
	Currency      string  `gorm:"column:currency;size:10;not null;default:'CNY'" json:"currency"`
	Description   string  `gorm:"column:description;size:500;not null;default:''" json:"description"`
	Status        int     `gorm:"column:status;not null;default:0;index:idx_pg_status;index:idx_pg_status_sort_id,priority:1" json:"status"`
	ApiURL        string  `gorm:"column:api_url;type:text;not null" json:"api_url"`
	PID           string  `gorm:"column:pid;type:text;not null" json:"pid"`
	ExtConfig     string  `gorm:"column:ext_config;type:text" json:"ext_config"` // JSON 扩展配置，含签名/汇率/目标手续费/查单等
	LogoURL       string  `gorm:"column:logo_url;type:text;not null" json:"logo_url"`
	SortOrder     int     `gorm:"column:sort_order;not null;default:0;index:idx_pg_sort_order;index:idx_pg_status_sort_id,priority:2" json:"sort_order"`
	MinAmount     float64 `gorm:"column:min_amount;type:decimal(10,2);not null;default:0" json:"min_amount"`
	MaxAmount     float64 `gorm:"column:max_amount;type:decimal(10,2);not null;default:10000" json:"max_amount"`
	FeeRate       int     `gorm:"column:fee_rate;not null;default:0" json:"fee_rate"`
	FeeMode       string  `gorm:"column:fee_mode;size:50;not null;default:''" json:"fee_mode"`
	MinLevel      int     `gorm:"column:min_level;not null;default:0" json:"min_level"`
	NotifyURL     string  `gorm:"column:notify_url;type:text;not null" json:"notify_url"`
	ExpireMinutes int     `gorm:"column:expire_minutes;not null;default:0" json:"expire_minutes"`
	CreateTime    int64   `gorm:"column:create_time;not null;default:0" json:"create_time"`
	UpdateTime    int64   `gorm:"column:update_time;not null;default:0" json:"update_time"`
}

// TableName 表名
func (PayGateway) TableName() string { return "pay_gateways" }

// CreatePayGateway 创建支付通道
func CreatePayGateway(gw *PayGateway) error {
	now := time.Now().Unix()
	gw.CreateTime = now
	gw.UpdateTime = now
	return db.DB.Create(gw).Error
}

// GetPayGatewayByID 根据ID获取支付通道
func GetPayGatewayByID(id uint64) (*PayGateway, error) {
	var gw PayGateway
	err := db.DB.Where("id = ?", id).First(&gw).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &gw, nil
}

// UpdatePayGateway 更新支付通道
func UpdatePayGateway(gw *PayGateway) error {
	gw.UpdateTime = time.Now().Unix()
	return db.DB.Save(gw).Error
}

// DeletePayGateway 删除支付通道
func DeletePayGateway(id uint64) error {
	return db.DB.Delete(&PayGateway{}, id).Error
}

func buildPayGatewayQuery(keyword string, onlyEnabled bool) *gorm.DB {
	q := db.DB.Model(&PayGateway{})
	if onlyEnabled {
		q = q.Where("status = ?", PayGatewayStatusEnabled)
	}
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("name LIKE ? OR description LIKE ? OR pay_type LIKE ?", kw, kw, kw)
	}
	return q
}

// GetPayGatewayList 分页获取支付通道列表
func GetPayGatewayList(page, pageSize int, keyword string, onlyEnabled bool) ([]PayGateway, int64, error) {
	q := buildPayGatewayQuery(keyword, onlyEnabled)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var gateways []PayGateway
	err := q.Order("sort_order ASC, id ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&gateways).Error
	return gateways, total, err
}

// GetEnabledPayGateways 获取所有启用的支付通道（不分页，用于用户端）
func GetEnabledPayGateways() ([]PayGateway, error) {
	var gateways []PayGateway
	err := db.DB.Where("status = ?", PayGatewayStatusEnabled).
		Order("sort_order ASC, id ASC").
		Find(&gateways).Error
	return gateways, err
}

// ExtConfigMap 解析 ext_config JSON，并填充已迁移字段的默认值。
// 返回值是 map[string]interface{}，数字保持 json.Number 以便精确转换。
func (g *PayGateway) ExtConfigMap() map[string]interface{} {
	if g == nil {
		return make(map[string]interface{})
	}
	m := parsePayGatewayExtConfigMap(g.ExtConfig)

	// 各字段默认值：仅当 ext_config 中不存在时填充
	defaults := map[string]interface{}{
		"sign_type":              "MD5",
		"key":                    "",
		"target_currency":        "",
		"exchange_rate_mode":     "system",
		"exchange_rate":          0,
		"exchange_fixed_amount":  0,
		"exchange_rate_source":   "",
		"target_fee_rate":        0,
		"target_fee_fixed":       0,
		"target_fee_mode":        "add",
		"active_query_enabled":   1,
		"query_interval_seconds": 120,
		"query_batch_size":       50,
	}
	for k, v := range defaults {
		if _, ok := m[k]; !ok {
			m[k] = v
		}
	}
	return m
}

// parsePayGatewayExtConfigMap 本地解析 ext_config JSON，避免 models 依赖 payment 包产生循环导入。
func parsePayGatewayExtConfigMap(extConfigJSON string) map[string]interface{} {
	result := make(map[string]interface{})
	s := strings.TrimSpace(extConfigJSON)
	if s == "" {
		return result
	}
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	if err := dec.Decode(&result); err != nil {
		return make(map[string]interface{})
	}
	return result
}

// GetSignType 返回签名算法，缺省 MD5
func (g *PayGateway) GetSignType() string {
	return payGatewayString(g.ExtConfigMap(), "sign_type", "MD5")
}

// GetKey 返回通道密钥
func (g *PayGateway) GetKey() string { return payGatewayString(g.ExtConfigMap(), "key", "") }

// GetTargetCurrency 返回目标币种
func (g *PayGateway) GetTargetCurrency() string {
	return payGatewayString(g.ExtConfigMap(), "target_currency", "")
}

// GetExchangeRateMode 返回汇率模式，缺省 system
func (g *PayGateway) GetExchangeRateMode() string {
	return payGatewayString(g.ExtConfigMap(), "exchange_rate_mode", "system")
}

// GetExchangeRate 返回固定汇率
func (g *PayGateway) GetExchangeRate() float64 {
	return payGatewayFloat(g.ExtConfigMap(), "exchange_rate", 0)
}

// GetExchangeFixedAmount 返回固定加额
func (g *PayGateway) GetExchangeFixedAmount() float64 {
	return payGatewayFloat(g.ExtConfigMap(), "exchange_fixed_amount", 0)
}

// GetExchangeRateSource 返回动态汇率源
func (g *PayGateway) GetExchangeRateSource() string {
	return payGatewayString(g.ExtConfigMap(), "exchange_rate_source", "")
}

// GetTargetFeeRate 返回目标手续费率
func (g *PayGateway) GetTargetFeeRate() int {
	return payGatewayInt(g.ExtConfigMap(), "target_fee_rate", 0)
}

// GetTargetFeeFixed 返回目标固定手续费
func (g *PayGateway) GetTargetFeeFixed() float64 {
	return payGatewayFloat(g.ExtConfigMap(), "target_fee_fixed", 0)
}

// GetTargetFeeMode 返回目标手续费模式，缺省 add
func (g *PayGateway) GetTargetFeeMode() string {
	return payGatewayString(g.ExtConfigMap(), "target_fee_mode", "add")
}

// GetActiveQueryEnabled 返回主动查单开关，缺省 1
func (g *PayGateway) GetActiveQueryEnabled() int {
	return payGatewayInt(g.ExtConfigMap(), "active_query_enabled", 1)
}

// GetQueryIntervalSeconds 返回查单间隔，缺省 120
func (g *PayGateway) GetQueryIntervalSeconds() int {
	v := payGatewayInt(g.ExtConfigMap(), "query_interval_seconds", 120)
	if v <= 0 {
		return 120
	}
	return v
}

// GetQueryBatchSize 返回查单批次，缺省 50
func (g *PayGateway) GetQueryBatchSize() int {
	v := payGatewayInt(g.ExtConfigMap(), "query_batch_size", 50)
	if v <= 0 {
		return 50
	}
	return v
}

// 类型转换辅助函数

func payGatewayString(m map[string]interface{}, key, def string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch val := v.(type) {
	case string:
		if strings.TrimSpace(val) == "" {
			return def
		}
		return val
	case json.Number:
		return val.String()
	case float64, float32, int, int64, int32, uint, uint64:
		return fmt.Sprintf("%v", v)
	case bool:
		return strconv.FormatBool(val)
	default:
		r := fmt.Sprintf("%v", v)
		if strings.TrimSpace(r) == "" {
			return def
		}
		return r
	}
}

func payGatewayFloat(m map[string]interface{}, key string, def float64) float64 {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch val := v.(type) {
	case json.Number:
		f, err := val.Float64()
		if err != nil {
			return def
		}
		return f
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		if err != nil {
			return def
		}
		return f
	default:
		return def
	}
}

func payGatewayInt(m map[string]interface{}, key string, def int) int {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch val := v.(type) {
	case json.Number:
		i, err := val.Int64()
		if err != nil {
			return def
		}
		return int(i)
	case float64:
		return int(val)
	case float32:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	case int32:
		return int(val)
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil {
			return def
		}
		return i
	case bool:
		if val {
			return 1
		}
		return 0
	default:
		return def
	}
}
