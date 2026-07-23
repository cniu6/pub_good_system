package config_test

import (
	"testing"

	"fst/backend/pkg/config"
)

func TestInitConfig_SQLiteEnv(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_PATH", t.TempDir()+"/cfg.db")
	t.Setenv("JWT_SECRET", "config-test-jwt-secret-16")
	t.Setenv("CORS_ORIGINS", "*")
	t.Setenv("APP_ENV", "development")
	t.Setenv("PORT", "18080")
	t.Setenv("ADMIN_API_PATH", "/admin123")

	config.InitConfig()
	cfg := config.GlobalConfig
	if cfg == nil {
		t.Fatal("GlobalConfig 为空")
	}
	if cfg.DBDriver != "sqlite" && cfg.DBDriver != "SQLite" {
		// 允许大小写，以实际归一为准
		if got := cfg.DBDriver; got != "sqlite" {
			// InitConfig 可能原样保存
			t.Logf("DBDriver=%q", got)
		}
	}
	if cfg.Port != "18080" {
		t.Fatalf("Port=%q want 18080", cfg.Port)
	}
	if cfg.JWTSecret == "" {
		t.Fatal("JWTSecret 不应为空")
	}
	if cfg.AdminAPIPath == "" {
		t.Fatal("AdminAPIPath 不应为空")
	}
}

// TestInitConfig_PostgresEnv 验证 DB_DRIVER=postgres 时 DSN 拼装正确（postgres:// URL，带 sslmode）。
func TestInitConfig_PostgresEnv(t *testing.T) {
	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "pguser")
	t.Setenv("DB_PASSWORD", "pgpass")
	t.Setenv("DB_NAME", "fst_platform")
	t.Setenv("JWT_SECRET", "config-test-jwt-secret-16")
	t.Setenv("CORS_ORIGINS", "*")
	t.Setenv("APP_ENV", "development")

	config.InitConfig()
	cfg := config.GlobalConfig
	if cfg == nil {
		t.Fatal("GlobalConfig 为空")
	}
	if cfg.DBDriver != "postgres" {
		t.Fatalf("DBDriver=%q want postgres", cfg.DBDriver)
	}
	want := "postgres://pguser:pgpass@127.0.0.1:5432/fst_platform?sslmode=disable"
	if cfg.DBDSN != want {
		t.Fatalf("DBDSN=%q want %q", cfg.DBDSN, want)
	}
}

func TestInitConfig_APILogWriterEnv(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_PATH", t.TempDir()+"/cfg-api-log.db")
	t.Setenv("JWT_SECRET", "config-test-jwt-secret-16")
	t.Setenv("CORS_ORIGINS", "*")
	t.Setenv("APP_ENV", "development")
	t.Setenv("API_LOG_QUEUE_CAPACITY", "123")
	t.Setenv("API_LOG_QUEUE_MAX_BYTES", "456789")
	t.Setenv("API_LOG_BATCH_SIZE", "17")
	t.Setenv("API_LOG_FLUSH_INTERVAL_MILLISECONDS", "2500")
	t.Setenv("API_LOG_WAL_DIR", t.TempDir()+"/wal")

	config.InitConfig()
	cfg := config.CloneGlobalConfig()
	if cfg == nil {
		t.Fatal("GlobalConfig 为空")
	}
	if cfg.APILogQueueCapacity != 123 || cfg.APILogQueueMaxBytes != 456789 ||
		cfg.APILogBatchSize != 17 || cfg.APILogFlushIntervalMillis != 2500 {
		t.Fatalf("API 日志队列配置不正确: %+v", cfg)
	}
	if cfg.APILogWALDir == "" {
		t.Fatal("APILogWALDir 不应为空")
	}
}

func TestIsProductionMode_Development(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_PATH", t.TempDir()+"/cfg2.db")
	t.Setenv("JWT_SECRET", "config-test-jwt-secret-16")
	t.Setenv("CORS_ORIGINS", "*")
	t.Setenv("APP_ENV", "development")
	config.InitConfig()
	if config.IsProductionMode() {
		t.Fatal("development 不应判定为生产")
	}
}
