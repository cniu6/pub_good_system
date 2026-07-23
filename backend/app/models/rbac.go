package models

import (
	"database/sql"
	"errors"
	"fmt"
	"fst/backend/pkg/db"
	"log"
	"time"

	"gorm.io/gorm"
)

// Role RBAC 角色（单组织 MVP：admin / operator / viewer）
type Role struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code        string `gorm:"column:code;size:50;not null;uniqueIndex:uk_roles_code" json:"code"`
	Name        string `gorm:"column:name;size:100;not null;default:''" json:"name"`
	Description string `gorm:"column:description;size:255;not null;default:''" json:"description"`
	CreateTime  int64  `gorm:"column:create_time;not null;default:0" json:"create_time"`
}

// TableName 表名
func (Role) TableName() string { return "roles" }

// Permission RBAC 权限点
type Permission struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code        string `gorm:"column:code;size:80;not null;uniqueIndex:uk_permissions_code" json:"code"`
	Name        string `gorm:"column:name;size:100;not null;default:''" json:"name"`
	Description string `gorm:"column:description;size:255;not null;default:''" json:"description"`
	CreateTime  int64  `gorm:"column:create_time;not null;default:0" json:"create_time"`
}

// TableName 表名
func (Permission) TableName() string { return "permissions" }

// RolePermission 角色-权限关联
type RolePermission struct {
	RoleID       uint64 `gorm:"column:role_id;primaryKey" json:"role_id"`
	PermissionID uint64 `gorm:"column:permission_id;primaryKey;index:idx_rp_permission" json:"permission_id"`
}

// TableName 表名
func (RolePermission) TableName() string { return "role_permissions" }

// UserRole 用户-角色关联
type UserRole struct {
	UserID     uint64 `gorm:"column:user_id;primaryKey" json:"user_id"`
	RoleID     uint64 `gorm:"column:role_id;primaryKey;index:idx_ur_role" json:"role_id"`
	CreateTime int64  `gorm:"column:create_time;not null;default:0" json:"create_time"`
}

// TableName 表名
func (UserRole) TableName() string { return "user_roles" }

// SeedRBACDefaults 播种默认角色/权限（建表由 GORM AutoMigrate 负责）
func SeedRBACDefaults() {
	now := time.Now().Unix()

	perms := []struct{ Code, Name, Desc string }{
		{"user:read", "用户只读", "查看用户列表/详情"},
		{"user:write", "用户写入", "创建/编辑/禁用用户"},
		{"finance:read", "财务只读", "查看余额/积分/提现"},
		{"finance:write", "财务写入", "调账/补单/提现审核"},
		{"payment:write", "支付写入", "支付补单/异常处理"},
		{"settings:write", "设置写入", "修改系统配置"},
		{"settings:read", "设置只读", "查看系统配置"},
	}
	for _, p := range perms {
		var existing Permission
		err := db.DB.Where("code = ?", p.Code).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.DB.Create(&Permission{
				Code: p.Code, Name: p.Name, Description: p.Desc, CreateTime: now,
			}).Error; err != nil {
				log.Printf("[Init] 播种权限 %s 失败: %v", p.Code, err)
			}
		} else if err != nil {
			log.Printf("[Init] 查询权限 %s 失败: %v", p.Code, err)
		}
	}

	roles := []struct{ Code, Name, Desc string }{
		{"admin", "管理员", "拥有全部权限"},
		{"operator", "运营", "只读 + 有限写入（用户/支付）"},
		{"viewer", "只读访客", "仅只读权限"},
	}
	for _, r := range roles {
		var role Role
		err := db.DB.Where("code = ?", r.Code).First(&role).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			role = Role{Code: r.Code, Name: r.Name, Description: r.Desc, CreateTime: now}
			if err := db.DB.Create(&role).Error; err != nil {
				log.Printf("[Init] 播种角色 %s 失败: %v", r.Code, err)
				continue
			}
		} else if err != nil {
			log.Printf("[Init] 查询角色 %s 失败: %v", r.Code, err)
			continue
		}
		_ = syncRolePermissions(role.ID, r.Code)
	}
}

