package models

import (
	"database/sql"
	"fst/backend/internal/db"
	"log"
	"time"
)

// UserScoreLog 会员积分变动表
type UserScoreLog struct {
	ID         uint64 `db:"id" json:"id"`
	UserID     uint64 `db:"user_id" json:"user_id"`
	Score      int64  `db:"score" json:"score"`        // 变更积分（正=增加，负=扣减）
	Before     int64  `db:"before" json:"before"`       // 变更前积分
	After      int64  `db:"after" json:"after"`         // 变更后积分
	Memo       string `db:"memo" json:"memo"`           // 备注
	CreateTime int64  `db:"create_time" json:"create_time"`
}

// InitUserScoreLogsTable 初始化积分变动日志表
func InitUserScoreLogsTable() {
	if db.CheckTableExists("user_score_logs") {
		return
	}

	schema := `CREATE TABLE IF NOT EXISTS user_score_logs (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID',
		score BIGINT NOT NULL DEFAULT 0 COMMENT '变更积分',
		` + "`before`" + ` BIGINT NOT NULL DEFAULT 0 COMMENT '变更前积分',
		` + "`after`" + ` BIGINT NOT NULL DEFAULT 0 COMMENT '变更后积分',
		memo VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
		create_time BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
		INDEX idx_user_id (user_id),
		INDEX idx_create_time (create_time)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

	_, err := db.DB.Exec(schema)
	if err != nil {
		log.Printf("[Init] Failed to create user_score_logs table: %v", err)
	} else {
		log.Println("[Init] Created user_score_logs table")
	}
}

// CreateUserScoreLog 创建积分变动记录
func CreateUserScoreLog(userID uint64, score, before, after int64, memo string) (*UserScoreLog, error) {
	now := time.Now().Unix()
	result, err := db.DB.Exec(
		"INSERT INTO user_score_logs (user_id, score, `before`, `after`, memo, create_time) VALUES (?, ?, ?, ?, ?, ?)",
		userID, score, before, after, memo, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &UserScoreLog{
		ID:         uint64(id),
		UserID:     userID,
		Score:      score,
		Before:     before,
		After:      after,
		Memo:       memo,
		CreateTime: now,
	}, nil
}

// GetUserScoreLogByID 获取指定ID的积分变动记录
func GetUserScoreLogByID(id uint64) (*UserScoreLog, error) {
	var logEntry UserScoreLog
	err := db.DB.Get(&logEntry, "SELECT id, user_id, score, `before`, `after`, memo, create_time FROM user_score_logs WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &logEntry, nil
}

// GetUserScoreLogList 获取积分变动列表（分页+搜索）
func GetUserScoreLogList(onlyUserID uint64, page, pageSize int, keyword string) ([]UserScoreLog, int64, error) {
	var logs []UserScoreLog
	var total int64

	where := "WHERE 1=1"
	args := []interface{}{}

	if onlyUserID > 0 {
		where += " AND user_id = ?"
		args = append(args, onlyUserID)
	}
	if keyword != "" {
		where += " AND (memo LIKE ? OR CAST(score AS CHAR) LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}

	// 总数
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := db.DB.Get(&total, "SELECT COUNT(*) FROM user_score_logs "+where, countArgs...)
	if err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (page - 1) * pageSize
	query := "SELECT id, user_id, score, `before`, `after`, memo, create_time FROM user_score_logs " + where + " ORDER BY create_time DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	err = db.DB.Select(&logs, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// DeleteUserScoreLog 删除积分变动记录
func DeleteUserScoreLog(id uint64) error {
	_, err := db.DB.Exec("DELETE FROM user_score_logs WHERE id = ?", id)
	return err
}

// UpdateUserScore 直接更新用户积分字段
func UpdateUserScore(userID uint64, newScore int64) error {
	now := time.Now().Unix()
	_, err := db.DB.Exec("UPDATE users SET score = ?, update_time = ? WHERE id = ?", newScore, now, userID)
	return err
}

// UpdateUserScoreTx 在事务中更新用户积分字段
func UpdateUserScoreTx(tx *sql.Tx, userID uint64, newScore int64) error {
	now := time.Now().Unix()
	_, err := tx.Exec("UPDATE users SET score = ?, update_time = ? WHERE id = ?", newScore, now, userID)
	return err
}

// CreateUserScoreLogTx 在事务中创建积分变动记录
func CreateUserScoreLogTx(tx *sql.Tx, userID uint64, score, before, after int64, memo string) (*UserScoreLog, error) {
	now := time.Now().Unix()
	result, err := tx.Exec(
		"INSERT INTO user_score_logs (user_id, score, `before`, `after`, memo, create_time) VALUES (?, ?, ?, ?, ?, ?)",
		userID, score, before, after, memo, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &UserScoreLog{
		ID:         uint64(id),
		UserID:     userID,
		Score:      score,
		Before:     before,
		After:      after,
		Memo:       memo,
		CreateTime: now,
	}, nil
}

// GetUserScoreForUpdate 在事务中锁定并读取用户积分（SELECT ... FOR UPDATE）
func GetUserScoreForUpdate(tx *sql.Tx, userID uint64) (int64, error) {
	var score int64
	err := tx.QueryRow("SELECT score FROM users WHERE id = ? AND delete_time IS NULL FOR UPDATE", userID).Scan(&score)
	return score, err
}
