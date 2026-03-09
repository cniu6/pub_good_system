package models

import (
	"database/sql"
	"fst/backend/internal/db"
	"log"
	"time"
)

// UserMoneyLog 会员余额变动表
type UserMoneyLog struct {
	ID         uint64  `db:"id" json:"id"`
	UserID     uint64  `db:"user_id" json:"user_id"`
	Money      float64 `db:"money" json:"money"`       // 变更金额（正=充值，负=扣款）
	Before     float64 `db:"before" json:"before"`      // 变更前余额
	After      float64 `db:"after" json:"after"`        // 变更后余额
	Memo       string  `db:"memo" json:"memo"`          // 备注
	CreateTime int64   `db:"create_time" json:"create_time"`
}

// InitUserMoneyLogsTable 初始化余额变动日志表
func InitUserMoneyLogsTable() {
	if db.CheckTableExists("user_money_logs") {
		db.EnsureIndex("user_money_logs", "idx_user_create_time", "ALTER TABLE user_money_logs ADD INDEX idx_user_create_time (user_id, create_time)")
		return
	}

	schema := `CREATE TABLE IF NOT EXISTS user_money_logs (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
		money DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '变更金额',
		` + "`before`" + ` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '变更前余额',
		` + "`after`" + ` DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '变更后余额',
		memo VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
		create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
		INDEX idx_user_id (user_id),
		INDEX idx_create_time (create_time),
		INDEX idx_user_create_time (user_id, create_time)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

	_, err := db.DB.Exec(schema)
	if err != nil {
		log.Printf("[Init] Failed to create user_money_logs table: %v", err)
	} else {
		log.Println("[Init] Created user_money_logs table")
	}
}

// CreateUserMoneyLog 创建余额变动记录
func CreateUserMoneyLog(userID uint64, money, before, after float64, memo string) (*UserMoneyLog, error) {
	now := time.Now().Unix()
	result, err := db.DB.Exec(
		"INSERT INTO user_money_logs (user_id, money, `before`, `after`, memo, create_time) VALUES (?, ?, ?, ?, ?, ?)",
		userID, money, before, after, memo, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &UserMoneyLog{
		ID:         uint64(id),
		UserID:     userID,
		Money:      money,
		Before:     before,
		After:      after,
		Memo:       memo,
		CreateTime: now,
	}, nil
}

// GetUserMoneyLogByID 获取指定ID的余额变动记录
func GetUserMoneyLogByID(id uint64) (*UserMoneyLog, error) {
	var logEntry UserMoneyLog
	err := db.DB.Get(&logEntry, "SELECT id, user_id, money, `before`, `after`, memo, create_time FROM user_money_logs WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &logEntry, nil
}

// GetUserMoneyLogList 获取余额变动列表（分页+搜索）
// 如果 onlyUserID > 0，则只返回该用户的记录（普通用户模式）
func GetUserMoneyLogList(onlyUserID uint64, page, pageSize int, keyword string) ([]UserMoneyLog, int64, error) {
	var logs []UserMoneyLog
	var total int64

	where := "WHERE 1=1"
	args := []interface{}{}

	if onlyUserID > 0 {
		where += " AND user_id = ?"
		args = append(args, onlyUserID)
	}
	if keyword != "" {
		where += " AND (memo LIKE ? OR CAST(money AS CHAR) LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}

	// 总数
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := db.DB.Get(&total, "SELECT COUNT(*) FROM user_money_logs "+where, countArgs...)
	if err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (page - 1) * pageSize
	query := "SELECT id, user_id, money, `before`, `after`, memo, create_time FROM user_money_logs " + where + " ORDER BY create_time DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	err = db.DB.Select(&logs, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// DeleteUserMoneyLog 删除余额变动记录
func DeleteUserMoneyLog(id uint64) error {
	_, err := db.DB.Exec("DELETE FROM user_money_logs WHERE id = ?", id)
	return err
}

// UpdateUserMoney 直接更新用户余额字段
func UpdateUserMoney(userID uint64, newMoney float64) error {
	now := time.Now().Unix()
	_, err := db.DB.Exec("UPDATE users SET money = ?, update_time = ? WHERE id = ?", newMoney, now, userID)
	return err
}

// UpdateUserMoneyTx 在事务中更新用户余额字段
func UpdateUserMoneyTx(tx *sql.Tx, userID uint64, newMoney float64) error {
	now := time.Now().Unix()
	_, err := tx.Exec("UPDATE users SET money = ?, update_time = ? WHERE id = ?", newMoney, now, userID)
	return err
}

// CreateUserMoneyLogTx 在事务中创建余额变动记录
func CreateUserMoneyLogTx(tx *sql.Tx, userID uint64, money, before, after float64, memo string) (*UserMoneyLog, error) {
	now := time.Now().Unix()
	result, err := tx.Exec(
		"INSERT INTO user_money_logs (user_id, money, `before`, `after`, memo, create_time) VALUES (?, ?, ?, ?, ?, ?)",
		userID, money, before, after, memo, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &UserMoneyLog{
		ID:         uint64(id),
		UserID:     userID,
		Money:      money,
		Before:     before,
		After:      after,
		Memo:       memo,
		CreateTime: now,
	}, nil
}

// GetUserMoneyForUpdate 在事务中锁定并读取用户余额（SELECT ... FOR UPDATE）
func GetUserMoneyForUpdate(tx *sql.Tx, userID uint64) (float64, error) {
	var money float64
	err := tx.QueryRow("SELECT money FROM users WHERE id = ? AND delete_time IS NULL FOR UPDATE", userID).Scan(&money)
	return money, err
}
