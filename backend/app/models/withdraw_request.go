package models

import (
	"fst/backend/pkg/db"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	WithdrawStatusPending  = 0 // 待审核
	WithdrawStatusApproved = 1 // 已审核待打款
	WithdrawStatusRejected = 2 // 已拒绝
	WithdrawStatusPaid     = 3 // 已打款完成
)

type WithdrawRequest struct {
	ID              uint64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID          uint64  `gorm:"column:user_id;not null;default:0;index:idx_wr_user_id;index:idx_wr_user_status_create,priority:1" json:"user_id"`
	Amount          float64 `gorm:"column:amount;type:decimal(10,2);not null;default:0" json:"amount"`
	AccountType     string  `gorm:"column:account_type;size:32;not null;default:'bank'" json:"account_type"`
	AccountName     string  `gorm:"column:account_name;size:100;not null;default:''" json:"account_name"`
	AccountNo       string  `gorm:"column:account_no;size:128;not null;default:''" json:"account_no"`
	RealName        string  `gorm:"column:real_name;size:100;not null;default:''" json:"real_name"`
	Remark          string  `gorm:"column:remark;size:255;not null;default:''" json:"remark"`
	Status          uint8   `gorm:"column:status;not null;default:0;index:idx_wr_status;index:idx_wr_user_status_create,priority:2;index:idx_wr_status_create,priority:1" json:"status"`
	BalanceDeducted bool    `gorm:"column:balance_deducted;not null;default:false" json:"balance_deducted"`
	ReviewRemark    string  `gorm:"column:review_remark;size:255;not null;default:''" json:"review_remark"`
	TransferRemark  string  `gorm:"column:transfer_remark;size:255;not null;default:''" json:"transfer_remark"`
	ReviewedAt      *int64  `gorm:"column:reviewed_at" json:"reviewed_at"`
	ReviewedBy      *uint64 `gorm:"column:reviewed_by" json:"reviewed_by"`
	PaidAt          *int64  `gorm:"column:paid_at" json:"paid_at"`
	PaidBy          *uint64 `gorm:"column:paid_by" json:"paid_by"`
	CreateTime      int64   `gorm:"column:create_time;not null;default:0;index:idx_wr_create_time;index:idx_wr_user_status_create,priority:3;index:idx_wr_status_create,priority:2" json:"create_time"`
	UpdateTime      int64   `gorm:"column:update_time;not null;default:0" json:"update_time"`
	DeleteTime      *int64  `gorm:"column:delete_time" json:"delete_time,omitempty"`
}

func (WithdrawRequest) TableName() string {
	return "withdraw_requests"
}

type WithdrawListQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
	UserID   uint64 `form:"user_id"`
	Status   *uint8 `form:"status"`
}

type WithdrawListResult struct {
	List     []WithdrawRequest `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

type WithdrawStatsResult struct {
	PendingCount  int64   `gorm:"column:pending_count" json:"pending_count"`
	ApprovedCount int64   `gorm:"column:approved_count" json:"approved_count"`
	RejectedCount int64   `gorm:"column:rejected_count" json:"rejected_count"`
	PaidCount     int64   `gorm:"column:paid_count" json:"paid_count"`
	PaidAmount    float64 `gorm:"column:paid_amount" json:"paid_amount"`
}

func CreateWithdrawRequest(req *WithdrawRequest) error {
	now := time.Now().Unix()
	req.CreateTime = now
	req.UpdateTime = now
	return db.DB.Create(req).Error
}

// CreateWithdrawRequestTx 在已有事务中创建提现申请（与预扣余额同事务时用）
func CreateWithdrawRequestTx(tx *gorm.DB, req *WithdrawRequest) error {
	now := time.Now().Unix()
	req.CreateTime = now
	req.UpdateTime = now
	return tx.Create(req).Error
}

func GetWithdrawRequestByID(id uint64) (*WithdrawRequest, error) {
	var item WithdrawRequest
	err := db.DB.Where("id = ? AND delete_time IS NULL", id).First(&item).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	return &item, nil
}

func GetWithdrawRequestByIDForUpdate(tx *gorm.DB, id uint64) (*WithdrawRequest, error) {
	var item WithdrawRequest
	err := db.ForUpdate(tx).
		Where("id = ? AND delete_time IS NULL", id).
		First(&item).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	return &item, nil
}

func buildWithdrawRequestQuery(query *WithdrawListQuery) *gorm.DB {
	q := db.DB.Model(&WithdrawRequest{}).Where("delete_time IS NULL")
	if query == nil {
		return q
	}
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	keyword := strings.TrimSpace(query.Keyword)
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("(account_name LIKE ? OR account_no LIKE ? OR real_name LIKE ? OR remark LIKE ?)", kw, kw, kw, kw)
	}
	return q
}

func GetWithdrawRequestList(query *WithdrawListQuery) (*WithdrawListResult, error) {
	if query == nil {
		query = &WithdrawListQuery{}
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	q := buildWithdrawRequestQuery(query)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	list := make([]WithdrawRequest, 0)
	offset := (page - 1) * pageSize
	if err := q.Order("create_time DESC").Limit(pageSize).Offset(offset).Find(&list).Error; err != nil {
		return nil, err
	}

	return &WithdrawListResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func GetWithdrawRequestStats(query *WithdrawListQuery) (*WithdrawStatsResult, error) {
	q := buildWithdrawRequestQuery(query)
	var result WithdrawStatsResult
	err := q.Select(`
		COALESCE(SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END), 0) AS pending_count,
		COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0) AS approved_count,
		COALESCE(SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END), 0) AS rejected_count,
		COALESCE(SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END), 0) AS paid_count,
		COALESCE(SUM(CASE WHEN status = 3 THEN amount ELSE 0 END), 0) AS paid_amount
	`).Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateWithdrawReviewTx 更新提现审核结果。
// clearBalanceDeducted：拒绝且此前已预扣余额（并已在同事务中退回）时传 true，
// 同步把 balance_deducted 清零，避免字段语义与实际余额状态不一致。
func UpdateWithdrawReviewTx(tx *gorm.DB, id uint64, status uint8, reviewRemark string, adminID uint64, clearBalanceDeducted bool) error {
	now := time.Now().Unix()
	updates := map[string]interface{}{
		"status":        status,
		"review_remark": reviewRemark,
		"reviewed_at":   now,
		"reviewed_by":   adminID,
		"update_time":   now,
	}
	if clearBalanceDeducted {
		updates["balance_deducted"] = false
	}
	return tx.Model(&WithdrawRequest{}).
		Where("id = ? AND delete_time IS NULL", id).
		Updates(updates).Error
}

func MarkWithdrawPaidTx(tx *gorm.DB, id uint64, transferRemark string, adminID uint64) error {
	now := time.Now().Unix()
	return tx.Model(&WithdrawRequest{}).
		Where("id = ? AND delete_time IS NULL", id).
		Updates(map[string]interface{}{
			"status":          WithdrawStatusPaid,
			"transfer_remark": transferRemark,
			"paid_at":         now,
			"paid_by":         adminID,
			"update_time":     now,
		}).Error
}

// ListWithdrawLegacyBalanceRisk 只读：已通过/已打款但未标记预扣的历史风险单（排查重复扣款）
func ListWithdrawLegacyBalanceRisk(limit int) ([]WithdrawRequest, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var list []WithdrawRequest
	err := db.DB.
		Where("delete_time IS NULL AND balance_deducted = ? AND status IN (?, ?)", false, WithdrawStatusApproved, WithdrawStatusPaid).
		Order("id ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}
