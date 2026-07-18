package models

import (
	"database/sql"
	"fst/backend/pkg/db"
	"log"
	"time"
)

// 幂等键状态
const (
	IdempotencyStatusProcessing = "processing" // 请求处理中（未成功前可在失败后释放重试）
	IdempotencyStatusCompleted  = "completed"  // 业务已成功，同 key 禁止再跑
)

type IdempotencyKey struct {
	ID          uint64 `db:"id"`
	IdemKey     string `db:"idem_key"`
	UserID      uint64 `db:"user_id"`
	Scope       string `db:"scope"`
	RequestHash string `db:"request_hash"`
	Status      string `db:"status"`
	ExpireAt    int64  `db:"expire_at"`
	CreateTime  int64  `db:"create_time"`
}

func InitIdempotencyKeysTable() {
	schema := `CREATE TABLE IF NOT EXISTS idempotency_keys (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		idem_key VARCHAR(120) NOT NULL DEFAULT '' COMMENT '幂等键',
		user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
		scope VARCHAR(80) NOT NULL DEFAULT '' COMMENT '作用域',
		request_hash CHAR(64) NOT NULL DEFAULT '' COMMENT '请求摘要',
		status VARCHAR(20) NOT NULL DEFAULT 'completed' COMMENT 'processing=处理中 completed=已成功',
		expire_at BIGINT NOT NULL DEFAULT 0 COMMENT '过期时间',
		create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
		UNIQUE KEY uk_user_scope_key (user_id, scope, idem_key),
		INDEX idx_expire_at (expire_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='接口幂等键表'`

	if _, err := db.Exec(schema); err != nil {
		log.Printf("[Init] Failed to create idempotency_keys table: %v", err)
	}

	// 兼容旧表：补 status 列。已有行视为 completed（旧逻辑等于「占坑即锁定」）
	if !db.CheckColumnExists("idempotency_keys", "status") {
		if _, err := db.Exec(
			"ALTER TABLE idempotency_keys ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'completed' COMMENT 'processing=处理中 completed=已成功' AFTER request_hash",
		); err != nil {
			log.Printf("[Init] Failed to add idempotency_keys.status: %v", err)
		} else {
			log.Println("[Init] Added idempotency_keys.status")
		}
	}
}

func CleanupExpiredIdempotencyKeys() (int64, error) {
	now := time.Now().Unix()
	result, err := db.Exec("DELETE FROM idempotency_keys WHERE expire_at > 0 AND expire_at <= ?", now)
	if err != nil {
		log.Printf("[Idempotency] cleanup failed: %v", err)
		return 0, err
	}
	return result.RowsAffected()
}

func DeleteExpiredIdempotencyKeyTx(tx *sql.Tx, idemKey string, userID uint64, scope string, now int64) error {
	_, err := tx.Exec(
		"DELETE FROM idempotency_keys WHERE idem_key = ? AND user_id = ? AND scope = ? AND expire_at > 0 AND expire_at <= ?",
		idemKey, userID, scope, now,
	)
	return err
}

// CreateIdempotencyKeyTx 占坑为 processing（业务成功后再 MarkCompleted）
func CreateIdempotencyKeyTx(tx *sql.Tx, idemKey string, userID uint64, scope string, requestHash string, expireAt int64) error {
	now := time.Now().Unix()
	_, err := tx.Exec(
		"INSERT INTO idempotency_keys (idem_key, user_id, scope, request_hash, status, expire_at, create_time) VALUES (?, ?, ?, ?, ?, ?, ?)",
		idemKey, userID, scope, requestHash, IdempotencyStatusProcessing, expireAt, now,
	)
	return err
}

func GetActiveIdempotencyKeyTx(tx *sql.Tx, idemKey string, userID uint64, scope string, now int64) (*IdempotencyKey, error) {
	var item IdempotencyKey
	err := tx.QueryRow(
		`SELECT id, idem_key, user_id, scope, request_hash, COALESCE(status, ?) AS status, expire_at, create_time
		 FROM idempotency_keys
		 WHERE idem_key = ? AND user_id = ? AND scope = ? AND (expire_at = 0 OR expire_at > ?)
		 LIMIT 1`,
		IdempotencyStatusCompleted, idemKey, userID, scope, now,
	).Scan(&item.ID, &item.IdemKey, &item.UserID, &item.Scope, &item.RequestHash, &item.Status, &item.ExpireAt, &item.CreateTime)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// DeleteIdempotencyKey 释放幂等键（业务失败时调用，允许同 key 重试）
func DeleteIdempotencyKey(idemKey string, userID uint64, scope string) error {
	_, err := db.Exec(
		"DELETE FROM idempotency_keys WHERE idem_key = ? AND user_id = ? AND scope = ?",
		idemKey, userID, scope,
	)
	return err
}

// MarkIdempotencyCompleted 业务成功后标记为 completed（真正锁定，禁止同 key 重放）
func MarkIdempotencyCompleted(idemKey string, userID uint64, scope string) error {
	_, err := db.Exec(
		"UPDATE idempotency_keys SET status = ? WHERE idem_key = ? AND user_id = ? AND scope = ?",
		IdempotencyStatusCompleted, idemKey, userID, scope,
	)
	return err
}

// DeleteStaleProcessingIdempotencyKeyTx 清理卡住的 processing（进程崩溃等），允许重试
func DeleteStaleProcessingIdempotencyKeyTx(tx *sql.Tx, idemKey string, userID uint64, scope string, olderThan int64) error {
	_, err := tx.Exec(
		"DELETE FROM idempotency_keys WHERE idem_key = ? AND user_id = ? AND scope = ? AND status = ? AND create_time > 0 AND create_time <= ?",
		idemKey, userID, scope, IdempotencyStatusProcessing, olderThan,
	)
	return err
}
