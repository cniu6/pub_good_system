package models

import (
	"database/sql"
	"fmt"
	"fst/backend/pkg/db"
	"log"
	"time"
)

// Role RBAC 角色（单组织 MVP：admin / operator / viewer）
type Role struct {
	ID          uint64 `db:"id" json:"id"`
	Code        string `db:"code" json:"code"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
	CreateTime  int64  `db:"create_time" json:"create_time"`
}

// Permission RBAC 权限点
type Permission struct {
	ID          uint64 `db:"id" json:"id"`
	Code        string `db:"code" json:"code"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
	CreateTime  int64  `db:"create_time" json:"create_time"`
}

// RolePermission 角色-权限关联
type RolePermission struct {
	RoleID       uint64 `db:"role_id" json:"role_id"`
	PermissionID uint64 `db:"permission_id" json:"permission_id"`
}

// UserRole 用户-角色关联
type UserRole struct {
	UserID     uint64 `db:"user_id" json:"user_id"`
	RoleID     uint64 `db:"role_id" json:"role_id"`
	CreateTime int64  `db:"create_time" json:"create_time"`
}

// InitRBACTables 创建 RBAC 表并播种默认角色/权限
func InitRBACTables() {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS roles (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(50) NOT NULL COMMENT '角色编码',
			name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '显示名',
			description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '说明',
			create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
			UNIQUE KEY uk_roles_code (code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='RBAC角色'`,
		`CREATE TABLE IF NOT EXISTS permissions (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			code VARCHAR(80) NOT NULL COMMENT '权限编码如 user:read',
			name VARCHAR(100) NOT NULL DEFAULT '' COMMENT '显示名',
			description VARCHAR(255) NOT NULL DEFAULT '' COMMENT '说明',
			create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
			UNIQUE KEY uk_permissions_code (code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='RBAC权限点'`,
		`CREATE TABLE IF NOT EXISTS role_permissions (
			role_id BIGINT UNSIGNED NOT NULL,
			permission_id BIGINT UNSIGNED NOT NULL,
			PRIMARY KEY (role_id, permission_id),
			INDEX idx_rp_permission (permission_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联'`,
		`CREATE TABLE IF NOT EXISTS user_roles (
			user_id BIGINT UNSIGNED NOT NULL,
			role_id BIGINT UNSIGNED NOT NULL,
			create_time BIGINT NOT NULL DEFAULT 0,
			PRIMARY KEY (user_id, role_id),
			INDEX idx_ur_role (role_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联'`,
	}
	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			log.Printf("[Init] RBAC 建表失败: %v", err)
		}
	}
	seedRBACDefaults()
}

func seedRBACDefaults() {
	now := time.Now().Unix()

	// 权限点种子
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
		var id uint64
		err := db.DB.Get(&id, "SELECT id FROM permissions WHERE code = ?", p.Code)
		if err == sql.ErrNoRows {
			_, err = db.Exec(
				`INSERT INTO permissions (code, name, description, create_time) VALUES (?, ?, ?, ?)`,
				p.Code, p.Name, p.Desc, now,
			)
			if err != nil {
				log.Printf("[Init] 播种权限 %s 失败: %v", p.Code, err)
			}
		}
	}

	// 角色种子
	roles := []struct{ Code, Name, Desc string }{
		{"admin", "管理员", "拥有全部权限"},
		{"operator", "运营", "只读 + 有限写入（用户/支付）"},
		{"viewer", "只读访客", "仅只读权限"},
	}
	for _, r := range roles {
		var id uint64
		err := db.DB.Get(&id, "SELECT id FROM roles WHERE code = ?", r.Code)
		if err == sql.ErrNoRows {
			res, err := db.Exec(
				`INSERT INTO roles (code, name, description, create_time) VALUES (?, ?, ?, ?)`,
				r.Code, r.Name, r.Desc, now,
			)
			if err != nil {
				log.Printf("[Init] 播种角色 %s 失败: %v", r.Code, err)
				continue
			}
			lid, _ := res.LastInsertId()
			id = uint64(lid)
		}
		_ = syncRolePermissions(id, r.Code)
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
		var permID uint64
		if err := db.DB.Get(&permID, "SELECT id FROM permissions WHERE code = ?", code); err != nil {
			continue
		}
		var exists int
		_ = db.DB.Get(&exists, "SELECT COUNT(1) FROM role_permissions WHERE role_id = ? AND permission_id = ?", roleID, permID)
		if exists > 0 {
			continue
		}
		_, _ = db.Exec(
			`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)`,
			roleID, permID,
		)
	}
	return nil
}

// ListRoles 列出全部角色
func ListRoles() ([]Role, error) {
	var list []Role
	err := db.DB.Select(&list, "SELECT id, code, name, description, create_time FROM roles ORDER BY id ASC")
	return list, err
}

// ListPermissions 列出全部权限点
func ListPermissions() ([]Permission, error) {
	var list []Permission
	err := db.DB.Select(&list, "SELECT id, code, name, description, create_time FROM permissions ORDER BY id ASC")
	return list, err
}

// GetRoleByCode 按编码取角色
func GetRoleByCode(code string) (*Role, error) {
	var role Role
	err := db.DB.Get(&role, "SELECT id, code, name, description, create_time FROM roles WHERE code = ?", code)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByID 按 ID 取角色
func GetRoleByID(id uint64) (*Role, error) {
	var role Role
	err := db.DB.Get(&role, "SELECT id, code, name, description, create_time FROM roles WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// AssignUserRole 为用户分配 RBAC 角色（替换该用户全部角色，MVP 单角色）
func AssignUserRole(userID, roleID uint64) error {
	now := time.Now().Unix()
	tx, err := db.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM user_roles WHERE user_id = ?", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`INSERT INTO user_roles (user_id, role_id, create_time) VALUES (?, ?, ?)`,
		userID, roleID, now,
	); err != nil {
		return err
	}
	return tx.Commit()
}

// ListUserRoles 列出用户已分配的 RBAC 角色
func ListUserRoles(userID uint64) ([]Role, error) {
	var list []Role
	err := db.DB.Select(&list, `
		SELECT r.id, r.code, r.name, r.description, r.create_time
		FROM roles r
		INNER JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = ?
		ORDER BY r.id ASC`, userID)
	return list, err
}

// UserHasPermissionCode 判断用户是否拥有某权限码（不含 users.role=admin 旁路，旁路在中间件）
func UserHasPermissionCode(userID uint64, permCode string) (bool, error) {
	var count int
	err := db.DB.Get(&count, `
		SELECT COUNT(1) FROM user_roles ur
		INNER JOIN role_permissions rp ON rp.role_id = ur.role_id
		INNER JOIN permissions p ON p.id = rp.permission_id
		WHERE ur.user_id = ? AND p.code = ?`, userID, permCode)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CountUserRoles 统计用户角色数（测试/诊断用）
func CountUserRoles(userID uint64) (int64, error) {
	var n int64
	err := db.DB.Get(&n, "SELECT COUNT(1) FROM user_roles WHERE user_id = ?", userID)
	return n, err
}

// EnsurePermissionExists 确保权限点存在（测试辅助）
func EnsurePermissionExists(code, name string) error {
	var id uint64
	err := db.DB.Get(&id, "SELECT id FROM permissions WHERE code = ?", code)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}
	_, err = db.Exec(
		`INSERT INTO permissions (code, name, description, create_time) VALUES (?, ?, ?, ?)`,
		code, name, "", time.Now().Unix(),
	)
	return err
}

// FormatRBACError 统一错误文案
func FormatRBACError(op string, err error) error {
	return fmt.Errorf("rbac %s: %w", op, err)
}
