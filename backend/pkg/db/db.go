package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fst/backend/pkg/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// DB 全局 sqlx 连接。建表/自迁移请走 internal/migrate.RunAutoMigrate，不要在这里顺带 Migrate。
var DB *sqlx.DB

// activeDriver 归一化后的驱动名：mysql / sqlite（由 InitDB 写入）。
var activeDriver string

// InitDB 仅建立数据库连接与连接池，不执行表迁移。
// 支持 DB_DRIVER=mysql（默认）与 DB_DRIVER=sqlite|sqlite3；不会在 MySQL 失败时偷偷切库。
func InitDB() {
	cfg := config.GlobalConfig
	if cfg == nil {
		log.Fatalf("[数据库配置错误] 数据库配置未初始化")
	}

	driver := normalizeDriver(cfg.DBDriver)
	activeDriver = driver

	switch driver {
	case "mysql":
		initMySQL(cfg)
	case "sqlite":
		initSQLite(cfg)
	default:
		log.Fatalf("[数据库配置错误] 不支持的 DB_DRIVER=%q，请使用 mysql 或 sqlite", cfg.DBDriver)
	}
}

func normalizeDriver(raw string) string {
	d := strings.ToLower(strings.TrimSpace(raw))
	switch d {
	case "", "mysql":
		return "mysql"
	case "sqlite", "sqlite3":
		return "sqlite"
	default:
		return d
	}
}

func initMySQL(cfg *config.Config) {
	var err error
	DB, err = sqlx.Connect("mysql", cfg.DBDSN)
	if err != nil {
		logMySQLConnectFailure(err, cfg)
		log.Fatalf("[数据库连接错误] 无法连接 MySQL，请检查服务与配置后重试")
	}

	DB.SetMaxOpenConns(100)
	DB.SetMaxIdleConns(10)

	log.Println("[DB] MySQL 连接已建立")
}

func initSQLite(cfg *config.Config) {
	dsn := strings.TrimSpace(cfg.DBDSN)
	if dsn == "" {
		log.Fatalf("[数据库配置错误] SQLite DSN 为空，请检查 DB_PATH / DB_NAME")
	}

	// 从 file: 路径里取出文件，确保父目录存在（Windows 友好）
	if err := ensureSQLiteDir(dsn); err != nil {
		log.Fatalf("[数据库配置错误] 无法创建 SQLite 数据目录: %v", err)
	}

	var err error
	// modernc.org/sqlite 注册名为 sqlite（纯 Go，Windows 无需 CGO）
	DB, err = sqlx.Connect("sqlite", dsn)
	if err != nil {
		log.Fatalf("[数据库连接错误] 无法打开 SQLite（DSN=%s）: %v", sanitizeSQLiteDSNForLog(dsn), err)
	}

	// SQLite 写锁敏感：默认单连接更稳；已开 WAL 时可略放宽
	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)

	log.Printf("[DB] SQLite 连接已建立（临时/本地模式，数据文件见 DB_PATH；生产请改用 MySQL）")
}

