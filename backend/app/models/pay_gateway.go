package models

import (
	"fst/backend/internal/db"
	"log"
	"time"
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
	ID          uint64  `db:"id" json:"id"`
	Name        string  `db:"name" json:"name"`                 // 通道名称
	Type        string  `db:"type" json:"type"`                 // 通道类型：epay
	PayType     string  `db:"pay_type" json:"pay_type"`         // 支付方式：alipay/wxpay/qqpay
	Description string  `db:"description" json:"description"`   // 描述/提示信息
	Status      int     `db:"status" json:"status"`             // 状态 0=禁用 1=启用
	ApiURL      string  `db:"api_url" json:"api_url"`           // 支付网关API地址
	PID         string  `db:"pid" json:"pid"`                   // 商户ID
	Key         string  `db:"key" json:"key,omitempty"`         // 商户密钥（用户侧隐藏）
	LogoURL     string  `db:"logo_url" json:"logo_url"`         // 通道Logo图片地址
	SortOrder   int     `db:"sort_order" json:"sort_order"`     // 排序（升序）
	MinAmount   float64 `db:"min_amount" json:"min_amount"`     // 最小充值金额
	MaxAmount   float64 `db:"max_amount" json:"max_amount"`     // 最大充值金额
	FeeRate     int     `db:"fee_rate" json:"fee_rate"`         // 手续费率（百分比 0-100）
	FeeMode     string  `db:"fee_mode" json:"fee_mode"`         // 手续费模式：add=加收 include=包含
	MinLevel    int     `db:"min_level" json:"min_level"`       // 最低等级限制（0=不限制）
	NotifyURL   string  `db:"notify_url" json:"notify_url"`     // 自定义回调地址（留空用全局）
	CreateTime  int64   `db:"create_time" json:"create_time"`
	UpdateTime  int64   `db:"update_time" json:"update_time"`
}

// InitPayGatewaysTable 初始化支付通道表
func InitPayGatewaysTable() {
	if db.CheckTableExists("pay_gateways") {
		db.EnsureIndex("pay_gateways", "idx_status_sort_id", "ALTER TABLE pay_gateways ADD INDEX idx_status_sort_id (status, sort_order, id)")
		return
	}

	schema := `CREATE TABLE IF NOT EXISTS pay_gateways (
		id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		name        VARCHAR(100)     NOT NULL DEFAULT '' COMMENT '通道名称',
		type        VARCHAR(50)      NOT NULL DEFAULT 'epay' COMMENT '通道类型',
		pay_type    VARCHAR(50)      NOT NULL DEFAULT '' COMMENT '支付方式',
		description VARCHAR(500)     NOT NULL DEFAULT '' COMMENT '描述/提示信息',
		status      TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态 0=禁用 1=启用',
		api_url     TEXT             NOT NULL COMMENT 'API地址',
		pid         TEXT             NOT NULL COMMENT '商户ID',
		` + "`key`" + `         TEXT             NOT NULL COMMENT '商户密钥',
		logo_url    TEXT             NOT NULL COMMENT 'Logo图片地址',
		sort_order  INT UNSIGNED     NOT NULL DEFAULT 0 COMMENT '排序',
		min_amount  DECIMAL(10,2)    NOT NULL DEFAULT 0.00 COMMENT '最小充值金额',
		max_amount  DECIMAL(10,2)    NOT NULL DEFAULT 10000.00 COMMENT '最大充值金额',
		fee_rate    INT              NOT NULL DEFAULT 0 COMMENT '手续费率(百分比0-100)',
		fee_mode    VARCHAR(50)      NOT NULL DEFAULT '' COMMENT '手续费模式',
		min_level   INT              NOT NULL DEFAULT 0 COMMENT '最低等级限制',
		notify_url  TEXT             NOT NULL COMMENT '自定义回调地址',
		create_time BIGINT           NOT NULL DEFAULT 0 COMMENT '创建时间',
		update_time BIGINT           NOT NULL DEFAULT 0 COMMENT '更新时间',
		INDEX idx_status (status),
		INDEX idx_sort_order (sort_order),
		INDEX idx_status_sort_id (status, sort_order, id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付通道表';`

	_, err := db.DB.Exec(schema)
	if err != nil {
		log.Printf("[Init] Failed to create pay_gateways table: %v", err)
	} else {
		log.Println("[Init] Created pay_gateways table")
	}
}

