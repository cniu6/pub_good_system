package db

import (
	"database/sql"
	"strings"
	"testing"
)

// TestPgShimDriverRegistered 确认 pg-shim 驱动已在 init() 里注册成功
// （避免以后有人重构 pg_shim.go 时手滑把 sql.Register 删掉，直到运行时连接 Postgres 才报错）。
func TestPgShimDriverRegistered(t *testing.T) {
	found := false
	for _, name := range sql.Drivers() {
		if name == pgShimDriverName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("驱动 %q 未注册，实际已注册: %v", pgShimDriverName, sql.Drivers())
	}
}

// TestAdaptForPostgresWire_CombinesFunctionAndPlaceholderRewrite 验证驱动层的组合入口：
// 反引号/MySQL函数转换 + 占位符 ?→$N 要同时生效，且顺序正确（函数转换在前，占位符在后，
// 避免函数转换过程中新引入的 ? 字符——目前实现里不会引入，但顺序本身要固定，防止以后改动踩坑）。
func TestAdaptForPostgresWire_CombinesFunctionAndPlaceholderRewrite(t *testing.T) {
	in := "UPDATE `users` SET money = money + ? WHERE id = ? FOR UPDATE"
	out := adaptForPostgresWire(in)
	want := "UPDATE users SET money = money + $1 WHERE id = $2 FOR UPDATE"
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
}

// 说明：本文件只测试 SQL 文本转换逻辑本身（纯字符串输入输出），不连真实 PostgreSQL。
// 本机没有可用的 Postgres 实例，无法端到端跑真机验证；上生产前请务必用真实 Postgres
// 过一遍 internal/migrate.RunAutoMigrate + 关键业务流程再放心用。

func TestAdaptMySQLQueryToPostgres_Backtick(t *testing.T) {
	out := AdaptMySQLQueryToPostgres("SELECT `id`, `username` FROM `users` WHERE `id` = ?")
	if strings.Contains(out, "`") {
		t.Fatalf("反引号未去除: %s", out)
	}
	if !strings.Contains(out, "SELECT id, username FROM users WHERE id = ?") {
		t.Fatalf("主体被破坏: %s", out)
	}
}

func TestAdaptMySQLQueryToPostgres_ForUpdateKeptAsIs(t *testing.T) {
	// Postgres 原生支持 FOR UPDATE，不应被剥掉（这点和 SQLite 版正好相反）
	in := "SELECT money FROM users WHERE id = ? FOR UPDATE"
	out := AdaptMySQLQueryToPostgres(in)
	if !strings.Contains(strings.ToUpper(out), "FOR UPDATE") {
		t.Fatalf("Postgres 原生支持 FOR UPDATE，不应被剥掉: %s", out)
	}
}

func TestAdaptMySQLQueryToPostgres_LockShareMode(t *testing.T) {
	out := AdaptMySQLQueryToPostgres("SELECT id FROM t WHERE id = ? LOCK IN SHARE MODE")
	if strings.Contains(strings.ToUpper(out), "LOCK IN SHARE MODE") {
		t.Fatalf("LOCK IN SHARE MODE 未转换: %s", out)
	}
	if !strings.Contains(strings.ToUpper(out), "FOR SHARE") {
		t.Fatalf("期望 FOR SHARE: %s", out)
	}
}

func TestAdaptMySQLQueryToPostgres_NowKeptAsIs(t *testing.T) {
	out := AdaptMySQLQueryToPostgres("SELECT * FROM t WHERE expires_at > NOW()")
	if !strings.Contains(strings.ToUpper(out), "NOW()") {
		t.Fatalf("Postgres 原生支持 NOW()，不应被替换: %s", out)
	}
}

func TestAdaptMySQLQueryToPostgres_DateSubNow(t *testing.T) {
	out := AdaptMySQLQueryToPostgres("DELETE FROM t WHERE updated_at < DATE_SUB(NOW(), INTERVAL 7 DAY)")
	if strings.Contains(strings.ToUpper(out), "DATE_SUB") {
		t.Fatalf("DATE_SUB 未替换: %s", out)
	}
	if !strings.Contains(out, "NOW() - INTERVAL '7 days'") {
		t.Fatalf("期望 NOW() - INTERVAL '7 days': %s", out)
	}
}

