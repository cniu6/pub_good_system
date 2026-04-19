package models

import (
	"database/sql"
	"fst/backend/pkg/db"
	"log"
	"time"
)

type IdempotencyKey struct {
	ID          uint64 `db:"id"`
	IdemKey     string `db:"idem_key"`
	UserID      uint64 `db:"user_id"`
	Scope       string `db:"scope"`
	RequestHash string `db:"request_hash"`
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
		expire_at BIGINT NOT NULL DEFAULT 0 COMMENT '过期时间',
		create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
		UNIQUE KEY uk_user_scope_key (user_id, scope, idem_key),
		INDEX idx_expire_at (expire_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='接口幂等键表';`

	if _, err := db.DB.Exec(schema); err != nil {
		log.Printf("[Init] Failed to create idempotency_keys table: %v", err)
	}
}

func CleanupExpiredIdempotencyKeys() (int64, error) {
	now := time.Now().Unix()
	result, err := db.DB.Exec("DELETE FROM idempotency_keys WHERE expire_at > 0 AND expire_at <= ?", now)
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

func CreateIdempotencyKeyTx(tx *sql.Tx, idemKey string, userID uint64, scope string, requestHash string, expireAt int64) error {
	now := time.Now().Unix()
	_, err := tx.Exec(
		"INSERT INTO idempotency_keys (idem_key, user_id, scope, request_hash, expire_at, create_time) VALUES (?, ?, ?, ?, ?, ?)",
		idemKey, userID, scope, requestHash, expireAt, now,
	)
	return err
}

func GetActiveIdempotencyKeyTx(tx *sql.Tx, idemKey string, userID uint64, scope string, now int64) (*IdempotencyKey, error) {
	var item IdempotencyKey
	err := tx.QueryRow(
		"SELECT id, idem_key, user_id, scope, request_hash, expire_at, create_time FROM idempotency_keys WHERE idem_key = ? AND user_id = ? AND scope = ? AND (expire_at = 0 OR expire_at > ?) LIMIT 1",
		idemKey, userID, scope, now,
	).Scan(&item.ID, &item.IdemKey, &item.UserID, &item.Scope, &item.RequestHash, &item.ExpireAt, &item.CreateTime)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

