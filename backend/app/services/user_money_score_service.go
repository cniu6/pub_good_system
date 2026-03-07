package services

import (
	"errors"
	"fst/backend/app/models"
	"fst/backend/utils"
)

// ========================================
// 余额变动服务
// ========================================

// ChangeUserMoney 变更用户余额（同时记录日志）
// amount 为正数=充值，负数=扣款
func ChangeUserMoney(userID uint64, amount float64, memo string) (*models.UserMoneyLog, error) {
	memo = utils.Clean_XSS(memo)

	user, err := models.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	beforeMoney := user.Money
	afterMoney := beforeMoney + amount

	if amount < 0 && afterMoney < 0 {
		return nil, errors.New("扣款金额超出用户余额")
	}
	if amount > 0 && afterMoney > 999999999999 {
		return nil, errors.New("充值金额超出上限")
	}

	// 更新用户余额
	if err := models.UpdateUserMoney(userID, afterMoney); err != nil {
		return nil, errors.New("更新用户余额失败: " + err.Error())
	}

	// 记录日志
	logEntry, err := models.CreateUserMoneyLog(userID, amount, beforeMoney, afterMoney, memo)
	if err != nil {
		return nil, errors.New("记录余额变动日志失败: " + err.Error())
	}

	return logEntry, nil
}

// SetUserMoney 直接设置用户余额（管理员用，同时记录日志）
func SetUserMoney(userID uint64, newMoney float64, memo string) (*models.UserMoneyLog, error) {
	memo = utils.Clean_XSS(memo)

	user, err := models.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	beforeMoney := user.Money
	amount := newMoney - beforeMoney

	if newMoney < 0 {
		return nil, errors.New("余额不能为负数")
	}

	// 更新用户余额
	if err := models.UpdateUserMoney(userID, newMoney); err != nil {
		return nil, errors.New("更新用户余额失败: " + err.Error())
	}

	// 记录日志
	logEntry, err := models.CreateUserMoneyLog(userID, amount, beforeMoney, newMoney, memo)
	if err != nil {
		return nil, errors.New("记录余额变动日志失败: " + err.Error())
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
func ChangeUserScore(userID uint64, amount int64, memo string) (*models.UserScoreLog, error) {
	memo = utils.Clean_XSS(memo)

	user, err := models.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	beforeScore := user.Score
	afterScore := beforeScore + amount

	if amount < 0 && afterScore < 0 {
		return nil, errors.New("扣减积分超出用户积分余额")
	}
	if amount > 0 && afterScore > 999999999999 {
		return nil, errors.New("增加积分超出上限")
	}

	// 更新用户积分
	if err := models.UpdateUserScore(userID, afterScore); err != nil {
		return nil, errors.New("更新用户积分失败: " + err.Error())
	}

	// 记录日志
	logEntry, err := models.CreateUserScoreLog(userID, amount, beforeScore, afterScore, memo)
	if err != nil {
		return nil, errors.New("记录积分变动日志失败: " + err.Error())
	}

	return logEntry, nil
}

// SetUserScore 直接设置用户积分（管理员用，同时记录日志）
func SetUserScore(userID uint64, newScore int64, memo string) (*models.UserScoreLog, error) {
	memo = utils.Clean_XSS(memo)

	user, err := models.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	beforeScore := user.Score
	amount := newScore - beforeScore

	if newScore < 0 {
		return nil, errors.New("积分不能为负数")
	}

	// 更新用户积分
	if err := models.UpdateUserScore(userID, newScore); err != nil {
		return nil, errors.New("更新用户积分失败: " + err.Error())
	}

	// 记录日志
	logEntry, err := models.CreateUserScoreLog(userID, amount, beforeScore, newScore, memo)
	if err != nil {
		return nil, errors.New("记录积分变动日志失败: " + err.Error())
	}

	return logEntry, nil
}

// GetUserScoreLogList 获取积分变动列表
func GetUserScoreLogList(onlyUserID uint64, page, pageSize int, keyword string) ([]models.UserScoreLog, int64, error) {
	keyword = utils.Clean_XSS(keyword)
	return models.GetUserScoreLogList(onlyUserID, page, pageSize, keyword)
}
