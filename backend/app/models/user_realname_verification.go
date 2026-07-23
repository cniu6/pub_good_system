package models

import (
	"database/sql"
	"fst/backend/pkg/db"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// RealnameVerification 实名认证记录
type RealnameVerification struct {
	ID               uint64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID           uint64  `gorm:"column:user_id;not null;index:idx_rv_user_id" json:"user_id"`
	RealName         string  `gorm:"column:real_name;size:100;not null" json:"real_name"`
	CertificateType  uint8   `gorm:"column:certificate_type;not null" json:"certificate_type"`
	CertificateNo    string  `gorm:"column:certificate_no;size:50;not null" json:"certificate_no"`
	CertUniqueKey    *string `gorm:"column:cert_unique_key;size:64;uniqueIndex:uk_realname_cert_unique_key" json:"-"`
	CertificateFront string  `gorm:"column:certificate_front;size:500;not null" json:"certificate_front"`
	CertificateBack  string  `gorm:"column:certificate_back;size:500;not null" json:"certificate_back"`
	Status           uint8   `gorm:"column:status;not null;default:0;index:idx_rv_status" json:"status"`
	RejectReason     string  `gorm:"column:reject_reason;size:255;not null;default:''" json:"reject_reason"`
	SubmittedAt      *int64  `gorm:"column:submitted_at;index:idx_rv_submitted_at" json:"submitted_at"`
	ReviewedAt       *int64  `gorm:"column:reviewed_at" json:"reviewed_at"`
	ReviewedBy       *uint64 `gorm:"column:reviewed_by" json:"reviewed_by"`
	CreateTime       *int64  `gorm:"column:create_time" json:"create_time"`
	UpdateTime       *int64  `gorm:"column:update_time" json:"update_time"`
	DeleteTime       *int64  `gorm:"column:delete_time" json:"-"`
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
	now := time.Now().Unix()
	verification.CreateTime = &now
	verification.UpdateTime = &now
	verification.SubmittedAt = &now
	if verification.Status != 1 && verification.Status != 2 {
		verification.Status = 0
	}
	verification.CertUniqueKey = BuildCertUniqueKey(verification.CertificateNo, verification.Status, false)
	return db.DB.Create(verification).Error
}

// GetRealnameVerificationByUserID 根据用户ID获取最新的实名认证记录
func GetRealnameVerificationByUserID(userID uint64) (*RealnameVerification, error) {
	var verification RealnameVerification
	err := db.DB.Where("user_id = ? AND delete_time IS NULL", userID).
		Order("id DESC").
		First(&verification).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	return &verification, nil
}

// GetRealnameVerificationByUserIDForUpdate 在事务中锁定用户最新实名认证记录
func GetRealnameVerificationByUserIDForUpdate(tx *gorm.DB, userID uint64) (*RealnameVerification, error) {
	var verification RealnameVerification
	err := db.ForUpdate(tx).
		Where("user_id = ? AND delete_time IS NULL", userID).
		Order("id DESC").
		First(&verification).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	return &verification, nil
}

// GetRealnameVerificationByID 根据ID获取实名认证记录
func GetRealnameVerificationByID(id uint64) (*RealnameVerification, error) {
	var verification RealnameVerification
	err := db.DB.Where("id = ? AND delete_time IS NULL", id).First(&verification).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	return &verification, nil
}

// GetRealnameVerificationByIDForUpdate 在事务中锁定实名认证记录
func GetRealnameVerificationByIDForUpdate(tx *gorm.DB, id uint64) (*RealnameVerification, error) {
	var verification RealnameVerification
	err := db.ForUpdate(tx).
		Where("id = ? AND delete_time IS NULL", id).
		First(&verification).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	return &verification, nil
}

// GetRealnameVerificationByIDIncludeDeleted 根据ID获取实名认证记录（包含已删除的）
func GetRealnameVerificationByIDIncludeDeleted(id uint64) (*RealnameVerification, error) {
	var verification RealnameVerification
	err := db.DB.Where("id = ?", id).First(&verification).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	return &verification, nil
}

// UpdateRealnameVerificationStatus 更新实名认证状态
func UpdateRealnameVerificationStatus(id uint64, status uint8, rejectReason string, reviewedBy uint64) error {
	now := time.Now().Unix()
	var certKey interface{}
	if status == 0 || status == 1 {
		var row RealnameVerification
		if err := db.DB.Select("certificate_no").
			Where("id = ? AND delete_time IS NULL", id).
			First(&row).Error; err != nil {
			return db.MapGormNotFound(err)
		}
		certKey = strings.ToUpper(strings.TrimSpace(row.CertificateNo))
	} else {
		certKey = nil
	}
	res := db.DB.Model(&RealnameVerification{}).
		Where("id = ? AND delete_time IS NULL", id).
		Updates(map[string]interface{}{
			"status":          status,
			"reject_reason":   rejectReason,
			"reviewed_at":     now,
			"reviewed_by":     reviewedBy,
			"cert_unique_key": certKey,
			"update_time":     now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateRealnameVerificationStatusTx 在事务中更新实名认证状态
func UpdateRealnameVerificationStatusTx(tx *gorm.DB, id uint64, status uint8, rejectReason string, reviewedBy uint64) error {
	now := time.Now().Unix()
	var certKey interface{}
	if status == 0 || status == 1 {
		var row RealnameVerification
		if err := tx.Select("certificate_no").
			Where("id = ? AND delete_time IS NULL", id).
			First(&row).Error; err != nil {
			return db.MapGormNotFound(err)
		}
		certKey = strings.ToUpper(strings.TrimSpace(row.CertificateNo))
	} else {
		certKey = nil
	}
	res := tx.Model(&RealnameVerification{}).
		Where("id = ? AND delete_time IS NULL", id).
		Updates(map[string]interface{}{
			"status":          status,
			"reject_reason":   rejectReason,
			"reviewed_at":     now,
			"reviewed_by":     reviewedBy,
			"cert_unique_key": certKey,
			"update_time":     now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SoftDeleteRealnameVerification 软删除实名认证记录
func SoftDeleteRealnameVerification(id uint64) error {
	now := time.Now().Unix()
	return db.DB.Model(&RealnameVerification{}).
		Where("id = ? AND delete_time IS NULL", id).
		Updates(map[string]interface{}{
			"delete_time":     now,
			"cert_unique_key": nil,
			"update_time":     now,
		}).Error
}

// SoftDeleteRealnameVerificationTx 在事务中软删除实名认证记录
func SoftDeleteRealnameVerificationTx(tx *gorm.DB, id uint64) error {
	now := time.Now().Unix()
	return tx.Model(&RealnameVerification{}).
		Where("id = ? AND delete_time IS NULL", id).
		Updates(map[string]interface{}{
			"delete_time":     now,
			"cert_unique_key": nil,
			"update_time":     now,
		}).Error
}

// CreateRealnameVerificationTx 在事务中创建实名认证记录
func CreateRealnameVerificationTx(tx *gorm.DB, verification *RealnameVerification) error {
	now := time.Now().Unix()
	verification.CreateTime = &now
	verification.UpdateTime = &now
	verification.SubmittedAt = &now
	if verification.Status != 1 && verification.Status != 2 {
		verification.Status = 0
	}
	verification.CertUniqueKey = BuildCertUniqueKey(verification.CertificateNo, verification.Status, false)
	return tx.Create(verification).Error
}

// EnsureRealnameCertUniqueConstraint 回填 cert_unique_key（列/唯一索引由 AutoMigrate + gorm tag 负责）。
func EnsureRealnameCertUniqueConstraint() {
	table := "user_realname_verifications"
	if !db.CheckTableExists(table) || !db.CheckColumnExists(table, "cert_unique_key") {
		return
	}
	// 回填有效记录唯一键
	if err := db.DB.Exec(`UPDATE user_realname_verifications
		SET cert_unique_key = UPPER(TRIM(certificate_no))
		WHERE delete_time IS NULL AND status IN (0, 1)
		  AND (cert_unique_key IS NULL OR cert_unique_key = '')
		  AND certificate_no IS NOT NULL AND TRIM(certificate_no) <> ''`).Error; err != nil {
		log.Printf("[Migrate] backfill cert_unique_key failed: %v", err)
	}
	// 拒绝/软删记录清空唯一键
	if err := db.DB.Exec(`UPDATE user_realname_verifications SET cert_unique_key = NULL
		WHERE (delete_time IS NOT NULL OR status = 2) AND cert_unique_key IS NOT NULL`).Error; err != nil {
		log.Printf("[Migrate] clear inactive cert_unique_key failed: %v", err)
	}
}

// CountOtherUsersByCertificateNoTx 在事务中统计「其他用户」使用同一证件号且状态为待审核/已通过的记录数，
// 用于提交前查重，防止同一证件号被多个账号占用实名认证。
func CountOtherUsersByCertificateNoTx(tx *gorm.DB, certificateNo string, excludeUserID uint64) (int64, error) {
	var count int64
	err := tx.Model(&RealnameVerification{}).
		Where("certificate_no = ? AND user_id != ? AND status IN (?, ?) AND delete_time IS NULL",
			certificateNo, excludeUserID, 0, 1).
		Count(&count).Error
	return count, err
}

// GetUserRealnameVerifications 获取用户的实名认证记录列表
func GetUserRealnameVerifications(userID uint64) ([]RealnameVerification, error) {
	var list []RealnameVerification
	err := db.DB.Where("user_id = ? AND delete_time IS NULL", userID).
		Order("id DESC").
		Find(&list).Error
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

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	q := db.DB.Model(&RealnameVerification{}).Where("delete_time IS NULL")
	if query.Keyword != "" {
		kw := "%" + query.Keyword + "%"
		q = q.Where("real_name LIKE ? OR certificate_no LIKE ?", kw, kw)
	}
	if query.Status != nil {
		q = q.Where("status = ?", *query.Status)
	}
	if query.UserID > 0 {
		q = q.Where("user_id = ?", query.UserID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Order("id DESC").Limit(query.PageSize).Offset(offset).Find(&list).Error; err != nil {
		return nil, err
	}

	return &RealnameVerificationListResult{
		List:     list,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}, nil
}
