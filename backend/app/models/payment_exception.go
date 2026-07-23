package models

import (
	"database/sql"
	"fmt"
	"fst/backend/pkg/db"
	"log"
	"strings"
	"time"
)

// 支付异常类型（与管理端筛选对齐）
const (
	PaymentExceptionSignFailed           = "sign_failed"              // 验签失败
	PaymentExceptionAmountMismatch       = "amount_mismatch"          // 金额不符
	PaymentExceptionBindingMismatch      = "binding_mismatch"         // 通道/商户/交易号绑定不符
	PaymentExceptionLateCallback         = "late_callback_recovered"  // 迟到回调已恢复到账
	PaymentExceptionRemoteLocalSaveFail  = "remote_local_save_failed" // 远程建单成功但本地落库失败
	PaymentExceptionReconcilePaid        = "reconcile_paid"           // 主动对账补单成功
	PaymentExceptionPermanentRejected    = "permanent_rejected"       // 永久错误已确认拒绝（网关可停重试）
	PaymentExceptionOrderMissing         = "order_missing"            // 回调订单不存在
	PaymentExceptionManualResolve        = "manual_resolve"           // 人工处理备注
)

// 异常处理状态
const (
	PaymentExceptionStatusOpen     = 0 // 待处理
	PaymentExceptionStatusResolved = 1 // 已处理
	PaymentExceptionStatusIgnored  = 2 // 已忽略
)

// PaymentException 支付异常审计记录
type PaymentException struct {
	ID             uint64  `db:"id" json:"id"`
	OrderNo        string  `db:"order_no" json:"order_no"`
	UserID         uint64  `db:"user_id" json:"user_id"`
	GatewayID      uint64  `db:"gateway_id" json:"gateway_id"`
	ExceptionType  string  `db:"exception_type" json:"exception_type"`
	Status         int     `db:"status" json:"status"`
	Source         string  `db:"source" json:"source"` // notify / reconcile / create / admin
	Message        string  `db:"message" json:"message"`
	Detail         string  `db:"detail" json:"detail"` // JSON 扩展（脱敏后）
	OrderStatus    int     `db:"order_status" json:"order_status"`
	TradeNo        string  `db:"trade_no" json:"trade_no"`
	ResolvedBy     uint64  `db:"resolved_by" json:"resolved_by"`
	ResolvedAt     *int64  `db:"resolved_at" json:"resolved_at"`
	ResolveRemark  string  `db:"resolve_remark" json:"resolve_remark"`
	CreateTime     int64   `db:"create_time" json:"create_time"`
	UpdateTime     int64   `db:"update_time" json:"update_time"`
}

