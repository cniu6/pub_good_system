package db

import (
	"regexp"
	"strings"
)

// IsSQLite 当前全局连接是否为 SQLite（含 sqlite3 别名归一化后）。
func IsSQLite() bool {
	return strings.EqualFold(activeDriver, "sqlite")
}

// IsMySQL 当前全局连接是否为 MySQL。
func IsMySQL() bool {
	return strings.EqualFold(activeDriver, "mysql")
}

// Q 按当前驱动适配 SQL：MySQL 原样返回；SQLite 下做常见 DML 方言转换。
// 业务里凡含 FOR UPDATE / NOW() / FROM_UNIXTIME / ON DUPLICATE KEY 等，请走本函数再执行。
func Q(query string) string {
	if !IsSQLite() {
		return query
	}
	return AdaptMySQLQueryToSQLite(query)
}

var (
	// FOR UPDATE / LOCK IN SHARE MODE：SQLite 无行锁语法，剥掉即可（单连接 + 事务串行）
	reForUpdate     = regexp.MustCompile(`(?i)\s+FOR\s+UPDATE\b`)
	reLockShareMode = regexp.MustCompile(`(?i)\s+LOCK\s+IN\s+SHARE\s+MODE\b`)
	// NOW() → CURRENT_TIMESTAMP（比较 datetime 列可用）
	reNowFunc = regexp.MustCompile(`(?i)\bNOW\s*\(\s*\)`)
	// DATE_SUB(NOW(), INTERVAL N DAY/HOUR/MINUTE/SECOND)
	reDateSubNow = regexp.MustCompile(`(?i)DATE_SUB\s*\(\s*NOW\s*\(\s*\)\s*,\s*INTERVAL\s+(\d+)\s+(DAY|HOUR|MINUTE|SECOND|DAYS|HOURS|MINUTES|SECONDS)\s*\)`)
	// DATE(FROM_UNIXTIME(col)) → date(col, 'unixepoch')
	reDateFromUnix = regexp.MustCompile(`(?i)DATE\s*\(\s*FROM_UNIXTIME\s*\(\s*([^)]+?)\s*\)\s*\)`)
	// FROM_UNIXTIME(col) → datetime(col, 'unixepoch')
	reFromUnix = regexp.MustCompile(`(?i)FROM_UNIXTIME\s*\(\s*([^)]+?)\s*\)`)
	// DATE_FORMAT(FROM_UNIXTIME(col), '%Y%m%d') 已先被 FROM_UNIXTIME 替换时需另处理；直接匹配整段更稳
	reDateFormatFromUnix = regexp.MustCompile(`(?i)DATE_FORMAT\s*\(\s*FROM_UNIXTIME\s*\(\s*([^)]+?)\s*\)\s*,\s*'%Y%m%d'\s*\)`)
	// UNIX_TIMESTAMP(col) → CAST(strftime('%s', col) AS INTEGER)
	reUnixTimestamp = regexp.MustCompile(`(?i)UNIX_TIMESTAMP\s*\(\s*([^)]+?)\s*\)`)
	// CAST(x AS UNSIGNED) → CAST(x AS INTEGER)
	reCastUnsigned = regexp.MustCompile(`(?i)\bAS\s+UNSIGNED\b`)
	// CHAR_LENGTH / CHARACTER_LENGTH → LENGTH（ASCII UUID 等场景两边语义一致）
	reCharLength = regexp.MustCompile(`(?i)\b(?:CHAR_LENGTH|CHARACTER_LENGTH)\s*\(`)
	// ON DUPLICATE KEY UPDATE → ON CONFLICT(first_col) DO UPDATE SET
	reOnDuplicate = regexp.MustCompile(`(?is)\bON\s+DUPLICATE\s+KEY\s+UPDATE\b`)
	reInsertCols  = regexp.MustCompile(`(?is)INSERT\s+INTO\s+[a-zA-Z0-9_` + "`" + `]+\s*\(\s*([a-zA-Z0-9_` + "`" + `,.\s]+?)\s*\)`)
)