func TestAdaptMySQLQueryToPostgres_FromUnixTime(t *testing.T) {
	out := AdaptMySQLQueryToPostgres("SELECT DATE(FROM_UNIXTIME(create_time)) AS day FROM users")
	if strings.Contains(strings.ToUpper(out), "FROM_UNIXTIME") {
		t.Fatalf("FROM_UNIXTIME 未替换: %s", out)
	}
	if !strings.Contains(out, "to_timestamp(create_time)") {
		t.Fatalf("期望 to_timestamp(create_time): %s", out)
	}

	out2 := AdaptMySQLQueryToPostgres("SELECT CAST(DATE_FORMAT(FROM_UNIXTIME(create_time), '%Y%m%d') AS UNSIGNED) AS day_key FROM api_access_logs")
	if strings.Contains(strings.ToUpper(out2), "FROM_UNIXTIME") || strings.Contains(strings.ToUpper(out2), "DATE_FORMAT") {
		t.Fatalf("DATE_FORMAT/FROM_UNIXTIME 未替换: %s", out2)
	}
	if !strings.Contains(out2, "to_char(to_timestamp(create_time), 'YYYYMMDD')") {
		t.Fatalf("期望 to_char(to_timestamp(...), 'YYYYMMDD'): %s", out2)
	}
	if strings.Contains(strings.ToUpper(out2), "UNSIGNED") {
		t.Fatalf("UNSIGNED 应改为 BIGINT: %s", out2)
	}
}

// TestAdaptMySQLQueryToPostgres_DateFormatFromUnixOfUnixTimestamp 覆盖 sms_logs/email_logs
// 这种 created_at 本身是 TIMESTAMP 列、多套了一层 UNIX_TIMESTAMP 的写法。
func TestAdaptMySQLQueryToPostgres_DateFormatFromUnixOfUnixTimestamp(t *testing.T) {
	out := AdaptMySQLQueryToPostgres(`SELECT CAST(DATE_FORMAT(FROM_UNIXTIME(UNIX_TIMESTAMP(created_at)), '%Y%m%d') AS UNSIGNED) AS day_key, COUNT(*) AS total_count FROM sms_logs GROUP BY day_key ORDER BY day_key ASC`)
	upper := strings.ToUpper(out)
	if strings.Contains(upper, "DATE_FORMAT") || strings.Contains(upper, "FROM_UNIXTIME") || strings.Contains(upper, "UNIX_TIMESTAMP") {
		t.Fatalf("DATE_FORMAT/FROM_UNIXTIME/UNIX_TIMESTAMP 未替换: %s", out)
	}
	if !strings.Contains(out, "to_char(created_at, 'YYYYMMDD')") {
		t.Fatalf("期望直接对 created_at to_char（不套 to_timestamp）: %s", out)
	}
}

func TestAdaptMySQLQueryToPostgres_UnixTimestamp(t *testing.T) {
	out := AdaptMySQLQueryToPostgres("SELECT UNIX_TIMESTAMP(created_at) FROM t")
	if strings.Contains(strings.ToUpper(out), "UNIX_TIMESTAMP") {
		t.Fatalf("UNIX_TIMESTAMP 未替换: %s", out)
	}
	if !strings.Contains(out, "EXTRACT(EPOCH FROM (created_at))") {
		t.Fatalf("期望 EXTRACT(EPOCH FROM (...)): %s", out)
	}
}

func TestAdaptMySQLQueryToPostgres_OnDuplicateKey(t *testing.T) {
	in := `INSERT INTO api_access_log_daily_stats (day_key, total_count, updated_at)
    VALUES (?, 1, ?)
    ON DUPLICATE KEY UPDATE
      total_count = total_count + 1,
      updated_at = ?`
	out := AdaptMySQLQueryToPostgres(in)
	if strings.Contains(strings.ToUpper(out), "ON DUPLICATE KEY") {
		t.Fatalf("ON DUPLICATE 未转换: %s", out)
	}
	if !strings.Contains(out, "ON CONFLICT(day_key) DO UPDATE SET") {
		t.Fatalf("期望 ON CONFLICT(day_key): %s", out)
	}
}

func TestAdaptMySQLQueryToPostgres_OnDuplicateValuesRef(t *testing.T) {
	in := `INSERT INTO t (k, v, read_at) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE read_at=VALUES(read_at)`
	out := AdaptMySQLQueryToPostgres(in)
	if !strings.Contains(out, "EXCLUDED.read_at") {
		t.Fatalf("期望 VALUES(read_at) 转成 EXCLUDED.read_at: %s", out)
	}
	if !strings.Contains(out, "VALUES (?, ?, ?)") {
		t.Fatalf("INSERT 占位符 VALUES 被误伤: %s", out)
	}
}