// InitPaymentExceptionsTable 初始化支付异常表
func InitPaymentExceptionsTable() {
	if db.CheckTableExists("payment_exceptions") {
		indexRepairs := map[string]string{
			"idx_pe_status_create": "ALTER TABLE payment_exceptions ADD INDEX idx_pe_status_create (status, create_time)",
			"idx_pe_order_no":      "ALTER TABLE payment_exceptions ADD INDEX idx_pe_order_no (order_no)",
			"idx_pe_type_status":   "ALTER TABLE payment_exceptions ADD INDEX idx_pe_type_status (exception_type, status)",
		}
		for indexName, alterSQL := range indexRepairs {
			db.EnsureIndex("payment_exceptions", indexName, alterSQL)
		}
		return
	}

	schema := `CREATE TABLE IF NOT EXISTS payment_exceptions (
		id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		order_no        VARCHAR(64)      NOT NULL DEFAULT '' COMMENT '系统订单号',
		user_id         BIGINT UNSIGNED  NOT NULL DEFAULT 0 COMMENT '用户ID',
		gateway_id      BIGINT UNSIGNED  NOT NULL DEFAULT 0 COMMENT '支付通道ID',
		exception_type  VARCHAR(64)      NOT NULL DEFAULT '' COMMENT '异常类型',
		status          TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0待处理 1已处理 2已忽略',
		source          VARCHAR(32)      NOT NULL DEFAULT '' COMMENT '来源',
		message         VARCHAR(500)     NOT NULL DEFAULT '' COMMENT '摘要',
		detail          TEXT             COMMENT '详情JSON(脱敏)',
		order_status    TINYINT          NOT NULL DEFAULT -1 COMMENT '记录时订单状态',
		trade_no        VARCHAR(64)      NOT NULL DEFAULT '' COMMENT '第三方交易号',
		resolved_by     BIGINT UNSIGNED  NOT NULL DEFAULT 0 COMMENT '处理人管理员ID',
		resolved_at     BIGINT           NULL DEFAULT NULL COMMENT '处理时间',
		resolve_remark  VARCHAR(500)     NOT NULL DEFAULT '' COMMENT '处理备注',
		create_time     BIGINT           NOT NULL DEFAULT 0 COMMENT '创建时间',
		update_time     BIGINT           NOT NULL DEFAULT 0 COMMENT '更新时间',
		INDEX idx_pe_status_create (status, create_time),
		INDEX idx_pe_order_no (order_no),
		INDEX idx_pe_type_status (exception_type, status),
		INDEX idx_pe_user_id (user_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付异常审计表';`

	if _, err := db.Exec(schema); err != nil {
		log.Printf("[Init] Failed to create payment_exceptions table: %v", err)
	} else {
		log.Println("[Init] Created payment_exceptions table")
	}
}

