package db

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// IsDuplicateKeyError 判断是否为唯一键冲突。
// 覆盖 MySQL 1062、SQLite UNIQUE、Postgres 23505。
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate entry") ||
		strings.Contains(msg, "unique constraint failed") ||
		strings.Contains(msg, "duplicate key value violates unique constraint") ||
		strings.Contains(msg, "sqlstate 23505") ||
		// pgx / libpq 常见包装：ERROR: duplicate key ... (SQLSTATE 23505)
		strings.Contains(msg, "(23505)")
}

// IsRetryableTransactionError 判断事务是否因临时锁冲突而可安全重试。
// 覆盖 MySQL 死锁/锁等待超时、Postgres 死锁/序列化冲突，以及 SQLite 的短暂锁占用。
func IsRetryableTransactionError(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1205 || mysqlErr.Number == 1213
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "deadlock found") ||
		strings.Contains(msg, "lock wait timeout") ||
		strings.Contains(msg, "sqlstate 40p01") ||
		strings.Contains(msg, "(40p01)") ||
		strings.Contains(msg, "sqlstate 40001") ||
		strings.Contains(msg, "(40001)") ||
		strings.Contains(msg, "database is locked") ||
		strings.Contains(msg, "database is busy")
}
