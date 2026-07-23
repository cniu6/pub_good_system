package models

import (
	"database/sql"
	"fst/backend/pkg/db"
	"time"

	"gorm.io/gorm"
)

// UserScoreLog 会员积分变动表
type UserScoreLog struct {
	ID         uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     uint64 `gorm:"column:user_id;not null;default:0;index:idx_usl_user_id;index:idx_usl_user_create_time,priority:1" json:"user_id"`
	Score      int64  `gorm:"column:score;not null;default:0" json:"score"`
	Before     int64  `gorm:"column:before;not null;default:0" json:"before"`
	After      int64  `gorm:"column:after;not null;default:0" json:"after"`
	Memo       string `gorm:"column:memo;size:255;not null;default:''" json:"memo"`
	CreateTime int64  `gorm:"column:create_time;not null;default:0;index:idx_usl_create_time;index:idx_usl_user_create_time,priority:2" json:"create_time"`
	DeleteTime *int64 `gorm:"column:delete_time" json:"delete_time,omitempty"`
}

func (UserScoreLog) TableName() string {
	return "user_score_logs"
}

// CreateUserScoreLog 创建积分变动记录
func CreateUserScoreLog(userID uint64, score, before, after int64, memo string) (*UserScoreLog, error) {
	now := time.Now().Unix()
	entry := &UserScoreLog{
		UserID:     userID,
		Score:      score,
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

// GetUserScoreLogByID 获取指定ID的积分变动记录
func GetUserScoreLogByID(id uint64) (*UserScoreLog, error) {
	var logEntry UserScoreLog
	err := db.DB.Where("id = ? AND delete_time IS NULL", id).First(&logEntry).Error
	if err != nil {
		return nil, db.MapGormNotFound(err)
	}
	return &logEntry, nil
}

// GetUserScoreLogList 获取积分变动列表（分页+搜索）
func GetUserScoreLogList(onlyUserID uint64, page, pageSize int, keyword string) ([]UserScoreLog, int64, error) {
	q := db.DB.Model(&UserScoreLog{}).Where("delete_time IS NULL")
	if onlyUserID > 0 {
		q = q.Where("user_id = ?", onlyUserID)
	}
	if keyword != "" {
		kw := "%" + keyword + "%"
		q = q.Where("(memo LIKE ? OR "+db.CastToText("score")+" LIKE ?)", kw, kw)
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

	var logs []UserScoreLog
	err := q.Order("create_time DESC").Limit(pageSize).Offset(offset).Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// DeleteUserScoreLog 软删除积分变动记录（财务审计：禁止物理删除）
// 记录不存在或已删除时返回 sql.ErrNoRows。
func DeleteUserScoreLog(id uint64) error {
	now := time.Now().Unix()
	res := db.DB.Model(&UserScoreLog{}).
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

// UpdateUserScore 直接更新用户积分字段
func UpdateUserScore(userID uint64, newScore int64) error {
	now := time.Now().Unix()
	return db.DB.Model(&User{}).
		Where("id = ? AND delete_time IS NULL", userID).
		Updates(map[string]interface{}{
			"score":       newScore,
			"update_time": now,
		}).Error
}

// UpdateUserScoreTx 在事务中更新用户积分字段
func UpdateUserScoreTx(tx *gorm.DB, userID uint64, newScore int64) error {
	now := time.Now().Unix()
	return tx.Model(&User{}).
		Where("id = ? AND delete_time IS NULL", userID).
		Updates(map[string]interface{}{
			"score":       newScore,
			"update_time": now,
		}).Error
}

// CreateUserScoreLogTx 在事务中创建积分变动记录
func CreateUserScoreLogTx(tx *gorm.DB, userID uint64, score, before, after int64, memo string) (*UserScoreLog, error) {
	now := time.Now().Unix()
	entry := &UserScoreLog{
		UserID:     userID,
		Score:      score,
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

// GetUserScoreForUpdate 在事务中锁定并读取用户积分（SELECT ... FOR UPDATE）
func GetUserScoreForUpdate(tx *gorm.DB, userID uint64) (int64, error) {
	var user User
	err := db.ForUpdate(tx).
		Select("score").
		Where("id = ? AND delete_time IS NULL", userID).
		First(&user).Error
	if err != nil {
		return 0, db.MapGormNotFound(err)
	}
	return user.Score, nil
}
