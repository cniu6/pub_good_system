package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fst/backend/pkg/config"

	"github.com/glebarez/sqlite"
	mysqlDriver "gorm.io/driver/mysql"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局 GORM 连接（业务唯一入口）。
var DB *gorm.DB

// activeDriver 归一化后的驱动名：mysql / sqlite / postgres。
var activeDriver string

// InitDB 建立 GORM 连接。不执行表迁移（由 migrate.RunAutoMigrate 负责）。
func InitDB() {
	cfg := config.GlobalConfig
	if cfg == nil {
		log.Fatalf("[数据库配置错误] 数据库配置未初始化")
	}

	driver := normalizeDriver(cfg.DBDriver)
	activeDriver = driver

	// IgnoreRecordNotFoundError：First/Take 查无是业务常态（幂等首查、未实名、无设置行等），
	// 不能当 Warn 刷屏；真正该报警的是 SQL/连接错误，仍会按 Warn 打出。
	gormCfg := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
		NowFunc: func() time.Time {
			return time.Now()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	var (
		gdb *gorm.DB
		err error
	)

	switch driver {
	case "mysql":
		gdb, err = openWithRetry("MySQL", func() (*gorm.DB, error) {
			return gorm.Open(mysqlDriver.Open(cfg.DBDSN), gormCfg)
		})
		if err != nil {
			logMySQLConnectFailure(err)
			log.Fatalf("[数据库连接错误] 无法连接 MySQL，请检查服务与配置后重试")
		}
		log.Println("[DB] MySQL 连接已建立（GORM）")
	case "sqlite":
		dsn := strings.TrimSpace(cfg.DBDSN)
		if dsn == "" {
			log.Fatalf("[数据库配置错误] SQLite DSN 为空，请检查 DB_PATH / DB_NAME")
		}
		if err := ensureSQLiteDir(dsn); err != nil {
			log.Fatalf("[数据库配置错误] 无法创建 SQLite 数据目录: %v", err)
		}
		gdb, err = gorm.Open(sqlite.Open(sqliteOpenName(dsn)), gormCfg)
		if err != nil {
			log.Fatalf("[数据库连接错误] 无法打开 SQLite（DSN=%s）: %v", sanitizeSQLiteDSNForLog(dsn), err)
		}
		log.Printf("[DB] SQLite 连接已建立（GORM / glebarez；生产请改用 MySQL/Postgres）")
	case "postgres":
		gdb, err = openWithRetry("PostgreSQL", func() (*gorm.DB, error) {
			return gorm.Open(postgresDriver.Open(cfg.DBDSN), gormCfg)
		})
		if err != nil {
			log.Fatalf("[数据库连接错误] 无法连接 PostgreSQL: %v", err)
		}
		log.Println("[DB] PostgreSQL 连接已建立（GORM）")
	default:
		log.Fatalf("[数据库配置错误] 不支持的 DB_DRIVER=%q，请使用 mysql / sqlite / postgres", cfg.DBDriver)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		log.Fatalf("[数据库连接错误] 获取底层 *sql.DB 失败: %v", err)
	}

	if driver == "sqlite" {
		// SQLite 单写者，多连接只会放大部分锁竞争，固定单连接。
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	} else {
		maxOpen := cfg.DBMaxOpenConns
		if maxOpen <= 0 {
			maxOpen = 100
		}
		maxIdle := cfg.DBMaxIdleConns
		if maxIdle <= 0 {
			maxIdle = 10
		}
		if maxIdle > maxOpen {
			maxIdle = maxOpen
		}
		sqlDB.SetMaxOpenConns(maxOpen)
		sqlDB.SetMaxIdleConns(maxIdle)
		// MySQL 会主动断开空闲连接（日志里的 closing bad idle connection: EOF）。
		// 空闲 1 分钟就回收，配合 ConnMaxLifetime（应小于 MySQL wait_timeout）避免池里全是死连接。
		sqlDB.SetConnMaxIdleTime(time.Minute)
		if cfg.DBConnMaxLifetimeMinutes > 0 {
			sqlDB.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetimeMinutes) * time.Minute)
		}
	}

	// MySQL 统一字符集/排序规则：DSN 已带 collation 参数，
	// 这里只执行 SET NAMES 确保连接池中复用的连接也使用统一排序规则。
	// 注意：不在此处 Set("gorm:table_options", ...) —— Set 返回的实例 clone==0，
	// 会导致后续链式调用共享同一个 Statement，污染全局 DB。
	// table_options 在 RunAutoMigrate 时临时设置。
	if driver == "mysql" {
		applyMySQLSessionCollation(gdb)
	}

	DB = gdb
}

