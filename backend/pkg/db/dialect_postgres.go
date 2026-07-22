package db

import (
	"regexp"
	"strconv"
	"strings"
)

// ========================================
// PostgreSQL 方言适配（DML）
//
// 占位符 ?→$N 的转换不在这里做：那个转换必须对*所有*到达数据库连接的 SQL 生效
// （包括大量没有经过 Q() / Exec() 包装、直接 tx.Exec("... ?", args...) 的写法），
// 所以统一放在驱动层兜底，见 pg_shim.go 的 RebindPostgresPlaceholders。
// 本文件只负责 MySQL 专属语法/函数 → PostgreSQL 等价写法的转换，与 dialect.go 里
// SQLite 那一套是同一治理思路的另一份实现（Postgres 语义更接近 MySQL，改动量小很多）。
// ========================================

var (
	// PostgreSQL 不认识反引号标识符；项目里表名/列名全部小写下划线，去掉反引号后
	// 按裸标识符解析语义不变（Postgres 未加引号的标识符会被自动折成小写）。
	rePgBacktick = regexp.MustCompile("`")
	// LOCK IN SHARE MODE（MySQL）→ FOR SHARE（Postgres 等价写法）
	rePgLockShareMode = regexp.MustCompile(`(?i)\s+LOCK\s+IN\s+SHARE\s+MODE\b`)
	// FROM_UNIXTIME(col) → to_timestamp(col)（返回 timestamptz，可比较可格式化）
	rePgFromUnix = regexp.MustCompile(`(?i)FROM_UNIXTIME\s*\(\s*([^)]+?)\s*\)`)
	// DATE(FROM_UNIXTIME(col)) → 直接转 date
	rePgDateFromUnix = regexp.MustCompile(`(?i)DATE\s*\(\s*FROM_UNIXTIME\s*\(\s*([^)]+?)\s*\)\s*\)`)
	// DATE_FORMAT(FROM_UNIXTIME(col), '%Y%m%d') → to_char(to_timestamp(col),'YYYYMMDD') 再转整数
	rePgDateFormatFromUnix = regexp.MustCompile(`(?i)DATE_FORMAT\s*\(\s*FROM_UNIXTIME\s*\(\s*([^)]+?)\s*\)\s*,\s*'%Y%m%d'\s*\)`)
	// DATE_FORMAT(FROM_UNIXTIME(UNIX_TIMESTAMP(col)), '%Y%m%d')：col 本身已是时间列，
	// 里外两层转换互相抵消，直接对 col 格式化即可（与 dialect.go 对 SQLite 的处理同理）。
	rePgDateFormatFromUnixOfUnixTimestamp = regexp.MustCompile(`(?i)DATE_FORMAT\s*\(\s*FROM_UNIXTIME\s*\(\s*UNIX_TIMESTAMP\s*\(\s*([^)]+?)\s*\)\s*\)\s*,\s*'%Y%m%d'\s*\)`)
	// UNIX_TIMESTAMP(col) → EXTRACT(EPOCH FROM col)，转整数秒
	rePgUnixTimestamp = regexp.MustCompile(`(?i)UNIX_TIMESTAMP\s*\(\s*([^)]+?)\s*\)`)
	// CAST(x AS UNSIGNED) → Postgres 无 UNSIGNED，用 BIGINT 近似
	rePgCastUnsigned = regexp.MustCompile(`(?i)\bAS\s+UNSIGNED\b`)
	// DATE_SUB(NOW(), INTERVAL N DAY/HOUR/...) → NOW() - INTERVAL 'N day'
	rePgDateSubNow = regexp.MustCompile(`(?i)DATE_SUB\s*\(\s*NOW\s*\(\s*\)\s*,\s*INTERVAL\s+(\d+)\s+(DAY|HOUR|MINUTE|SECOND|DAYS|HOURS|MINUTES|SECONDS)\s*\)`)
)