// AdaptMySQLQueryToSQLite 把业务里常见的 MySQL DML 转成 SQLite 可执行语句。
// 仅覆盖本项目已用到的写法；不是完整方言引擎。本地临时库用。
func AdaptMySQLQueryToSQLite(mysqlSQL string) string {
	s := strings.TrimSpace(mysqlSQL)
	if s == "" {
		return s
	}

	// 先处理带 FROM_UNIXTIME 的 DATE_FORMAT，避免被拆开后对不上
	s = reDateFormatFromUnix.ReplaceAllString(s, "CAST(strftime('%Y%m%d', $1, 'unixepoch') AS INTEGER)")
	s = reDateFromUnix.ReplaceAllString(s, "date($1, 'unixepoch')")
	s = reFromUnix.ReplaceAllString(s, "datetime($1, 'unixepoch')")
	s = reUnixTimestamp.ReplaceAllString(s, "CAST(strftime('%s', $1) AS INTEGER)")
	s = reCastUnsigned.ReplaceAllString(s, "AS INTEGER")
	s = reCharLength.ReplaceAllString(s, "LENGTH(")

	s = reDateSubNow.ReplaceAllStringFunc(s, func(m string) string {
		sub := reDateSubNow.FindStringSubmatch(m)
		if len(sub) < 3 {
			return m
		}
		n, unit := sub[1], strings.ToLower(sub[2])
		unit = strings.TrimSuffix(unit, "s") // days → day
		return "datetime('now', '-" + n + " " + unit + "')"
	})
	s = reNowFunc.ReplaceAllString(s, "CURRENT_TIMESTAMP")

	s = reForUpdate.ReplaceAllString(s, "")
	s = reLockShareMode.ReplaceAllString(s, "")

	if reOnDuplicate.MatchString(s) {
		s = adaptOnDuplicateKey(s)
	}

	// LOWER(UUID())：SQLite 无 UUID()，迁移回填场景用随机占位意义不大，改为空串函数避免语法炸
	// 仅当整段是 UPDATE ... SET request_id = LOWER(UUID()) 这类时才替换
	if strings.Contains(strings.ToUpper(s), "UUID()") {
		s = regexp.MustCompile(`(?i)LOWER\s*\(\s*UUID\s*\(\s*\)\s*\)`).ReplaceAllString(s, "lower(hex(randomblob(16)))")
		s = regexp.MustCompile(`(?i)\bUUID\s*\(\s*\)`).ReplaceAllString(s, "lower(hex(randomblob(16)))")
	}

	return strings.TrimSpace(s)
}

// adaptOnDuplicateKey 将 ON DUPLICATE KEY UPDATE 转为 ON CONFLICT(首列) DO UPDATE SET。
// 本项目 upsert 约定：INSERT 列清单的第一列即唯一键（stat_key / day_key / route_path 等）。
func adaptOnDuplicateKey(sql string) string {
	conflictCol := ""
	if m := reInsertCols.FindStringSubmatch(sql); len(m) >= 2 {
		cols := strings.Split(m[1], ",")
		if len(cols) > 0 {
			conflictCol = strings.TrimSpace(cols[0])
			conflictCol = strings.Trim(conflictCol, "`\"")
		}
	}
	if conflictCol == "" {
		// 解析失败时无法安全转换；保持原样让错误暴露，便于排查
		return sql
	}
	return reOnDuplicate.ReplaceAllString(sql, "ON CONFLICT("+conflictCol+") DO UPDATE SET")
}

