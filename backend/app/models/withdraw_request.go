package models

import (
	"database/sql"
	"fst/backend/pkg/db"
	"log"
	"strings"
	"time"
)

const (
	WithdrawStatusPending  = 0 // 待审核
	WithdrawStatusApproved = 1 // 已审核待打款
	WithdrawStatusRejected = 2 // 已拒绝
	WithdrawStatusPaid     = 3 // 已打款完成
)

type WithdrawRequest struct {
	ID              uint64  `db:"id" json:"id"`
	UserID          uint64  `db:"user_id" json:"user_id"`
	Amount          float64 `db:"amount" json:"amount"`
	AccountType     string  `db:"account_type" json:"account_type"`
	AccountName     string  `db:"account_name" json:"account_name"`
	AccountNo       string  `db:"account_no" json:"account_no"`
	RealName        string  `db:"real_name" json:"real_name"`
	Remark          string  `db:"remark" json:"remark"`
	Status          uint8   `db:"status" json:"status"`
	BalanceDeducted bool    `db:"balance_deducted" json:"balance_deducted"`
	ReviewRemark    string  `db:"review_remark" json:"review_remark"`
	TransferRemark  string  `db:"transfer_remark" json:"transfer_remark"`
	ReviewedAt      *int64  `db:"reviewed_at" json:"reviewed_at"`
	ReviewedBy      *uint64 `db:"reviewed_by" json:"reviewed_by"`
	PaidAt          *int64  `db:"paid_at" json:"paid_at"`
	PaidBy          *uint64 `db:"paid_by" json:"paid_by"`
	CreateTime      int64   `db:"create_time" json:"create_time"`
	UpdateTime      int64   `db:"update_time" json:"update_time"`
	DeleteTime      *int64  `db:"delete_time" json:"delete_time,omitempty"`
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
	PendingCount  int64   `db:"pending_count" json:"pending_count"`
	ApprovedCount int64   `db:"approved_count" json:"approved_count"`
	RejectedCount int64   `db:"rejected_count" json:"rejected_count"`
	PaidCount     int64   `db:"paid_count" json:"paid_count"`
	PaidAmount    float64 `db:"paid_amount" json:"paid_amount"`
}

func InitWithdrawRequestsTable() {
	if db.CheckTableExists("withdraw_requests") {
		db.EnsureIndex("withdraw_requests", "idx_user_status_create", "ALTER TABLE withdraw_requests ADD INDEX idx_user_status_create (user_id, status, create_time)")
		db.EnsureIndex("withdraw_requests", "idx_status_create", "ALTER TABLE withdraw_requests ADD INDEX idx_status_create (status, create_time)")
		if !db.CheckColumnExists("withdraw_requests", "balance_deducted") {
			if _, err := db.Exec("ALTER TABLE withdraw_requests ADD COLUMN balance_deducted TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已预扣余额:0=否,1=是' AFTER status"); err != nil {
				log.Printf("[Init] Failed to add withdraw_requests.balance_deducted: %v", err)
			} else {
				log.Printf("[Init] Added withdraw_requests.balance_deducted column")
			}
		}
		return
	}

	schema := `CREATE TABLE IF NOT EXISTS withdraw_requests (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
		amount DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '提现金额',
		account_type VARCHAR(32) NOT NULL DEFAULT 'bank' COMMENT '收款方式',
		account_name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '收款账户名称',
		account_no VARCHAR(128) NOT NULL DEFAULT '' COMMENT '收款账号',
		real_name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '收款人姓名',
		remark VARCHAR(255) NOT NULL DEFAULT '' COMMENT '用户备注',
		status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '状态:0=待审核,1=已审核待打款,2=已拒绝,3=已打款',
		balance_deducted TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已预扣余额:0=否,1=是',
		review_remark VARCHAR(255) NOT NULL DEFAULT '' COMMENT '审核备注',
		transfer_remark VARCHAR(255) NOT NULL DEFAULT '' COMMENT '打款备注',
		reviewed_at BIGINT NULL DEFAULT NULL COMMENT '审核时间',
		reviewed_by BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '审核管理员ID',
		paid_at BIGINT NULL DEFAULT NULL COMMENT '打款时间',
		paid_by BIGINT UNSIGNED NULL DEFAULT NULL COMMENT '打款管理员ID',
		create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
		update_time BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
		delete_time BIGINT NULL DEFAULT NULL COMMENT '删除时间',
		INDEX idx_user_id (user_id),
		INDEX idx_status (status),
		INDEX idx_create_time (create_time),
		INDEX idx_user_status_create (user_id, status, create_time),
		INDEX idx_status_create (status, create_time)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户提现申请表';`

	if _, err := db.Exec(schema); err != nil {
		log.Printf("[Init] Failed to create withdraw_requests table: %v", err)
	} else {
		log.Println("[Init] Created withdraw_requests table")
	}
}

