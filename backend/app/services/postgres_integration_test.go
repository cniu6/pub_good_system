//go:build integration

package services

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/migrate"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
)

// 真实 PostgreSQL 上的服务层回归测试：SQLite 下 ForUpdate 是空操作且写事务串行，
// 行锁与并发问题只能在真实库上暴露。
//
//	set FST_PG_DSN=postgres://postgres@127.0.0.1:55432/fst_test?sslmode=disable
//	go test -tags integration ./backend/app/services/ -run Postgres -count=1
func setupPostgres(t *testing.T) {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("FST_PG_DSN"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("TEST_POSTGRES_DSN"))
	}
	if dsn == "" {
		t.Skip("未设置 FST_PG_DSN / TEST_POSTGRES_DSN，跳过 PostgreSQL 集成测试")
	}

	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("JWT_SECRET", "test-jwt-secret-at-least-16-chars")
	t.Setenv("ADMIN_JWT_SECRET", "test-admin-jwt-secret-16")
	t.Setenv("CORS_ORIGINS", "*")
	t.Setenv("APP_ENV", "development")
	t.Setenv("ENABLE_SWAGGER", "false")
	t.Setenv("ADMIN_API_PATH", "/admin")

	config.InitConfig()
	config.UpdateGlobalConfig(func(cfg *config.Config) {
		cfg.DBDriver = "postgres"
		cfg.DBDSN = dsn
	})
	db.InitDB()
	t.Cleanup(func() { _ = db.Close() })

	// 每个用例从空 schema 开始，避免数据互相污染（测试库专用，切勿指向业务库）。
	for _, stmt := range []string{"DROP SCHEMA public CASCADE", "CREATE SCHEMA public"} {
		if err := db.DB.Exec(stmt).Error; err != nil {
			t.Fatalf("重置 schema 失败(%s): %v", stmt, err)
		}
	}
	migrate.RunAutoMigrate()
}

// TestPostgresWithdrawCreate 提现创建在 PG 上必须能成功，且同一用户的第二笔待处理申请仍被拦截。
func TestPostgresWithdrawCreate(t *testing.T) {
	setupPostgres(t)

	old := GlobalSettingsService
	defer func() { GlobalSettingsService = old }()
	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"withdraw_enabled": {Key: "withdraw_enabled", Value: "true"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	user := testutil.CreateTestUser(t, "pg-withdraw-user")
	req := func() *CreateWithdrawRequest {
		return &CreateWithdrawRequest{Amount: 10, AccountType: "bank", AccountName: "test", AccountNo: "123456", RealName: "测试用户"}
	}

	svc := NewWithdrawService()
	if _, err := svc.Create(user.ID, req()); err != nil {
		t.Fatalf("PG 上首笔提现应创建成功，实际: %v", err)
	}
	if _, err := svc.Create(user.ID, req()); err == nil || !IsClientError(err) {
		t.Fatalf("已有待处理提现时应拒绝第二笔申请，实际: %v", err)
	}
}

// TestPostgresLastAdminConcurrentDisable 两个管理员被并发禁用时，必须至少保留一个启用管理员。
func TestPostgresLastAdminConcurrentDisable(t *testing.T) {
	setupPostgres(t)
	svc := NewUserService()

	for round := 0; round < 30; round++ {
		a := testutil.CreateTestAdmin(t, fmt.Sprintf("pg-admin-a-%d", round))
		b := testutil.CreateTestAdmin(t, fmt.Sprintf("pg-admin-b-%d", round))
		// 每轮只保留 a、b 两个启用管理员，制造「最后两个」并发竞争。
		if err := db.DB.Model(&models.User{}).
			Where("role = ? AND id NOT IN ?", "admin", []uint64{a.ID, b.ID}).
			Update("status", 0).Error; err != nil {
			t.Fatalf("准备数据失败: %v", err)
		}

		start := make(chan struct{})
		var wg sync.WaitGroup
		for _, id := range []uint64{a.ID, b.ID} {
			wg.Add(1)
			go func(id uint64) {
				defer wg.Done()
				<-start
				_ = svc.UpdateStatus(id, 0)
			}(id)
		}
		close(start)
		wg.Wait()

		var active int64
		if err := db.DB.Model(&models.User{}).
			Where("role = ? AND status = ? AND delete_time IS NULL", "admin", 1).
			Count(&active).Error; err != nil {
			t.Fatalf("统计管理员失败: %v", err)
		}
		if active < 1 {
			t.Fatalf("第 %d 轮并发禁用后已没有启用中的管理员", round)
		}
	}
}
