package models

import (
	"database/sql"
	"fst/backend/pkg/db"
	"log"
	"strings"
	"time"
)

// RealnameVerification 实名认证记录
type RealnameVerification struct {
	ID               uint64  `db:"id" json:"id"`
	UserID           uint64  `db:"user_id" json:"user_id"`
	RealName         string  `db:"real_name" json:"real_name"`                 // 真实姓名
	CertificateType  uint8   `db:"certificate_type" json:"certificate_type"`   // 证件类型: 1=身份证, 2=护照, 3=军官证
	CertificateNo    string  `db:"certificate_no" json:"certificate_no"`       // 证件号码
	CertUniqueKey    *string `db:"cert_unique_key" json:"-"`                   // 有效证件唯一键（待审/通过时有值，拒绝/软删为 NULL）
	CertificateFront string  `db:"certificate_front" json:"certificate_front"` // 证件正面照URL
	CertificateBack  string  `db:"certificate_back" json:"certificate_back"`   // 证件背面照URL
	Status           uint8   `db:"status" json:"status"`                       // 状态: 0=待审核, 1=通过, 2=拒绝
	RejectReason     string  `db:"reject_reason" json:"reject_reason"`         // 拒绝原因
	SubmittedAt      *int64  `db:"submitted_at" json:"submitted_at"`           // 提交时间
	ReviewedAt       *int64  `db:"reviewed_at" json:"reviewed_at"`             // 审核时间
	ReviewedBy       *uint64 `db:"reviewed_by" json:"reviewed_by"`             // 审核人ID
	CreateTime       *int64  `db:"create_time" json:"create_time"`
	UpdateTime       *int64  `db:"update_time" json:"update_time"`
	DeleteTime       *int64  `db:"delete_time" json:"-"`
}

// BuildCertUniqueKey 有效实名记录的唯一键：待审/通过时返回规范化证件号，否则返回 nil（允许多条拒绝记录）
func BuildCertUniqueKey(certificateNo string, status uint8, deleted bool) *string {
	if deleted || (status != 0 && status != 1) {
		return nil
	}
	key := strings.ToUpper(strings.TrimSpace(certificateNo))
	if key == "" {
		return nil
	}
	return &key
}

func (r *RealnameVerification) TableName() string {
	return "user_realname_verifications"
}

// CreateRealnameVerification 创建实名认证记录
func CreateRealnameVerification(verification *RealnameVerification) error {
	query := `INSERT INTO user_realname_verifications (
		user_id, real_name, certificate_type, certificate_no, cert_unique_key,
		certificate_front, certificate_back, status, reject_reason, submitted_at, reviewed_at, reviewed_by, create_time, update_time
	) VALUES (
		:user_id, :real_name, :certificate_type, :certificate_no, :cert_unique_key,
		:certificate_front, :certificate_back, :status, :reject_reason, :submitted_at, :reviewed_at, :reviewed_by, :create_time, :update_time
	)`

	now := time.Now().Unix()
	verification.CreateTime = &now
	verification.UpdateTime = &now
	verification.SubmittedAt = &now
	// 默认走待审核；当服务层根据配置显式写入“已通过”时，这里保留该状态不覆盖。
	if verification.Status != 1 && verification.Status != 2 {
		verification.Status = 0
	}
	verification.CertUniqueKey = BuildCertUniqueKey(verification.CertificateNo, verification.Status, false)

	result, err := db.DB.NamedExec(query, verification)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	verification.ID = uint64(id)
	return nil
}

