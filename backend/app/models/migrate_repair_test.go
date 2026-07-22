package models_test

import (
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

// TestInitSMSTemplatesTable_RepairsSignName 旧表缺 sign_name 时应能补列。
func TestInitSMSTemplatesTable_RepairsSignName(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	if _, err := db.Exec(`DROP TABLE IF EXISTS sms_templates`); err != nil {
		t.Fatalf("删表失败: %v", err)
	}
	_, err := db.Exec(`
CREATE TABLE sms_templates (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	lang TEXT NOT NULL DEFAULT 'zh-CN',
	content TEXT NOT NULL,
	created_at TEXT,
	updated_at TEXT
)`)
	if err != nil {
		t.Fatalf("建旧版 sms_templates 失败: %v", err)
	}
	if db.CheckColumnExists("sms_templates", "sign_name") {
		t.Fatal("旧表不应已有 sign_name")
	}

	models.InitSMSTemplatesTable()
	if !db.CheckColumnExists("sms_templates", "sign_name") {
		t.Fatal("补列后仍缺 sign_name")
	}
	if !db.CheckColumnExists("sms_templates", "description") {
		t.Fatal("补列后仍缺 description")
	}
}

// TestRepairVerificationCodes_EmailToContactSQLite SQLite 下 email→contact 应能拷贝数据。
func TestRepairVerificationCodes_EmailToContactSQLite(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	if _, err := db.Exec(`DROP TABLE IF EXISTS verification_codes`); err != nil {
		t.Fatalf("删表失败: %v", err)
	}
	_, err := db.Exec(`
CREATE TABLE verification_codes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL,
	code TEXT NOT NULL,
	code_type TEXT NOT NULL,
	expires_at TEXT NOT NULL,
	is_used INTEGER NOT NULL DEFAULT 0,
	is_deleted INTEGER NOT NULL DEFAULT 0,
	created_at TEXT,
	updated_at TEXT
)`)
	if err != nil {
		t.Fatalf("建旧版 verification_codes 失败: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO verification_codes (email, code, code_type, expires_at) VALUES ('a@b.com','123456','register','2099-01-01')`); err != nil {
		t.Fatalf("插入旧数据失败: %v", err)
	}

	models.InitVerificationCodeTable()

	if !db.CheckColumnExists("verification_codes", "contact") {
		t.Fatal("迁移后仍缺 contact")
	}
	var contact string
	if err := db.DB.Get(&contact, `SELECT contact FROM verification_codes WHERE code = '123456'`); err != nil {
		t.Fatalf("读 contact 失败: %v", err)
	}
	if contact != "a@b.com" {
		t.Fatalf("contact 未从 email 拷贝, got=%q", contact)
	}
	// 旧表缺 attempts 时，repair 路径应一并补上（防猜解字段）
	if !db.CheckColumnExists("verification_codes", "attempts") {
		t.Fatal("迁移后仍缺 attempts")
	}
}
