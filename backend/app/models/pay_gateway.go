package models

import (
	"database/sql"
	"errors"
	"fst/backend/pkg/db"
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

// PayGateway 支付通道模型
type PayGateway struct {
	ID                   uint64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name                 string  `gorm:"column:name;size:100;not null;default:''" json:"name"`
	Type                 string  `gorm:"column:type;size:50;not null;default:'epay'" json:"type"`
	PayType              string  `gorm:"column:pay_type;size:50;not null;default:''" json:"pay_type"`
	SignType             string  `gorm:"column:sign_type;size:50;not null;default:''" json:"sign_type"`
	Version              string  `gorm:"column:version;size:50;not null;default:''" json:"version"`
	Device               string  `gorm:"column:device;size:50;not null;default:'pc'" json:"device"`
	Currency             string  `gorm:"column:currency;size:10;not null;default:'CNY'" json:"currency"`
	Description          string  `gorm:"column:description;size:500;not null;default:''" json:"description"`
	Status               int     `gorm:"column:status;not null;default:0;index:idx_pg_status;index:idx_pg_status_sort_id,priority:1" json:"status"`
	ApiURL               string  `gorm:"column:api_url;type:text;not null" json:"api_url"`
	PID                  string  `gorm:"column:pid;type:text;not null" json:"pid"`
	Key                  string  `gorm:"column:key;type:text;not null" json:"key,omitempty"` // 兼容旧通道：单密钥模式
	ExtConfig            string  `gorm:"column:ext_config;type:text" json:"ext_config"`      // JSON 扩展配置，多版本/多密钥
	LogoURL              string  `gorm:"column:logo_url;type:text;not null" json:"logo_url"`
	SortOrder            int     `gorm:"column:sort_order;not null;default:0;index:idx_pg_sort_order;index:idx_pg_status_sort_id,priority:2" json:"sort_order"`
	MinAmount            float64 `gorm:"column:min_amount;type:decimal(10,2);not null;default:0" json:"min_amount"`
	MaxAmount            float64 `gorm:"column:max_amount;type:decimal(10,2);not null;default:10000" json:"max_amount"`
	FeeRate              int     `gorm:"column:fee_rate;not null;default:0" json:"fee_rate"`
	FeeMode              string  `gorm:"column:fee_mode;size:50;not null;default:''" json:"fee_mode"`
	MinLevel             int     `gorm:"column:min_level;not null;default:0" json:"min_level"`
	NotifyURL            string  `gorm:"column:notify_url;type:text;not null" json:"notify_url"`
	ExpireMinutes        int     `gorm:"column:expire_minutes;not null;default:0" json:"expire_minutes"`
	ActiveQueryEnabled   int     `gorm:"column:active_query_enabled;not null;default:1" json:"active_query_enabled"`
	QueryIntervalSeconds int     `gorm:"column:query_interval_seconds;not null;default:120" json:"query_interval_seconds"`
	QueryBatchSize       int     `gorm:"column:query_batch_size;not null;default:50" json:"query_batch_size"`
	CreateTime           int64   `gorm:"column:create_time;not null;default:0" json:"create_time"`
	UpdateTime           int64   `gorm:"column:update_time;not null;default:0" json:"update_time"`
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
