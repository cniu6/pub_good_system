package models

import (
	"database/sql"
	"errors"
	"fst/backend/pkg/db"
	"log"
	"time"
)

// RepairVerificationCodeTable 存量库修复：email→contact 数据迁移、删除错误旧列，
// 并对缺失的常规列做轻量 ADD（避免对残缺旧表跑 AutoMigrate 触发 SQLite 重建失败）。
func RepairVerificationCodeTable() {
	if !db.CheckTableExists("verification_codes") {
		return
	}
	migrateVerificationCodesEmailToContact()
	dropLegacyVerificationCodeColumns()
	ensureVerificationCodeColumns()
}

func ensureVerificationCodeColumns() {
	// 只补 AutoMigrate 正常路径会有、但存量残缺表可能缺的列
	alters := []struct {
		col string
		sql string
	}{
		{"attempts", "ALTER TABLE verification_codes ADD COLUMN attempts INTEGER NOT NULL DEFAULT 0"},
		{"is_used", "ALTER TABLE verification_codes ADD COLUMN is_used INTEGER NOT NULL DEFAULT 0"},
		{"is_deleted", "ALTER TABLE verification_codes ADD COLUMN is_deleted INTEGER NOT NULL DEFAULT 0"},
		{"code_type", "ALTER TABLE verification_codes ADD COLUMN code_type VARCHAR(20) NOT NULL DEFAULT 'register'"},
		{"expires_at", "ALTER TABLE verification_codes ADD COLUMN expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP"},
		{"created_at", "ALTER TABLE verification_codes ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP"},
		{"updated_at", "ALTER TABLE verification_codes ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP"},
	}
	for _, a := range alters {
		if db.CheckColumnExists("verification_codes", a.col) {
			continue
		}
		if err := db.DB.Exec(a.sql).Error; err != nil {
			log.Printf("[Migrate] 补 verification_codes.%s 失败: %v", a.col, err)
		} else {
			log.Printf("[Migrate] 已补 verification_codes.%s", a.col)
		}
	}
}

// migrateVerificationCodesEmailToContact 把旧列 email 的数据拷到 contact。
// AutoMigrate 可能已先加好空 contact 列，因此「两列都在」时仍要做空 contact 回填。
func migrateVerificationCodesEmailToContact() {
	hasEmail := db.CheckColumnExists("verification_codes", "email")
	if !hasEmail {
		return
	}
	hasContact := db.CheckColumnExists("verification_codes", "contact")

	if !hasContact {
		if db.IsMySQL() {
			if err := db.DB.Exec("ALTER TABLE verification_codes CHANGE COLUMN email contact VARCHAR(255) NOT NULL COMMENT '联系方式(邮箱或手机号)'").Error; err != nil {
				log.Printf("[Migrate] verification_codes.email→contact 改名失败: %v", err)
				return
			}
			log.Println("[Migrate] 已将 verification_codes.email 改名为 contact")
			return
		}
		if err := db.DB.Exec("ALTER TABLE verification_codes ADD COLUMN contact VARCHAR(255) NOT NULL DEFAULT ''").Error; err != nil {
			log.Printf("[Migrate] 补 verification_codes.contact 失败: %v", err)
			return
		}
		hasContact = true
	}

	if !hasContact {
		return
	}
	res := db.DB.Exec("UPDATE verification_codes SET contact = email WHERE (contact = '' OR contact IS NULL) AND email IS NOT NULL AND email <> ''")
	if res.Error != nil {
		log.Printf("[Migrate] 拷贝 email→contact 失败: %v", res.Error)
		return
	}
	if res.RowsAffected > 0 {
		log.Printf("[Migrate] 已从 email 回填 contact，影响 %d 行", res.RowsAffected)
	}
}

func dropLegacyVerificationCodeColumns() {
	for _, col := range []string{"type", "expire_at"} {
		if !db.CheckColumnExists("verification_codes", col) {
			continue
		}
		// SQLite 对部分保留字/探测可能误报，删除失败只打日志不阻断
		if err := db.DB.Exec("ALTER TABLE verification_codes DROP COLUMN " + db.QuoteIdent(col)).Error; err != nil {
			log.Printf("[Migrate] 删除旧列 verification_codes.%s 跳过: %v", col, err)
		} else {
			log.Printf("[Migrate] 已删除旧列 verification_codes.%s", col)
		}
	}
}

