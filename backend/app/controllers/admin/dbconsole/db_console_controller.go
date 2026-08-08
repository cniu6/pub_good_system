package dbconsole

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
	"fst/backend/utils"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
)

// DBConsoleController 管理端数据库控制台（表预览 / 只读 SQL / SQLite 备份）
type DBConsoleController struct{}

func NewDBConsoleController() *DBConsoleController {
	return &DBConsoleController{}
}

// 表名仅允许字母数字下划线，防止注入
var tableNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// 只读语句前缀。WITH 可携带 INSERT/UPDATE/DELETE，PRAGMA 也可能修改 SQLite 状态，
// 因此不能将它们视为只读语句。
var readOnlyPrefixes = []string{"SELECT", "SHOW", "EXPLAIN"}

// 写模式下仍禁止的危险关键字（整词匹配）
var dangerousSQLKeywords = []string{
	"DROP DATABASE", "DROP SCHEMA", "ATTACH", "DETACH",
	"LOAD_EXTENSION", "VACUUM INTO", "COPY PROGRAM",
}

// sensitiveColumnRe 用于禁止经控制台直接写 password/token 等敏感列（产品防护，非日志脱敏）。
var sensitiveColumnRe = regexp.MustCompile(`(?i)(password|passwd|secret|token|api[_ -]?key|access[_ -]?key|private[_ -]?key|credential)`)

const maxDBConsoleRowsAffected = 1000

