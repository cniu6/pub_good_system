package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// TestSQLiteCriticalDML 用真实 SQLite 跑审计里会炸的几类语句，确认适配后可执行。
func TestSQLiteCriticalDML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "crit.db")
	dsn := "file:" + filepath.ToSlash(path) + "?_pragma=busy_timeout(5000)&_txlock=immediate"

	xdb, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer xdb.Close()
	xdb.SetMaxOpenConns(1)

	// 挂到全局，使 Q()/IsSQLite 生效
	oldDB, oldDriver := DB, activeDriver
	DB = xdb
	activeDriver = "sqlite"
	defer func() {
		DB = oldDB
		activeDriver = oldDriver
	}()

	mustExec := func(sql string, args ...any) {
		t.Helper()
		if _, err := Exec(sql, args...); err != nil {
			t.Fatalf("Exec failed sql=%s err=%v", sql, err)
		}
	}

	mustExec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		money REAL NOT NULL DEFAULT 0,
		score INTEGER NOT NULL DEFAULT 0,
		login_failure INTEGER NOT NULL DEFAULT 0,
		delete_time INTEGER,
		create_time INTEGER NOT NULL DEFAULT 0,
		last_login_time INTEGER NOT NULL DEFAULT 0,
		status INTEGER NOT NULL DEFAULT 1
	)`)
	mustExec(`CREATE TABLE payment_orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_no TEXT NOT NULL UNIQUE,
		status INTEGER NOT NULL DEFAULT 0,
		pay_amount REAL NOT NULL DEFAULT 0,
		paid_at INTEGER
	)`)
	mustExec(`CREATE TABLE verification_codes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		contact TEXT NOT NULL,
		code_type TEXT NOT NULL,
		code TEXT NOT NULL,
		is_used INTEGER NOT NULL DEFAULT 0,
		is_deleted INTEGER NOT NULL DEFAULT 0,
		expires_at TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`)
	mustExec(`CREATE TABLE system_settings (
		setting_key TEXT PRIMARY KEY,
		setting_value TEXT NOT NULL DEFAULT '',
		updated_at TEXT
	)`)
	mustExec(`CREATE TABLE api_access_log_daily_stats (
		day_key INTEGER PRIMARY KEY,
		total_count INTEGER NOT NULL DEFAULT 0,
		updated_at INTEGER NOT NULL DEFAULT 0
	)`)
	mustExec(`CREATE TABLE operation_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		create_time INTEGER NOT NULL DEFAULT 0,
		created_at TEXT
	)`)

	mustExec(`INSERT INTO users (money, score, create_time, last_login_time) VALUES (10.5, 3, ?, ?)`, time.Now().Unix()-3600, time.Now().Unix())
	mustExec(`INSERT INTO payment_orders (order_no, status, pay_amount, paid_at) VALUES ('O1', 1, 9.9, ?)`, time.Now().Unix())
	mustExec(`INSERT INTO verification_codes (contact, code_type, code, expires_at, created_at, updated_at) VALUES ('a@b.c','email','123456', datetime('now','+1 hour'), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`)
	mustExec(`INSERT INTO system_settings (setting_key, setting_value) VALUES ('site_name','fst')`)
	mustExec(`INSERT INTO operation_logs (create_time, created_at) VALUES (0, CURRENT_TIMESTAMP)`)

	tx, err := DB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	var money float64
	if err := tx.QueryRow(Q("SELECT money FROM users WHERE id = ? AND delete_time IS NULL FOR UPDATE"), 1).Scan(&money); err != nil {
		t.Fatalf("FOR UPDATE 适配后仍失败: %v", err)
	}
	_ = tx.Commit()

	var cnt int
	if err := DB.Get(&cnt, Q("SELECT COUNT(*) FROM verification_codes WHERE expires_at > NOW() AND is_deleted = 0")); err != nil {
		t.Fatalf("NOW() 适配后仍失败: %v", err)
	}
	if cnt < 1 {
		t.Fatalf("期望有未过期验证码")
	}

	if _, err := Exec("UPDATE system_settings SET setting_value = ?, updated_at = NOW() WHERE setting_key = ?", "x", "site_name"); err != nil {
		t.Fatalf("settings NOW() 失败: %v", err)
	}

	if _, err := Exec(`INSERT INTO api_access_log_daily_stats (day_key, total_count, updated_at) VALUES (?, 1, ?)
		ON DUPLICATE KEY UPDATE total_count = total_count + 1, updated_at = ?`, 20260718, 1, 2); err != nil {
		t.Fatalf("ON DUPLICATE 适配失败: %v", err)
	}
	if _, err := Exec(`INSERT INTO api_access_log_daily_stats (day_key, total_count, updated_at) VALUES (?, 1, ?)
		ON DUPLICATE KEY UPDATE total_count = total_count + 1, updated_at = ?`, 20260718, 3, 4); err != nil {
		t.Fatalf("ON CONFLICT 第二次失败: %v", err)
	}
	var total int64
	if err := DB.Get(&total, "SELECT total_count FROM api_access_log_daily_stats WHERE day_key = ?", 20260718); err != nil {
		t.Fatalf("read daily: %v", err)
	}
	if total != 2 {
		t.Fatalf("期望 upsert 累加为 2，实际 %d", total)
	}

	type row struct {
		Day   string `db:"day"`
		Value int64  `db:"value"`
	}
	var rows []row
	q := Q("SELECT DATE(FROM_UNIXTIME(create_time)) AS day, COUNT(*) AS value FROM users WHERE create_time >= ? GROUP BY DATE(FROM_UNIXTIME(create_time))")
	if err := DB.Select(&rows, q, time.Now().Add(-48*time.Hour).Unix()); err != nil {
		t.Fatalf("仪表盘 FROM_UNIXTIME 适配失败: %v sql=%s", err, q)
	}
	if len(rows) < 1 {
		t.Fatalf("期望有趋势数据")
	}

	if _, err := Exec("UPDATE operation_logs SET create_time = UNIX_TIMESTAMP(created_at) WHERE create_time = 0 AND created_at IS NOT NULL"); err != nil {
		t.Fatalf("UNIX_TIMESTAMP 适配失败: %v", err)
	}

	// 清理临时文件提示（TempDir 会删）
	_ = os.Remove(path)
}