var (
	reColComment     = regexp.MustCompile(`(?i)\s+COMMENT\s+'([^']*)'`)
	reTableComment   = regexp.MustCompile(`(?i)\s+COMMENT\s*=\s*'([^']*)'`)
	reEngineClause   = regexp.MustCompile(`(?i)\s*ENGINE\s*=\s*\w+(\s+DEFAULT\s+CHARSET\s*=\s*\w+)?(\s+COLLATE\s*=\s*\w+)?`)
	reDefaultCharset = regexp.MustCompile(`(?i)\s*DEFAULT\s+CHARSET\s*=\s*\w+`)
	reCollate        = regexp.MustCompile(`(?i)\s*COLLATE\s*=\s*\w+`)
	reUnsigned       = regexp.MustCompile(`(?i)\s+UNSIGNED`)
	reAfterCol       = regexp.MustCompile(`(?i)\s+AFTER\s+[a-zA-Z0-9_]+`)
	reOnUpdateTS     = regexp.MustCompile(`(?i)\s+ON\s+UPDATE\s+CURRENT_TIMESTAMP`)
	reMediumText     = regexp.MustCompile(`(?i)\bMEDIUMTEXT\b`)
	reLongText       = regexp.MustCompile(`(?i)\bLONGTEXT\b`)
	reTinyIntSized   = regexp.MustCompile(`(?i)\bTINYINT\s*\(\s*\d+\s*\)`)
	reAutoIncPK      = regexp.MustCompile(`(?i)\b([a-zA-Z_][a-zA-Z0-9_]*)\s+BIGINT(?:\s+UNSIGNED)?\s+AUTO_INCREMENT\s+PRIMARY\s+KEY`)
	reAutoInc        = regexp.MustCompile(`(?i)\s+AUTO_INCREMENT\b`)
	reAddIndex       = regexp.MustCompile(`(?i)^\s*ALTER\s+TABLE\s+([a-zA-Z0-9_]+)\s+ADD\s+(UNIQUE\s+)?(?:INDEX|KEY)\s+([a-zA-Z0-9_]+)\s*\(([^)]+)\)\s*;?\s*$`)
	reCreateTable    = regexp.MustCompile(`(?is)^\s*CREATE\s+TABLE\s+(IF\s+NOT\s+EXISTS\s+)?([a-zA-Z0-9_]+)\s*\((.*)\)\s*(.*?)\s*;?\s*$`)
)

// isDDL 粗判是否为建表/改表/建索引类语句（仅 SQLite 适配时使用）。
func isDDL(query string) bool {
	q := strings.TrimSpace(strings.ToUpper(query))
	return strings.HasPrefix(q, "CREATE TABLE") ||
		strings.HasPrefix(q, "ALTER TABLE") ||
		strings.HasPrefix(q, "CREATE INDEX") ||
		strings.HasPrefix(q, "CREATE UNIQUE INDEX")
}

// AdaptMySQLDDLToSQLite 把项目里常见的 MySQL DDL 转成 SQLite 可执行语句列表。
// 说明：仅作本地临时缓解，不是完整方言兼容层。
func AdaptMySQLDDLToSQLite(mysqlSQL string) []string {
	sql := strings.TrimSpace(mysqlSQL)
	if sql == "" {
		return nil
	}

	upper := strings.ToUpper(sql)
	// 历史库字段重命名：SQLite 下新库不会走到，直接跳过避免启动失败
	if strings.Contains(upper, "CHANGE COLUMN") || strings.Contains(upper, "CHANGE `") {
		return nil
	}

	if m := reAddIndex.FindStringSubmatch(sql); m != nil {
		unique := ""
		if strings.TrimSpace(m[2]) != "" {
			unique = "UNIQUE "
		}
		table, indexName, cols := m[1], m[3], m[4]
		return []string{
			"CREATE " + unique + "INDEX IF NOT EXISTS " + indexName + " ON " + table + " (" + cols + ")",
		}
	}

	if reCreateTable.MatchString(sql) {
		return adaptCreateTable(sql)
	}

	// 其它 ALTER / CREATE INDEX：做通用清洗后单条执行
	cleaned := cleanMySQLTypeNoise(sql)
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.TrimRight(cleaned, ";")
	if cleaned == "" {
		return nil
	}
	return []string{cleaned}
}

func adaptCreateTable(mysqlSQL string) []string {
	m := reCreateTable.FindStringSubmatch(mysqlSQL)
	if m == nil {
		return []string{cleanMySQLTypeNoise(mysqlSQL)}
	}
	ifNotExists := m[1]
	tableName := m[2]
	body := m[3]

	parts := splitSQLDefs(body)
	colDefs := make([]string, 0, len(parts))
	indexStmts := make([]string, 0)

	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		up := strings.ToUpper(p)

		// 表内 UNIQUE KEY / UNIQUE INDEX name (cols)
		if strings.HasPrefix(up, "UNIQUE KEY ") || strings.HasPrefix(up, "UNIQUE INDEX ") {
			name, cols, ok := parseNamedIndex(p)
			if ok {
				indexStmts = append(indexStmts,
					"CREATE UNIQUE INDEX IF NOT EXISTS "+name+" ON "+tableName+" ("+cols+")")
			}
			continue
		}
		// 表内 INDEX / KEY name (cols)
		if strings.HasPrefix(up, "INDEX ") || strings.HasPrefix(up, "KEY ") {
			name, cols, ok := parseNamedIndex(p)
			if ok {
				indexStmts = append(indexStmts,
					"CREATE INDEX IF NOT EXISTS "+name+" ON "+tableName+" ("+cols+")")
			}
			continue
		}
		// PRIMARY KEY (cols) 单独行：保留
		if strings.HasPrefix(up, "PRIMARY KEY") {
			colDefs = append(colDefs, cleanMySQLTypeNoise(p))
			continue
		}

		colDefs = append(colDefs, cleanMySQLTypeNoise(p))
	}

	create := "CREATE TABLE " + ifNotExists + tableName + " (\n\t" + strings.Join(colDefs, ",\n\t") + "\n)"
	out := []string{create}
	out = append(out, indexStmts...)
	return out
}