// GetRealnameVerificationByUserID 根据用户ID获取最新的实名认证记录
func GetRealnameVerificationByUserID(userID uint64) (*RealnameVerification, error) {
	var verification RealnameVerification
	err := db.DB.Get(&verification,
		"SELECT * FROM user_realname_verifications WHERE user_id = ? AND delete_time IS NULL ORDER BY id DESC LIMIT 1",
		userID)
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

// GetRealnameVerificationByUserIDForUpdate 在事务中锁定用户最新实名认证记录
func GetRealnameVerificationByUserIDForUpdate(tx *sql.Tx, userID uint64) (*RealnameVerification, error) {
	var verification RealnameVerification
	err := tx.QueryRow(
		db.Q(`SELECT id, user_id, real_name, certificate_type, certificate_no, certificate_front, certificate_back,
		        status, reject_reason, submitted_at, reviewed_at, reviewed_by, create_time, update_time, delete_time
		   FROM user_realname_verifications
		  WHERE user_id = ? AND delete_time IS NULL
		  ORDER BY id DESC
		  LIMIT 1
		  FOR UPDATE`),
		userID,
	).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.RealName,
		&verification.CertificateType,
		&verification.CertificateNo,
		&verification.CertificateFront,
		&verification.CertificateBack,
		&verification.Status,
		&verification.RejectReason,
		&verification.SubmittedAt,
		&verification.ReviewedAt,
		&verification.ReviewedBy,
		&verification.CreateTime,
		&verification.UpdateTime,
		&verification.DeleteTime,
	)
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

// GetRealnameVerificationByID 根据ID获取实名认证记录
func GetRealnameVerificationByID(id uint64) (*RealnameVerification, error) {
	var verification RealnameVerification
	err := db.DB.Get(&verification,
		"SELECT * FROM user_realname_verifications WHERE id = ? AND delete_time IS NULL",
		id)
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

// GetRealnameVerificationByIDForUpdate 在事务中锁定实名认证记录
func GetRealnameVerificationByIDForUpdate(tx *sql.Tx, id uint64) (*RealnameVerification, error) {
	var verification RealnameVerification
	err := tx.QueryRow(
		db.Q(`SELECT id, user_id, real_name, certificate_type, certificate_no, certificate_front, certificate_back,
		        status, reject_reason, submitted_at, reviewed_at, reviewed_by, create_time, update_time, delete_time
		   FROM user_realname_verifications
		  WHERE id = ? AND delete_time IS NULL
		  FOR UPDATE`),
		id,
	).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.RealName,
		&verification.CertificateType,
		&verification.CertificateNo,
		&verification.CertificateFront,
		&verification.CertificateBack,
		&verification.Status,
		&verification.RejectReason,
		&verification.SubmittedAt,
		&verification.ReviewedAt,
		&verification.ReviewedBy,
		&verification.CreateTime,
		&verification.UpdateTime,
		&verification.DeleteTime,
	)
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

// GetRealnameVerificationByIDIncludeDeleted 根据ID获取实名认证记录（包含已删除的）
func GetRealnameVerificationByIDIncludeDeleted(id uint64) (*RealnameVerification, error) {
	var verification RealnameVerification
	err := db.DB.Get(&verification,
		"SELECT * FROM user_realname_verifications WHERE id = ?",
		id)
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

// UpdateRealnameVerificationStatus 更新实名认证状态
func UpdateRealnameVerificationStatus(id uint64, status uint8, rejectReason string, reviewedBy uint64) error {
	now := time.Now().Unix()
	// 拒绝时清空唯一键，允许该证件号再次被其他用户申请；通过/待审保留唯一键
	var certKey interface{}
	if status == 0 || status == 1 {
		var no string
		if err := db.DB.Get(&no, "SELECT certificate_no FROM user_realname_verifications WHERE id = ? AND delete_time IS NULL", id); err != nil {
			return err
		}
		certKey = strings.ToUpper(strings.TrimSpace(no))
	} else {
		certKey = nil
	}
	query := `UPDATE user_realname_verifications
			  SET status = ?, reject_reason = ?, reviewed_at = ?, reviewed_by = ?, cert_unique_key = ?, update_time = ?
			  WHERE id = ? AND delete_time IS NULL`
	result, err := db.Exec(query, status, rejectReason, now, reviewedBy, certKey, now, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateRealnameVerificationStatusTx 在事务中更新实名认证状态
func UpdateRealnameVerificationStatusTx(tx *sql.Tx, id uint64, status uint8, rejectReason string, reviewedBy uint64) error {
	now := time.Now().Unix()
	var certKey interface{}
	if status == 0 || status == 1 {
		var no string
		if err := tx.QueryRow(db.Q("SELECT certificate_no FROM user_realname_verifications WHERE id = ? AND delete_time IS NULL"), id).Scan(&no); err != nil {
			return err
		}
		certKey = strings.ToUpper(strings.TrimSpace(no))
	} else {
		certKey = nil
	}
	result, err := tx.Exec(
		db.Q(`UPDATE user_realname_verifications
		    SET status = ?, reject_reason = ?, reviewed_at = ?, reviewed_by = ?, cert_unique_key = ?, update_time = ?
		  WHERE id = ? AND delete_time IS NULL`),
		status, rejectReason, now, reviewedBy, certKey, now, id,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SoftDeleteRealnameVerification 软删除实名认证记录
func SoftDeleteRealnameVerification(id uint64) error {
	now := time.Now().Unix()
	query := `UPDATE user_realname_verifications SET delete_time = ?, cert_unique_key = NULL, update_time = ? WHERE id = ? AND delete_time IS NULL`
	_, err := db.Exec(query, now, now, id)
	return err
}

// SoftDeleteRealnameVerificationTx 在事务中软删除实名认证记录
func SoftDeleteRealnameVerificationTx(tx *sql.Tx, id uint64) error {
	now := time.Now().Unix()
	_, err := tx.Exec(
		db.Q("UPDATE user_realname_verifications SET delete_time = ?, cert_unique_key = NULL, update_time = ? WHERE id = ? AND delete_time IS NULL"),
		now, now, id,
	)
	return err
}

// CreateRealnameVerificationTx 在事务中创建实名认证记录
func CreateRealnameVerificationTx(tx *sql.Tx, verification *RealnameVerification) error {
	query := `INSERT INTO user_realname_verifications (
		user_id, real_name, certificate_type, certificate_no, cert_unique_key,
		certificate_front, certificate_back, status, reject_reason, submitted_at, reviewed_at, reviewed_by, create_time, update_time
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	)`

	now := time.Now().Unix()
	verification.CreateTime = &now
	verification.UpdateTime = &now
	verification.SubmittedAt = &now
	if verification.Status != 1 && verification.Status != 2 {
		verification.Status = 0
	}
	verification.CertUniqueKey = BuildCertUniqueKey(verification.CertificateNo, verification.Status, false)

	result, err := tx.Exec(
		db.Q(query),
		verification.UserID,
		verification.RealName,
		verification.CertificateType,
		verification.CertificateNo,
		verification.CertUniqueKey,
		verification.CertificateFront,
		verification.CertificateBack,
		verification.Status,
		verification.RejectReason,
		verification.SubmittedAt,
		verification.ReviewedAt,
		verification.ReviewedBy,
		verification.CreateTime,
		verification.UpdateTime,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	verification.ID = uint64(id)
	return nil
}

// EnsureRealnameCertUniqueConstraint 补列 + 回填 + 唯一索引，保障并发下有效证件号不可重复占用
func EnsureRealnameCertUniqueConstraint() {
	table := "user_realname_verifications"
	if !db.CheckTableExists(table) {
		return
	}
	if !db.CheckColumnExists(table, "cert_unique_key") {
		alter := "ALTER TABLE user_realname_verifications ADD COLUMN cert_unique_key VARCHAR(64) NULL DEFAULT NULL COMMENT '有效证件唯一键(待审/通过)'"
		if _, err := db.Exec(alter); err != nil {
			log.Printf("[Init] add cert_unique_key failed: %v", err)
			return
		}
		log.Println("[Init] Added user_realname_verifications.cert_unique_key")
	}
	// 回填有效记录唯一键（冲突时跳过，留给业务层处理）
	if _, err := db.Exec(`UPDATE user_realname_verifications
		SET cert_unique_key = UPPER(TRIM(certificate_no))
		WHERE delete_time IS NULL AND status IN (0, 1)
		  AND (cert_unique_key IS NULL OR cert_unique_key = '')
		  AND certificate_no IS NOT NULL AND TRIM(certificate_no) <> ''`); err != nil {
		log.Printf("[Init] backfill cert_unique_key failed: %v", err)
	}
	// 拒绝/软删记录清空唯一键
	if _, err := db.Exec(`UPDATE user_realname_verifications SET cert_unique_key = NULL
		WHERE (delete_time IS NOT NULL OR status = 2) AND cert_unique_key IS NOT NULL`); err != nil {
		log.Printf("[Init] clear inactive cert_unique_key failed: %v", err)
	}
	db.EnsureIndex(table, "uk_realname_cert_unique_key",
		"ALTER TABLE user_realname_verifications ADD UNIQUE INDEX uk_realname_cert_unique_key (cert_unique_key)")
}

// CountOtherUsersByCertificateNoTx 在事务中统计「其他用户」使用同一证件号且状态为待审核/已通过的记录数，
// 用于提交前查重，防止同一证件号被多个账号占用实名认证。
func CountOtherUsersByCertificateNoTx(tx *sql.Tx, certificateNo string, excludeUserID uint64) (int64, error) {
	var count int64
	err := tx.QueryRow(
		db.Q("SELECT COUNT(*) FROM user_realname_verifications WHERE certificate_no = ? AND user_id != ? AND status IN (0, 1) AND delete_time IS NULL"),
		certificateNo, excludeUserID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetUserRealnameVerifications 获取用户的实名认证记录列表
func GetUserRealnameVerifications(userID uint64) ([]RealnameVerification, error) {
	var list []RealnameVerification
	err := db.DB.Select(&list,
		"SELECT * FROM user_realname_verifications WHERE user_id = ? AND delete_time IS NULL ORDER BY id DESC",
		userID)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// RealnameVerificationListQuery 实名认证列表查询参数
type RealnameVerificationListQuery struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
	Keyword  string `form:"keyword" json:"keyword"`
	Status   *uint8 `form:"status" json:"status"`
	UserID   uint64 `form:"user_id" json:"user_id"`
}

// RealnameVerificationListResult 实名认证列表返回结果
type RealnameVerificationListResult struct {
	List     []RealnameVerification `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// GetRealnameVerificationList 获取实名认证列表（管理员）
func GetRealnameVerificationList(query *RealnameVerificationListQuery) (*RealnameVerificationListResult, error) {
	if query == nil {
		query = &RealnameVerificationListQuery{}
	}

	var list []RealnameVerification
	var total int64

	// 默认分页参数
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	// 构建查询条件
	where := "WHERE delete_time IS NULL"
	args := []interface{}{}

	if query.Keyword != "" {
		where += " AND (real_name LIKE ? OR certificate_no LIKE ?)"
		kw := "%" + query.Keyword + "%"
		args = append(args, kw, kw)
	}
	if query.Status != nil {
		where += " AND status = ?"
		args = append(args, *query.Status)
	}
	if query.UserID > 0 {
		where += " AND user_id = ?"
		args = append(args, query.UserID)
	}

	// 查询总数
	count_query := "SELECT COUNT(*) FROM user_realname_verifications " + where
	err := db.DB.Get(&total, count_query, args...)
	if err != nil {
		return nil, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	list_query := "SELECT * FROM user_realname_verifications " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, query.PageSize, offset)

	err = db.DB.Select(&list, list_query, args...)
	if err != nil {
		return nil, err
	}

	return &RealnameVerificationListResult{
		List:     list,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}