// CreatePaymentException 写入一条支付异常记录
func CreatePaymentException(ex *PaymentException) error {
	if ex == nil {
		return fmt.Errorf("exception is nil")
	}
	now := time.Now().Unix()
	ex.CreateTime = now
	ex.UpdateTime = now
	ex.TradeNo = NormalizeTradeNo(ex.TradeNo)
	if ex.ExceptionType == "" {
		ex.ExceptionType = PaymentExceptionPermanentRejected
	}
	result, err := db.Exec(
		`INSERT INTO payment_exceptions
		 (order_no, user_id, gateway_id, exception_type, status, source, message, detail, order_status, trade_no,
		  resolved_by, resolved_at, resolve_remark, create_time, update_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ex.OrderNo, ex.UserID, ex.GatewayID, ex.ExceptionType, ex.Status, ex.Source, ex.Message, ex.Detail,
		ex.OrderStatus, ex.TradeNo, ex.ResolvedBy, ex.ResolvedAt, ex.ResolveRemark, ex.CreateTime, ex.UpdateTime,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	ex.ID = uint64(id)
	return nil
}

// CreatePaymentExceptionTx 事务内写入异常记录
func CreatePaymentExceptionTx(tx *sql.Tx, ex *PaymentException) error {
	if ex == nil {
		return fmt.Errorf("exception is nil")
	}
	now := time.Now().Unix()
	ex.CreateTime = now
	ex.UpdateTime = now
	ex.TradeNo = NormalizeTradeNo(ex.TradeNo)
	result, err := tx.Exec(
		db.Q(`INSERT INTO payment_exceptions
		 (order_no, user_id, gateway_id, exception_type, status, source, message, detail, order_status, trade_no,
		  resolved_by, resolved_at, resolve_remark, create_time, update_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`),
		ex.OrderNo, ex.UserID, ex.GatewayID, ex.ExceptionType, ex.Status, ex.Source, ex.Message, ex.Detail,
		ex.OrderStatus, ex.TradeNo, ex.ResolvedBy, ex.ResolvedAt, ex.ResolveRemark, ex.CreateTime, ex.UpdateTime,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	ex.ID = uint64(id)
	return nil
}

// ListPaymentExceptions 分页查询支付异常
func ListPaymentExceptions(page, pageSize int, status *int, exceptionType, orderNo string, userID uint64) ([]PaymentException, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	where := []string{"1=1"}
	args := []interface{}{}
	if status != nil {
		where = append(where, "status = ?")
		args = append(args, *status)
	}
	if exceptionType != "" {
		where = append(where, "exception_type = ?")
		args = append(args, exceptionType)
	}
	if orderNo != "" {
		where = append(where, "order_no LIKE ?")
		args = append(args, "%"+strings.TrimSpace(orderNo)+"%")
	}
	if userID > 0 {
		where = append(where, "user_id = ?")
		args = append(args, userID)
	}
	whereSQL := strings.Join(where, " AND ")

	var total int64
	if err := db.DB.Get(&total, "SELECT COUNT(*) FROM payment_exceptions WHERE "+whereSQL, args...); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(
		"SELECT id, order_no, user_id, gateway_id, exception_type, status, source, message, detail, order_status, trade_no, resolved_by, resolved_at, resolve_remark, create_time, update_time FROM payment_exceptions WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?",
		whereSQL,
	)
	args = append(args, pageSize, (page-1)*pageSize)
	var list []PaymentException
	if err := db.DB.Select(&list, query, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetPaymentExceptionByID 按 ID 取异常
func GetPaymentExceptionByID(id uint64) (*PaymentException, error) {
	var ex PaymentException
	err := db.DB.Get(&ex, `SELECT id, order_no, user_id, gateway_id, exception_type, status, source, message, detail, order_status, trade_no, resolved_by, resolved_at, resolve_remark, create_time, update_time
		FROM payment_exceptions WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return &ex, nil
}

// ResolvePaymentException 人工处理/忽略异常
func ResolvePaymentException(id, adminID uint64, status int, remark string) error {
	if status != PaymentExceptionStatusResolved && status != PaymentExceptionStatusIgnored {
		return fmt.Errorf("非法处理状态")
	}
	now := time.Now().Unix()
	_, err := db.Exec(
		`UPDATE payment_exceptions SET status = ?, resolved_by = ?, resolved_at = ?, resolve_remark = ?, update_time = ? WHERE id = ? AND status = ?`,
		status, adminID, now, strings.TrimSpace(remark), now, id, PaymentExceptionStatusOpen,
	)
	return err
}

// ListPaymentOrdersForReconcile 扫描待对账订单：待支付 + 近期取消/失败（可迟到回调恢复）
func ListPaymentOrdersForReconcile(limit int, canceledLookbackSec int64) ([]PaymentOrder, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if canceledLookbackSec <= 0 {
		canceledLookbackSec = 7 * 24 * 3600 // 默认回看 7 天
	}
	now := time.Now().Unix()
	cutoff := now - canceledLookbackSec

	var orders []PaymentOrder
	// 待支付：有通道、临近过期或已创建一段时间；取消/失败：近期且有通道
	query := `
		SELECT id, order_no, user_id, gateway_id, trade_no, payment_channel, payment_type,
		       amount, fee, pay_amount, subject, status, notify_count, pay_url, paid_at,
		       expire_at, client_ip, extra, create_time, update_time
		FROM payment_orders
		WHERE gateway_id > 0 AND (
			(status = ? AND (expire_at = 0 OR expire_at <= ? OR create_time <= ?))
			OR (status IN (?, ?) AND update_time >= ?)
		)
		ORDER BY update_time ASC
		LIMIT ?`
	// 待支付：过期前后 30 分钟窗口，或创建超过 2 分钟仍未付（给网关查单机会）
	pendingWindowEnd := now + 30*60
	pendingCreatedBefore := now - 120
	err := db.DB.Select(&orders, query,
		PaymentStatusPending, pendingWindowEnd, pendingCreatedBefore,
		PaymentStatusCanceled, PaymentStatusFailed, cutoff,
		limit,
	)
	if err != nil {
		return nil, err
	}
	normalizePaymentOrders(orders)
	return orders, nil
}
