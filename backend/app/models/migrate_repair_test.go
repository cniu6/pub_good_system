package models_test

import (
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

// TestRepairVerificationCodes_EmailToContactSQLite SQLite 下 email→contact 应能拷贝数据。
func TestRepairVerificationCodes_EmailToContactSQLite(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	if err := db.DB.Exec(`DROP TABLE IF EXISTS verification_codes`).Error; err != nil {
		t.Fatalf("删表失败: %v", err)
	}
	if err := db.DB.Exec(`
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
)`).Error; err != nil {
		t.Fatalf("建旧版 verification_codes 失败: %v", err)
	}
	if err := db.DB.Exec(`INSERT INTO verification_codes (email, code, code_type, expires_at) VALUES ('a@b.com','123456','register','2099-01-01')`).Error; err != nil {
		t.Fatalf("插入旧数据失败: %v", err)
	}

	models.RepairVerificationCodeTable()

	if !db.CheckColumnExists("verification_codes", "contact") {
		t.Fatal("迁移后仍缺 contact")
	}
	var contact string
	if err := db.DB.Raw(`SELECT contact FROM verification_codes WHERE code = '123456'`).Scan(&contact).Error; err != nil {
		t.Fatalf("读 contact 失败: %v", err)
	}
	if contact != "a@b.com" {
		t.Fatalf("contact 未从 email 拷贝, got=%q", contact)
	}
	if !db.CheckColumnExists("verification_codes", "attempts") {
		t.Fatal("迁移后仍缺 attempts")
	}
}
