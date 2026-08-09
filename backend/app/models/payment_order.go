package models

import (
	crypto_rand "crypto/rand"
	"fmt"
	"fst/backend/pkg/db"
	"math/big"
	"strings"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

// 原子自增序列号，防止同一纳秒内碰撞
var orderSeq uint64

func NormalizeTradeNo(tradeNo string) string {
	trimmed := strings.TrimSpace(tradeNo)
	if trimmed == "" {
		return ""
	}

	normalized := strings.ToUpper(trimmed)
	normalized = strings.Trim(normalized, " /\\-_:;,.#")
	replacer := strings.NewReplacer("_", "", "-", "", "/", "", "\\", "", " ", "", ":", "", ";", "", ".", "", "#", "")
	compact := replacer.Replace(normalized)

	switch compact {
	case "TRADENO", "OUTTRADENO", "NULL", "UNDEFINED", "NONE", "NIL", "NA":
		return ""
	default:
		// 外部交易号可能超长，落库前静默截断到列宽
		return clampBytes(trimmed, 64)
	}
}

func normalizePaymentOrder(order *PaymentOrder) {
	if order == nil {
		return
	}
	order.TradeNo = NormalizeTradeNo(order.TradeNo)
}

func normalizePaymentOrders(orders []PaymentOrder) {
	for i := range orders {
		orders[i].TradeNo = NormalizeTradeNo(orders[i].TradeNo)
	}
}

// 订单状态常量
const (
	PaymentStatusPending  = 0 // 待支付
	PaymentStatusPaid     = 1 // 已支付
	PaymentStatusCanceled = 2 // 已取消
	PaymentStatusRefunded = 3 // 已退款
	PaymentStatusFailed   = 4 // 支付失败
)

const RealPaidOrderFilterSQL = "gateway_id > 0 AND payment_channel <> 'admin' AND payment_type <> 'manual' AND trade_no <> '' AND UPPER(TRIM(trade_no)) NOT IN ('MANUAL', 'TRADENO', 'OUTTRADENO', 'NULL', 'UNDEFINED', 'NONE', 'NIL', 'NA')"

// PaymentOrder 支付订单
type PaymentOrder struct {
	ID             uint64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderNo        string  `gorm:"column:order_no;size:64;not null;uniqueIndex:idx_order_no" json:"order_no"`
	UserID         uint64  `gorm:"column:user_id;not null;default:0;index:idx_po_user_id;index:idx_po_user_status_create,priority:1" json:"user_id"`
	GatewayID      uint64  `gorm:"column:gateway_id;not null;default:0;index:idx_po_gateway_id;index:idx_po_gateway_status,priority:1" json:"gateway_id"`
	TradeNo        string  `gorm:"column:trade_no;size:64;not null;default:'';index:idx_po_trade_no" json:"trade_no"`
	PaymentChannel string  `gorm:"column:payment_channel;size:20;not null;default:'epay'" json:"payment_channel"`
	PaymentType    string  `gorm:"column:payment_type;size:20;not null;default:'alipay'" json:"payment_type"`
	Amount         float64 `gorm:"column:amount;type:decimal(10,2);not null;default:0" json:"amount"`
	Fee            float64 `gorm:"column:fee;type:decimal(10,2);not null;default:0" json:"fee"`
	PayAmount      float64 `gorm:"column:pay_amount;type:decimal(10,2);not null;default:0" json:"pay_amount"`
	Subject        string  `gorm:"column:subject;size:255;not null;default:''" json:"subject"`
	Status         int     `gorm:"column:status;not null;default:0;index:idx_po_status;index:idx_po_gateway_status,priority:2;index:idx_po_status_expire,priority:1;index:idx_po_user_status_create,priority:2" json:"status"`
	NotifyCount    int     `gorm:"column:notify_count;not null;default:0" json:"notify_count"`
	PayURL         string  `gorm:"column:pay_url;type:text" json:"pay_url"`
	PaidAt         *int64  `gorm:"column:paid_at" json:"paid_at"`
	ExpireAt       int64   `gorm:"column:expire_at;not null;default:0;index:idx_po_status_expire,priority:2" json:"expire_at"`
	ClientIP       string  `gorm:"column:client_ip;size:64;not null;default:''" json:"client_ip"`
	Extra          string  `gorm:"column:extra;type:text" json:"extra"`
	LastQueryAt    *int64  `gorm:"column:last_query_at" json:"last_query_at"`
	QueryCount     int     `gorm:"column:query_count;not null;default:0" json:"query_count"`
	CreateTime     int64   `gorm:"column:create_time;not null;default:0;index:idx_po_create_time;index:idx_po_user_status_create,priority:3" json:"create_time"`
	UpdateTime     int64   `gorm:"column:update_time;not null;default:0" json:"update_time"`
}

func (PaymentOrder) TableName() string {
	return "payment_orders"
}

// GenerateOrderNo 生成唯一订单号: P + 年月日时分秒 + 4位序列 + 4位密码学随机数
// 使用原子自增序列 + crypto/rand 保证高并发下不碰撞
func GenerateOrderNo() string {
	now := time.Now()
	seq := atomic.AddUint64(&orderSeq, 1) % 10000
	rnd, _ := crypto_rand.Int(crypto_rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("P%s%04d%04d", now.Format("20060102150405"), seq, rnd.Int64())
}

// CreatePaymentOrder 创建支付订单
func CreatePaymentOrder(order *PaymentOrder) error {
	now := time.Now().Unix()
	order.TradeNo = NormalizeTradeNo(order.TradeNo)
	order.CreateTime = now
	order.UpdateTime = now
	return db.DB.Create(order).Error
}

// CreatePaymentOrderTx 在已有事务中创建支付订单（与余额操作同事务时用）
func CreatePaymentOrderTx(tx *gorm.DB, order *PaymentOrder) error {
	now := time.Now().Unix()
	order.TradeNo = NormalizeTradeNo(order.TradeNo)
	order.CreateTime = now
	order.UpdateTime = now
	return tx.Create(order).Error
}

// GetPaymentOrderByOrderNo 按系统订单号查询
func GetPaymentOrderByOrderNo(orderNo string) (*PaymentOrder, error) {
	var order PaymentOrder
	err := db.DB.Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	normalizePaymentOrder(&order)
	return &order, nil
}

// GetPaymentOrderByID 按ID查询
func GetPaymentOrderByID(id uint64) (*PaymentOrder, error) {
	var order PaymentOrder
	err := db.DB.Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	normalizePaymentOrder(&order)
	return &order, nil
}

func GetPaymentOrderByIDForUpdate(tx *gorm.DB, id uint64) (*PaymentOrder, error) {
	var order PaymentOrder
	err := db.ForUpdate(tx).Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	normalizePaymentOrder(&order)
	return &order, nil
}

// GetPaymentOrderForUpdate 在事务中锁定订单（SELECT ... FOR UPDATE）
func GetPaymentOrderForUpdate(tx *gorm.DB, orderNo string) (*PaymentOrder, error) {
	var order PaymentOrder
	err := db.ForUpdate(tx).Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	normalizePaymentOrder(&order)
	return &order, nil
}

// canTransitionPaymentStatus 校验订单状态流转是否合法。
// 允许 canceled/failed → paid：过期取消或本地落库失败后，网关迟到成功回调/主动对账仍可安全恢复到账
// （前提是服务层已完成验签、金额与通道绑定校验，且余额入账与状态变更同事务）。
func canTransitionPaymentStatus(fromStatus, toStatus int) bool {
	if fromStatus == toStatus {
		return true
	}

	switch fromStatus {
	case PaymentStatusPending:
		return toStatus == PaymentStatusPaid || toStatus == PaymentStatusCanceled || toStatus == PaymentStatusFailed
	case PaymentStatusCanceled, PaymentStatusFailed:
		// 迟到成功支付：仅允许恢复为已支付，禁止回退到 pending 等其它状态
		return toStatus == PaymentStatusPaid
	case PaymentStatusPaid:
		return toStatus == PaymentStatusRefunded
	default:
		return false
	}
}

// UpdatePaymentOrderStatusTx 在事务中更新订单状态
// 仅当 tradeNo 非空时才更新 trade_no 字段，避免覆盖已保存的第三方交易号
func UpdatePaymentOrderStatusTx(tx *gorm.DB, orderNo string, status int, tradeNo string) error {
	var current PaymentOrder
	if err := db.ForUpdate(tx).Where("order_no = ?", orderNo).First(&current).Error; err != nil {
		return db.MapGormNotFound(err)
	}
	if !canTransitionPaymentStatus(current.Status, status) {
		return fmt.Errorf("Invalid order status transition: %d -> %d", current.Status, status)
	}

	tradeNo = NormalizeTradeNo(tradeNo)
	now := time.Now().Unix()
	updates := map[string]interface{}{
		"status":       status,
		"notify_count": gorm.Expr("notify_count + 1"),
		"update_time":  now,
	}
	if status == PaymentStatusPaid {
		updates["paid_at"] = now
	} else {
		updates["paid_at"] = nil
	}
	if tradeNo != "" {
		updates["trade_no"] = tradeNo
	}
	return tx.Model(&PaymentOrder{}).Where("order_no = ?", orderNo).Updates(updates).Error
}

func UpdatePaymentOrderPaymentInfo(orderNo, tradeNo, payURL string) error {
	tradeNo = NormalizeTradeNo(tradeNo)
	now := time.Now().Unix()
	updates := map[string]interface{}{
		"pay_url":     payURL,
		"update_time": now,
	}
	if tradeNo != "" {
		updates["trade_no"] = tradeNo
	}
	return db.DB.Model(&PaymentOrder{}).Where("order_no = ?", orderNo).Updates(updates).Error
}

// UpdatePaymentOrderStatus 更新订单状态（非事务）
// 仅当 tradeNo 非空时才更新 trade_no 字段，避免覆盖已保存的第三方交易号
func UpdatePaymentOrderStatus(orderNo string, status int, tradeNo string) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		return UpdatePaymentOrderStatusTx(tx, orderNo, status, tradeNo)
	})
}