func TestAdaptMySQLQueryToPostgres_UUID(t *testing.T) {
	out := AdaptMySQLQueryToPostgres("UPDATE api_access_logs SET request_id = LOWER(UUID()) WHERE request_id IS NULL")
	// 注意：gen_random_uuid() 本身也含 "UUID()" 子串，这里要判断的是 MySQL 的 LOWER(UUID()) 写法已消失
	if strings.Contains(strings.ToUpper(out), "LOWER(UUID())") {
		t.Fatalf("UUID 未替换: %s", out)
	}
	if !strings.Contains(out, "gen_random_uuid()") {
		t.Fatalf("期望 gen_random_uuid(): %s", out)
	}
}

func TestRebindPostgresPlaceholders_Basic(t *testing.T) {
	out := RebindPostgresPlaceholders("UPDATE users SET money = ?, score = ? WHERE id = ?")
	want := "UPDATE users SET money = $1, score = $2 WHERE id = $3"
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
}

// TestRebindPostgresPlaceholders_SkipsQuestionMarkInStringLiteral 验证字符串字面量里的 ?
// 不会被误当占位符（比如硬编码的 like 模式、备注文案）。
func TestRebindPostgresPlaceholders_SkipsQuestionMarkInStringLiteral(t *testing.T) {
	in := "SELECT * FROM t WHERE memo = 'what?' AND id = ?"
	out := RebindPostgresPlaceholders(in)
	want := "SELECT * FROM t WHERE memo = 'what?' AND id = $1"
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
}

func TestRebindPostgresPlaceholders_NoPlaceholderNoop(t *testing.T) {
	in := "SELECT COUNT(*) FROM users"
	if out := RebindPostgresPlaceholders(in); out != in {
		t.Fatalf("无占位符时应原样返回: %s", out)
	}
}

func TestAdaptCreateTableToPostgres_Basic(t *testing.T) {
	mysql := `CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL COMMENT '用户名',
			status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态',
			UNIQUE KEY idx_users_username (username),
			INDEX idx_users_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

	stmts := AdaptMySQLDDLToPostgres(mysql)
	if len(stmts) < 3 {
		t.Fatalf("期望至少 1 条建表 + 2 条索引，实际 %d: %#v", len(stmts), stmts)
	}
	create := stmts[0]
	if !strings.Contains(create, "BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY") {
		t.Fatalf("主键未转换为 Postgres IDENTITY: %s", create)
	}
	upper := strings.ToUpper(create)
	if strings.Contains(upper, "ENGINE") || strings.Contains(upper, "COMMENT") || strings.Contains(upper, "UNSIGNED") {
		t.Fatalf("清洗不彻底: %s", create)
	}
	if !strings.Contains(upper, "SMALLINT") {
		t.Fatalf("TINYINT 应转成 SMALLINT: %s", create)
	}
	joined := strings.Join(stmts[1:], "\n")
	if !strings.Contains(joined, "CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username") {
		t.Fatalf("缺少唯一索引: %s", joined)
	}
	if !strings.Contains(joined, "CREATE INDEX IF NOT EXISTS idx_users_status") {
		t.Fatalf("缺少普通索引: %s", joined)
	}
}

func TestAdaptAddIndexToPostgres(t *testing.T) {
	mysql := "ALTER TABLE email_logs ADD INDEX idx_email_logs_created_at (created_at)"
	stmts := AdaptMySQLDDLToPostgres(mysql)
	if len(stmts) != 1 {
		t.Fatalf("期望 1 条，实际 %#v", stmts)
	}
	want := "CREATE INDEX IF NOT EXISTS idx_email_logs_created_at ON email_logs (created_at)"
	if stmts[0] != want {
		t.Fatalf("got %q want %q", stmts[0], want)
	}
}

func TestAdaptChangeColumnSkippedForPostgres(t *testing.T) {
	mysql := "ALTER TABLE verification_codes CHANGE COLUMN email contact VARCHAR(255) NOT NULL COMMENT 'x'"
	stmts := AdaptMySQLDDLToPostgres(mysql)
	if stmts != nil {
		t.Fatalf("CHANGE COLUMN 应跳过，实际 %#v", stmts)
	}
}

func TestAdaptAddColumnToPostgres(t *testing.T) {
	mysql := "ALTER TABLE users ADD COLUMN group_id BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '分组ID' AFTER id"
	stmts := AdaptMySQLDDLToPostgres(mysql)
	if len(stmts) != 1 {
		t.Fatalf("期望 1 条，实际 %#v", stmts)
	}
	s := stmts[0]
	upper := strings.ToUpper(s)
	if strings.Contains(upper, "AFTER") || strings.Contains(upper, "COMMENT") || strings.Contains(upper, "UNSIGNED") {
		t.Fatalf("清洗不彻底: %s", s)
	}
}
