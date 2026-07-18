package db

import (
	"strings"
	"testing"
)

func TestAdaptMySQLQuery_ForUpdate(t *testing.T) {
	in := "SELECT money FROM users WHERE id = ? AND delete_time IS NULL FOR UPDATE"
	out := AdaptMySQLQueryToSQLite(in)
	if strings.Contains(strings.ToUpper(out), "FOR UPDATE") {
		t.Fatalf("应剥掉 FOR UPDATE: %s", out)
	}
	if !strings.Contains(out, "SELECT money FROM users") {
		t.Fatalf("主体被破坏: %s", out)
	}
}

func TestAdaptMySQLQuery_NowAndDateSub(t *testing.T) {
	out := AdaptMySQLQueryToSQLite("SELECT * FROM t WHERE expires_at > NOW()")
	if strings.Contains(strings.ToUpper(out), "NOW()") {
		t.Fatalf("NOW 未替换: %s", out)
	}
	if !strings.Contains(out, "CURRENT_TIMESTAMP") {
		t.Fatalf("期望 CURRENT_TIMESTAMP: %s", out)
	}

	out2 := AdaptMySQLQueryToSQLite("DELETE FROM t WHERE updated_at < DATE_SUB(NOW(), INTERVAL 7 DAY)")
	if strings.Contains(strings.ToUpper(out2), "DATE_SUB") || strings.Contains(strings.ToUpper(out2), "NOW()") {
		t.Fatalf("DATE_SUB/NOW 未替换: %s", out2)
	}
	if !strings.Contains(out2, "datetime('now', '-7 day')") {
		t.Fatalf("期望 datetime modifier: %s", out2)
	}
}

func TestAdaptMySQLQuery_FromUnixTime(t *testing.T) {
	out := AdaptMySQLQueryToSQLite("SELECT DATE(FROM_UNIXTIME(create_time)) AS day FROM users GROUP BY DATE(FROM_UNIXTIME(create_time))")
	if strings.Contains(strings.ToUpper(out), "FROM_UNIXTIME") {
		t.Fatalf("FROM_UNIXTIME 未替换: %s", out)
	}
	if !strings.Contains(out, "date(create_time, 'unixepoch')") {
		t.Fatalf("期望 date(..., unixepoch): %s", out)
	}

	out2 := AdaptMySQLQueryToSQLite("SELECT CAST(DATE_FORMAT(FROM_UNIXTIME(create_time), '%Y%m%d') AS UNSIGNED) AS day_key FROM api_access_logs")
	if strings.Contains(strings.ToUpper(out2), "FROM_UNIXTIME") || strings.Contains(strings.ToUpper(out2), "DATE_FORMAT") {
		t.Fatalf("DATE_FORMAT/FROM_UNIXTIME 未替换: %s", out2)
	}
	if !strings.Contains(out2, "strftime('%Y%m%d'") {
		t.Fatalf("期望 strftime: %s", out2)
	}
	if strings.Contains(strings.ToUpper(out2), "UNSIGNED") {
		t.Fatalf("UNSIGNED 应改为 INTEGER: %s", out2)
	}
}

func TestAdaptMySQLQuery_OnDuplicateKey(t *testing.T) {
	in := `INSERT INTO api_access_log_daily_stats (day_key, total_count, updated_at)
    VALUES (?, 1, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      updated_at = ?`
	out := AdaptMySQLQueryToSQLite(in)
	if strings.Contains(strings.ToUpper(out), "ON DUPLICATE KEY") {
		t.Fatalf("ON DUPLICATE 未转换: %s", out)
	}
	if !strings.Contains(out, "ON CONFLICT(day_key) DO UPDATE SET") {
		t.Fatalf("期望 ON CONFLICT(day_key): %s", out)
	}
}

func TestAdaptMySQLQuery_UUID(t *testing.T) {
	out := AdaptMySQLQueryToSQLite("UPDATE api_access_logs SET request_id = LOWER(UUID()) WHERE request_id IS NULL")
	if strings.Contains(strings.ToUpper(out), "UUID()") {
		t.Fatalf("UUID 未替换: %s", out)
	}
	if !strings.Contains(out, "randomblob") {
		t.Fatalf("期望 randomblob: %s", out)
	}
}

func TestAdaptMySQLQuery_CharLength(t *testing.T) {
	out := AdaptMySQLQueryToSQLite("SELECT id FROM auto_job_runs WHERE run_uid = '' OR CHAR_LENGTH(run_uid) <> 36")
	if strings.Contains(strings.ToUpper(out), "CHAR_LENGTH") {
		t.Fatalf("CHAR_LENGTH 未替换: %s", out)
	}
	if !strings.Contains(out, "LENGTH(run_uid)") {
		t.Fatalf("期望 LENGTH(run_uid): %s", out)
	}

	out2 := AdaptMySQLQueryToSQLite("SELECT CHARACTER_LENGTH(name) FROM t")
	if strings.Contains(strings.ToUpper(out2), "CHARACTER_LENGTH") {
		t.Fatalf("CHARACTER_LENGTH 未替换: %s", out2)
	}
	if !strings.Contains(out2, "LENGTH(name)") {
		t.Fatalf("期望 LENGTH(name): %s", out2)
	}
}