// IncrementNotifyCount 增加通知次数
func IncrementNotifyCount(orderNo string) error {
	return db.DB.Model(&PaymentOrder{}).
		Where("order_no = ?", orderNo).
		Updates(map[string]interface{}{
			"notify_count": gorm.Expr("notify_count + 1"),
			"update_time":  time.Now().Unix(),
		}).Error
}

// UpdatePaymentOrderQueryAttempt 记录一次主动查单尝试
func UpdatePaymentOrderQueryAttempt(orderNo string) error {
	now := time.Now().Unix()
	return db.DB.Model(&PaymentOrder{}).
		Where("order_no = ?", orderNo).
		Updates(map[string]interface{}{
			"last_query_at": now,
			"query_count":   gorm.Expr("query_count + 1"),
		}).Error
}

// GetPaymentOrderList 分页获取订单列表
// userID > 0 时只查该用户的订单
func GetPaymentOrderList(userID uint64, page, pageSize int, status int, keyword string) ([]PaymentOrder, int64, error) {
	var orders []PaymentOrder
	var total int64

	q := db.DB.Model(&PaymentOrder{})
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("order_no LIKE ? OR trade_no LIKE ? OR subject LIKE ?", kw, kw, kw)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := q.Order("create_time DESC").Limit(pageSize).Offset(offset).Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}
	normalizePaymentOrders(orders)

	return orders, total, nil
}