// CreatePayGateway 创建支付通道
func CreatePayGateway(gw *PayGateway) error {
	now := time.Now().Unix()
	gw.CreateTime = now
	gw.UpdateTime = now

	result, err := db.DB.Exec(
		"INSERT INTO pay_gateways (name, type, pay_type, description, status, api_url, pid, `key`, logo_url, sort_order, min_amount, max_amount, fee_rate, fee_mode, min_level, notify_url, create_time, update_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		gw.Name, gw.Type, gw.PayType, gw.Description, gw.Status,
		gw.ApiURL, gw.PID, gw.Key, gw.LogoURL, gw.SortOrder,
		gw.MinAmount, gw.MaxAmount, gw.FeeRate, gw.FeeMode,
		gw.MinLevel, gw.NotifyURL, gw.CreateTime, gw.UpdateTime,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	gw.ID = uint64(id)
	return nil
}

// GetPayGatewayByID 根据ID获取支付通道
func GetPayGatewayByID(id uint64) (*PayGateway, error) {
	var gw PayGateway
	err := db.DB.Get(&gw, "SELECT id, name, type, pay_type, description, status, api_url, pid, `key`, logo_url, sort_order, min_amount, max_amount, fee_rate, fee_mode, min_level, notify_url, create_time, update_time FROM pay_gateways WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &gw, nil
}

// UpdatePayGateway 更新支付通道
func UpdatePayGateway(gw *PayGateway) error {
	gw.UpdateTime = time.Now().Unix()
	_, err := db.DB.Exec(
		"UPDATE pay_gateways SET name=?, type=?, pay_type=?, description=?, status=?, api_url=?, pid=?, `key`=?, logo_url=?, sort_order=?, min_amount=?, max_amount=?, fee_rate=?, fee_mode=?, min_level=?, notify_url=?, update_time=? WHERE id=?",
		gw.Name, gw.Type, gw.PayType, gw.Description, gw.Status,
		gw.ApiURL, gw.PID, gw.Key, gw.LogoURL, gw.SortOrder,
		gw.MinAmount, gw.MaxAmount, gw.FeeRate, gw.FeeMode,
		gw.MinLevel, gw.NotifyURL, gw.UpdateTime, gw.ID,
	)
	return err
}

// DeletePayGateway 删除支付通道
func DeletePayGateway(id uint64) error {
	_, err := db.DB.Exec("DELETE FROM pay_gateways WHERE id = ?", id)
	return err
}

// GetPayGatewayList 分页获取支付通道列表
func GetPayGatewayList(page, pageSize int, keyword string, onlyEnabled bool) ([]PayGateway, int64, error) {
	var gateways []PayGateway
	var total int64

	where := "WHERE 1=1"
	args := []interface{}{}

	if onlyEnabled {
		where += " AND status = ?"
		args = append(args, PayGatewayStatusEnabled)
	}
	if keyword != "" {
		where += " AND (name LIKE ? OR description LIKE ? OR pay_type LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}

	// 总数
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := db.DB.Get(&total, "SELECT COUNT(*) FROM pay_gateways "+where, countArgs...)
	if err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (page - 1) * pageSize
	query := "SELECT id, name, type, pay_type, description, status, api_url, pid, `key`, logo_url, sort_order, min_amount, max_amount, fee_rate, fee_mode, min_level, notify_url, create_time, update_time FROM pay_gateways " + where + " ORDER BY sort_order ASC, id ASC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	err = db.DB.Select(&gateways, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return gateways, total, nil
}

// GetEnabledPayGateways 获取所有启用的支付通道（不分页，用于用户端）
func GetEnabledPayGateways() ([]PayGateway, error) {
	var gateways []PayGateway
	err := db.DB.Select(&gateways, "SELECT id, name, type, pay_type, description, status, api_url, pid, `key`, logo_url, sort_order, min_amount, max_amount, fee_rate, fee_mode, min_level, notify_url, create_time, update_time FROM pay_gateways WHERE status = ? ORDER BY sort_order ASC, id ASC", PayGatewayStatusEnabled)
	if err != nil {
		return nil, err
	}
	return gateways, nil
}
