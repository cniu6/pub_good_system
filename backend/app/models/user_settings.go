package models

import (
	"database/sql"
	"errors"
	"fst/backend/pkg/db"
	"time"

	"gorm.io/gorm"
)

// UserSettings 用户设置模型
type UserSettings struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID      uint64 `gorm:"column:user_id;uniqueIndex" json:"user_id"`
	Theme       string `gorm:"column:theme" json:"theme"`
	NotifyEmail bool   `gorm:"column:notify_email" json:"notify_email"`
	CreatedAt   int64  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at" json:"updated_at"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

// GetUserSettings 获取用户设置
func GetUserSettings(userID uint64) (*UserSettings, error) {
	var settings UserSettings
	err := db.DB.Where("user_id = ?", userID).First(&settings).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// SaveUserSettings 保存用户设置（upsert）
func SaveUserSettings(settings *UserSettings) error {
	now := time.Now().Unix()
	settings.UpdatedAt = now

	result := db.DB.Model(&UserSettings{}).Where("user_id = ?", settings.UserID).Updates(map[string]interface{}{
		"theme":        settings.Theme,
		"notify_email": settings.NotifyEmail,
		"updated_at":   now,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		settings.CreatedAt = now
		return db.DB.Create(settings).Error
	}
	return nil
}
