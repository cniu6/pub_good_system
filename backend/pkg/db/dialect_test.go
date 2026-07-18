package db

import (
	"strings"
	"testing"
)

func TestAdaptCreateTable_Basic(t *testing.T) {
	mysql := `CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL COMMENT '用户名',
			status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态',
			UNIQUE KEY idx_users_username (username),
			INDEX idx_users_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

	stmts := AdaptMySQLDDLToSQLite(mysql)
	if len(stmts) < 3 {
		t.Fatalf("期望至少 1 条建表 + 2 条索引，实际 %d: %#v", len(stmts), stmts)
	}
	create := stmts[0]
	if !strings.Contains(create, "INTEGER PRIMARY KEY AUTOINCREMENT") {
		t.Fatalf("主键未转换为 SQLite AUTOINCREMENT: %s", create)
	}
	if strings.Contains(strings.ToUpper(create), "ENGINE") {
		t.Fatalf("仍含 ENGINE: %s", create)
	}
	if strings.Contains(strings.ToUpper(create), "COMMENT") {
		t.Fatalf("仍含 COMMENT: %s", create)
	}
	if strings.Contains(strings.ToUpper(create), "UNSIGNED") {
		t.Fatalf("仍含 UNSIGNED: %s", create)
	}
	joined := strings.Join(stmts[1:], "\n")
	if !strings.Contains(joined, "CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username") {
		t.Fatalf("缺少唯一索引: %s", joined)
	}
	if !strings.Contains(joined, "CREATE INDEX IF NOT EXISTS idx_users_status") {
		t.Fatalf("缺少普通索引: %s", joined)
	}
}

func TestAdaptAddIndex(t *testing.T) {
	mysql := "ALTER TABLE email_logs ADD INDEX idx_email_logs_created_at (created_at)"
	stmts := AdaptMySQLDDLToSQLite(mysql)
	if len(stmts) != 1 {
		t.Fatalf("期望 1 条，实际 %#v", stmts)
	}
	want := "CREATE INDEX IF NOT EXISTS idx_email_logs_created_at ON email_logs (created_at)"
	if stmts[0] != want {
		t.Fatalf("got %q want %q", stmts[0], want)
	}
}

func TestAdaptChangeColumnSkipped(t *testing.T) {
	mysql := "ALTER TABLE verification_codes CHANGE COLUMN email contact VARCHAR(255) NOT NULL COMMENT 'x'"
	stmts := AdaptMySQLDDLToSQLite(mysql)
	if stmts != nil {
		t.Fatalf("CHANGE COLUMN 应跳过，实际 %#v", stmts)
	}
}

func TestAdaptAddColumn(t *testing.T) {
	mysql := "ALTER TABLE users ADD COLUMN group_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '分组ID' AFTER id"
	stmts := AdaptMySQLDDLToSQLite(mysql)
	if len(stmts) != 1 {
		t.Fatalf("期望 1 条，实际 %#v", stmts)
	}
	s := stmts[0]
	if strings.Contains(strings.ToUpper(s), "AFTER") || strings.Contains(strings.ToUpper(s), "COMMENT") || strings.Contains(strings.ToUpper(s), "UNSIGNED") {
		t.Fatalf("清洗不彻底: %s", s)
	}
}
