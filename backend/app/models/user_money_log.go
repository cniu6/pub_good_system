package models

import (
	"database/sql"
	"fmt"
	"fst/backend/pkg/db"
	"math"
	"time"

	"gorm.io/gorm"
)

// UserMoneyLog 会员余额变动表
type UserMoneyLog struct {
	ID         uint64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     uint64  `gorm:"column:user_id;not null;default:0;index:idx_uml_user_id;index:idx_uml_user_create_time,priority:1" json:"user_id"`
	Money      float64 `gorm:"column:money;type:decimal(10,2);not null;default:0" json:"money"`
	Before     float64 `gorm:"column:before;type:decimal(10,2);not null;default:0" json:"before"`
	After      float64 `gorm:"column:after;type:decimal(10,2);not null;default:0" json:"after"`
	Memo       string  `gorm:"column:memo;size:255;not null;default:''" json:"memo"`
	CreateTime int64   `gorm:"column:create_time;not null;default:0;index:idx_uml_create_time;index:idx_uml_user_create_time,priority:2" json:"create_time"`
	DeleteTime *int64  `gorm:"column:delete_time" json:"delete_time,omitempty"`
}

func (UserMoneyLog) TableName() string {
	return "user_money_logs"
}

// CreateUserMoneyLog 创建余额变动记录
func CreateUserMoneyLog(userID uint64, money, before, after float64, memo string) (*UserMoneyLog, error) {
	now := time.Now().Unix()
	entry := &UserMoneyLog{
		UserID:     userID,
		Money:      money,
		Before:     before,
		After:      after,
		Memo:       memo,
		CreateTime: now,
	}
	if err := db.DB.Create(entry).Error; err != nil {
		return nil, err
	}
	return entry, nil
}

// GetUserMoneyLogByID 获取指定ID的余额变动记录
func GetUserMoneyLogByID(id uint64) (*UserMoneyLog, error) {
	var logEntry UserMoneyLog
	err := db.DB.Where("id = ? AND delete_time IS NULL", id).First(&logEntry).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	return &logEntry, nil
}

// GetUserMoneyLogList 获取余额变动列表（分页+搜索）
// 如果 onlyUserID > 0，则只返回该用户的记录（普通用户模式）
func GetUserMoneyLogList(onlyUserID uint64, page, pageSize int, keyword string) ([]UserMoneyLog, int64, error) {
	q := db.DB.Model(&UserMoneyLog{}).Where("delete_time IS NULL")
	if onlyUserID > 0 {
		q = q.Where("user_id = ?", onlyUserID)
	}
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("(memo LIKE ? OR "+db.CastToText("money")+" LIKE ?)", kw, kw)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var logs []UserMoneyLog
	err := q.Order("create_time DESC").Limit(pageSize).Offset(offset).Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// DeleteUserMoneyLog 软删除余额变动记录（财务审计：禁止物理删除）
// 记录不存在或已删除时返回 sql.ErrNoRows。
func DeleteUserMoneyLog(id uint64) error {
	now := time.Now().Unix()
	res := db.DB.Model(&UserMoneyLog{}).
		Where("id = ? AND delete_time IS NULL", id).
		Update("delete_time", now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateUserMoney 直接更新用户余额字段（写入前按分规范化）
func UpdateUserMoney(userID uint64, newMoney float64) error {
	normalized, err := normalizeMoneyYuan(newMoney)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	return db.DB.Model(&User{}).
		Where("id = ? AND delete_time IS NULL", userID).
		Updates(map[string]interface{}{
			"money":       normalized,
			"update_time": now,
		}).Error
}

// UpdateUserMoneyTx 在事务中更新用户余额字段（写入前按分规范化）
func UpdateUserMoneyTx(tx *gorm.DB, userID uint64, newMoney float64) error {
	normalized, err := normalizeMoneyYuan(newMoney)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	return tx.Model(&User{}).
		Where("id = ? AND delete_time IS NULL", userID).
		Updates(map[string]interface{}{
			"money":       normalized,
			"update_time": now,
		}).Error
}

// normalizeMoneyYuan 余额落库前统一规范到分精度（避免 float 脏值写入 DECIMAL）
func normalizeMoneyYuan(yuan float64) (float64, error) {
	if math.IsNaN(yuan) || math.IsInf(yuan, 0) {
		return 0, fmt.Errorf("金额非法")
	}
	fen := int64(math.Round(yuan * 100))
	return float64(fen) / 100.0, nil
}

// CreateUserMoneyLogTx 在事务中创建余额变动记录
func CreateUserMoneyLogTx(tx *gorm.DB, userID uint64, money, before, after float64, memo string) (*UserMoneyLog, error) {
	now := time.Now().Unix()
	entry := &UserMoneyLog{
		UserID:     userID,
		Money:      money,
		Before:     before,
		After:      after,
		Memo:       memo,
		CreateTime: now,
	}
	if err := tx.Create(entry).Error; err != nil {
		return nil, err
	}
	return entry, nil
}

// GetUserMoneyForUpdate 在事务中锁定并读取用户余额（SELECT ... FOR UPDATE）
func GetUserMoneyForUpdate(tx *gorm.DB, userID uint64) (float64, error) {
	var user User
	err := db.ForUpdate(tx).
		Select("money").
		Where("id = ? AND delete_time IS NULL", userID).
		First(&user).Error
	if err != nil {
		return 0, db.MapGormNotFound(err)
	}
	return user.Money, nil
}