// extractSQLitePathFromDSN 从 DBDSN 解析 SQLite 文件路径（与 pkg/db 逻辑对齐，不导出其未导出函数）
func extractSQLitePathFromDSN(dsn string) string {
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

func writeDBConsoleAudit(c *gin.Context, action, reqBody, respBody string, status int) {
	uid, name := utils.GetAdminAuditUser(c)
	// 操作日志明文落库（与全局操作日志策略一致，不做字段脱敏）
	rb, sb := reqBody, respBody
	_ = models.CreateOperationLog(&models.OperationLog{
		UserID:       uid,
		Username:     name,
		Module:       "数据库控制台",
		Action:       action,
		Method:       c.Request.Method,
		Path:         c.FullPath(),
		IP:           utils.GetClientIP(c),
		UserAgent:    c.Request.UserAgent(),
		HandlerName:  "admin.(*DBConsoleController)." + action,
		RequestBody:  &rb,
		ResponseBody: &sb,
		StatusCode:   status,
	})
}

// Info GET /db/info
// @Summary 数据库信息
// @Tags Admin-数据库
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/db/info [get]
func (ctrl *DBConsoleController) Info(c *gin.Context) {
	driver := db.DriverName()
	writeDBConsoleAudit(c, "info", "", driver, 200)
	utils.Success(c, gin.H{
		"driver":           driver,
		"backup_supported": db.IsSQLite() && !config.IsProductionMode(),
		"write_enabled":    config.IsAdminDBWriteEnabled(),
	})
}

// listTableNames 按驱动列出当前库表名
func listTableNames() ([]string, error) {
	if db.DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	var names []string
	switch {
	case db.IsSQLite():
		rows, err := db.DB.Raw(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`).Rows()
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err != nil {
				return nil, err
			}
			names = append(names, n)
		}
	case db.IsPostgres():
		rows, err := db.DB.Raw(`SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = current_schema() ORDER BY tablename`).Rows()
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err != nil {
				return nil, err
			}
			names = append(names, n)
		}
	default: // MySQL
		rows, err := db.DB.Raw(`SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() ORDER BY TABLE_NAME`).Rows()
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err != nil {
				return nil, err
			}
			names = append(names, n)
		}
	}
	if names == nil {
		names = []string{}
	}
	return names, nil
}

func tableExistsInList(name string, list []string) bool {
	for _, n := range list {
		if strings.EqualFold(n, name) {
			return true
		}
	}
	return false
}

// Tables GET /db/tables
// @Summary 数据库表列表
// @Tags Admin-数据库
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/db/tables [get]
func (ctrl *DBConsoleController) Tables(c *gin.Context) {
	names, err := listTableNames()
	if err != nil {
		log.Printf("[ADMIN][DB] list tables failed: %v", err)
		utils.Fail(c, 500, "获取表列表失败")
		return
	}
	writeDBConsoleAudit(c, "tables", "", fmt.Sprintf("count=%d", len(names)), 200)
	utils.Success(c, gin.H{"list": names})
}

// TableRows GET /db/tables/:name/rows
// @Summary 表数据预览
// @Tags Admin-数据库
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/db/tables/{name}/rows [get]
func (ctrl *DBConsoleController) TableRows(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	if !tableNameRe.MatchString(name) {
		utils.Fail(c, 400, "非法表名")
		return
	}
	names, err := listTableNames()
	if err != nil {
		utils.Fail(c, 500, "获取表列表失败")
		return
	}
	if !tableExistsInList(name, names) {
		utils.Fail(c, 404, "表不存在")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// 表名已校验，安全拼接
	quoted := quoteIdent(name)
	var total int64
	if err := db.DB.Raw("SELECT COUNT(*) FROM " + quoted).Scan(&total).Error; err != nil {
		log.Printf("[ADMIN][DB] count rows table=%s: %v", name, err)
		utils.Fail(c, 500, "查询失败")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	rows, err := db.DB.WithContext(ctx).Raw(fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", quoted), pageSize, offset).Rows()
	if err != nil {
		log.Printf("[ADMIN][DB] preview rows table=%s: %v", name, err)
		utils.Fail(c, 500, "查询失败")
		return
	}
	defer rows.Close()

	cols, data, err := scanRowsLimited(rows, 100)
	if err != nil {
		utils.Fail(c, 500, "读取结果失败")
		return
	}
	writeDBConsoleAudit(c, "table_rows", name, fmt.Sprintf("rows=%d", len(data)), 200)
	utils.Success(c, gin.H{
		"table":     name,
		"columns":   cols,
		"rows":      data,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type dbColumnMeta struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Nullable      bool   `json:"nullable"`
	DefaultValue  string `json:"default_value"`
	Comment       string `json:"comment"`
	PrimaryKey    bool   `json:"primary_key"`
	AutoIncrement bool   `json:"auto_increment"`
}

type dbIndexMeta struct {
	Name       string   `json:"name"`
	Columns    []string `json:"columns"`
	Unique     bool     `json:"unique"`
	PrimaryKey bool     `json:"primary_key"`
}

type dbForeignKeyMeta struct {
	Name       string   `json:"name"`
	Columns    []string `json:"columns"`
	RefTable   string   `json:"ref_table"`
	RefColumns []string `json:"ref_columns"`
}

type dbTableMeta struct {
	Table       string             `json:"table"`
	Comment     string             `json:"comment"`
	Columns     []dbColumnMeta     `json:"columns"`
	Indexes     []dbIndexMeta      `json:"indexes"`
	ForeignKeys []dbForeignKeyMeta `json:"foreign_keys"`
}

func validateDBConsoleTable(name string) error {
	if !tableNameRe.MatchString(name) {
		return fmt.Errorf("非法表名")
	}
	names, err := listTableNames()
	if err != nil {
		return err
	}
	if !tableExistsInList(name, names) {
		return fmt.Errorf("表不存在")
	}
	return nil
}

func buildTableMeta(name string) (*dbTableMeta, error) {
	if err := validateDBConsoleTable(name); err != nil {
		return nil, err
	}
	columnTypes, err := db.DB.Migrator().ColumnTypes(name)
	if err != nil {
		return nil, err
	}
	meta := &dbTableMeta{
		Table:       name,
		Columns:     make([]dbColumnMeta, 0, len(columnTypes)),
		Indexes:     []dbIndexMeta{},
		ForeignKeys: []dbForeignKeyMeta{},
	}
	for _, columnType := range columnTypes {
		column := dbColumnMeta{Name: columnType.Name(), Type: columnType.DatabaseTypeName()}
		if value, ok := columnType.ColumnType(); ok && value != "" {
			column.Type = value
		}
		column.Nullable, _ = columnType.Nullable()
		column.DefaultValue, _ = columnType.DefaultValue()
		column.Comment, _ = columnType.Comment()
		column.PrimaryKey, _ = columnType.PrimaryKey()
		column.AutoIncrement, _ = columnType.AutoIncrement()
		meta.Columns = append(meta.Columns, column)
	}
	indexes, err := db.DB.Migrator().GetIndexes(name)
	if err != nil {
		return nil, err
	}
	for _, index := range indexes {
		unique, _ := index.Unique()
		primaryKey, _ := index.PrimaryKey()
		meta.Indexes = append(meta.Indexes, dbIndexMeta{
			Name:       index.Name(),
			Columns:    index.Columns(),
			Unique:     unique,
			PrimaryKey: primaryKey,
		})
	}
	sort.Slice(meta.Indexes, func(i, j int) bool {
		return meta.Indexes[i].Name < meta.Indexes[j].Name
	})
	meta.Comment = lookupTableComment(name)
	foreignKeys, err := lookupForeignKeys(name)
	if err != nil {
		return nil, err
	}
	meta.ForeignKeys = foreignKeys
	return meta, nil
}

func lookupTableComment(name string) string {
	var comment sql.NullString
	var err error
	switch {
	case db.IsMySQL():
		err = db.DB.Raw(
			`SELECT TABLE_COMMENT FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`,
			name,
		).Scan(&comment).Error
	case db.IsPostgres():
		err = db.DB.Raw(
			`SELECT COALESCE(obj_description(c.oid, 'pg_class'), '') FROM pg_class c WHERE c.oid = ?::regclass`,
			name,
		).Scan(&comment).Error
	}
	if err != nil {
		log.Printf("[ADMIN][DB] lookup table comment table=%s: %v", name, err)
		return ""
	}
	return comment.String
}

func lookupForeignKeys(name string) ([]dbForeignKeyMeta, error) {
	var query string
	var args []interface{}
	switch {
	case db.IsSQLite():
		query = "PRAGMA foreign_key_list(" + quoteIdent(name) + ")"
	case db.IsPostgres():
		query = `
SELECT tc.constraint_name AS fk_name, kcu.column_name AS fk_column, ccu.table_name AS ref_table, ccu.column_name AS ref_column
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage ccu
  ON ccu.constraint_name = tc.constraint_name AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_schema = current_schema() AND tc.table_name = ?
ORDER BY tc.constraint_name, kcu.ordinal_position`
		args = []interface{}{name}
	default:
		query = `
SELECT kcu.CONSTRAINT_NAME AS fk_name, kcu.COLUMN_NAME AS fk_column,
       kcu.REFERENCED_TABLE_NAME AS ref_table, kcu.REFERENCED_COLUMN_NAME AS ref_column
FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
WHERE kcu.TABLE_SCHEMA = DATABASE() AND kcu.TABLE_NAME = ? AND kcu.REFERENCED_TABLE_NAME IS NOT NULL
ORDER BY kcu.CONSTRAINT_NAME, kcu.ORDINAL_POSITION`
		args = []interface{}{name}
	}
	rows, err := db.DB.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, data, err := scanRowsLimited(rows, 200)
	if err != nil {
		return nil, err
	}
	_ = columns
	grouped := make(map[string]*dbForeignKeyMeta)
	order := make([]string, 0)
	for _, row := range data {
		var key, column, refTable, refColumn string
		if db.IsSQLite() {
			key = fmt.Sprintf("fk_%v", row["id"])
			column = fmt.Sprint(row["from"])
			refTable = fmt.Sprint(row["table"])
			refColumn = fmt.Sprint(row["to"])
		} else {
			key = fmt.Sprint(row["fk_name"])
			column = fmt.Sprint(row["fk_column"])
			refTable = fmt.Sprint(row["ref_table"])
			refColumn = fmt.Sprint(row["ref_column"])
		}
		if grouped[key] == nil {
			grouped[key] = &dbForeignKeyMeta{Name: key, Columns: []string{}, RefTable: refTable, RefColumns: []string{}}
			order = append(order, key)
		}
		grouped[key].Columns = append(grouped[key].Columns, column)
		grouped[key].RefColumns = append(grouped[key].RefColumns, refColumn)
	}
	result := make([]dbForeignKeyMeta, 0, len(order))
	for _, key := range order {
		result = append(result, *grouped[key])
	}
	return result, nil
}

// TableMeta GET /db/tables/:name/meta
// @Summary 表结构元信息
// @Tags Admin-数据库
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/db/tables/{name}/meta [get]
func (ctrl *DBConsoleController) TableMeta(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	meta, err := buildTableMeta(name)
	if err != nil {
		if err.Error() == "表不存在" || err.Error() == "非法表名" {
			utils.Fail(c, 400, err.Error())
		} else {
			log.Printf("[ADMIN][DB] get meta table=%s: %v", name, err)
			utils.Fail(c, 500, "读取表结构失败")
		}
		return
	}
	writeDBConsoleAudit(c, "table_meta", name, "ok", 200)
	utils.Success(c, meta)
}

func buildDDL(name string, meta *dbTableMeta) (string, error) {
	switch {
	case db.IsSQLite():
		var ddl sql.NullString
		if err := db.DB.Raw(`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?`, name).Scan(&ddl).Error; err != nil {
			return "", err
		}
		return ddl.String, nil
	case db.IsMySQL():
		rows, err := db.DB.Raw("SHOW CREATE TABLE " + quoteIdent(name)).Rows()
		if err != nil {
			return "", err
		}
		defer rows.Close()
		_, data, err := scanRowsLimited(rows, 1)
		if err != nil || len(data) == 0 {
			return "", err
		}
		for _, value := range data[0] {
			if text, ok := value.(string); ok && strings.Contains(strings.ToUpper(text), "CREATE TABLE") {
				return text, nil
			}
		}
		return "", fmt.Errorf("未返回建表语句")
	default:
		lines := make([]string, 0, len(meta.Columns)+1)
		primaryColumns := make([]string, 0)
		for _, column := range meta.Columns {
			line := fmt.Sprintf("  %s %s", quoteIdent(column.Name), column.Type)
			if !column.Nullable {
				line += " NOT NULL"
			}
			if column.DefaultValue != "" {
				line += " DEFAULT " + column.DefaultValue
			}
			lines = append(lines, line)
			if column.PrimaryKey {
				primaryColumns = append(primaryColumns, quoteIdent(column.Name))
			}
		}
		if len(primaryColumns) > 0 {
			lines = append(lines, "  PRIMARY KEY ("+strings.Join(primaryColumns, ", ")+")")
		}
		ddl := "CREATE TABLE " + quoteIdent(name) + " (\n" + strings.Join(lines, ",\n") + "\n);"
		for _, index := range meta.Indexes {
			if index.PrimaryKey || len(index.Columns) == 0 {
				continue
			}
			columns := make([]string, 0, len(index.Columns))
			for _, column := range index.Columns {
				columns = append(columns, quoteIdent(column))
			}
			prefix := "CREATE INDEX "
			if index.Unique {
				prefix = "CREATE UNIQUE INDEX "
			}
			ddl += "\n" + prefix + quoteIdent(index.Name) + " ON " + quoteIdent(name) + " (" + strings.Join(columns, ", ") + ");"
		}
		return ddl, nil
	}
}

// TableDDL GET /db/tables/:name/ddl
// @Summary 表建表语句
// @Tags Admin-数据库
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/db/tables/{name}/ddl [get]
func (ctrl *DBConsoleController) TableDDL(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	meta, err := buildTableMeta(name)
	if err != nil {
		utils.Fail(c, 400, "读取表结构失败")
		return
	}
	ddl, err := buildDDL(name, meta)
	if err != nil {
		log.Printf("[ADMIN][DB] get ddl table=%s: %v", name, err)
		utils.Fail(c, 500, "生成 DDL 失败")
		return
	}
	writeDBConsoleAudit(c, "table_ddl", name, "ok", 200)
	utils.Success(c, gin.H{"table": name, "ddl": ddl})
}

func quoteIdent(name string) string {
	if db.IsPostgres() {
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	}
	// MySQL / SQLite 用反引号
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func scanRowsLimited(rows *sql.Rows, limit int) ([]string, []map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	out := make([]map[string]interface{}, 0)
	for rows.Next() && len(out) < limit {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		row := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			row[col] = normalizeSQLValue(vals[i])
		}
		out = append(out, row)
	}
	return cols, out, rows.Err()
}

func normalizeSQLValue(v interface{}) interface{} {
	switch t := v.(type) {
	case nil:
		return nil
	case []byte:
		// 尽量当 UTF-8 文本；不可打印则返回十六进制摘要
		s := string(t)
		if isPrintableUTF8(s) {
			return s
		}
		if len(t) > 64 {
			return fmt.Sprintf("0x%x…(%d bytes)", t[:32], len(t))
		}
		return fmt.Sprintf("0x%x", t)
	default:
		return t
	}
}

func isPrintableUTF8(s string) bool {
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

type execSQLRequest struct {
	SQL        string `json:"sql"`
	AllowWrite bool   `json:"allow_write"`
}

type dbRowCreateRequest struct {
	Values map[string]interface{} `json:"values"`
}

type dbRowUpdateRequest struct {
	PrimaryKey map[string]interface{} `json:"primary_key"`
	Values     map[string]interface{} `json:"values"`
}

type dbRowDeleteRequest struct {
	PrimaryKey map[string]interface{} `json:"primary_key"`
}

func requireDBConsoleWrite(c *gin.Context, action, requestBody string) bool {
	if config.IsAdminDBWriteEnabled() {
		return true
	}
	writeDBConsoleAudit(c, action+"_blocked", requestBody, "write disabled", 403)
	utils.Fail(c, 403, "当前环境的数据库控制台仅支持只读操作")
	return false
}

func validateRowValues(meta *dbTableMeta, values map[string]interface{}, allowPrimaryKey bool) (map[string]string, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("至少填写一个字段")
	}
	columns := make(map[string]dbColumnMeta, len(meta.Columns))
	for _, column := range meta.Columns {
		columns[strings.ToLower(column.Name)] = column
	}
	resolved := make(map[string]string, len(values))
	for name := range values {
		column, ok := columns[strings.ToLower(strings.TrimSpace(name))]
		if !ok {
			return nil, fmt.Errorf("字段 %s 不存在", name)
		}
		if sensitiveColumnRe.MatchString(column.Name) {
			return nil, fmt.Errorf("敏感字段 %s 不支持通过控制台写入", column.Name)
		}
		if column.PrimaryKey && !allowPrimaryKey {
			return nil, fmt.Errorf("主键字段不支持修改")
		}
		resolved[name] = column.Name
	}
	return resolved, nil
}

func resolvePrimaryKey(meta *dbTableMeta, values map[string]interface{}) ([]string, []interface{}, error) {
	primaryColumns := make([]string, 0)
	for _, column := range meta.Columns {
		if column.PrimaryKey {
			primaryColumns = append(primaryColumns, column.Name)
		}
	}
	if len(primaryColumns) == 0 {
		return nil, nil, fmt.Errorf("该表没有主键，不能执行行级修改")
	}
	valuesByLowerName := make(map[string]interface{}, len(values))
	for name, value := range values {
		valuesByLowerName[strings.ToLower(name)] = value
	}
	args := make([]interface{}, 0, len(primaryColumns))
	for _, column := range primaryColumns {
		value, ok := valuesByLowerName[strings.ToLower(column)]
		if !ok || value == nil {
			return nil, nil, fmt.Errorf("缺少主键字段 %s", column)
		}
		args = append(args, value)
	}
	return primaryColumns, args, nil
}

func executeDBConsoleWrite(ctx context.Context, query string, args ...interface{}) (int64, error) {
	tx := db.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer func() {
		_ = tx.Rollback().Error
	}()
	result := tx.Exec(query, args...)
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected > maxDBConsoleRowsAffected {
		return 0, fmt.Errorf("影响行数超过上限 %d，已回滚", maxDBConsoleRowsAffected)
	}
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

func marshalAuditPayload(value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

// CreateTableRow POST /db/tables/:name/rows
// @Summary 表数据新增
// @Tags Admin-数据库
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/db/tables/{name}/rows [post]
func (ctrl *DBConsoleController) CreateTableRow(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	var req dbRowCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	requestBody := marshalAuditPayload(req)
	if !requireDBConsoleWrite(c, "row_create", requestBody) {
		return
	}
	meta, err := buildTableMeta(name)
	if err != nil {
		utils.Fail(c, 400, "读取表结构失败")
		return
	}
	resolved, err := validateRowValues(meta, req.Values, true)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	inputNames := make([]string, 0, len(resolved))
	for inputName := range resolved {
		inputNames = append(inputNames, inputName)
	}
	sort.Strings(inputNames)
	columns := make([]string, 0, len(inputNames))
	placeholders := make([]string, 0, len(inputNames))
	args := make([]interface{}, 0, len(inputNames))
	for _, inputName := range inputNames {
		columns = append(columns, quoteIdent(resolved[inputName]))
		placeholders = append(placeholders, "?")
		args = append(args, req.Values[inputName])
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", quoteIdent(name), strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	affected, err := executeDBConsoleWrite(ctx, query, args...)
	if err != nil {
		writeDBConsoleAudit(c, "row_create_error", requestBody, err.Error(), 400)
		utils.Fail(c, 400, "新增失败: "+err.Error())
		return
	}
	writeDBConsoleAudit(c, "row_create", requestBody, fmt.Sprintf("rows_affected=%d", affected), 200)
	utils.Success(c, gin.H{"rows_affected": affected})
}

// UpdateTableRow PATCH /db/tables/:name/rows
// @Summary 表数据更新
// @Tags Admin-数据库
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/db/tables/{name}/rows [patch]
func (ctrl *DBConsoleController) UpdateTableRow(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	var req dbRowUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	requestBody := marshalAuditPayload(req)
	if !requireDBConsoleWrite(c, "row_update", requestBody) {
		return
	}
	meta, err := buildTableMeta(name)
	if err != nil {
		utils.Fail(c, 400, "读取表结构失败")
		return
	}
	resolved, err := validateRowValues(meta, req.Values, false)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	primaryColumns, primaryArgs, err := resolvePrimaryKey(meta, req.PrimaryKey)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	inputNames := make([]string, 0, len(resolved))
	for inputName := range resolved {
		inputNames = append(inputNames, inputName)
	}
	sort.Strings(inputNames)
	assignments := make([]string, 0, len(inputNames))
	args := make([]interface{}, 0, len(inputNames)+len(primaryArgs))
	for _, inputName := range inputNames {
		assignments = append(assignments, quoteIdent(resolved[inputName])+" = ?")
		args = append(args, req.Values[inputName])
	}
	where := make([]string, 0, len(primaryColumns))
	for _, column := range primaryColumns {
		where = append(where, quoteIdent(column)+" = ?")
	}
	args = append(args, primaryArgs...)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", quoteIdent(name), strings.Join(assignments, ", "), strings.Join(where, " AND "))
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	affected, err := executeDBConsoleWrite(ctx, query, args...)
	if err != nil {
		writeDBConsoleAudit(c, "row_update_error", requestBody, err.Error(), 400)
		utils.Fail(c, 400, "更新失败: "+err.Error())
		return
	}
	writeDBConsoleAudit(c, "row_update", requestBody, fmt.Sprintf("rows_affected=%d", affected), 200)
	utils.Success(c, gin.H{"rows_affected": affected})
}

// DeleteTableRow DELETE /db/tables/:name/rows
// @Summary 表数据删除
// @Tags Admin-数据库
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/db/tables/{name}/rows [delete]
func (ctrl *DBConsoleController) DeleteTableRow(c *gin.Context) {
	name := strings.TrimSpace(c.Param("name"))
	var req dbRowDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	requestBody := marshalAuditPayload(req)
	if !requireDBConsoleWrite(c, "row_delete", requestBody) {
		return
	}
	meta, err := buildTableMeta(name)
	if err != nil {
		utils.Fail(c, 400, "读取表结构失败")
		return
	}
	primaryColumns, args, err := resolvePrimaryKey(meta, req.PrimaryKey)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	where := make([]string, 0, len(primaryColumns))
	for _, column := range primaryColumns {
		where = append(where, quoteIdent(column)+" = ?")
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", quoteIdent(name), strings.Join(where, " AND "))
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	affected, err := executeDBConsoleWrite(ctx, query, args...)
	if err != nil {
		writeDBConsoleAudit(c, "row_delete_error", requestBody, err.Error(), 400)
		utils.Fail(c, 400, "删除失败: "+err.Error())
		return
	}
	writeDBConsoleAudit(c, "row_delete", requestBody, fmt.Sprintf("rows_affected=%d", affected), 200)
	utils.Success(c, gin.H{"rows_affected": affected})
}

func firstSQLKeyword(sqlText string) string {
	s := strings.TrimSpace(sqlText)
	// 去掉前导注释
	for {
		if strings.HasPrefix(s, "--") {
			if i := strings.Index(s, "\n"); i >= 0 {
				s = strings.TrimSpace(s[i+1:])
				continue
			}
			return ""
		}
		if strings.HasPrefix(s, "/*") {
			if i := strings.Index(s, "*/"); i >= 0 {
				s = strings.TrimSpace(s[i+2:])
				continue
			}
			return ""
		}
		break
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return strings.ToUpper(fields[0])
}

func isReadOnlySQL(sqlText string) bool {
	kw := firstSQLKeyword(sqlText)
	for _, p := range readOnlyPrefixes {
		if kw == p {
			return true
		}
	}
	return false
}

// queryReadOnlySQL 在数据库只读事务中执行查询，避免 SQL 方言特性绕过前缀校验后写库。
func queryReadOnlySQL(ctx context.Context, sqlText string) ([]string, []map[string]interface{}, error) {
	tx := db.DB.WithContext(ctx).Begin(&sql.TxOptions{ReadOnly: true})
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	defer func() {
		_ = tx.Rollback().Error
	}()

	rows, err := tx.Raw(sqlText).Rows()
	if err != nil {
		return nil, nil, err
	}
	cols, data, scanErr := scanRowsLimited(rows, 200)
	closeErr := rows.Close()
	if scanErr != nil {
		return nil, nil, scanErr
	}
	if closeErr != nil {
		return nil, nil, closeErr
	}
	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}
	return cols, data, nil
}

func containsDangerousSQL(sqlText string) bool {
	upper := strings.ToUpper(sqlText)
	for _, kw := range dangerousSQLKeywords {
		if strings.Contains(upper, kw) {
			return true
		}
	}
	return false
}

func isSingleSQLStatement(sqlText string) bool {
	trimmed := strings.TrimSpace(sqlText)
	trimmed = strings.TrimSuffix(trimmed, ";")
	return !strings.Contains(trimmed, ";")
}

func isAllowedWriteSQL(sqlText string) bool {
	switch firstSQLKeyword(sqlText) {
	case "INSERT":
		return true
	case "UPDATE", "DELETE":
		upper := " " + strings.ToUpper(sqlText) + " "
		return strings.Contains(upper, " WHERE ")
	default:
		return false
	}
}

// Backup GET /db/backup — 仅 SQLite 文件备份
// @Summary 下载 SQLite 数据库备份
// @Tags Admin-数据库
// @Security BearerAuth
// @Router /v1/admin/db/backup [get]
func (ctrl *DBConsoleController) Backup(c *gin.Context) {
	if config.IsProductionMode() {
		writeDBConsoleAudit(c, "backup_blocked", "", "production read-only", 403)
		utils.Fail(c, 403, "生产环境禁止下载数据库备份")
		return
	}
	if !db.IsSQLite() {
		utils.Fail(c, 400, "暂仅支持 SQLite 文件备份")
		return
	}
	cfg := config.GlobalConfig
	if cfg == nil {
		utils.Fail(c, 500, "配置未就绪")
		return
	}
	path := extractSQLitePathFromDSN(cfg.DBDSN)
	if path == "" {
		utils.Fail(c, 400, "无法解析 SQLite 文件路径")
		return
	}
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() {
		utils.Fail(c, 404, "数据库文件不存在")
		return
	}
	writeDBConsoleAudit(c, "backup", path, "ok", 200)
	filename := filepath.Base(path)
	if filename == "" || filename == "." {
		filename = "fst.db"
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", "application/octet-stream")
	http.ServeFile(c.Writer, c.Request, path)
}

// RegisterRoutes 注册数据库控制台路由（管理端 AdminOnly 已保护）
func (ctrl *DBConsoleController) RegisterRoutes(adminGroup *gin.RouterGroup) {
	g := adminGroup.Group("/db")
	{
		g.GET("/info", ctrl.Info)
		g.GET("/tables", ctrl.Tables)
		g.GET("/tables/:name/rows", ctrl.TableRows)
		g.GET("/tables/:name/meta", ctrl.TableMeta)
		g.GET("/tables/:name/ddl", ctrl.TableDDL)
		g.POST("/tables/:name/rows", ctrl.CreateTableRow)
		g.PATCH("/tables/:name/rows", ctrl.UpdateTableRow)
		g.DELETE("/tables/:name/rows", ctrl.DeleteTableRow)
		g.POST("/sql", ctrl.execSQLEntry)
		g.GET("/backup", ctrl.Backup)
	}
}

// Query 执行 SQL（查询/写操作）
// @Summary 数据库 SQL 执行
// @Tags Admin-数据库
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/db/sql [post]
func (ctrl *DBConsoleController) execSQLEntry(c *gin.Context) {
	var req execSQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	c.Set("_db_sql_req", req)
	ctrl.execSQLFromCtx(c)
}

func (ctrl *DBConsoleController) execSQLFromCtx(c *gin.Context) {
	v, ok := c.Get("_db_sql_req")
	if !ok {
		utils.Fail(c, 400, "参数错误")
		return
	}
	req := v.(execSQLRequest)
	// 复用 ExecSQL 逻辑：临时把 body 已解析结果走内部方法
	ctrl.execSQLParsed(c, req)
}

func (ctrl *DBConsoleController) execSQLParsed(c *gin.Context, req execSQLRequest) {
	sqlText := strings.TrimSpace(req.SQL)
	if sqlText == "" {
		utils.Fail(c, 400, "SQL 不能为空")
		return
	}
	if len(sqlText) > 20000 {
		utils.Fail(c, 400, "SQL 过长")
		return
	}

	if req.AllowWrite {
		if !requireDBConsoleWrite(c, "sql_exec", sqlText) {
			return
		}
		if !isSingleSQLStatement(sqlText) {
			writeDBConsoleAudit(c, "sql_blocked", sqlText, "multiple statements", 400)
			utils.Fail(c, 400, "写操作仅允许单条 SQL 语句")
			return
		}
		if containsDangerousSQL(sqlText) {
			writeDBConsoleAudit(c, "sql_blocked", sqlText, "dangerous keyword", 400)
			utils.Fail(c, 400, "禁止执行危险语句")
			return
		}
		if !isAllowedWriteSQL(sqlText) {
			writeDBConsoleAudit(c, "sql_blocked", sqlText, "write statement not allowed", 400)
			utils.Fail(c, 400, "写操作仅允许带 WHERE 条件的 UPDATE/DELETE 或 INSERT")
			return
		}
	} else {
		if !isReadOnlySQL(sqlText) {
			writeDBConsoleAudit(c, "sql_blocked", sqlText, "read-only mode", 400)
			utils.Fail(c, 400, "只读模式仅允许 SELECT/SHOW/EXPLAIN")
			return
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	start := time.Now()
	if req.AllowWrite {
		affected, err := executeDBConsoleWrite(ctx, sqlText)
		if err != nil {
			// 原始 SQL 错误只进审计/服务端日志，避免把库内部细节回给客户端
			writeDBConsoleAudit(c, "sql_error", sqlText, err.Error(), 400)
			log.Printf("[DBConsole] sql write failed: %v", err)
			utils.Fail(c, 400, "执行失败")
			return
		}
		writeDBConsoleAudit(c, "sql_exec", sqlText, fmt.Sprintf("rows_affected=%d", affected), 200)
		utils.Success(c, gin.H{
			"columns":       []string{},
			"rows":          []interface{}{},
			"rows_affected": affected,
			"duration_ms":   time.Since(start).Milliseconds(),
		})
		return
	}

	cols, data, err := queryReadOnlySQL(ctx, sqlText)
	if err != nil {
		writeDBConsoleAudit(c, "sql_error", sqlText, err.Error(), 400)
		log.Printf("[DBConsole] sql query failed: %v", err)
		utils.Fail(c, 400, "执行失败")
		return
	}
	writeDBConsoleAudit(c, "sql_query", utils.TruncateString(sqlText, 2000), fmt.Sprintf("rows=%d", len(data)), 200)
	utils.Success(c, gin.H{
		"columns":     cols,
		"rows":        data,
		"row_count":   len(data),
		"truncated":   len(data) >= 200,
		"duration_ms": time.Since(start).Milliseconds(),
	})
}
