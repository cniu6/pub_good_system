package migrate_test

import (
	"testing"

	"fst/backend/internal/migrate"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

// TestDropObsoleteFinanceApprovalArtifacts 验证废弃审批表/设置行会被自迁移清掉。
func TestDropObsoleteFinanceApprovalArtifacts(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	if err := db.DB.Exec(`CREATE TABLE approval_requests (id INTEGER PRIMARY KEY)`).Error; err != nil {
		t.Fatalf("建废弃表失败: %v", err)
	}
	if err := db.DB.Exec(
		`INSERT INTO system_settings (setting_key, setting_value, setting_type, category, label) VALUES (?,?,?,?,?)`,
		"finance_dual_approval", "false", "boolean", "payment", "财务双人复核",
	).Error; err != nil {
		t.Fatalf("插入废弃设置失败: %v", err)
	}

	migrate.RunAutoMigrate()

	if db.CheckTableExists("approval_requests") {
		t.Fatal("approval_requests 应已被删除")
	}
	var n int64
	if err := db.DB.Raw(`SELECT COUNT(*) FROM system_settings WHERE setting_key = ?`, "finance_dual_approval").Scan(&n).Error; err != nil {
		t.Fatalf("查询设置失败: %v", err)
	}
	if n != 0 {
		t.Fatalf("finance_dual_approval 应已删除，仍有 %d 行", n)
	}
}

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
