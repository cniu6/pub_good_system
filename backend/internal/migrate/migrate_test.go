package migrate_test

import (
	"testing"

	"fst/backend/internal/migrate"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

func TestRunAutoMigrate_IdempotentOnSQLite(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	tables := []string{
		"users", "system_settings", "payment_orders", "withdraw_requests",
		"operation_logs", "api_access_logs", "verification_codes", "user_sessions",
		"idempotency_keys", "user_money_logs", "user_score_logs",
	}
	for _, name := range tables {
		if !db.CheckTableExists(name) {
			t.Fatalf("迁移后缺少表: %s", name)
		}
	}

	// 再跑一次应不炸（IF NOT EXISTS / 幂等）
	migrate.RunAutoMigrate()
	if !db.CheckTableExists("users") {
		t.Fatal("二次迁移后 users 消失")
	}
	if !db.CheckColumnExists("users", "username") {
		t.Fatal("users.username 列探测失败")
	}
}
