// Package testutil 提供后端测试共用的 SQLite 临时库启动/清理工具。
// 每个测试应独立调用 SetupSQLite，避免互相污染全局 DB。
package testutil

import (
	"path/filepath"
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/migrate"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
)

// SetupSQLite 在临时目录创建 SQLite，初始化配置与全量迁移。
// 返回清理函数（关闭 DB）。测试内请：defer cleanup()
func SetupSQLite(t *testing.T) func() {
	t.Helper()

	dir := t.TempDir()
	dbFile := filepath.Join(dir, "test.db")

	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_PATH", dbFile)
	t.Setenv("JWT_SECRET", "test-jwt-secret-at-least-16-chars")
	t.Setenv("ADMIN_JWT_SECRET", "test-admin-jwt-secret-16")
	t.Setenv("CORS_ORIGINS", "*")
	t.Setenv("APP_ENV", "development")
	t.Setenv("ENABLE_SWAGGER", "false")
	t.Setenv("ADMIN_API_PATH", "/admin")

	// 避免读到磁盘上的真实 .env 干扰：godotenv 通常不覆盖已有环境变量，上面 Setenv 优先
	config.InitConfig()
	db.InitDB()
	if !db.IsSQLite() {
		t.Fatalf("期望 sqlite，实际 %s", db.DriverName())
	}
	migrate.RunAutoMigrate()

	return func() {
		if db.DB != nil {
			_ = db.DB.Close()
			db.DB = nil
		}
	}
}

// CreateTestUser 插入一个可用测试用户，返回带 ID 的对象。
func CreateTestUser(t *testing.T, username string) *models.User {
	t.Helper()
	u := &models.User{
		Username: username,
		Nickname: username,
		Email:    username + "@example.test",
		Password: "$2a$10$testhashnotusedforloginxxxxxxxxxxx",
		Money:    100,
		Score:    20,
		Role:     "user",
		Status:   1,
	}
	if err := models.CreateUser(u); err != nil {
		t.Fatalf("CreateUser(%s) 失败: %v", username, err)
	}
	if u.ID == 0 {
		t.Fatalf("CreateUser 未回填 ID")
	}
	return u
}

// CreateTestAdmin 插入管理员用户。
func CreateTestAdmin(t *testing.T, username string) *models.User {
	t.Helper()
	u := CreateTestUser(t, username)
	if _, err := db.Exec(`UPDATE users SET role = ? WHERE id = ?`, "admin", u.ID); err != nil {
		t.Fatalf("提升管理员失败: %v", err)
	}
	u.Role = "admin"
	return u
}
