package services

import (
	"errors"
	"fst/backend/app/models"
	"fst/backend/internal/db"
	"fst/backend/utils"
)

// ========================================
// 余额变动服务
// ========================================

// ChangeUserMoney 变更用户余额（同时记录日志）
// amount 为正数=充值，负数=扣款
// 使用数据库事务 + SELECT ... FOR UPDATE 防止并发竞态
func ChangeUserMoney(userID uint64, amount float64, memo string) (*models.UserMoneyLog, error) {
	memo = utils.Clean_XSS(memo)

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, errors.New("开启事务失败: " + err.Error())
	}
	defer tx.Rollback()

	beforeMoney, err := models.GetUserMoneyForUpdate(tx, userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	afterMoney := beforeMoney + amount

	if amount < 0 && afterMoney < 0 {
		return nil, errors.New("扣款金额超出用户余额")
	}
	if amount > 0 && afterMoney > 999999999999 {
		return nil, errors.New("充值金额超出上限")
	}

	if err := models.UpdateUserMoneyTx(tx, userID, afterMoney); err != nil {
		return nil, errors.New("更新用户余额失败: " + err.Error())
	}

	logEntry, err := models.CreateUserMoneyLogTx(tx, userID, amount, beforeMoney, afterMoney, memo)
	if err != nil {
		return nil, errors.New("记录余额变动日志失败: " + err.Error())
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("提交事务失败: " + err.Error())
	}

	return logEntry, nil
}

// SetUserMoney 直接设置用户余额（管理员用，同时记录日志）
// 使用数据库事务 + SELECT ... FOR UPDATE 防止并发竞态
func SetUserMoney(userID uint64, newMoney float64, memo string) (*models.UserMoneyLog, error) {
	memo = utils.Clean_XSS(memo)

	if newMoney < 0 {
		return nil, errors.New("余额不能为负数")
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, errors.New("开启事务失败: " + err.Error())
	}
	defer tx.Rollback()

	beforeMoney, err := models.GetUserMoneyForUpdate(tx, userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	amount := newMoney - beforeMoney

	if err := models.UpdateUserMoneyTx(tx, userID, newMoney); err != nil {
		return nil, errors.New("更新用户余额失败: " + err.Error())
	}

	logEntry, err := models.CreateUserMoneyLogTx(tx, userID, amount, beforeMoney, newMoney, memo)
	if err != nil {
		return nil, errors.New("记录余额变动日志失败: " + err.Error())
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("提交事务失败: " + err.Error())
	}

	return logEntry, nil
}

// GetUserMoneyLogList 获取余额变动列表
func GetUserMoneyLogList(onlyUserID uint64, page, pageSize int, keyword string) ([]models.UserMoneyLog, int64, error) {
	keyword = utils.Clean_XSS(keyword)
	return models.GetUserMoneyLogList(onlyUserID, page, pageSize, keyword)
}

// ========================================
// 积分变动服务
// ========================================

// ChangeUserScore 变更用户积分（同时记录日志）
// amount 为正数=增加，负数=扣减
// 使用数据库事务 + SELECT ... FOR UPDATE 防止并发竞态
func ChangeUserScore(userID uint64, amount int64, memo string) (*models.UserScoreLog, error) {
	memo = utils.Clean_XSS(memo)

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, errors.New("开启事务失败: " + err.Error())
	}
	defer tx.Rollback()

	beforeScore, err := models.GetUserScoreForUpdate(tx, userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	afterScore := beforeScore + amount

	if amount < 0 && afterScore < 0 {
		return nil, errors.New("扣减积分超出用户积分余额")
	}
	if amount > 0 && afterScore > 999999999999 {
		return nil, errors.New("增加积分超出上限")
	}

	if err := models.UpdateUserScoreTx(tx, userID, afterScore); err != nil {
		return nil, errors.New("更新用户积分失败: " + err.Error())
	}

	logEntry, err := models.CreateUserScoreLogTx(tx, userID, amount, beforeScore, afterScore, memo)
	if err != nil {
		return nil, errors.New("记录积分变动日志失败: " + err.Error())
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("提交事务失败: " + err.Error())
	}

	return logEntry, nil
}

// SetUserScore 直接设置用户积分（管理员用，同时记录日志）
// 使用数据库事务 + SELECT ... FOR UPDATE 防止并发竞态
func SetUserScore(userID uint64, newScore int64, memo string) (*models.UserScoreLog, error) {
	memo = utils.Clean_XSS(memo)

	if newScore < 0 {
		return nil, errors.New("积分不能为负数")
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, errors.New("开启事务失败: " + err.Error())
	}
	defer tx.Rollback()

	beforeScore, err := models.GetUserScoreForUpdate(tx, userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	amount := newScore - beforeScore

	if err := models.UpdateUserScoreTx(tx, userID, newScore); err != nil {
		return nil, errors.New("更新用户积分失败: " + err.Error())
	}

	logEntry, err := models.CreateUserScoreLogTx(tx, userID, amount, beforeScore, newScore, memo)
	if err != nil {
		return nil, errors.New("记录积分变动日志失败: " + err.Error())
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("提交事务失败: " + err.Error())
	}

	return logEntry, nil
}

// GetUserScoreLogList 获取积分变动列表
func GetUserScoreLogList(onlyUserID uint64, page, pageSize int, keyword string) ([]models.UserScoreLog, int64, error) {
	keyword = utils.Clean_XSS(keyword)
	return models.GetUserScoreLogList(onlyUserID, page, pageSize, keyword)
}
