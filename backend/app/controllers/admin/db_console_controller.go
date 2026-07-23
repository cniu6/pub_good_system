package admin

import (
	"context"
	"database/sql"
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

// 只读语句前缀
var readOnlyPrefixes = []string{"SELECT", "SHOW", "EXPLAIN", "PRAGMA", "WITH"}

// 写模式下仍禁止的危险关键字（整词匹配）
var dangerousSQLKeywords = []string{
	"DROP DATABASE", "DROP SCHEMA", "ATTACH", "DETACH",
	"LOAD_EXTENSION", "VACUUM INTO", "COPY PROGRAM",
}

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

func adminAuditUser(c *gin.Context) (uint64, string) {
	uid, _ := utils.GetUserID(c)
	var name string
	if v, ok := c.Get("username"); ok {
		if s, ok2 := v.(string); ok2 {
			name = s
		}
	}
	return uid, name
}

func writeDBConsoleAudit(c *gin.Context, action, reqBody, respBody string, status int) {
	uid, name := adminAuditUser(c)
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
func (ctrl *DBConsoleController) Info(c *gin.Context) {
	driver := db.DriverName()
	utils.Success(c, gin.H{
		"driver":          driver,
		"backup_supported": db.IsSQLite(),
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
func (ctrl *DBConsoleController) Tables(c *gin.Context) {
	names, err := listTableNames()
	if err != nil {
		log.Printf("[ADMIN][DB] list tables failed: %v", err)
		utils.Fail(c, 500, "获取表列表失败")
		return
	}
	utils.Success(c, gin.H{"list": names})
}

// TableRows GET /db/tables/:name/rows
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
	utils.Success(c, gin.H{
		"table":     name,
		"columns":   cols,
		"rows":      data,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
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
	SQL       string `json:"sql"`
	AllowWrite bool  `json:"allow_write"`
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

func containsDangerousSQL(sqlText string) bool {
	upper := strings.ToUpper(sqlText)
	for _, kw := range dangerousSQLKeywords {
		if strings.Contains(upper, kw) {
			return true
		}
	}
	return false
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// Backup GET /db/backup — 仅 SQLite 文件备份
func (ctrl *DBConsoleController) Backup(c *gin.Context) {
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
		g.POST("/sql", ctrl.execSQLEntry)
		g.GET("/backup", ctrl.Backup)
	}
}

// execSQLEntry 预读 body 后执行 SQL（写操作仍由 allow_write + handler 内校验）
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
		if containsDangerousSQL(sqlText) {
			writeDBConsoleAudit(c, "sql_blocked", sqlText, "dangerous keyword", 400)
			utils.Fail(c, 400, "禁止执行危险语句")
			return
		}
	} else {
		if !isReadOnlySQL(sqlText) {
			writeDBConsoleAudit(c, "sql_blocked", sqlText, "read-only mode", 400)
			utils.Fail(c, 400, "只读模式仅允许 SELECT/SHOW/EXPLAIN/PRAGMA/WITH")
			return
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	start := time.Now()
	rows, err := db.DB.WithContext(ctx).Raw(sqlText).Rows()
	if err != nil {
		if req.AllowWrite {
			res := db.DB.WithContext(ctx).Exec(sqlText)
			if res.Error != nil {
				writeDBConsoleAudit(c, "sql_error", sqlText, res.Error.Error(), 400)
				utils.Fail(c, 400, "执行失败: "+res.Error.Error())
				return
			}
			writeDBConsoleAudit(c, "sql_exec", sqlText, fmt.Sprintf("rows_affected=%d", res.RowsAffected), 200)
			utils.Success(c, gin.H{
				"columns":       []string{},
				"rows":          []interface{}{},
				"rows_affected": res.RowsAffected,
				"duration_ms":   time.Since(start).Milliseconds(),
			})
			return
		}
		writeDBConsoleAudit(c, "sql_error", sqlText, err.Error(), 400)
		utils.Fail(c, 400, "执行失败: "+err.Error())
		return
	}
	defer rows.Close()

	cols, data, scanErr := scanRowsLimited(rows, 200)
	if scanErr != nil {
		writeDBConsoleAudit(c, "sql_error", sqlText, scanErr.Error(), 500)
		utils.Fail(c, 500, "读取结果失败")
		return
	}
	writeDBConsoleAudit(c, "sql_query", truncateStr(sqlText, 2000), fmt.Sprintf("rows=%d", len(data)), 200)
	utils.Success(c, gin.H{
		"columns":     cols,
		"rows":        data,
		"row_count":   len(data),
		"truncated":  len(data) >= 200,
		"duration_ms": time.Since(start).Milliseconds(),
	})
}
