package db_test

import (
	"os"
	"path/filepath"
	"testing"

	"fst/backend/internal/migrate"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
)

// TestSQLiteBootstrap 验证：显式配置 sqlite 时能连库并跑完自迁移（临时缓解路径）。
func TestSQLiteBootstrap(t *testing.T) {
	dir := t.TempDir()
	dbFile := filepath.Join(dir, "fst_test.db")

	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_PATH", dbFile)
	t.Setenv("JWT_SECRET", "test-jwt-secret-at-least-16")
	t.Setenv("CORS_ORIGINS", "*")
	t.Setenv("APP_ENV", "development")

	config.InitConfig()
	db.InitDB()
	defer func() {
		if db.DB != nil {
			_ = db.DB.Close()
		}
	}()

	if !db.IsSQLite() {
		t.Fatalf("期望 sqlite，实际 %s", db.DriverName())
	}

	migrate.RunAutoMigrate()

	if !db.CheckTableExists("users") {
		t.Fatal("users 表未创建")
	}
	if !db.CheckTableExists("system_settings") {
		t.Fatal("system_settings 表未创建")
	}
	if !db.CheckColumnExists("users", "username") {
		t.Fatal("users.username 列探测失败")
	}

	if _, err := os.Stat(dbFile); err != nil {
		t.Fatalf("数据文件未生成: %v", err)
	}
}
