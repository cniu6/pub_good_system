package models

import (
	"database/sql"
	"errors"
	"fst/backend/pkg/db"
	"log"
	"time"
)

// UserLevelCap 用户等级能力上限（按 users.level 匹配）
type UserLevelCap struct {
	Level         uint64 `gorm:"column:level;primaryKey" json:"level"`
	Name          string `gorm:"column:name;size:100;not null;default:''" json:"name"`
	AllowAPIKey   bool   `gorm:"column:allow_api_key;not null;default:true" json:"allow_api_key"`
	AllowRecharge bool   `gorm:"column:allow_recharge;not null;default:true" json:"allow_recharge"`
	AllowWithdraw bool   `gorm:"column:allow_withdraw;not null;default:true" json:"allow_withdraw"`
	MenuFlags     string `gorm:"column:menu_flags;type:text" json:"menu_flags"` // JSON 字符串，预留给菜单开关
	CreateTime    int64  `gorm:"column:create_time;not null;default:0" json:"create_time"`
}

// TableName 表名
func (UserLevelCap) TableName() string { return "user_level_caps" }

// SeedUserLevelCaps 播种默认等级（1=普通，2=VIP）
func SeedUserLevelCaps() {
	now := time.Now().Unix()
	defaults := []UserLevelCap{
		{Level: 1, Name: "默认", AllowAPIKey: true, AllowRecharge: true, AllowWithdraw: true, MenuFlags: "{}", CreateTime: now},
		{Level: 2, Name: "VIP", AllowAPIKey: true, AllowRecharge: true, AllowWithdraw: true, MenuFlags: "{}", CreateTime: now},
	}
	for _, d := range defaults {
		var existing UserLevelCap
		err := db.FindOne(db.DB.Where("level = ?", d.Level), &existing)
		if errors.Is(err, sql.ErrNoRows) {
			if err := db.DB.Create(&d).Error; err != nil {
				log.Printf("[Init] 播种用户等级 %d 失败: %v", d.Level, err)
			}
		} else if err != nil {
			log.Printf("[Init] 查询用户等级 %d 失败: %v", d.Level, err)
		}
	}
}

// ListUserLevelCaps 列出全部等级能力
func ListUserLevelCaps() ([]UserLevelCap, error) {
	var list []UserLevelCap
	err := db.DB.Order("level ASC").Find(&list).Error
	return list, err
}

// GetUserLevelCap 按等级取能力；不存在时返回宽松默认（兼容未配置等级）
func GetUserLevelCap(level uint64) (*UserLevelCap, error) {
	if level == 0 {
		level = 1
	}
	var cap UserLevelCap
	err := db.FindOne(db.DB.Where("level = ?", level), &cap)
	if errors.Is(err, sql.ErrNoRows) {
		// 未配置：默认全部允许，避免误伤存量用户
		return &UserLevelCap{
			Level: level, Name: "默认", AllowAPIKey: true, AllowRecharge: true, AllowWithdraw: true, MenuFlags: "{}",
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &cap, nil
}

// UpdateUserLevelCap 更新等级能力（按 level upsert 字段）
func UpdateUserLevelCap(level uint64, name string, allowAPIKey, allowRecharge, allowWithdraw bool, menuFlags string) error {
	if level == 0 {
		return errors.New("Invalid level")
	}
	now := time.Now().Unix()
	var existing UserLevelCap
	err := db.FindOne(db.DB.Where("level = ?", level), &existing)
	if errors.Is(err, sql.ErrNoRows) {
		return db.DB.Create(&UserLevelCap{
			Level: level, Name: name,
			AllowAPIKey: allowAPIKey, AllowRecharge: allowRecharge, AllowWithdraw: allowWithdraw,
			MenuFlags: menuFlags, CreateTime: now,
		}).Error
	}
	if err != nil {
		return err
	}
	return db.DB.Model(&existing).Updates(map[string]interface{}{
		"name":           name,
		"allow_api_key":  allowAPIKey,
		"allow_recharge": allowRecharge,
		"allow_withdraw": allowWithdraw,
		"menu_flags":     menuFlags,
	}).Error
}

// CheckUserLevelAllows 检查用户当前等级是否允许某能力；返回 (ok, 拒绝文案)
func CheckUserLevelAllows(userID uint64, feature string) (bool, string) {
	user, err := GetUserByID(userID)
	if err != nil || user == nil {
		return false, "用户不存在"
	}
	// 系统管理员不受等级限制
	if user.Role == "admin" {
		return true, ""
	}
	cap, err := GetUserLevelCap(user.Level)
	if err != nil {
		return false, "等级配置读取失败"
	}
	switch feature {
	case "api_key":
		if !cap.AllowAPIKey {
			return false, "当前用户等级不允许使用 API Key"
		}
	case "recharge":
		if !cap.AllowRecharge {
			return false, "当前用户等级不允许充值"
		}
	case "withdraw":
		if !cap.AllowWithdraw {
			return false, "当前用户等级不允许提现"
		}
	default:
		return true, ""
	}
	return true, ""
}