// VerificationCode 验证码模型
type VerificationCode struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Contact   string    `gorm:"column:contact;size:255;not null;default:'';index:idx_vc_contact_type_active_created,priority:1;index:idx_vc_contact_code_type_active,priority:1" json:"contact"`
	Code      string    `gorm:"column:code;size:32;not null;default:'';index:idx_vc_contact_code_type_active,priority:2" json:"code"`
	CodeType  string    `gorm:"column:code_type;size:20;not null;default:'register';index:idx_vc_contact_type_active_created,priority:2;index:idx_vc_contact_code_type_active,priority:3" json:"code_type"`
	ExpiresAt time.Time `gorm:"column:expires_at;index:idx_vc_expires_at" json:"expires_at"`
	IsUsed    int       `gorm:"column:is_used;not null;default:0;index:idx_vc_contact_type_active_created,priority:3;index:idx_vc_contact_code_type_active,priority:4" json:"is_used"`
	IsDeleted int       `gorm:"column:is_deleted;not null;default:0;index:idx_vc_contact_type_active_created,priority:4;index:idx_vc_contact_code_type_active,priority:5" json:"is_deleted"`
	Attempts  int       `gorm:"column:attempts;not null;default:0" json:"attempts"`
	CreatedAt time.Time `gorm:"column:created_at;index:idx_vc_contact_type_active_created,priority:5" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (VerificationCode) TableName() string {
	return "verification_codes"
}

// CreateVerificationCode 创建验证码记录。contact 可以是邮箱或手机号。
func CreateVerificationCode(contact, code, codeType string, expiresAt time.Time) error {
	if err := db.DB.Model(&VerificationCode{}).
		Where("contact = ? AND code_type = ? AND is_deleted = 0 AND is_used = 0", contact, codeType).
		Update("is_deleted", 1).Error; err != nil {
		return err
	}

	vc := VerificationCode{
		Contact:   contact,
		Code:      code,
		CodeType:  codeType,
		ExpiresAt: expiresAt,
		IsUsed:    0,
		IsDeleted: 0,
	}
	// created_at/updated_at 交给数据库默认值，避免 GORM 零值覆盖
	return db.DB.Omit("CreatedAt", "UpdatedAt").Create(&vc).Error
}

// HasRecentVerificationCode 是否在 since 之后发过同类型未用验证码（限流）
func HasRecentVerificationCode(contact, codeType string, since time.Time) (bool, error) {
	var count int64
	err := db.DB.Model(&VerificationCode{}).
		Where("contact = ? AND code_type = ? AND is_used = 0 AND is_deleted = 0 AND created_at >= ?", contact, codeType, since).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetValidVerificationCode 获取有效验证码（未使用、未过期、未软删）。
// 查无是常态，用 FindOne 避免 First 刷 record not found。
func GetValidVerificationCode(contact, codeType string) (*VerificationCode, error) {
	var vc VerificationCode
	err := db.FindOne(db.DB.Where(
		"contact = ? AND code_type = ? AND is_used = 0 AND is_deleted = 0 AND expires_at > ?",
		contact, codeType, time.Now(),
	).Order("created_at DESC"), &vc)
	if err != nil {
		return nil, err
	}
	return &vc, nil
}

// MarkVerificationCodeAsUsed 标记验证码为已使用
func MarkVerificationCodeAsUsed(id uint64) error {
	return db.DB.Model(&VerificationCode{}).Where("id = ?", id).Update("is_used", 1).Error
}

// maxVerificationAttempts 单个验证码允许的最大验证失败次数
const maxVerificationAttempts = 5

// ConsumeVerificationCode 校验并消费验证码。成功返回 (true, nil)；错误码返回 (false, nil) 并累加 attempts。
func ConsumeVerificationCode(contact, code, codeType string) (bool, error) {
	now := time.Now()
	result := db.DB.Exec(
		`UPDATE verification_codes
		 SET is_used = 1
		 WHERE id = (
		 	SELECT id FROM (
		 		SELECT id FROM verification_codes
		 		WHERE contact = ? AND code = ? AND code_type = ? AND is_used = 0 AND is_deleted = 0 AND expires_at > ?
		 		ORDER BY created_at DESC LIMIT 1
		 	) AS latest
		 ) AND is_used = 0`,
		contact, code, codeType, now,
	)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return true, nil
	}

	if err := db.DB.Exec(
		`UPDATE verification_codes
		 SET is_deleted = CASE WHEN attempts + 1 >= ? THEN 1 ELSE is_deleted END,
		     attempts = attempts + 1
		 WHERE id = (
		 	SELECT id FROM (
		 		SELECT id FROM verification_codes
		 		WHERE contact = ? AND code_type = ? AND is_used = 0 AND is_deleted = 0 AND expires_at > ?
		 		ORDER BY created_at DESC LIMIT 1
		 	) AS latest
		 )`,
		maxVerificationAttempts, contact, codeType, now,
	).Error; err != nil {
		return false, err
	}

	return false, nil
}

// MarkVerificationCodeAsDeleted 软删除验证码
func MarkVerificationCodeAsDeleted(id uint64) error {
	return db.DB.Model(&VerificationCode{}).Where("id = ?", id).Update("is_deleted", 1).Error
}

// DeleteVerificationCodesByContact 彻底删除指定联系方式的验证码（注册/重置成功后清理）
func DeleteVerificationCodesByContact(contact string, codeType string) error {
	q := db.DB.Where("contact = ?", contact)
	if codeType != "" {
		q = q.Where("code_type = ?", codeType)
	}
	return q.Delete(&VerificationCode{}).Error
}

// SoftDeleteExpiredCodes 软删除已过期的验证码
func SoftDeleteExpiredCodes() (int64, error) {
	result := db.DB.Model(&VerificationCode{}).
		Where("expires_at <= ? AND is_deleted = 0", time.Now()).
		Update("is_deleted", 1)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// CleanupOldVerificationCodes 清理 7 天前的已删/已用记录（硬删除）
func CleanupOldVerificationCodes() (int64, error) {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	result := db.DB.Where("(is_deleted = 1 OR is_used = 1) AND updated_at < ?", cutoff).
		Delete(&VerificationCode{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// VerifyCode 验证验证码是否正确（直接匹配）。查无是常态，用 FindOne。
func VerifyCode(contact, code, codeType string) (bool, uint64, error) {
	var vc VerificationCode
	err := db.FindOne(db.DB.Select("id", "code", "expires_at").
		Where("contact = ? AND code = ? AND code_type = ? AND is_used = 0 AND is_deleted = 0", contact, code, codeType).
		Order("created_at DESC"), &vc)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, sql.ErrNoRows
	}
	if err != nil {
		return false, 0, err
	}
	if vc.ExpiresAt.Before(time.Now()) {
		return false, 0, nil
	}
	return true, vc.ID, nil
}
