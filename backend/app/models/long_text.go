package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// LongText 跨库长文本字段。
//
// MySQL 的 TEXT 上限只有 65,535 字节，日志/公告正文需要 MEDIUMTEXT；
// PostgreSQL 没有 MEDIUMTEXT，使用 TEXT；SQLite 接受任意类型名并归入 TEXT affinity。
// 通过 GormDBDataType 按驱动选择类型，避免 AutoMigrate 把已有 mediumtext 改成 text。
type LongText string

func (LongText) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	// SQLite 接受任意类型名并按 TEXT affinity 处理，继续用 mediumtext 可避免
	// 已有 SQLite 开发库被 AutoMigrate 重建；只有 PostgreSQL 必须用 text。
	if db != nil && db.Dialector != nil && db.Dialector.Name() == "postgres" {
		return "text"
	}
	return "mediumtext"
}

// NewLongTextPtr 把 string 转成 *LongText，供日志字段赋值。
func NewLongTextPtr(value string) *LongText {
	v := LongText(value)
	return &v
}

// LongTextFromPtr 把 *string 转成 *LongText；nil 保持 nil。
func LongTextFromPtr(value *string) *LongText {
	if value == nil {
		return nil
	}
	return NewLongTextPtr(*value)
}
