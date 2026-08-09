// normalize_collations 一键统一 MySQL 数据库/表/字符串列的字符集与排序规则。
// 默认目标：utf8mb4 + 根据 MySQL 版本自动选择最优 collation。
// 支持 --dry-run 预览、--apply 真正执行、--collation 手动指定。
// 只处理 MySQL；SQLite/Postgres 无需也不支持此操作。
package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var versionRe = regexp.MustCompile(`^(\d+)\.(\d+)`)

type options struct {
	Apply      bool
	DryRun     bool
	Collation  string
	Workers    int
	SkipDB     bool
	SkipTables []string
}

type tableInfo struct {
	TableName      string
	TableCharset   string
	TableCollation string
	NeedsConvert   bool
}

type columnInfo struct {
	TableName  string
	ColumnName string
	ColumnType string
	Charset    string
	Collation  string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintf(os.Stderr, "失败: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}

	config.InitConfig()
	db.InitDB()

	if !db.IsMySQL() {
		return errors.New("当前只支持 MySQL；SQLite/Postgres 不需要也不支持字符集排序规则迁移")
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("获取底层 sql.DB 失败: %w", err)
	}

	// 目标 collation：用户显式 > 自动按 MySQL 版本选择
	target, err := resolveTargetCollation(sqlDB, opts.Collation)
	if err != nil {
		return err
	}

	dbName, err := currentDatabase(sqlDB)
	if err != nil {
		return err
	}

	log.Printf("当前数据库: %s", dbName)
	log.Printf("目标字符集: utf8mb4, 目标排序规则: %s", target)

	// 查询数据库默认排序规则
	dbCharset, dbCollation, err := databaseCollation(sqlDB, dbName)
	if err != nil {
		return err
	}
	log.Printf("数据库当前: %s / %s", dbCharset, dbCollation)

	// 查询表信息
	tables, err := listTables(sqlDB, dbName)
	if err != nil {
		return err
	}

	skipSet := make(map[string]struct{})
	for _, t := range opts.SkipTables {
		skipSet[t] = struct{}{}
	}

	var needTables []*tableInfo
	var skipped []*tableInfo
	for _, t := range tables {
		if _, ok := skipSet[t.TableName]; ok {
			skipped = append(skipped, t)
			continue
		}
		if t.TableCharset != "utf8mb4" || t.TableCollation != target {
			t.NeedsConvert = true
			needTables = append(needTables, t)
		}
	}

	// 再查还有没有其他列的排序规则不一致（表级转换能处理大部分，这里用于显示/兜底）
	columns, err := listColumns(sqlDB, dbName, target)
	if err != nil {
		return err
	}

	log.Println("\n=== 执行计划 ===")
	if !opts.SkipDB && (dbCharset != "utf8mb4" || dbCollation != target) {
		log.Printf("[数据库] ALTER DATABASE %s CHARACTER SET utf8mb4 COLLATE %s;", dbName, target)
	} else if !opts.SkipDB {
		log.Printf("[数据库] 无需修改（已是 utf8mb4 / %s）", target)
	} else {
		log.Println("[数据库] 已跳过 (--skip-db)")
	}

	log.Printf("\n涉及表数量: %d", len(needTables))
	for _, t := range needTables {
		log.Printf("  [表] %s: %s / %s -> utf8mb4 / %s", t.TableName, t.TableCharset, t.TableCollation, target)
	}
	for _, t := range skipped {
		log.Printf("  [跳过] %s (用户显式排除)", t.TableName)
	}

	if len(columns) > 0 {
		log.Printf("\n列级异常数量: %d（表级 CONVERT TO 会自动修复）", len(columns))
		for _, c := range columns[:min(20, len(columns))] {
			log.Printf("  %s.%s: %s / %s", c.TableName, c.ColumnName, c.Charset, c.Collation)
		}
		if len(columns) > 20 {
			log.Printf("  ... 还有 %d 条", len(columns)-20)
		}
	}

	if !opts.Apply {
		log.Println("\n=== 本次为 --dry-run，未执行任何修改 ===")
		log.Println("确认无误后，追加 --apply 执行")
		return nil
	}

	// 执行阶段
	log.Println("\n=== 开始执行 ===")

	if !opts.SkipDB && (dbCharset != "utf8mb4" || dbCollation != target) {
		sql := fmt.Sprintf("ALTER DATABASE %s CHARACTER SET utf8mb4 COLLATE %s", quoteIdentifier(dbName), target)
		if err := execSQL(sqlDB, sql); err != nil {
			return fmt.Errorf("修改数据库默认排序规则失败: %w", err)
		}
		log.Printf("[数据库] 已设置为 utf8mb4 / %s", target)
	}

	workers := opts.Workers
	if workers <= 0 {
		workers = 1
	}
	if workers > len(needTables) {
		workers = len(needTables)
	}

	type job struct {
		idx int
		tbl *tableInfo
	}

	jobs := make(chan job, len(needTables))
	for i, t := range needTables {
		jobs <- job{idx: i + 1, tbl: t}
	}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(workers)
	start := time.Now()

	errorsCh := make(chan error, len(needTables))
	var successCount, failCount int
	var mu sync.Mutex

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for j := range jobs {
				t := j.tbl
				sql := fmt.Sprintf("ALTER TABLE %s CONVERT TO CHARACTER SET utf8mb4 COLLATE %s",
					quoteIdentifier(t.TableName), target)
				log.Printf("[%d/%d][worker-%d] 处理 %s ...", j.idx, len(needTables), workerID+1, t.TableName)
				if err := execSQL(sqlDB, sql); err != nil {
					log.Printf("[%d/%d] %s 失败: %v", j.idx, len(needTables), t.TableName, err)
					errorsCh <- fmt.Errorf("%s: %w", t.TableName, err)
					mu.Lock()
					failCount++
					mu.Unlock()
					continue
				}
				mu.Lock()
				successCount++
				mu.Unlock()
				log.Printf("[%d/%d] %s 完成", j.idx, len(needTables), t.TableName)
			}
		}(w)
	}

	wg.Wait()
	close(errorsCh)

	elapsed := time.Since(start)
	log.Printf("\n=== 执行结束（耗时 %s）===", elapsed)
	log.Printf("成功: %d / 失败: %d / 总计: %d", successCount, failCount, len(needTables))

	if failCount > 0 {
		log.Println("失败的表：")
		for err := range errorsCh {
			log.Printf("  - %s", err)
		}
		return errors.New("部分表迁移失败")
	}

	log.Println("全部表排序规则已统一，支持 emoji 等 4 字节字符")
	return nil
}

