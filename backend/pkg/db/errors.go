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