// parseNamedIndex 解析 `INDEX idx (a,b)` / `UNIQUE KEY uk (a)` 等形式。
func parseNamedIndex(def string) (name, cols string, ok bool) {
	def = strings.TrimSpace(def)
	up := strings.ToUpper(def)
	switch {
	case strings.HasPrefix(up, "UNIQUE KEY "):
		def = strings.TrimSpace(def[len("UNIQUE KEY "):])
	case strings.HasPrefix(up, "UNIQUE INDEX "):
		def = strings.TrimSpace(def[len("UNIQUE INDEX "):])
	case strings.HasPrefix(up, "INDEX "):
		def = strings.TrimSpace(def[len("INDEX "):])
	case strings.HasPrefix(up, "KEY "):
		def = strings.TrimSpace(def[len("KEY "):])
	default:
		return "", "", false
	}
	lp := strings.Index(def, "(")
	rp := strings.LastIndex(def, ")")
	if lp <= 0 || rp <= lp {
		return "", "", false
	}
	name = strings.TrimSpace(def[:lp])
	cols = strings.TrimSpace(def[lp+1 : rp])
	if name == "" || cols == "" {
		return "", "", false
	}
	return name, cols, true
}

// splitSQLDefs 按顶层逗号拆分 CREATE TABLE 字段/索引定义。
func splitSQLDefs(body string) []string {
	var parts []string
	var b strings.Builder
	depth := 0
	inSingle := false
	for i := 0; i < len(body); i++ {
		ch := body[i]
		switch {
		case ch == '\'' && !inSingle:
			inSingle = true
			b.WriteByte(ch)
		case ch == '\'' && inSingle:
			// 简单处理 '' 转义
			if i+1 < len(body) && body[i+1] == '\'' {
				b.WriteByte(ch)
				b.WriteByte(body[i+1])
				i++
				continue
			}
			inSingle = false
			b.WriteByte(ch)
		case inSingle:
			b.WriteByte(ch)
		case ch == '(':
			depth++
			b.WriteByte(ch)
		case ch == ')':
			depth--
			b.WriteByte(ch)
		case ch == ',' && depth == 0:
			parts = append(parts, b.String())
			b.Reset()
		default:
			b.WriteByte(ch)
		}
	}
	if s := b.String(); strings.TrimSpace(s) != "" {
		parts = append(parts, s)
	}
	return parts
}

// cleanMySQLTypeNoise 去掉 SQLite 不认识的 MySQL 修饰，并做类型近似映射。
func cleanMySQLTypeNoise(sql string) string {
	s := sql
	s = reColComment.ReplaceAllString(s, "")
	s = reTableComment.ReplaceAllString(s, "")
	s = reEngineClause.ReplaceAllString(s, "")
	s = reDefaultCharset.ReplaceAllString(s, "")
	s = reCollate.ReplaceAllString(s, "")
	s = reAfterCol.ReplaceAllString(s, "")
	s = reOnUpdateTS.ReplaceAllString(s, "")
	s = reUnsigned.ReplaceAllString(s, "")
	s = reMediumText.ReplaceAllString(s, "TEXT")
	s = reLongText.ReplaceAllString(s, "TEXT")
	s = reTinyIntSized.ReplaceAllString(s, "INTEGER")
	// BIGINT AUTO_INCREMENT PRIMARY KEY → INTEGER PRIMARY KEY AUTOINCREMENT
	s = reAutoIncPK.ReplaceAllString(s, "$1 INTEGER PRIMARY KEY AUTOINCREMENT")
	// 残留 AUTO_INCREMENT（非 PK）直接去掉，避免语法错误
	s = reAutoInc.ReplaceAllString(s, "")
	// 收尾空白
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, ";")
	return s
}