// CancelExpiredOrders 取消过期未支付的订单
func CancelExpiredOrders() (int64, error) {
	now := time.Now().Unix()
	result := db.DB.Model(&PaymentOrder{}).
		Where("status = ? AND expire_at > 0 AND expire_at < ?", PaymentStatusPending, now).
		Updates(map[string]interface{}{
			"status":      PaymentStatusCanceled,
			"update_time": now,
		})
	return result.RowsAffected, result.Error
}

// GetPaymentStats 获取支付统计
type PaymentStats struct {
	TotalOrders   int64   `gorm:"column:total_orders" json:"total_orders"`
	PaidOrders    int64   `gorm:"column:paid_orders" json:"paid_orders"`
	TotalAmount   float64 `gorm:"column:total_amount" json:"total_amount"`
	TodayOrders   int64   `gorm:"column:today_orders" json:"today_orders"`
	TodayAmount   float64 `gorm:"column:today_amount" json:"today_amount"`
	MonthAmount   float64 `gorm:"column:month_amount" json:"month_amount"`
	YearAmount    float64 `gorm:"column:year_amount" json:"year_amount"`
	PendingOrders int64   `gorm:"column:pending_orders" json:"pending_orders"`
}

func GetPaymentStats() (*PaymentStats, error) {
	var stats PaymentStats
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()
	yearStart := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location()).Unix()

	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_orders,
			COALESCE(SUM(CASE WHEN status = 1 AND %s THEN 1 ELSE 0 END), 0) as paid_orders,
			COALESCE(SUM(CASE WHEN status = 1 AND %s THEN pay_amount ELSE 0 END), 0) as total_amount,
			COALESCE(SUM(CASE WHEN status = 1 AND %s AND paid_at >= ? THEN 1 ELSE 0 END), 0) as today_orders,
			COALESCE(SUM(CASE WHEN status = 1 AND %s AND paid_at >= ? THEN pay_amount ELSE 0 END), 0) as today_amount,
			COALESCE(SUM(CASE WHEN status = 1 AND %s AND paid_at >= ? THEN pay_amount ELSE 0 END), 0) as month_amount,
			COALESCE(SUM(CASE WHEN status = 1 AND %s AND paid_at >= ? THEN pay_amount ELSE 0 END), 0) as year_amount,
			COALESCE(SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END), 0) as pending_orders
		FROM payment_orders
	`, RealPaidOrderFilterSQL, RealPaidOrderFilterSQL, RealPaidOrderFilterSQL, RealPaidOrderFilterSQL, RealPaidOrderFilterSQL, RealPaidOrderFilterSQL)

	err := db.DB.Raw(query, todayStart, todayStart, monthStart, yearStart).Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// DeletePaymentOrder 删除订单（仅管理员）
func DeletePaymentOrder(id uint64) error {
	return db.DB.Delete(&PaymentOrder{}, id).Error
}

func CountPendingOrdersByGatewayID(gatewayID uint64) (int64, error) {
	var count int64
	err := db.DB.Model(&PaymentOrder{}).
		Where("gateway_id = ? AND status = ?", gatewayID, PaymentStatusPending).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountPendingOrdersByUserIDTx 在事务中统计用户待支付订单数（配合用户行锁一起使用，防止并发建单绕过限流）
func CountPendingOrdersByUserIDTx(tx *gorm.DB, userID uint64) (int64, error) {
	var count int64
	err := tx.Model(&PaymentOrder{}).
		Where("user_id = ? AND status = ?", userID, PaymentStatusPending).
		Count(&count).Error
	return count, err
}
