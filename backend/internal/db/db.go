package db

import (
	"fmt"
	"fst/backend/internal/db/database"
	"strings"

	"gorm.io/gorm"
)

// DB 简化的数据库操作工具（ThinkPHP 风格链式查询，草稿用）
type DB struct {
	db           *gorm.DB
	table        string
	where        []whereCondition
	order        string
	limit        int
	offset       int
	selectFields []string
}

// whereCondition WHERE 条件
type whereCondition struct {
	field    string
	operator string
	value    any
	logic    string // AND / OR
}

var globalDB *DB

// InitDB 初始化全局简化 DB 包装
func InitDB() {
	gormDB := database.InitDB()
	globalDB = NewDB(gormDB)
}

// GetDB 获取可链式调用的 DB 克隆实例
func GetDB() *DB {
	if globalDB == nil {
		InitDB()
	}
	return globalDB.Clone()
}

// GetGormDB 获取原生 GORM
func GetGormDB() *gorm.DB {
	return database.GetDB()
}

// Table 按表名创建查询
func Table(tableName string) *DB {
	return GetDB().Table(tableName)
}

// NewDB 创建包装实例
func NewDB(db *gorm.DB) *DB {
	return &DB{
		db:    db,
		where: make([]whereCondition, 0),
	}
}

// Table 设置表名
func (s *DB) Table(tableName string) *DB {
	s.table = tableName
	return s
}

// Where 添加 AND 条件
func (s *DB) Where(field string, operator string, value any) *DB {
	s.where = append(s.where, whereCondition{field: field, operator: operator, value: value, logic: "AND"})
	return s
}

// OrWhere 添加 OR 条件
func (s *DB) OrWhere(field string, operator string, value any) *DB {
	s.where = append(s.where, whereCondition{field: field, operator: operator, value: value, logic: "OR"})
	return s
}

// WhereIn IN 条件
func (s *DB) WhereIn(field string, values []any) *DB {
	s.where = append(s.where, whereCondition{field: field, operator: "IN", value: values, logic: "AND"})
	return s
}

// WhereLike LIKE 条件（自动加 %）
func (s *DB) WhereLike(field string, value string) *DB {
	s.where = append(s.where, whereCondition{field: field, operator: "LIKE", value: "%" + value + "%", logic: "AND"})
	return s
}

// OrderBy 设置排序（调用方需保证字段已白名单，防注入）
func (s *DB) OrderBy(order string) *DB {
	s.order = order
	return s
}

// Limit 限制条数
func (s *DB) Limit(limit int) *DB {
	s.limit = limit
	return s
}

// Offset 偏移
func (s *DB) Offset(offset int) *DB {
	s.offset = offset
	return s
}

// Page 分页（页码从 1 开始）
func (s *DB) Page(page, pageSize int) *DB {
	s.limit = pageSize
	s.offset = (page - 1) * pageSize
	return s
}

// Select 查询字段
func (s *DB) Select(fields ...string) *DB {
	s.selectFields = fields
	return s
}

func (s *DB) buildQuery(query *gorm.DB) *gorm.DB {
	if s.table != "" {
		query = query.Table(s.table)
	}
	if len(s.selectFields) > 0 {
		query = query.Select(s.selectFields)
	}

	for i, condition := range s.where {
		var whereClause string
		var args []any

		switch condition.operator {
		case "IN":
			if values, ok := condition.value.([]any); ok {
				placeholders := make([]string, len(values))
				for j := range values {
					placeholders[j] = "?"
				}
				whereClause = fmt.Sprintf("%s IN (%s)", condition.field, strings.Join(placeholders, ","))
				args = values
			}
		case "LIKE":
			whereClause = fmt.Sprintf("%s LIKE ?", condition.field)
			args = []any{condition.value}
		default:
			whereClause = fmt.Sprintf("%s %s ?", condition.field, condition.operator)
			args = []any{condition.value}
		}

		if i == 0 {
			query = query.Where(whereClause, args...)
		} else if condition.logic == "OR" {
			query = query.Or(whereClause, args...)
		} else {
			query = query.Where(whereClause, args...)
		}
	}

	if s.order != "" {
		query = query.Order(s.order)
	}
	if s.limit > 0 {
		query = query.Limit(s.limit)
	}
	if s.offset > 0 {
		query = query.Offset(s.offset)
	}
	return query
}

// Find 查询多条
func (s *DB) Find(dest any) error {
	return s.buildQuery(s.db).Find(dest).Error
}

// First 查询单条
func (s *DB) First(dest any) error {
	return s.buildQuery(s.db).First(dest).Error
}

// Count 统计
func (s *DB) Count() (int64, error) {
	var count int64
	err := s.buildQuery(s.db).Count(&count).Error
	return count, err
}

// Create 插入
func (s *DB) Create(value any) error {
	return s.db.Create(value).Error
}

// Update 按当前条件更新
func (s *DB) Update(updates any) error {
	return s.buildQuery(s.db).Updates(updates).Error
}

// Delete 软删除
func (s *DB) Delete(value any) error {
	return s.buildQuery(s.db).Delete(value).Error
}

// Exists 是否存在
func (s *DB) Exists() (bool, error) {
	count, err := s.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Reset 重置条件
func (s *DB) Reset() *DB {
	s.table = ""
	s.where = make([]whereCondition, 0)
	s.order = ""
	s.limit = 0
	s.offset = 0
	s.selectFields = nil
	return s
}

// Clone 克隆（深拷贝 where）
func (s *DB) Clone() *DB {
	newDB := &DB{
		db:           s.db,
		table:        s.table,
		where:        make([]whereCondition, len(s.where)),
		order:        s.order,
		limit:        s.limit,
		offset:       s.offset,
		selectFields: s.selectFields,
	}
	copy(newDB.where, s.where)
	return newDB
}

// Transaction 事务
func (s *DB) Transaction(fn func(*DB) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return fn(&DB{db: tx, where: make([]whereCondition, 0)})
	})
}

// Begin 开启事务
func (s *DB) Begin() *DB {
	tx := s.db.Begin()
	return &DB{db: tx, table: s.table, where: make([]whereCondition, len(s.where)), order: s.order, limit: s.limit, offset: s.offset}
}

// Commit 提交
func (s *DB) Commit() error {
	return s.db.Commit().Error
}

// Rollback 回滚
func (s *DB) Rollback() error {
	return s.db.Rollback().Error
}

// Pluck 抽取列
func (s *DB) Pluck(column string, dest any) error {
	return s.buildQuery(s.db).Pluck(column, dest).Error
}