func CreateWithdrawRequest(req *WithdrawRequest) error {
	now := time.Now().Unix()
	req.CreateTime = now
	req.UpdateTime = now
	result, err := db.Exec(
		`INSERT INTO withdraw_requests (user_id, amount, account_type, account_name, account_no, real_name, remark, status, balance_deducted, review_remark, transfer_remark, reviewed_at, reviewed_by, paid_at, paid_by, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.UserID, req.Amount, req.AccountType, req.AccountName, req.AccountNo, req.RealName, req.Remark,
		req.Status, req.BalanceDeducted, req.ReviewRemark, req.TransferRemark, req.ReviewedAt, req.ReviewedBy, req.PaidAt, req.PaidBy,
		req.CreateTime, req.UpdateTime, req.DeleteTime,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	req.ID = uint64(id)
	return nil
}

// CreateWithdrawRequestTx 在已有事务中创建提现申请（与预扣余额同事务时用）
func CreateWithdrawRequestTx(tx *sql.Tx, req *WithdrawRequest) error {
	now := time.Now().Unix()
	req.CreateTime = now
	req.UpdateTime = now
	result, err := tx.Exec(
		`INSERT INTO withdraw_requests (user_id, amount, account_type, account_name, account_no, real_name, remark, status, balance_deducted, review_remark, transfer_remark, reviewed_at, reviewed_by, paid_at, paid_by, create_time, update_time, delete_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.UserID, req.Amount, req.AccountType, req.AccountName, req.AccountNo, req.RealName, req.Remark,
		req.Status, req.BalanceDeducted, req.ReviewRemark, req.TransferRemark, req.ReviewedAt, req.ReviewedBy, req.PaidAt, req.PaidBy,
		req.CreateTime, req.UpdateTime, req.DeleteTime,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	req.ID = uint64(id)
	return nil
}

func GetWithdrawRequestByID(id uint64) (*WithdrawRequest, error) {
	var item WithdrawRequest
	err := db.DB.Get(&item, "SELECT * FROM withdraw_requests WHERE id = ? AND delete_time IS NULL", id)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func GetWithdrawRequestByIDForUpdate(tx *sql.Tx, id uint64) (*WithdrawRequest, error) {
	var item WithdrawRequest
	err := tx.QueryRow(
		db.Q(`SELECT id, user_id, amount, account_type, account_name, account_no, real_name, remark, status, balance_deducted, review_remark, transfer_remark, reviewed_at, reviewed_by, paid_at, paid_by, create_time, update_time, delete_time
		 FROM withdraw_requests WHERE id = ? AND delete_time IS NULL FOR UPDATE`), id,
	).Scan(
		&item.ID, &item.UserID, &item.Amount, &item.AccountType, &item.AccountName, &item.AccountNo, &item.RealName,
		&item.Remark, &item.Status, &item.BalanceDeducted, &item.ReviewRemark, &item.TransferRemark, &item.ReviewedAt, &item.ReviewedBy,
		&item.PaidAt, &item.PaidBy, &item.CreateTime, &item.UpdateTime, &item.DeleteTime,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
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

	where := " WHERE delete_time IS NULL "
	args := []interface{}{}

	if query.UserID > 0 {
		where += " AND user_id = ?"
		args = append(args, query.UserID)
	}
	if query.Status != nil {
		where += " AND status = ?"
		args = append(args, *query.Status)
	}
	keyword := strings.TrimSpace(query.Keyword)
	if keyword != "" {
		where += " AND (account_name LIKE ? OR account_no LIKE ? OR real_name LIKE ? OR remark LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw, kw)
	}

	var total int64
	if err := db.DB.Get(&total, "SELECT COUNT(*) FROM withdraw_requests"+where, args...); err != nil {
		return nil, err
	}

	list := make([]WithdrawRequest, 0)
	offset := (page - 1) * pageSize
	listArgs := append(append([]interface{}{}, args...), pageSize, offset)
	if err := db.DB.Select(&list, "SELECT * FROM withdraw_requests"+where+" ORDER BY create_time DESC LIMIT ? OFFSET ?", listArgs...); err != nil {
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
	if query == nil {
		query = &WithdrawListQuery{}
	}

	where := " WHERE delete_time IS NULL "
	args := []interface{}{}

	if query.UserID > 0 {
		where += " AND user_id = ?"
		args = append(args, query.UserID)
	}
	if query.Status != nil {
		where += " AND status = ?"
		args = append(args, *query.Status)
	}
	keyword := strings.TrimSpace(query.Keyword)
	if keyword != "" {
		where += " AND (account_name LIKE ? OR account_no LIKE ? OR real_name LIKE ? OR remark LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw, kw)
	}

	var result WithdrawStatsResult
	querySQL := `SELECT
		COALESCE(SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END), 0) AS pending_count,
		COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0) AS approved_count,
		COALESCE(SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END), 0) AS rejected_count,
		COALESCE(SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END), 0) AS paid_count,
		COALESCE(SUM(CASE WHEN status = 3 THEN amount ELSE 0 END), 0) AS paid_amount
		FROM withdraw_requests` + where
	if err := db.DB.Get(&result, querySQL, args...); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateWithdrawReviewTx 更新提现审核结果。
// clearBalanceDeducted：拒绝且此前已预扣余额（并已在同事务中退回）时传 true，
// 同步把 balance_deducted 清零，避免字段语义与实际余额状态不一致。
func UpdateWithdrawReviewTx(tx *sql.Tx, id uint64, status uint8, reviewRemark string, adminID uint64, clearBalanceDeducted bool) error {
	now := time.Now().Unix()
	if clearBalanceDeducted {
		_, err := tx.Exec(
			"UPDATE withdraw_requests SET status = ?, review_remark = ?, reviewed_at = ?, reviewed_by = ?, balance_deducted = 0, update_time = ? WHERE id = ? AND delete_time IS NULL",
			status, reviewRemark, now, adminID, now, id,
		)
		return err
	}
	_, err := tx.Exec(
		"UPDATE withdraw_requests SET status = ?, review_remark = ?, reviewed_at = ?, reviewed_by = ?, update_time = ? WHERE id = ? AND delete_time IS NULL",
		status, reviewRemark, now, adminID, now, id,
	)
	return err
}

func MarkWithdrawPaidTx(tx *sql.Tx, id uint64, transferRemark string, adminID uint64) error {
	now := time.Now().Unix()
	_, err := tx.Exec(
		"UPDATE withdraw_requests SET status = ?, transfer_remark = ?, paid_at = ?, paid_by = ?, update_time = ? WHERE id = ? AND delete_time IS NULL",
		WithdrawStatusPaid, transferRemark, now, adminID, now, id,
	)
	return err
}