// AdaptMySQLQueryToPostgres 把业务里常见的 MySQL DML 写法转成 PostgreSQL 可执行语句。
// 仅覆盖本项目已用到的写法，不是完整方言引擎；占位符转换见 pg_shim.go（驱动层兜底，
// 对所有 SQL 生效，不管是否调用过本函数)。
func AdaptMySQLQueryToPostgres(mysqlSQL string) string {
	s := strings.TrimSpace(mysqlSQL)
	if s == "" {
		return s
	}

	s = rePgBacktick.ReplaceAllString(s, "")

	// 双重嵌套的先替换，避免被拆开后正则因括号不配对漏转换（原理同 dialect.go 里 SQLite 那份）
	s = rePgDateFormatFromUnixOfUnixTimestamp.ReplaceAllString(s, "CAST(to_char($1, 'YYYYMMDD') AS BIGINT)")
	s = rePgDateFormatFromUnix.ReplaceAllString(s, "CAST(to_char(to_timestamp($1), 'YYYYMMDD') AS BIGINT)")
	s = rePgDateFromUnix.ReplaceAllString(s, "CAST(to_timestamp($1) AS DATE)")
	s = rePgFromUnix.ReplaceAllString(s, "to_timestamp($1)")
	s = rePgUnixTimestamp.ReplaceAllString(s, "CAST(EXTRACT(EPOCH FROM ($1)) AS BIGINT)")
	s = rePgCastUnsigned.ReplaceAllString(s, "AS BIGINT")
	// CHAR_LENGTH / CHARACTER_LENGTH：Postgres 原生支持，无需转换
	// NOW()：Postgres 原生支持，无需转换
	// FOR UPDATE：Postgres 原生支持，无需剥除（不像 SQLite）

	s = rePgDateSubNow.ReplaceAllStringFunc(s, func(m string) string {
		sub := rePgDateSubNow.FindStringSubmatch(m)
		if len(sub) < 3 {
			return m
		}
		n, unit := sub[1], strings.ToLower(sub[2])
		if !strings.HasSuffix(unit, "s") {
			unit += "s"
		}
		return "(NOW() - INTERVAL '" + n + " " + unit + "')"
	})

	s = rePgLockShareMode.ReplaceAllString(s, " FOR SHARE")

	if reOnDuplicate.MatchString(s) {
		s = adaptOnDuplicateKeyPostgres(s)
	}

	// LOWER(UUID()) / UUID()：Postgres 13+ 内置 gen_random_uuid()
	if strings.Contains(strings.ToUpper(s), "UUID()") {
		s = regexp.MustCompile(`(?i)LOWER\s*\(\s*UUID\s*\(\s*\)\s*\)`).ReplaceAllString(s, "lower(gen_random_uuid()::text)")
		s = regexp.MustCompile(`(?i)\bUUID\s*\(\s*\)`).ReplaceAllString(s, "gen_random_uuid()::text")
	}

	return strings.TrimSpace(s)
}

// adaptOnDuplicateKeyPostgres 将 ON DUPLICATE KEY UPDATE 转为 ON CONFLICT(首列) DO UPDATE SET。
// 冲突列的启发式规则与 SQLite 版共用「INSERT 列清单第一列即唯一键」的项目约定，见 dialect.go 注释。
func adaptOnDuplicateKeyPostgres(sql string) string {
	conflictCol := ""
	if m := reInsertCols.FindStringSubmatch(sql); len(m) >= 2 {
		cols := strings.Split(m[1], ",")
		if len(cols) > 0 {
			conflictCol = strings.TrimSpace(cols[0])
			conflictCol = strings.Trim(conflictCol, "`\"")
		}
	}
	if conflictCol == "" {
		return sql
	}
	out := reOnDuplicate.ReplaceAllString(sql, "ON CONFLICT("+conflictCol+") DO UPDATE SET")
	// MySQL upsert 里 SET 子句的 VALUES(col) 在 Postgres 应写作 EXCLUDED.col。
	out = reValuesRef.ReplaceAllString(out, "EXCLUDED.$1")
	return out
}

// ========================================
// PostgreSQL 方言适配（DDL）
// ========================================

var (
	rePgTinyIntSized = regexp.MustCompile(`(?i)\bTINYINT\s*\(\s*\d+\s*\)`)
	rePgTinyIntPlain = regexp.MustCompile(`(?i)\bTINYINT\b`)
	// BIGINT [UNSIGNED] AUTO_INCREMENT PRIMARY KEY → BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY
	// （Postgres 10+ 标准写法，比 BIGSERIAL 更贴近 ANSI，且不用操心历史遗留的序列命名冲突）
	rePgAutoIncPK = regexp.MustCompile(`(?i)\b([a-zA-Z_][a-zA-Z0-9_]*)\s+BIGINT(?:\s+UNSIGNED)?\s+AUTO_INCREMENT\s+PRIMARY\s+KEY`)
)

