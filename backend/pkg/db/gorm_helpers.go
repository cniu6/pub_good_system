package db

import (
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ForUpdate 在事务查询上加行锁（MySQL/Postgres：FOR UPDATE；SQLite 无行锁语义，不加子句）。
func ForUpdate(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return tx
	}
	if IsSQLite() {
		return tx
	}
	return tx.Clauses(clause.Locking{Strength: "UPDATE"})
}

// MapGormNotFound 将 gorm.ErrRecordNotFound 转为 sql.ErrNoRows，兼容既有业务判断。
func MapGormNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return sql.ErrNoRows
	}
	return err
}

// WithTx 在事务中执行 fn（封装 GORM Transaction）。
func WithTx(fn func(tx *gorm.DB) error) error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}
	return DB.Transaction(fn)
}

// CastToText 把表达式转成可 LIKE / 字符串比较的类型。
// 注意：PostgreSQL 里 CAST(x AS CHAR) 等价 char(1) 会截断成 1 个字符，必须用 TEXT；
// MySQL 则用 CAST(x AS CHAR)（MySQL 的 CAST 不支持 TEXT 目标类型）。
func CastToText(expr string) string {
	if IsMySQL() {
		return "CAST(" + expr + " AS CHAR)"
	}
	return "CAST(" + expr + " AS TEXT)"
}

// QuoteIdent 按当前驱动给标识符加引号（DDL 用）。
func QuoteIdent(name string) string {
	if IsMySQL() {
		return "`" + name + "`"
	}
	return `"` + name + `"`
}
