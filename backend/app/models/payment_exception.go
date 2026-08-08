package models

import (
	"database/sql"
	"errors"
	"fmt"
	"fst/backend/pkg/db"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 支付异常类型（与管理端筛选对齐）
const (
	PaymentExceptionSignFailed          = "sign_failed"              // 验签失败
	PaymentExceptionAmountMismatch      = "amount_mismatch"          // 金额不符
	PaymentExceptionBindingMismatch     = "binding_mismatch"         // 通道/商户/交易号绑定不符
	PaymentExceptionLateCallback        = "late_callback_recovered"  // 迟到回调已恢复到账
	PaymentExceptionRemoteLocalSaveFail = "remote_local_save_failed" // 远程建单成功但本地落库失败
	PaymentExceptionReconcilePaid       = "reconcile_paid"           // 主动对账补单成功
	PaymentExceptionPermanentRejected   = "permanent_rejected"       // 永久错误已确认拒绝（网关可停重试）
	PaymentExceptionOrderMissing        = "order_missing"            // 回调订单不存在
	PaymentExceptionManualResolve       = "manual_resolve"           // 人工处理备注
)

// 异常处理状态
const (
	PaymentExceptionStatusOpen     = 0 // 待处理
	PaymentExceptionStatusResolved = 1 // 已处理
	PaymentExceptionStatusIgnored  = 2 // 已忽略
)

// PaymentException 支付异常审计记录
type PaymentException struct {
	ID            uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderNo       string `gorm:"column:order_no;size:64;not null;default:'';index:idx_pe_order_no" json:"order_no"`
	UserID        uint64 `gorm:"column:user_id;not null;default:0;index:idx_pe_user_id" json:"user_id"`
	GatewayID     uint64 `gorm:"column:gateway_id;not null;default:0" json:"gateway_id"`
	ExceptionType string `gorm:"column:exception_type;size:64;not null;default:'';index:idx_pe_type_status,priority:1" json:"exception_type"`
	Status        int    `gorm:"column:status;not null;default:0;index:idx_pe_status_create,priority:1;index:idx_pe_type_status,priority:2" json:"status"`
	Source        string `gorm:"column:source;size:32;not null;default:''" json:"source"` // notify / reconcile / create / admin
	Message       string `gorm:"column:message;size:500;not null;default:''" json:"message"`
	Detail        string `gorm:"column:detail;type:text" json:"detail"` // JSON 扩展（脱敏后）
	OrderStatus   int    `gorm:"column:order_status;not null;default:-1" json:"order_status"`
	TradeNo       string `gorm:"column:trade_no;size:64;not null;default:''" json:"trade_no"`
	ResolvedBy    uint64 `gorm:"column:resolved_by;not null;default:0" json:"resolved_by"`
	ResolvedAt    *int64 `gorm:"column:resolved_at" json:"resolved_at"`
	ResolveRemark string `gorm:"column:resolve_remark;size:500;not null;default:''" json:"resolve_remark"`
	CreateTime    int64  `gorm:"column:create_time;not null;default:0;index:idx_pe_status_create,priority:2" json:"create_time"`
	UpdateTime    int64  `gorm:"column:update_time;not null;default:0" json:"update_time"`
}

// TableName 表名
func (PaymentException) TableName() string { return "payment_exceptions" }

func preparePaymentException(ex *PaymentException) {
	now := time.Now().Unix()
	ex.CreateTime = now
	ex.UpdateTime = now
	ex.OrderNo = clampBytes(ex.OrderNo, 64)
	ex.TradeNo = NormalizeTradeNo(ex.TradeNo)
	ex.ExceptionType = clampBytes(ex.ExceptionType, 64)
	ex.Source = clampBytes(ex.Source, 32)
	ex.Message = clampBytes(ex.Message, 500)
	ex.ResolveRemark = clampBytes(ex.ResolveRemark, 500)
	if ex.ExceptionType == "" {
		ex.ExceptionType = PaymentExceptionPermanentRejected
	}
}

// CreatePaymentException 写入一条支付异常记录
func CreatePaymentException(ex *PaymentException) error {
	if ex == nil {
		return fmt.Errorf("exception is nil")
	}
	preparePaymentException(ex)
	return db.DB.Create(ex).Error
}

// CreatePaymentExceptionTx 事务内写入异常记录
func CreatePaymentExceptionTx(tx *gorm.DB, ex *PaymentException) error {
	if ex == nil {
		return fmt.Errorf("exception is nil")
	}
	preparePaymentException(ex)
	return tx.Create(ex).Error
}

// ListPaymentExceptions 分页查询支付异常
func ListPaymentExceptions(page, pageSize int, status *int, exceptionType, orderNo string, userID uint64) ([]PaymentException, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	q := db.DB.Model(&PaymentException{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if exceptionType != "" {
		q = q.Where("exception_type = ?", exceptionType)
	}
	if orderNo != "" {
		q = q.Where("order_no LIKE ?", "%"+strings.TrimSpace(orderNo)+"%")
	}
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []PaymentException
	err := q.Order("id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

// GetPaymentExceptionByID 按 ID 取异常
func GetPaymentExceptionByID(id uint64) (*PaymentException, error) {
	var ex PaymentException
	err := db.DB.Where("id = ?", id).First(&ex).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &ex, nil
}

// ResolvePaymentException 人工处理/忽略异常
func ResolvePaymentException(id, adminID uint64, status int, remark string) error {
	if status != PaymentExceptionStatusResolved && status != PaymentExceptionStatusIgnored {
		return fmt.Errorf("Invalid processing status")
	}
	now := time.Now().Unix()
	return db.DB.Model(&PaymentException{}).
		Where("id = ? AND status = ?", id, PaymentExceptionStatusOpen).
		Updates(map[string]any{
			"status":         status,
			"resolved_by":    adminID,
			"resolved_at":    now,
			"resolve_remark": strings.TrimSpace(remark),
			"update_time":    now,
		}).Error
}

// ListPaymentOrdersForReconcile 扫描待对账订单：待支付 + 近期已取消。
// 不再扫 failed：本地已判定失败后继续查网关只会空转；迟到成功仍走 notify 回调入账。
// 已有永久查单失败异常的订单也不再扫。
func ListPaymentOrdersForReconcile(limit int, canceledLookbackSec int64) ([]PaymentOrder, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if canceledLookbackSec <= 0 {
		canceledLookbackSec = 7 * 24 * 3600 // 默认回看 7 天
	}
	now := time.Now().Unix()
	cutoff := now - canceledLookbackSec
	pendingWindowEnd := now + 30*60
	pendingCreatedBefore := now - 120

	var orders []PaymentOrder
	err := db.DB.Where(
		`gateway_id > 0 AND (
			(status = ? AND (expire_at = 0 OR expire_at <= ? OR create_time <= ?))
			OR (status = ? AND update_time >= ?)
		) AND order_no NOT IN (
			SELECT order_no FROM payment_exceptions
			WHERE exception_type IN (?, ?) AND order_no <> ''
		)`,
		PaymentStatusPending, pendingWindowEnd, pendingCreatedBefore,
		PaymentStatusCanceled, cutoff,
		PaymentExceptionOrderMissing, PaymentExceptionPermanentRejected,
	).Order("update_time ASC").Limit(limit).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	normalizePaymentOrders(orders)
	return orders, nil
}