// openWithRetry 在 MySQL/PostgreSQL 暂时不可用时短暂重试：
// 数据库重启或网络抖动不该让进程启动失败，但重试后仍失败必须退出，
// 避免“进程活着、端口在听，但所有业务请求都报 500”。
func openWithRetry(name string, open func() (*gorm.DB, error)) (*gorm.DB, error) {
	const attempts = 3
	const delay = 2 * time.Second

	var gdb *gorm.DB
	var err error
	for i := 1; i <= attempts; i++ {
		gdb, err = open()
		if err == nil {
			return gdb, nil
		}
		if i < attempts {
			log.Printf("[DB] %s 连接失败（第 %d/%d 次），%s 后重试: %v", name, i, attempts, delay, err)
			time.Sleep(delay)
		}
	}
	return nil, err
}

// applyMySQLSessionCollation 对连接池中的连接执行 SET NAMES，统一会话排序规则。
// 在原始 gdb（clone==1）上调用 Exec，Exec 内部会创建新实例，不污染 gdb 本身。
func applyMySQLSessionCollation(gdb *gorm.DB) {
	cfg := config.GlobalConfig
	collation := config.DefaultMySQLCollation
	if cfg != nil && cfg.DBCollation != "" {
		collation = cfg.DBCollation
	}
	if err := gdb.Exec("SET NAMES utf8mb4 COLLATE " + collation).Error; err != nil {
		log.Printf("[DB] 设置 MySQL 会话排序规则失败: %v", err)
	}
}

func sqliteOpenName(dsn string) string {
	if p := extractSQLiteFilePath(dsn); p != "" {
		return p
	}
	return dsn
}

func normalizeDriver(raw string) string {
	d := strings.ToLower(strings.TrimSpace(raw))
	switch d {
	case "", "mysql":
		return "mysql"
	case "sqlite", "sqlite3":
		return "sqlite"
	case "postgres", "postgresql", "pg":
		return "postgres"
	default:
		return d
	}
}

func logMySQLConnectFailure(err error) {
	log.Printf("[数据库连接错误] 无法连接 MySQL: %v", err)
	log.Println("[临时缓解] 本地暂时没有 MySQL（或连不上）时，可改用 SQLite：")
	log.Println("  1) 根目录 .env：DB_DRIVER=sqlite")
	log.Println("  2) 可选：DB_PATH=./fst.db")
	log.Println("  3) 保存后重启后端")
}

func ensureSQLiteDir(dsn string) error {
	path := extractSQLiteFilePath(dsn)
	if path == "" {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func extractSQLiteFilePath(dsn string) string {
	s := strings.TrimSpace(dsn)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "file:") {
		s = strings.TrimPrefix(s, "file:")
		s = strings.TrimPrefix(s, "//")
		if i := strings.Index(s, "?"); i >= 0 {
			s = s[:i]
		}
	} else if i := strings.Index(s, "?"); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return filepath.FromSlash(s)
}

func sanitizeSQLiteDSNForLog(dsn string) string {
	if i := strings.Index(dsn, "?"); i >= 0 {
		return dsn[:i]
	}
	return dsn
}

// Close 关闭底层连接。
func Close() error {
	if DB == nil {
		return nil
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	DB = nil
	return sqlDB.Close()
}

// CheckTableExists 当前库是否存在指定表。
func CheckTableExists(tableName string) bool {
	if DB == nil {
		return false
	}
	return DB.Migrator().HasTable(tableName)
}

// CheckColumnExists 指定表是否存在指定列。
func CheckColumnExists(tableName, columnName string) bool {
	if DB == nil {
		return false
	}
	return DB.Migrator().HasColumn(tableName, columnName)
}

// CheckIndexExists 指定表是否存在指定索引。
func CheckIndexExists(tableName, indexName string) bool {
	if DB == nil {
		return false
	}
	return DB.Migrator().HasIndex(tableName, indexName)
}

// GetDB 返回全局 GORM 连接。
func GetDB() *gorm.DB {
	return DB
}

// SQLDB 返回底层 database/sql 连接（Ping/Stats 等用）。
func SQLDB() (*sql.DB, error) {
	if DB == nil {
		return nil, fmt.Errorf("Database not initialized")
	}
	return DB.DB()
}

// DriverName 返回当前归一化驱动名。
func DriverName() string {
	return activeDriver
}

// IsSQLite 当前是否 SQLite。
func IsSQLite() bool {
	return strings.EqualFold(activeDriver, "sqlite")
}

// IsMySQL 当前是否 MySQL。
func IsMySQL() bool {
	return strings.EqualFold(activeDriver, "mysql")
}

// IsPostgres 当前是否 PostgreSQL。
func IsPostgres() bool {
	return strings.EqualFold(activeDriver, "postgres")
}