func parseOptions(args []string) (*options, error) {
	fs := flag.NewFlagSet("normalize_collations", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	opts := &options{}
	fs.BoolVar(&opts.Apply, "apply", false, "真正执行迁移（默认只打印计划）")
	fs.StringVar(&opts.Collation, "collation", "", "手动指定目标 collation，留空则按 MySQL 版本自动选择")
	fs.IntVar(&opts.Workers, "workers", 1, "并发处理表数量（慎用，大表会锁表）")
	fs.BoolVar(&opts.SkipDB, "skip-db", false, "跳过 ALTER DATABASE")
	var skipTables string
	fs.StringVar(&skipTables, "skip-tables", "", "跳过指定表，多个用逗号分隔")

	fs.Usage = func() {
		exe := filepath.Base(os.Args[0])
		fmt.Fprintf(fs.Output(), "用法:\n")
		fmt.Fprintf(fs.Output(), "  先预览：go run ./tools/normalize_collations\n")
		fmt.Fprintf(fs.Output(), "  再执行：go run ./tools/normalize_collations -apply\n")
		fmt.Fprintf(fs.Output(), "  指定规则：go run ./tools/normalize_collations -collation utf8mb4_0900_ai_ci -apply\n")
		fmt.Fprintf(fs.Output(), "  并发执行：go run ./tools/normalize_collations -apply -workers 4\n\n")
		fmt.Fprintf(fs.Output(), "  工具 %s [参数]\n\n", exe)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if fs.NArg() > 0 {
		return nil, fmt.Errorf("存在未识别参数: %s", strings.Join(fs.Args(), " "))
	}

	if skipTables != "" {
		for _, t := range strings.Split(skipTables, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				opts.SkipTables = append(opts.SkipTables, t)
			}
		}
	}

	return opts, nil
}

// resolveTargetCollation 选择最合适的统一排序规则。
// 优先级：--collation 显式值 > 项目配置 DBCollation > 按 MySQL 版本自动选择。
func resolveTargetCollation(sqlDB *sql.DB, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	if cfg := config.GetGlobalConfig(); cfg != nil && cfg.DBCollation != "" {
		return cfg.DBCollation, nil
	}

	var version string
	if err := sqlDB.QueryRow("SELECT VERSION()").Scan(&version); err != nil {
		return "", fmt.Errorf("查询 MySQL 版本失败: %w", err)
	}

	m := versionRe.FindStringSubmatch(version)
	if len(m) < 3 {
		return "", fmt.Errorf("无法解析 MySQL 版本: %s", version)
	}

	major := m[1]
	minor := m[2]

	switch {
	// MySQL 8.0+ 推荐基于 Unicode 9.0 的最新默认 collation
	case major == "8" || (major == "5" && minor == "7" && isMariaDBLike(version)):
		return "utf8mb4_0900_ai_ci", nil
	case major == "5" && minor == "7":
		return "utf8mb4_unicode_ci", nil
	default:
		// MariaDB 10.x/11.x 也支持 utf8mb4_unicode_ci，安全性最高
		return "utf8mb4_unicode_ci", nil
	}
}

// isMariaDBLike 简单判断 VERSION() 里是否有 MariaDB 字样。
func isMariaDBLike(v string) bool {
	return strings.Contains(strings.ToLower(v), "mariadb")
}

func currentDatabase(sqlDB *sql.DB) (string, error) {
	var name string
	if err := sqlDB.QueryRow("SELECT DATABASE()").Scan(&name); err != nil {
		return "", fmt.Errorf("查询当前数据库失败: %w", err)
	}
	if name == "" {
		return "", errors.New("当前没有选中的数据库，请检查 DSN 里是否包含 database 名称")
	}
	return name, nil
}

func databaseCollation(sqlDB *sql.DB, dbName string) (string, string, error) {
	var charset, collation string
	err := sqlDB.QueryRow(
		"SELECT DEFAULT_CHARACTER_SET_NAME, DEFAULT_COLLATION_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?",
		dbName,
	).Scan(&charset, &collation)
	if err != nil {
		return "", "", fmt.Errorf("查询数据库默认排序规则失败: %w", err)
	}
	return charset, collation, nil
}

func listTables(sqlDB *sql.DB, dbName string) ([]*tableInfo, error) {
	rows, err := sqlDB.Query(
		`SELECT TABLE_NAME, CCSA.CHARACTER_SET_NAME, CCSA.COLLATION_NAME
		 FROM information_schema.TABLES T
		 JOIN information_schema.COLLATION_CHARACTER_SET_APPLICABILITY CCSA
		   ON CCSA.COLLATION_NAME = T.TABLE_COLLATION
		 WHERE T.TABLE_SCHEMA = ? AND T.TABLE_TYPE = 'BASE TABLE'`,
		dbName,
	)
	if err != nil {
		return nil, fmt.Errorf("查询表信息失败: %w", err)
	}
	defer rows.Close()

	var tables []*tableInfo
	for rows.Next() {
		var t tableInfo
		if err := rows.Scan(&t.TableName, &t.TableCharset, &t.TableCollation); err != nil {
			return nil, err
		}
		tables = append(tables, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(tables, func(i, j int) bool { return tables[i].TableName < tables[j].TableName })
	return tables, nil
}

// listColumns 返回当前不是 utf8mb4 或目标排序规则的字符串列。
func listColumns(sqlDB *sql.DB, dbName, target string) ([]columnInfo, error) {
	rows, err := sqlDB.Query(
		`SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE, CHARACTER_SET_NAME, COLLATION_NAME
		 FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = ?
		   AND DATA_TYPE IN ('varchar', 'char', 'text', 'tinytext', 'mediumtext', 'longtext', 'enum', 'set')
		   AND (CHARACTER_SET_NAME != 'utf8mb4' OR COLLATION_NAME != ?)`,
		dbName, target,
	)
	if err != nil {
		return nil, fmt.Errorf("查询列信息失败: %w", err)
	}
	defer rows.Close()

	var cols []columnInfo
	for rows.Next() {
		var c columnInfo
		if err := rows.Scan(&c.TableName, &c.ColumnName, &c.ColumnType, &c.Charset, &c.Collation); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(cols, func(i, j int) bool {
		if cols[i].TableName != cols[j].TableName {
			return cols[i].TableName < cols[j].TableName
		}
		return cols[i].ColumnName < cols[j].ColumnName
	})
	return cols, nil
}

func execSQL(sqlDB *sql.DB, sql string) error {
	_, err := sqlDB.Exec(sql)
	return err
}

func quoteIdentifier(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "``") + "`"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
