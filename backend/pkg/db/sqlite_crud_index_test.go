package db_test

import (
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/migrate"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

// TestSQLiteMigrateAndIndexes 只覆盖 pkg/db 独特职责：
// 表探测、索引探测（GORM AutoMigrate 落库）、二次迁移幂等。
// 业务 CRUD 由 app/models、app/services 各自测试，不在此重复。
func TestSQLiteMigrateAndIndexes(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	needTables := []string{
		"users", "email_logs", "email_templates", "verification_codes",
		"user_realname_verifications", "auto_job_definitions", "auto_job_runs",
		"system_settings", "user_settings", "user_sessions",
		"user_money_logs", "user_score_logs", "operation_logs", "api_access_logs",
		"sms_logs", "sms_templates", "payment_orders", "withdraw_requests",
		"idempotency_keys", "pay_gateways",
		"email_log_stats", "sms_log_stats", "operation_log_stats", "api_access_log_stats",
	}
	for _, name := range needTables {
		if !db.CheckTableExists(name) {
			t.Fatalf("缺表: %s", name)
		}
	}

	indexes := []struct{ table, index string }{
		{"users", "idx_users_username"},
		{"users", "idx_users_email"},
		{"users", "idx_users_status"},
		{"email_logs", "idx_email_logs_to"},
		{"verification_codes", "idx_vc_contact_type_active_created"},
		{"user_realname_verifications", "uk_realname_cert_unique_key"},
		{"auto_job_definitions", "idx_auto_job_def_enabled"},
		{"auto_job_runs", "idx_auto_job_runs_job_started"},
	}
	var missing []string
	for _, item := range indexes {
		if !db.CheckIndexExists(item.table, item.index) {
			missing = append(missing, item.table+"."+item.index)
		}
	}
	if len(missing) > 0 {
		var names []string
		_ = db.DB.Raw(`SELECT name FROM sqlite_master WHERE type='index' ORDER BY name`).Scan(&names).Error
		t.Fatalf("缺索引: %v\n当前索引=%v", missing, names)
	}

	migrate.RunAutoMigrate()
	if !db.CheckTableExists("users") || !db.CheckIndexExists("users", "idx_users_username") {
		t.Fatal("二次迁移后 users/索引异常")
	}

	u1 := testutil.CreateTestUser(t, "idx-unique-1")
	if err := models.CreateUser(&models.User{
		Username: u1.Username, Email: "other-" + u1.Email, Password: "x", Role: "user", Status: 1,
	}); err == nil {
		t.Fatal("同 username 应被唯一索引拒绝")
	}
}