// cleanMySQLTypeNoiseForPostgres 去掉 SQLite 不认识、Postgres 也不认识的 MySQL 修饰，并做类型近似映射。
// 与 dialect.go 里的 cleanMySQLTypeNoise（SQLite 版）共用同一批「结构性」清洗规则
// （COMMENT / ENGINE / CHARSET / COLLATE / AFTER / ON UPDATE / UNSIGNED / MEDIUMTEXT / LONGTEXT /
// AUTO_INCREMENT 残留），只有「类型怎么落地」这一步各写各的：
//   - TINYINT → SMALLINT（Postgres 无 1 字节整型，SMALLINT 是最小兼容类型）
//   - BIGINT AUTO_INCREMENT PRIMARY KEY → BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY
func cleanMySQLTypeNoiseForPostgres(sql string) string {
	s := rePgBacktick.ReplaceAllString(sql, "")
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
	s = rePgTinyIntSized.ReplaceAllString(s, "SMALLINT")
	s = rePgTinyIntPlain.ReplaceAllString(s, "SMALLINT")
	s = rePgAutoIncPK.ReplaceAllString(s, "$1 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY")
	s = reAutoInc.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, ";")
	return s
}

// AdaptMySQLDDLToPostgres 把项目里常见的 MySQL DDL 转成 PostgreSQL 可执行语句列表。
// 结构解析（CREATE TABLE 拆列/拆内联索引、ALTER TABLE ADD INDEX 拆分）复用 dialect.go
// 里 dialect-agnostic 的 splitSQLDefs / parseNamedIndex / reCreateTable / reAddIndex，
// 只有「列定义怎么清洗」这一层走 Postgres 专属的 cleanMySQLTypeNoiseForPostgres。
// 说明：仅作已知写法的适配，不是完整方言引擎。
func AdaptMySQLDDLToPostgres(mysqlSQL string) []string {
	sql := strings.TrimSpace(mysqlSQL)
	if sql == "" {
		return nil
	}

	upper := strings.ToUpper(sql)
	// 历史库字段重命名（MySQL CHANGE COLUMN 语法）：Postgres 新库不会走到这条老迁移路径，跳过避免启动失败。
	// 如果未来真的需要在 Postgres 上跑这条历史迁移，请改写成 RENAME COLUMN + ALTER COLUMN TYPE 两步。
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
		return adaptCreateTableGeneric(sql, cleanMySQLTypeNoiseForPostgres)
	}

	cleaned := cleanMySQLTypeNoiseForPostgres(sql)
	cleaned = strings.TrimRight(strings.TrimSpace(cleaned), ";")
	if cleaned == "" {
		return nil
	}
	return []string{cleaned}
}

// RebindPostgresPlaceholders 将 SQL 里的 ? 占位符按出现顺序转换为 Postgres 的 $1,$2,...。
// 会跳过单引号字符串字面量里的 ?（比如硬编码的 JSON/时间格式片段），避免误伤。
// 放在这里（而不是 AdaptMySQLQueryToPostgres）是因为它必须对*所有*到达数据库连接的 SQL
// 生效，不管业务代码有没有调用 db.Q()；实际接入点在 pg_shim.go 的驱动层。
func RebindPostgresPlaceholders(sql string) string {
	if !strings.ContainsRune(sql, '?') {
		return sql
	}
	var b strings.Builder
	b.Grow(len(sql) + 8)
	inSingle := false
	n := 0
	for i := 0; i < len(sql); i++ {
		ch := sql[i]
		switch {
		case ch == '\'' && !inSingle:
			inSingle = true
			b.WriteByte(ch)
		case ch == '\'' && inSingle:
			if i+1 < len(sql) && sql[i+1] == '\'' {
				b.WriteByte(ch)
				b.WriteByte(sql[i+1])
				i++
				continue
			}
			inSingle = false
			b.WriteByte(ch)
		case inSingle:
			b.WriteByte(ch)
		case ch == '?':
			n++
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
		default:
			b.WriteByte(ch)
		}
	}
	return b.String()
}