// syncRolePermissions 按角色编码同步默认权限集（幂等）
func syncRolePermissions(roleID uint64, roleCode string) error {
	var codes []string
	switch roleCode {
	case "admin":
		codes = []string{"user:read", "user:write", "finance:read", "finance:write", "payment:write", "settings:read", "settings:write"}
	case "operator":
		codes = []string{"user:read", "user:write", "finance:read", "payment:write", "settings:read"}
	case "viewer":
		codes = []string{"user:read", "finance:read", "settings:read"}
	default:
		return nil
	}
	for _, code := range codes {
		var perm Permission
		if err := db.DB.Where("code = ?", code).First(&perm).Error; err != nil {
			continue
		}
		var count int64
		db.DB.Model(&RolePermission{}).
			Where("role_id = ? AND permission_id = ?", roleID, perm.ID).
			Count(&count)
		if count > 0 {
			continue
		}
		_ = db.DB.Create(&RolePermission{RoleID: roleID, PermissionID: perm.ID}).Error
	}
	return nil
}

// ListRoles 列出全部角色
func ListRoles() ([]Role, error) {
	var list []Role
	err := db.DB.Order("id ASC").Find(&list).Error
	return list, err
}

// ListPermissions 列出全部权限点
func ListPermissions() ([]Permission, error) {
	var list []Permission
	err := db.DB.Order("id ASC").Find(&list).Error
	return list, err
}

// GetRoleByCode 按编码取角色
func GetRoleByCode(code string) (*Role, error) {
	var role Role
	err := db.DB.Where("code = ?", code).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByID 按 ID 取角色
func GetRoleByID(id uint64) (*Role, error) {
	var role Role
	err := db.DB.Where("id = ?", id).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// AssignUserRole 为用户分配 RBAC 角色（替换该用户全部角色，MVP 单角色）
func AssignUserRole(userID, roleID uint64) error {
	now := time.Now().Unix()
	return db.WithTx(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&UserRole{}).Error; err != nil {
			return err
		}
		return tx.Create(&UserRole{UserID: userID, RoleID: roleID, CreateTime: now}).Error
	})
}

// ListUserRoles 列出用户已分配的 RBAC 角色
func ListUserRoles(userID uint64) ([]Role, error) {
	var list []Role
	err := db.DB.Table("roles r").
		Select("r.id, r.code, r.name, r.description, r.create_time").
		Joins("INNER JOIN user_roles ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		Order("r.id ASC").
		Scan(&list).Error
	return list, err
}

// UserHasPermissionCode 判断用户是否拥有某权限码（不含 users.role=admin 旁路，旁路在中间件）
func UserHasPermissionCode(userID uint64, permCode string) (bool, error) {
	var count int64
	err := db.DB.Table("user_roles ur").
		Joins("INNER JOIN role_permissions rp ON rp.role_id = ur.role_id").
		Joins("INNER JOIN permissions p ON p.id = rp.permission_id").
		Where("ur.user_id = ? AND p.code = ?", userID, permCode).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountUserRoles 统计用户角色数（测试/诊断用）
func CountUserRoles(userID uint64) (int64, error) {
	var n int64
	err := db.DB.Model(&UserRole{}).Where("user_id = ?", userID).Count(&n).Error
	return n, err
}

// EnsurePermissionExists 确保权限点存在（测试辅助）
func EnsurePermissionExists(code, name string) error {
	var existing Permission
	err := db.DB.Where("code = ?", code).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.DB.Create(&Permission{
		Code: code, Name: name, Description: "", CreateTime: time.Now().Unix(),
	}).Error
}

// FormatRBACError 统一错误文案
func FormatRBACError(op string, err error) error {
	return fmt.Errorf("rbac %s: %w", op, err)
}
