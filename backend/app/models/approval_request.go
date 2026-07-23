package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fst/backend/pkg/db"
	"time"

	"gorm.io/gorm"
)

// 审批状态
const (
	ApprovalStatusPending  = "pending"
	ApprovalStatusApproved = "approved"
	ApprovalStatusRejected = "rejected"
)

// 审批类型（高危财务）
const (
	ApprovalTypeForcePaymentComplete = "force_payment_complete"
	ApprovalTypeMoneyAdjust          = "money_adjust"
)

// ApprovalRequest 高危操作审批/审计记录
type ApprovalRequest struct {
	ID          uint64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Type        string  `gorm:"column:type;size:64;not null;default:''" json:"type"`
	PayloadJSON string  `gorm:"column:payload_json;type:text;not null" json:"payload_json"`
	Status      string  `gorm:"column:status;size:20;not null;default:'pending'" json:"status"`
	RequesterID uint64  `gorm:"column:requester_id;not null;default:0" json:"requester_id"`
	ReviewerID  *uint64 `gorm:"column:reviewer_id" json:"reviewer_id"`
	Comment     string  `gorm:"column:comment;size:500;not null;default:''" json:"comment"`
	CreateTime  int64   `gorm:"column:create_time;not null;default:0" json:"create_time"`
	ReviewTime  *int64  `gorm:"column:review_time" json:"review_time"`
}

// TableName 表名
func (ApprovalRequest) TableName() string { return "approval_requests" }

// CreateApprovalRequest 创建审批记录
func CreateApprovalRequest(req *ApprovalRequest) error {
	now := time.Now().Unix()
	req.CreateTime = now
	if req.Status == "" {
		req.Status = ApprovalStatusPending
	}
	return db.DB.Create(req).Error
}

// GetApprovalRequestByID 按 ID 取审批
func GetApprovalRequestByID(id uint64) (*ApprovalRequest, error) {
	var item ApprovalRequest
	err := db.DB.Where("id = ?", id).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ApproveApprovalRequest 审批通过（双人复核）
func ApproveApprovalRequest(id, reviewerID uint64, comment string) error {
	now := time.Now().Unix()
	r := db.DB.Model(&ApprovalRequest{}).
		Where("id = ? AND status = ?", id, ApprovalStatusPending).
		Updates(map[string]any{
			"status":      ApprovalStatusApproved,
			"reviewer_id": reviewerID,
			"comment":     comment,
			"review_time": now,
		})
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// RejectApprovalRequest 审批拒绝
func RejectApprovalRequest(id, reviewerID uint64, comment string) error {
	now := time.Now().Unix()
	r := db.DB.Model(&ApprovalRequest{}).
		Where("id = ? AND status = ?", id, ApprovalStatusPending).
		Updates(map[string]any{
			"status":      ApprovalStatusRejected,
			"reviewer_id": reviewerID,
			"comment":     comment,
			"review_time": now,
		})
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ListPendingApprovals 待审批列表
func ListPendingApprovals(limit int) ([]ApprovalRequest, error) {
	if limit <= 0 {
		limit = 50
	}
	var list []ApprovalRequest
	err := db.DB.Where("status = ?", ApprovalStatusPending).
		Order("id DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

// MarshalApprovalPayload 序列化载荷
func MarshalApprovalPayload(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