// logMySQLConnectFailure 打印中文失败原因 + 可改用 SQLite 的明确配置指引（不自动切库）。
func logMySQLConnectFailure(err error, cfg *config.Config) {
	log.Printf("[数据库连接错误] 无法连接 MySQL: %v", err)
	log.Println("[临时缓解] 本地暂时没有 MySQL（或连不上）时，可改用 SQLite，步骤如下：")
	log.Println("  1) 在项目根目录 .env 设置：DB_DRIVER=sqlite")
	log.Println("  2) 可选：DB_PATH=./data/fst.db  （默认即为 data/fst.db，相对运行目录）")
	log.Println("  3) 保存后重启后端；SQLite 仅建议开发/临时使用，生产请恢复 MySQL")
	if cfg != nil && strings.TrimSpace(cfg.DBDSN) != "" {
		// 不打印密码：DSN 形如 user:pass@tcp(...)，只提示驱动与主机概念
		log.Println("  当前仍按 MySQL 连接；不会自动切换，以免静默改库")
	}
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

// extractSQLiteFilePath 从 modernc DSN（file:路径?参数 或裸路径）解析出文件系统路径。
func extractSQLiteFilePath(dsn string) string {
	s := strings.TrimSpace(dsn)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "file:") {
		s = strings.TrimPrefix(s, "file:")
		// file:///C:/x 或 file:C:/x 或 file:./data/x.db
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
	// URI 里是斜杠，转成本地路径
	return filepath.FromSlash(s)
}

func sanitizeSQLiteDSNForLog(dsn string) string {
	if i := strings.Index(dsn, "?"); i >= 0 {
		return dsn[:i]
	}
	return dsn
}

// Exec 执行 SQL。SQLite 下：DDL 做 MySQL→SQLite 适配；DML 走 Q() 做常见方言转换。
func Exec(query string, args ...interface{}) (sql.Result, error) {
	if DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	if IsSQLite() && isDDL(query) {
		stmts := AdaptMySQLDDLToSQLite(query)
		if len(stmts) == 0 {
			// 例如 CHANGE COLUMN：SQLite 临时模式直接跳过
			return nil, nil
		}
		var res sql.Result
		var err error
		for _, stmt := range stmts {
			// DDL 一般无占位参数；若有 args，仅第一条带上（兼容极少场景）
			if res == nil && len(args) > 0 {
				res, err = DB.Exec(stmt, args...)
			} else {
				res, err = DB.Exec(stmt)
			}
			if err != nil {
				return res, err
			}
		}
		return res, nil
	}
	return DB.Exec(Q(query), args...)
}

// CheckTableExists 当前库是否存在指定表。
func CheckTableExists(tableName string) bool {
	if DB == nil {
		return false
	}
	if IsSQLite() {
		var name string
		err := DB.Get(&name, `SELECT name FROM sqlite_master WHERE type='table' AND name=?`, tableName)
		return err == nil && name != ""
	}
	var count int
	query := `SELECT COUNT(*) FROM information_schema.tables 
			  WHERE table_schema = DATABASE() AND table_name = ?`
	err := DB.Get(&count, query, tableName)
	return err == nil && count > 0
}

// CheckColumnExists 指定表是否存在指定列。
func CheckColumnExists(tableName, columnName string) bool {
	if DB == nil {
		return false
	}
	if IsSQLite() {
		var count int
		// pragma_table_info 可作为表值函数查询
		err := DB.Get(&count, `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name=?`, tableName, columnName)
		return err == nil && count > 0
	}
	var count int
	query := `SELECT COUNT(*) FROM information_schema.columns 
			  WHERE table_schema = DATABASE() 
			  AND table_name = ? 
			  AND column_name = ?`
	err := DB.Get(&count, query, tableName, columnName)
	return err == nil && count > 0
}

// CheckIndexExists 指定表是否存在指定索引。
func CheckIndexExists(tableName, indexName string) bool {
	if DB == nil {
		return false
	}
	if IsSQLite() {
		var count int
		err := DB.Get(&count,
			`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=? AND tbl_name=?`,
			indexName, tableName)
		return err == nil && count > 0
	}
	var count int
	query := `SELECT COUNT(*) FROM information_schema.statistics 
			  WHERE table_schema = DATABASE() 
			  AND table_name = ? 
			  AND index_name = ?`
	err := DB.Get(&count, query, tableName, indexName)
	return err == nil && count > 0
}

// EnsureIndex 表存在且索引缺失时执行 ALTER/CREATE 补索引。
func EnsureIndex(tableName, indexName, alterSQL string) {
	if !CheckTableExists(tableName) || CheckIndexExists(tableName, indexName) {
		return
	}
	_, err := Exec(alterSQL)
	if err != nil {
		log.Printf("[DB] 补索引 '%s' on '%s' 失败: %v", indexName, tableName, err)
	} else {
		log.Printf("[DB] 已补索引 '%s' on '%s'", indexName, tableName)
	}
}

// GetDB 返回全局 sqlx 连接。
func GetDB() *sqlx.DB {
	return DB
}

// DriverName 返回当前归一化驱动名（mysql / sqlite），未初始化时为空。
func DriverName() string {
	return activeDriver
}
