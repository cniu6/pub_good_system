package models

import (
	"fst/backend/internal/db"
	"log"
	"time"
)

// InitVerificationCodeTable 初始化验证码表（如果不存在）
func InitVerificationCodeTable() {
	if !db.CheckTableExists("verification_codes") {
		// 表不存在，创建新表
		schema := `CREATE TABLE IF NOT EXISTS verification_codes (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			email VARCHAR(255) NOT NULL COMMENT '邮箱地址',
			code VARCHAR(10) NOT NULL COMMENT '验证码',
			code_type VARCHAR(20) NOT NULL COMMENT '类型:register=注册,reset_password=重置密码',
			expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
			is_used TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已使用:0=未使用,1=已使用',
			is_deleted TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否软删除:0=正常,1=已删除',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
			INDEX idx_email_type (email, code_type),
			INDEX idx_expires_at (expires_at),
			INDEX idx_is_deleted (is_deleted)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

		_, err := db.DB.Exec(schema)
		if err != nil {
			log.Printf("[Init] Failed to create verification_codes table: %v", err)
		} else {
			log.Println("[Init] Created verification_codes table")
		}
		return
	}

	// 表已存在，检查并修复字段
	repairVerificationCodeTable()
}

// repairVerificationCodeTable 检查并修复表字段
func repairVerificationCodeTable() {
	// 定义需要删除的错误字段名（之前版本创建的错误字段）
	wrongColumns := []string{"type", "expire_at"}
	for _, col := range wrongColumns {
		if db.CheckColumnExists("verification_codes", col) {
			_, err := db.DB.Exec("ALTER TABLE verification_codes DROP COLUMN " + col)
			if err != nil {
				log.Printf("[Init] Failed to drop old '%s' column: %v", col, err)
			} else {
				log.Printf("[Init] Dropped old '%s' column", col)
			}
		}
	}

	// 定义需要的正确字段
	requiredColumns := map[string]string{
		"code_type":  "ALTER TABLE verification_codes ADD COLUMN code_type VARCHAR(20) NOT NULL DEFAULT 'register' COMMENT '类型:register=注册,reset_password=重置密码'",
		"expires_at": "ALTER TABLE verification_codes ADD COLUMN expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '过期时间'",
		"is_used":    "ALTER TABLE verification_codes ADD COLUMN is_used TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已使用:0=未使用,1=已使用'",
		"is_deleted": "ALTER TABLE verification_codes ADD COLUMN is_deleted TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否软删除:0=正常,1=已删除'",
		"created_at": "ALTER TABLE verification_codes ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'",
		"updated_at": "ALTER TABLE verification_codes ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'",
	}

	for col, alterSQL := range requiredColumns {
		if !db.CheckColumnExists("verification_codes", col) {
			_, err := db.DB.Exec(alterSQL)
			if err != nil {
				log.Printf("[Init] Failed to add column %s: %v", col, err)
			} else {
				log.Printf("[Init] Added column %s to verification_codes table", col)
			}
		}
	}
}

// VerificationCode 验证码模型
type VerificationCode struct {
	ID        uint64    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Code      string    `db:"code" json:"code"`
	CodeType  string    `db:"code_type" json:"code_type"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	IsUsed    int       `db:"is_used" json:"is_used"`
	IsDeleted int       `db:"is_deleted" json:"is_deleted"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// CreateVerificationCode 创建验证码记录
func CreateVerificationCode(email, code, codeType string, expiresAt time.Time) error {
	// 先将该邮箱该类型的旧验证码标记为软删除
	_, err := db.DB.Exec(
		"UPDATE verification_codes SET is_deleted = 1 WHERE email = ? AND code_type = ? AND is_deleted = 0 AND is_used = 0",
		email, codeType,
	)
	if err != nil {
		return err
	}

	query := `INSERT INTO verification_codes (email, code, code_type, expires_at, is_used, is_deleted) 
			  VALUES (?, ?, ?, ?, 0, 0)`
	_, err = db.DB.Exec(query, email, code, codeType, expiresAt)
	return err
}

// GetValidVerificationCode 获取有效的验证码（未使用、未过期、未软删除）
func GetValidVerificationCode(email, codeType string) (*VerificationCode, error) {
	var vc VerificationCode
	query := `SELECT * FROM verification_codes 
			  WHERE email = ? AND code_type = ? AND is_used = 0 AND is_deleted = 0 AND expires_at > NOW()
			  ORDER BY created_at DESC LIMIT 1`
	err := db.DB.Get(&vc, query, email, codeType)
	if err != nil {
		return nil, err
	}
	return &vc, nil
}

// MarkVerificationCodeAsUsed 标记验证码为已使用
func MarkVerificationCodeAsUsed(id uint64) error {
	_, err := db.DB.Exec("UPDATE verification_codes SET is_used = 1 WHERE id = ?", id)
	return err
}

// MarkVerificationCodeAsDeleted 软删除验证码
func MarkVerificationCodeAsDeleted(id uint64) error {
	_, err := db.DB.Exec("UPDATE verification_codes SET is_deleted = 1 WHERE id = ?", id)
	return err
}

// DeleteVerificationCodesByEmail 彻底删除指定邮箱的所有验证码记录（用于注册/重置成功后清理）
func DeleteVerificationCodesByEmail(email string, codeType string) error {
	query := `DELETE FROM verification_codes WHERE email = ?`
	args := []interface{}{email}
	if codeType != "" {
		query += ` AND code_type = ?`
		args = append(args, codeType)
	}
	_, err := db.DB.Exec(query, args...)
	return err
}

// SoftDeleteExpiredCodes 软删除已过期的验证码
func SoftDeleteExpiredCodes() error {
	_, err := db.DB.Exec("UPDATE verification_codes SET is_deleted = 1 WHERE expires_at <= NOW() AND is_deleted = 0")
	return err
}

// CleanupOldVerificationCodes 清理7天前的已删除或已使用记录（硬删除）
func CleanupOldVerificationCodes() error {
	_, err := db.DB.Exec(
		"DELETE FROM verification_codes WHERE (is_deleted = 1 OR is_used = 1) AND updated_at < DATE_SUB(NOW(), INTERVAL 7 DAY)",
	)
	return err
}

// VerifyCode 验证验证码是否正确（改进版：直接匹配代码）
func VerifyCode(email, code, codeType string) (bool, uint64, error) {
	var vc VerificationCode
	query := `SELECT id, code, expires_at FROM verification_codes 
			  WHERE email = ? AND code = ? AND code_type = ? AND is_used = 0 AND is_deleted = 0 
			  ORDER BY created_at DESC LIMIT 1`
	err := db.DB.Get(&vc, query, email, code, codeType)
	if err != nil {
		return false, 0, err
	}

	// 检查是否过期
	if vc.ExpiresAt.Before(time.Now()) {
		return false, 0, nil
	}

	return true, vc.ID, nil
}
