package models

import (
	"database/sql"
	"errors"
	"fst/backend/pkg/db"
	"log"
	"time"

	"gorm.io/gorm"
)

// 幂等键状态
const (
	IdempotencyStatusProcessing = "processing" // 请求处理中（未成功前可在失败后释放重试）
	IdempotencyStatusCompleted  = "completed"  // 业务已成功，同 key 禁止再跑
)

// IdempotencyKey 接口幂等键
type IdempotencyKey struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	IdemKey     string `gorm:"column:idem_key;size:120;not null;default:'';uniqueIndex:uk_user_scope_key,priority:3" json:"idem_key"`
	UserID      uint64 `gorm:"column:user_id;not null;default:0;uniqueIndex:uk_user_scope_key,priority:1" json:"user_id"`
	Scope       string `gorm:"column:scope;size:80;not null;default:'';uniqueIndex:uk_user_scope_key,priority:2" json:"scope"`
	RequestHash string `gorm:"column:request_hash;size:64;not null;default:''" json:"request_hash"`
	Status      string `gorm:"column:status;size:20;not null;default:'completed'" json:"status"`
	ExpireAt    int64  `gorm:"column:expire_at;not null;default:0;index:idx_expire_at" json:"expire_at"`
	CreateTime  int64  `gorm:"column:create_time;not null;default:0" json:"create_time"`
}

// TableName 表名
func (IdempotencyKey) TableName() string { return "idempotency_keys" }

// CleanupExpiredIdempotencyKeys 清理过期幂等键
func CleanupExpiredIdempotencyKeys() (int64, error) {
	now := time.Now().Unix()
	result := db.DB.Where("expire_at > 0 AND expire_at <= ?", now).Delete(&IdempotencyKey{})
	if result.Error != nil {
		log.Printf("[Idempotency] cleanup failed: %v", result.Error)
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// DeleteExpiredIdempotencyKeyTx 事务内删除已过期幂等键
func DeleteExpiredIdempotencyKeyTx(tx *gorm.DB, idemKey string, userID uint64, scope string, now int64) error {
	return tx.Where(
		"idem_key = ? AND user_id = ? AND scope = ? AND expire_at > 0 AND expire_at <= ?",
		idemKey, userID, scope, now,
	).Delete(&IdempotencyKey{}).Error
}

// CreateIdempotencyKeyTx 占坑为 processing（业务成功后再 MarkCompleted）
func CreateIdempotencyKeyTx(tx *gorm.DB, idemKey string, userID uint64, scope string, requestHash string, expireAt int64) error {
	now := time.Now().Unix()
	item := &IdempotencyKey{
		IdemKey:     idemKey,
		UserID:      userID,
		Scope:       scope,
		RequestHash: requestHash,
		Status:      IdempotencyStatusProcessing,
		ExpireAt:    expireAt,
		CreateTime:  now,
	}
	return tx.Create(item).Error
}

// GetActiveIdempotencyKeyTx 取未过期的有效幂等键
func GetActiveIdempotencyKeyTx(tx *gorm.DB, idemKey string, userID uint64, scope string, now int64) (*IdempotencyKey, error) {
	var item IdempotencyKey
	err := tx.Where(
		"idem_key = ? AND user_id = ? AND scope = ? AND (expire_at = 0 OR expire_at > ?)",
		idemKey, userID, scope, now,
	).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	if item.Status == "" {
		item.Status = IdempotencyStatusCompleted
	}
	return &item, nil
}

// DeleteIdempotencyKey 释放幂等键（业务失败时调用，允许同 key 重试）
func DeleteIdempotencyKey(idemKey string, userID uint64, scope string) error {
	return db.DB.Where("idem_key = ? AND user_id = ? AND scope = ?", idemKey, userID, scope).
		Delete(&IdempotencyKey{}).Error
}

// MarkIdempotencyCompleted 业务成功后标记为 completed（真正锁定，禁止同 key 重放）
func MarkIdempotencyCompleted(idemKey string, userID uint64, scope string) error {
	return db.DB.Model(&IdempotencyKey{}).
		Where("idem_key = ? AND user_id = ? AND scope = ?", idemKey, userID, scope).
		Update("status", IdempotencyStatusCompleted).Error
}

// ForceUpsertIdempotencyCompleted 强制写入 completed。
// 用于「业务已成功但 MarkCompleted 失败」的兜底，避免 processing 僵死后被清理、同 key 再跑业务。
func ForceUpsertIdempotencyCompleted(idemKey string, userID uint64, scope string, requestHash string, expireAt int64) error {
	now := time.Now().Unix()
	if expireAt <= 0 {
		expireAt = now + 600
	}
	if requestHash == "" {
		requestHash = "force-completed"
	}
	r := db.DB.Model(&IdempotencyKey{}).
		Where("idem_key = ? AND user_id = ? AND scope = ?", idemKey, userID, scope).
		Updates(map[string]any{
			"status":       IdempotencyStatusCompleted,
			"request_hash": requestHash,
			"expire_at":    expireAt,
		})
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected > 0 {
		return nil
	}
	err := db.DB.Create(&IdempotencyKey{
		IdemKey:     idemKey,
		UserID:      userID,
		Scope:       scope,
		RequestHash: requestHash,
		Status:      IdempotencyStatusCompleted,
		ExpireAt:    expireAt,
		CreateTime:  now,
	}).Error
	if err == nil {
		return nil
	}
	return db.DB.Model(&IdempotencyKey{}).
		Where("idem_key = ? AND user_id = ? AND scope = ?", idemKey, userID, scope).
		Updates(map[string]any{
			"status":       IdempotencyStatusCompleted,
			"request_hash": requestHash,
			"expire_at":    expireAt,
		}).Error
}

// DeleteStaleProcessingIdempotencyKeyTx 清理卡住的 processing（进程崩溃等），允许重试
func DeleteStaleProcessingIdempotencyKeyTx(tx *gorm.DB, idemKey string, userID uint64, scope string, olderThan int64) error {
	return tx.Where(
		"idem_key = ? AND user_id = ? AND scope = ? AND status = ? AND create_time > 0 AND create_time <= ?",
		idemKey, userID, scope, IdempotencyStatusProcessing, olderThan,
	).Delete(&IdempotencyKey{}).Error
}
