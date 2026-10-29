package config

import (
	"strings"
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
)

func TestBuildDSNFromValues(t *testing.T) {
	t.Run("mysql", func(t *testing.T) {
		dsn := buildDSNFromValues(dbDSNValues{
			Driver:    "mysql",
			Host:      "mysql.internal",
			Port:      "3307",
			User:      "app_user",
			Password:  "app_pass",
			Name:      "app_db",
			Collation: "utf8mb4_general_ci",
		})
		parsed, err := mysqlDriver.ParseDSN(dsn)
		if err != nil {
			t.Fatalf("解析 MySQL DSN 失败: %v (%s)", err, dsn)
		}
		if parsed.Addr != "mysql.internal:3307" || parsed.User != "app_user" ||
			parsed.Passwd != "app_pass" || parsed.DBName != "app_db" {
			t.Fatalf("MySQL DSN 字段不匹配: %+v", parsed)
		}
		if parsed.Collation != "utf8mb4_general_ci" || !parsed.ParseTime {
			t.Fatalf("MySQL DSN 参数不匹配: collation=%q parseTime=%v", parsed.Collation, parsed.ParseTime)
		}
		if parsed.Loc == nil || parsed.Loc.String() != time.Local.String() {
			t.Fatalf("MySQL DSN loc=%v want %v", parsed.Loc, time.Local)
		}
	})

	t.Run("postgres", func(t *testing.T) {
		// 用实际使用的 pgx 驱动解析，校验解析后的值，而不是只比对拼出来的字符串。
		// 覆盖空格、@、/、:、+ 等在 URL userinfo 中容易被误编码的字符。
		dsn := buildDSNFromValues(dbDSNValues{
			Driver:   "postgres",
			Host:     "pg.internal",
			Port:     "6432",
			User:     "pg user",
			Password: "p@ss/wo:rd+ x",
			Name:     "app db",
			SSLMode:  "require",
		})
		parsed, err := pgx.ParseConfig(dsn)
		if err != nil {
			t.Fatalf("pgx 解析 Postgres DSN 失败: %v (%s)", err, dsn)
		}
		if parsed.User != "pg user" || parsed.Password != "p@ss/wo:rd+ x" {
			t.Fatalf("用户名/密码解析不一致: user=%q password=%q dsn=%s", parsed.User, parsed.Password, dsn)
		}
		if parsed.Host != "pg.internal" || parsed.Port != 6432 || parsed.Database != "app db" {
			t.Fatalf("地址/库名解析不一致: host=%q port=%d db=%q", parsed.Host, parsed.Port, parsed.Database)
		}
	})

	t.Run("sqlite", func(t *testing.T) {
		dir := t.TempDir()
		dsn := buildDSNFromValues(dbDSNValues{
			Driver:     "sqlite",
			SQLitePath: dir + "/app.db",
		})
		if !strings.HasPrefix(dsn, "file:") || !strings.Contains(dsn, "busy_timeout(5000)") ||
			!strings.Contains(dsn, "journal_mode(WAL)") || !strings.Contains(dsn, "foreign_keys(1)") {
			t.Fatalf("SQLite DSN 不匹配: %s", dsn)
		}
	})

	t.Run("driver aliases", func(t *testing.T) {
		cases := map[string]string{
			"":           "mysql",
			"mysql":      "mysql",
			"sqlite3":    "sqlite",
			"postgresql": "postgres",
			"pg":         "postgres",
		}
		for input, want := range cases {
			if got := normalizeDBDriver(input); got != want {
				t.Fatalf("normalizeDBDriver(%q)=%q want %q", input, got, want)
			}
		}
	})
}

func TestBuildDSNPerDriverEnvPrecedence(t *testing.T) {
	t.Setenv("DB_DRIVER", "mysql")
	t.Setenv("DB_MYSQL_HOST", "mysql.internal")
	t.Setenv("DB_MYSQL_PORT", "3307")
	t.Setenv("DB_MYSQL_USER", "mysql_user")
	t.Setenv("DB_MYSQL_PASSWORD", "mysql_pass")
	t.Setenv("DB_MYSQL_NAME", "mysql_db")
	t.Setenv("DB_MYSQL_COLLATION", "utf8mb4_general_ci")
	t.Setenv("DB_HOST", "legacy_host")
	t.Setenv("DB_PORT", "13306")
	t.Setenv("DB_USER", "legacy_user")
	t.Setenv("DB_PASSWORD", "legacy_pass")
	t.Setenv("DB_NAME", "legacy_db")

	parsed, err := mysqlDriver.ParseDSN(buildDSN())
	if err != nil {
		t.Fatalf("解析 MySQL DSN 失败: %v", err)
	}
	if parsed.Addr != "mysql.internal:3307" || parsed.User != "mysql_user" ||
		parsed.Passwd != "mysql_pass" || parsed.DBName != "mysql_db" {
		t.Fatalf("分驱动变量未优先: %+v", parsed)
	}
}

func TestBuildDSNLegacyFallback(t *testing.T) {
	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("DB_POSTGRES_HOST", "")
	t.Setenv("DB_POSTGRES_PORT", "")
	t.Setenv("DB_POSTGRES_USER", "")
	t.Setenv("DB_POSTGRES_PASSWORD", "")
	t.Setenv("DB_POSTGRES_NAME", "")
	t.Setenv("DB_POSTGRES_SSLMODE", "")
	t.Setenv("DB_HOST", "legacy-pg")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "legacy_user")
	t.Setenv("DB_PASSWORD", "legacy_pass")
	t.Setenv("DB_NAME", "legacy_db")
	t.Setenv("DB_SSLMODE", "require")

	want := "postgres://legacy_user:legacy_pass@legacy-pg:5432/legacy_db?connect_timeout=5&sslmode=require"
	if got := buildDSN(); got != want {
		t.Fatalf("旧变量回退 DSN=%q want %q", got, want)
	}
}
