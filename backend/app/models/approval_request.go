package models

import (
	"database/sql"
	"encoding/json"
	"fst/backend/pkg/db"
	"log"
	"time"
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
	ID          uint64  `db:"id" json:"id"`
	Type        string  `db:"type" json:"type"`
	PayloadJSON string  `db:"payload_json" json:"payload_json"`
	Status      string  `db:"status" json:"status"`
	RequesterID uint64  `db:"requester_id" json:"requester_id"`
	ReviewerID  *uint64 `db:"reviewer_id" json:"reviewer_id"`
	Comment     string  `db:"comment" json:"comment"`
	CreateTime  int64   `db:"create_time" json:"create_time"`
	ReviewTime  *int64  `db:"review_time" json:"review_time"`
}

// InitApprovalRequestsTable 初始化审批表
func InitApprovalRequestsTable() {
	schema := `CREATE TABLE IF NOT EXISTS approval_requests (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		type VARCHAR(64) NOT NULL DEFAULT '' COMMENT '审批类型',
		payload_json TEXT NOT NULL COMMENT '请求载荷JSON',
		status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/approved/rejected',
		requester_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '申请人',
		reviewer_id BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '审批人',
		comment VARCHAR(500) NOT NULL DEFAULT '' COMMENT '备注',
		create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
		review_time BIGINT NULL DEFAULT NULL COMMENT '审批时间',
		INDEX idx_ar_status_create (status, create_time),
		INDEX idx_ar_type_status (type, status),
		INDEX idx_ar_requester (requester_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='高危操作审批/审计'`

	if _, err := db.Exec(schema); err != nil {
		log.Printf("[Init] Failed to create approval_requests: %v", err)
	}
}

// CreateApprovalRequest 创建审批记录
func CreateApprovalRequest(req *ApprovalRequest) error {
	now := time.Now().Unix()
	req.CreateTime = now
	if req.Status == "" {
		req.Status = ApprovalStatusPending
	}
	res, err := db.Exec(`
		INSERT INTO approval_requests (type, payload_json, status, requester_id, reviewer_id, comment, create_time, review_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.Type, req.PayloadJSON, req.Status, req.RequesterID, req.ReviewerID, req.Comment, req.CreateTime, req.ReviewTime,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	req.ID = uint64(id)
	return nil
}

// GetApprovalRequestByID 按 ID 取审批
func GetApprovalRequestByID(id uint64) (*ApprovalRequest, error) {
	var item ApprovalRequest
	err := db.DB.Get(&item, `SELECT id, type, payload_json, status, requester_id, reviewer_id, comment, create_time, review_time
		FROM approval_requests WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ApproveApprovalRequest 审批通过（双人复核）
func ApproveApprovalRequest(id, reviewerID uint64, comment string) error {
	now := time.Now().Unix()
	res, err := db.Exec(`
		UPDATE approval_requests
		SET status = ?, reviewer_id = ?, comment = ?, review_time = ?
		WHERE id = ? AND status = ?`,
		ApprovalStatusApproved, reviewerID, comment, now, id, ApprovalStatusPending,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// RejectApprovalRequest 审批拒绝
func RejectApprovalRequest(id, reviewerID uint64, comment string) error {
	now := time.Now().Unix()
	res, err := db.Exec(`
		UPDATE approval_requests
		SET status = ?, reviewer_id = ?, comment = ?, review_time = ?
		WHERE id = ? AND status = ?`,
		ApprovalStatusRejected, reviewerID, comment, now, id, ApprovalStatusPending,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
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
	err := db.DB.Select(&list, `
		SELECT id, type, payload_json, status, requester_id, reviewer_id, comment, create_time, review_time
		FROM approval_requests WHERE status = ? ORDER BY id DESC LIMIT ?`,
		ApprovalStatusPending, limit)
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
