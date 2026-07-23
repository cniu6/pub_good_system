package models

import (
	crypto_rand "crypto/rand"
	"database/sql"
	"fmt"
	"fst/backend/pkg/db"
	"log"
	"math/big"
	"strings"
	"sync/atomic"
	"time"
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
		return trimmed
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
	ID             uint64  `db:"id" json:"id"`
	OrderNo        string  `db:"order_no" json:"order_no"`               // 系统订单号
	UserID         uint64  `db:"user_id" json:"user_id"`                 // 用户ID
	GatewayID      uint64  `db:"gateway_id" json:"gateway_id"`           // 支付通道ID
	TradeNo        string  `db:"trade_no" json:"trade_no"`               // 第三方交易号
	PaymentChannel string  `db:"payment_channel" json:"payment_channel"` // 支付通道类型：epay
	PaymentType    string  `db:"payment_type" json:"payment_type"`       // 支付方式：alipay/wxpay/qqpay
	Amount         float64 `db:"amount" json:"amount"`                   // 充值金额（用户希望到账的金额）
	Fee            float64 `db:"fee" json:"fee"`                         // 手续费
	PayAmount      float64 `db:"pay_amount" json:"pay_amount"`           // 实际支付金额
	Subject        string  `db:"subject" json:"subject"`                 // 订单标题
	Status         int     `db:"status" json:"status"`                   // 状态：0=待支付,1=已支付,2=已取消,3=已退款,4=失败
	NotifyCount    int     `db:"notify_count" json:"notify_count"`       // 回调通知次数
	PayURL         string  `db:"pay_url" json:"pay_url"`                 // 支付链接
	PaidAt         *int64  `db:"paid_at" json:"paid_at"`                 // 支付完成时间
	ExpireAt       int64   `db:"expire_at" json:"expire_at"`             // 订单过期时间
	ClientIP       string  `db:"client_ip" json:"client_ip"`             // 下单客户端IP
	Extra          string  `db:"extra" json:"extra"`                     // 扩展信息（JSON）
	CreateTime     int64   `db:"create_time" json:"create_time"`
	UpdateTime     int64   `db:"update_time" json:"update_time"`
}

