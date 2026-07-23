//go:build integration

package db_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/migrate"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
)

// TestPostgresIntegration 真机 PostgreSQL 集成测试。
//
// 重要：PostgreSQL 适配目前只做过字符串转换层单测，**在本测试于真实 PG 上通过之前，
// 不能视为生产已认证**。请用真实实例跑：
//
//	set FST_PG_DSN=postgres://user:pass@127.0.0.1:5432/fst_test?sslmode=disable
//	go test -tags integration ./backend/pkg/db/ -run Postgres -count=1
//
// 也可使用 TEST_POSTGRES_DSN（与 FST_PG_DSN 二选一）。未设置时 Skip，不影响默认 CI。
func TestPostgresIntegration(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("FST_PG_DSN"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("TEST_POSTGRES_DSN"))
	}
	if dsn == "" {
		t.Skip("Postgres 未生产认证：请设置 FST_PG_DSN 或 TEST_POSTGRES_DSN 后再跑本集成测试")
	}

	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("JWT_SECRET", "postgres-integration-test-secret-please-change")
	t.Setenv("CORS_ORIGINS", "*")
	t.Setenv("APP_ENV", "development")
	t.Setenv("DB_SSLMODE", "disable")

	config.InitConfig()
	config.UpdateGlobalConfig(func(cfg *config.Config) {
		cfg.DBDriver = "postgres"
		cfg.DBDSN = dsn
	})

	db.InitDB()
	t.Cleanup(func() {
		if db.DB != nil {
			_ = db.DB.Close()
		}
	})

	if !db.IsPostgres() {
		t.Fatalf("期望 postgres 驱动，实际 %q", db.DriverName())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.DB.PingContext(ctx); err != nil {
		t.Fatalf("Postgres Ping 失败: %v", err)
	}

	// 全量自迁移：真机验证 DDL 方言适配
	migrate.RunAutoMigrate()

	user := &models.User{
		Username: "pg-integration-user",
		Nickname: "PG 集成测试",
		Email:    "pg-integration@example.test",
		Password: "not-used-in-test",
		Money:    10,
		Score:    1,
		Role:     "user",
		Status:   1,
	}
	if err := models.CreateUser(user); err != nil {
		t.Fatalf("models.CreateUser 失败（Postgres 尚未生产认证直到此步通过）: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("CreateUser 未回填 ID")
	}

	got, err := models.GetUserByID(user.ID)
	if err != nil || got == nil || got.Username != user.Username {
		t.Fatalf("GetUserByID 失败: got=%+v err=%v", got, err)
	}

	t.Logf("Postgres 集成测试通过（driver=%s user_id=%d）。可视为真机冒烟通过，生产前仍建议覆盖核心业务路径。", db.DriverName(), user.ID)
}