// InitPaymentOrdersTable 初始化支付订单表
func InitPaymentOrdersTable() {
	if db.CheckTableExists("payment_orders") {
		indexRepairs := map[string]string{
			"idx_gateway_status":     "ALTER TABLE payment_orders ADD INDEX idx_gateway_status (gateway_id, status)",
			"idx_status_expire":      "ALTER TABLE payment_orders ADD INDEX idx_status_expire (status, expire_at)",
			"idx_user_status_create": "ALTER TABLE payment_orders ADD INDEX idx_user_status_create (user_id, status, create_time)",
		}

		for indexName, alterSQL := range indexRepairs {
			db.EnsureIndex("payment_orders", indexName, alterSQL)
		}
		return
	}

	schema := `CREATE TABLE IF NOT EXISTS payment_orders (
		id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		order_no        VARCHAR(64)      NOT NULL COMMENT '系统订单号',
		user_id         BIGINT UNSIGNED  NOT NULL DEFAULT 0 COMMENT '用户ID',
		gateway_id      BIGINT UNSIGNED  NOT NULL DEFAULT 0 COMMENT '支付通道ID',
		trade_no        VARCHAR(64)      NOT NULL DEFAULT '' COMMENT '第三方交易号',
		payment_channel VARCHAR(20)      NOT NULL DEFAULT 'epay' COMMENT '支付通道类型',
		payment_type    VARCHAR(20)      NOT NULL DEFAULT 'alipay' COMMENT '支付方式',
		amount          DECIMAL(10,2)    NOT NULL DEFAULT 0.00 COMMENT '充值金额',
		fee             DECIMAL(10,2)    NOT NULL DEFAULT 0.00 COMMENT '手续费',
		pay_amount      DECIMAL(10,2)    NOT NULL DEFAULT 0.00 COMMENT '实际支付金额',
		subject         VARCHAR(255)     NOT NULL DEFAULT '' COMMENT '订单标题',
		status          TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态:0=待支付,1=已支付,2=已取消,3=已退款,4=失败',
		notify_count    INT UNSIGNED     NOT NULL DEFAULT 0 COMMENT '回调通知次数',
		pay_url         TEXT             COMMENT '支付链接',
		paid_at         BIGINT           NULL DEFAULT NULL COMMENT '支付完成时间',
		expire_at       BIGINT           NOT NULL DEFAULT 0 COMMENT '订单过期时间',
		client_ip       VARCHAR(50)      NOT NULL DEFAULT '' COMMENT '下单客户端IP',
		extra           TEXT             COMMENT '扩展信息JSON',
		create_time     BIGINT           NOT NULL DEFAULT 0 COMMENT '创建时间',
		update_time     BIGINT           NOT NULL DEFAULT 0 COMMENT '更新时间',
		UNIQUE KEY idx_order_no (order_no),
		INDEX idx_user_id (user_id),
		INDEX idx_gateway_id (gateway_id),
		INDEX idx_status (status),
		INDEX idx_gateway_status (gateway_id, status),
		INDEX idx_status_expire (status, expire_at),
		INDEX idx_user_status_create (user_id, status, create_time),
		INDEX idx_trade_no (trade_no),
		INDEX idx_create_time (create_time)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付订单表';`

	_, err := db.Exec(schema)
	if err != nil {
		log.Printf("[Init] Failed to create payment_orders table: %v", err)
	} else {
		log.Println("[Init] Created payment_orders table")
	}
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

	result, err := db.Exec(
		`INSERT INTO payment_orders (order_no, user_id, gateway_id, trade_no, payment_channel, payment_type, amount, fee, pay_amount, subject, status, notify_count, pay_url, paid_at, expire_at, client_ip, extra, create_time, update_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		order.OrderNo, order.UserID, order.GatewayID, order.TradeNo, order.PaymentChannel, order.PaymentType,
		order.Amount, order.Fee, order.PayAmount, order.Subject, order.Status, order.NotifyCount,
		order.PayURL, order.PaidAt, order.ExpireAt, order.ClientIP, order.Extra,
		order.CreateTime, order.UpdateTime,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	order.ID = uint64(id)
	return nil
}

// CreatePaymentOrderTx 在已有事务中创建支付订单（与余额操作同事务时用）
func CreatePaymentOrderTx(tx *sql.Tx, order *PaymentOrder) error {
	now := time.Now().Unix()
	order.TradeNo = NormalizeTradeNo(order.TradeNo)
	order.CreateTime = now
	order.UpdateTime = now

	result, err := tx.Exec(
		`INSERT INTO payment_orders (order_no, user_id, gateway_id, trade_no, payment_channel, payment_type, amount, fee, pay_amount, subject, status, notify_count, pay_url, paid_at, expire_at, client_ip, extra, create_time, update_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		order.OrderNo, order.UserID, order.GatewayID, order.TradeNo, order.PaymentChannel, order.PaymentType,
		order.Amount, order.Fee, order.PayAmount, order.Subject, order.Status, order.NotifyCount,
		order.PayURL, order.PaidAt, order.ExpireAt, order.ClientIP, order.Extra,
		order.CreateTime, order.UpdateTime,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	order.ID = uint64(id)
	return nil
}

// GetPaymentOrderByOrderNo 按系统订单号查询
func GetPaymentOrderByOrderNo(orderNo string) (*PaymentOrder, error) {
	var order PaymentOrder
	err := db.DB.Get(&order, "SELECT * FROM payment_orders WHERE order_no = ?", orderNo)
	if err != nil {
		return nil, err
	}
	normalizePaymentOrder(&order)
	return &order, nil
}

// GetPaymentOrderByID 按ID查询
func GetPaymentOrderByID(id uint64) (*PaymentOrder, error) {
	var order PaymentOrder
	err := db.DB.Get(&order, "SELECT * FROM payment_orders WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	normalizePaymentOrder(&order)
	return &order, nil
}

func GetPaymentOrderByIDForUpdate(tx *sql.Tx, id uint64) (*PaymentOrder, error) {
	var order PaymentOrder
	err := tx.QueryRow(
		db.Q("SELECT id, order_no, user_id, gateway_id, trade_no, payment_channel, payment_type, amount, fee, pay_amount, subject, status, notify_count, COALESCE(pay_url,''), paid_at, expire_at, client_ip, COALESCE(extra,''), create_time, update_time FROM payment_orders WHERE id = ? FOR UPDATE"),
		id,
	).Scan(
		&order.ID, &order.OrderNo, &order.UserID, &order.GatewayID, &order.TradeNo,
		&order.PaymentChannel, &order.PaymentType, &order.Amount, &order.Fee, &order.PayAmount, &order.Subject,
		&order.Status, &order.NotifyCount, &order.PayURL, &order.PaidAt, &order.ExpireAt,
		&order.ClientIP, &order.Extra, &order.CreateTime, &order.UpdateTime,
	)
	if err != nil {
		return nil, err
	}
	normalizePaymentOrder(&order)
	return &order, nil
}

// GetPaymentOrderForUpdate 在事务中锁定订单（SELECT ... FOR UPDATE）
func GetPaymentOrderForUpdate(tx *sql.Tx, orderNo string) (*PaymentOrder, error) {
	var order PaymentOrder
	err := tx.QueryRow(
		db.Q("SELECT id, order_no, user_id, gateway_id, trade_no, payment_channel, payment_type, amount, fee, pay_amount, subject, status, notify_count, COALESCE(pay_url,''), paid_at, expire_at, client_ip, COALESCE(extra,''), create_time, update_time FROM payment_orders WHERE order_no = ? FOR UPDATE"),
		orderNo,
	).Scan(
		&order.ID, &order.OrderNo, &order.UserID, &order.GatewayID, &order.TradeNo,
		&order.PaymentChannel, &order.PaymentType, &order.Amount, &order.Fee, &order.PayAmount, &order.Subject,
		&order.Status, &order.NotifyCount, &order.PayURL, &order.PaidAt, &order.ExpireAt,
		&order.ClientIP, &order.Extra, &order.CreateTime, &order.UpdateTime,
	)
	if err != nil {
		return nil, err
	}
	normalizePaymentOrder(&order)
	return &order, nil
}

// canTransitionPaymentStatus 校验订单状态流转是否合法。
// 允许 canceled/failed → paid：过期取消或本地落库失败后，网关迟到成功回调/主动对账仍可安全恢复到账
//（前提是服务层已完成验签、金额与通道绑定校验，且余额入账与状态变更同事务）。
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
func UpdatePaymentOrderStatusTx(tx *sql.Tx, orderNo string, status int, tradeNo string) error {
	var currentStatus int
	if err := tx.QueryRow(db.Q("SELECT status FROM payment_orders WHERE order_no = ? FOR UPDATE"), orderNo).Scan(&currentStatus); err != nil {
		return err
	}
	if !canTransitionPaymentStatus(currentStatus, status) {
		return fmt.Errorf("非法订单状态流转: %d -> %d", currentStatus, status)
	}

	tradeNo = NormalizeTradeNo(tradeNo)
	now := time.Now().Unix()
	var paidAt *int64
	if status == PaymentStatusPaid {
		paidAt = &now
	}
	if tradeNo != "" {
		_, err := tx.Exec(
			"UPDATE payment_orders SET status = ?, trade_no = ?, paid_at = ?, notify_count = notify_count + 1, update_time = ? WHERE order_no = ?",
			status, tradeNo, paidAt, now, orderNo,
		)
		return err
	}
	_, err := tx.Exec(
		"UPDATE payment_orders SET status = ?, paid_at = ?, notify_count = notify_count + 1, update_time = ? WHERE order_no = ?",
		status, paidAt, now, orderNo,
	)
	return err
}

func UpdatePaymentOrderPaymentInfo(orderNo, tradeNo, payURL string) error {
	tradeNo = NormalizeTradeNo(tradeNo)
	now := time.Now().Unix()
	if tradeNo != "" {
		_, err := db.Exec(
			"UPDATE payment_orders SET trade_no = ?, pay_url = ?, update_time = ? WHERE order_no = ?",
			tradeNo,
			payURL,
			now,
			orderNo,
		)
		return err
	}
	_, err := db.Exec(
		"UPDATE payment_orders SET pay_url = ?, update_time = ? WHERE order_no = ?",
		payURL,
		now,
		orderNo,
	)
	return err
}

// UpdatePaymentOrderStatus 更新订单状态（非事务）
// 仅当 tradeNo 非空时才更新 trade_no 字段，避免覆盖已保存的第三方交易号
func UpdatePaymentOrderStatus(orderNo string, status int, tradeNo string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := UpdatePaymentOrderStatusTx(tx, orderNo, status, tradeNo); err != nil {
		return err
	}

	return tx.Commit()
}

// IncrementNotifyCount 增加通知次数
func IncrementNotifyCount(orderNo string) error {
	_, err := db.Exec("UPDATE payment_orders SET notify_count = notify_count + 1, update_time = ? WHERE order_no = ?", time.Now().Unix(), orderNo)
	return err
}

// GetPaymentOrderList 分页获取订单列表
// userID > 0 时只查该用户的订单
func GetPaymentOrderList(userID uint64, page, pageSize int, status int, keyword string) ([]PaymentOrder, int64, error) {
	var orders []PaymentOrder
	var total int64

	where := "WHERE 1=1"
	args := []interface{}{}

	if userID > 0 {
		where += " AND user_id = ?"
		args = append(args, userID)
	}
	if status >= 0 {
		where += " AND status = ?"
		args = append(args, status)
	}
	if keyword != "" {
		where += " AND (order_no LIKE ? OR trade_no LIKE ? OR subject LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}

	// 总数
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := db.DB.Get(&total, "SELECT COUNT(*) FROM payment_orders "+where, countArgs...)
	if err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (page - 1) * pageSize
	query := "SELECT * FROM payment_orders " + where + " ORDER BY create_time DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	err = db.DB.Select(&orders, query, args...)
	if err != nil {
		return nil, 0, err
	}
	normalizePaymentOrders(orders)

	return orders, total, nil
}

// CancelExpiredOrders 取消过期未支付的订单
func CancelExpiredOrders() (int64, error) {
	now := time.Now().Unix()
	result, err := db.Exec(
		"UPDATE payment_orders SET status = ?, update_time = ? WHERE status = ? AND expire_at > 0 AND expire_at < ?",
		PaymentStatusCanceled, now, PaymentStatusPending, now,
	)
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return affected, nil
}

// GetPaymentStats 获取支付统计
type PaymentStats struct {
	TotalOrders   int64   `db:"total_orders" json:"total_orders"`
	PaidOrders    int64   `db:"paid_orders" json:"paid_orders"`
	TotalAmount   float64 `db:"total_amount" json:"total_amount"`
	TodayOrders   int64   `db:"today_orders" json:"today_orders"`
	TodayAmount   float64 `db:"today_amount" json:"today_amount"`
	MonthAmount   float64 `db:"month_amount" json:"month_amount"`
	YearAmount    float64 `db:"year_amount" json:"year_amount"`
	PendingOrders int64   `db:"pending_orders" json:"pending_orders"`
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

	err := db.DB.Get(&stats, query, todayStart, todayStart, monthStart, yearStart)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// DeletePaymentOrder 删除订单（仅管理员）
func DeletePaymentOrder(id uint64) error {
	_, err := db.Exec("DELETE FROM payment_orders WHERE id = ?", id)
	return err
}

func CountPendingOrdersByGatewayID(gatewayID uint64) (int64, error) {
	var count int64
	err := db.DB.Get(&count, "SELECT COUNT(*) FROM payment_orders WHERE gateway_id = ? AND status = ?", gatewayID, PaymentStatusPending)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountPendingOrdersByUserIDTx 在事务中统计用户待支付订单数（配合用户行锁一起使用，防止并发建单绕过限流）
func CountPendingOrdersByUserIDTx(tx *sql.Tx, userID uint64) (int64, error) {
	var count int64
	err := tx.QueryRow(db.Q("SELECT COUNT(*) FROM payment_orders WHERE user_id = ? AND status = ?"), userID, PaymentStatusPending).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
