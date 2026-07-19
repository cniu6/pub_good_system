package services

import (
	"database/sql"
	"errors"
	"fst/backend/app/models"
	"fst/backend/pkg/db"
	"fst/backend/utils"
	"math"
	"strings"
)

// MoneyOperationRequest 统一余额操作请求
type MoneyOperationRequest struct {
	Amount      float64
	Memo        string
	Operation   string
	OrderNo     string
	TradeNo     string
	OrderStatus int
}

// ========================================
// 余额变动服务
// ========================================

// ChangeUserMoney 变更用户余额（同时记录日志）
// amount 为正数=充值，负数=扣款
// 内部通过 ExecuteBalanceOp(OpChangeAndLog) 实现事务安全
func ChangeUserMoney(userID uint64, amount float64, memo string) (*models.UserMoneyLog, error) {
	memo = utils.Clean_XSS(memo)
	result, err := utils.ExecuteBalanceOp(&utils.BalanceReq{
		UserID: userID,
		Amount: amount,
		Memo:   memo,
	}, utils.OpChangeAndLog)
	if err != nil {
		return nil, err
	}
	return result.MoneyLog, nil
}

// ChangeUserMoneyI18n 变更用户余额（多语言备注版本）
func ChangeUserMoneyI18n(userID uint64, amount float64, memoI18n map[string]string) (*models.UserMoneyLog, error) {
	result, err := utils.ExecuteBalanceOp(&utils.BalanceReq{
		UserID:   userID,
		Amount:   amount,
		MemoI18n: memoI18n,
	}, utils.OpChangeAndLog)
	if err != nil {
		return nil, err
	}
	return result.MoneyLog, nil
}

// SetUserMoney 直接设置用户余额（管理员用，同时记录日志）
// 内部先按「分」计算差值，再通过 ExecuteBalanceOp(OpChangeAndLog) 处理
func SetUserMoney(userID uint64, newMoney float64, memo string) (*models.UserMoneyLog, error) {
	memo = utils.Clean_XSS(memo)

	if newMoney < 0 {
		return nil, errors.New("余额不能为负数")
	}
	// 拒绝非法浮点，避免管理员误传 NaN/Inf 污染余额
	if math.IsNaN(newMoney) || math.IsInf(newMoney, 0) {
		return nil, errors.New("金额非法")
	}

	newFen, err := utils.YuanToFen(newMoney)
	if err != nil {
		return nil, err
	}
	if newFen > utils.MoneyMaxFen {
		return nil, errors.New("余额超出上限")
	}
	newMoney = utils.FenToYuan(newFen)

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, errors.New("开启事务失败: " + err.Error())
	}
	defer tx.Rollback()

	beforeMoney, err := models.GetUserMoneyForUpdate(tx, userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	beforeFen, err := utils.YuanToFen(beforeMoney)
	if err != nil {
		return nil, errors.New("用户余额非法")
	}

	// 差值按分计算，再转回元交给统一入口
	amount := utils.FenToYuan(newFen - beforeFen)

	result, err := utils.ExecuteBalanceOpTx(tx, &utils.BalanceReq{
		UserID: userID,
		Amount: amount,
		Memo:   memo,
	}, utils.OpChangeAndLog)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("提交事务失败: " + err.Error())
	}

	return result.MoneyLog, nil
}

// GetUserMoneyLogList 获取余额变动列表
func GetUserMoneyLogList(onlyUserID uint64, page, pageSize int, keyword string) ([]models.UserMoneyLog, int64, error) {
	keyword = utils.Clean_XSS(keyword)
	return models.GetUserMoneyLogList(onlyUserID, page, pageSize, keyword)
}

// AddUserMoneyLogOnly 仅添加余额变动日志（不修改余额）
// amount 为正数=充值，负数=扣款
func AddUserMoneyLogOnly(userID uint64, amount float64, memo string) (*models.UserMoneyLog, error) {
	memo = utils.Clean_XSS(memo)
	result, err := utils.ExecuteBalanceOp(&utils.BalanceReq{
		UserID: userID,
		Amount: amount,
		Memo:   memo,
	}, utils.OpLogOnly)
	if err != nil {
		return nil, err
	}
	return result.MoneyLog, nil
}

// OperateUserMoney 统一余额操作入口（支持余额/日志/订单的交集与并集）
func OperateUserMoney(userID uint64, req MoneyOperationRequest) (*utils.BalanceResult, error) {
	if userID == 0 {
		return nil, errors.New("用户ID不能为空")
	}

	opType, needOrder, err := mapMoneyOperationType(req.Operation)
	if err != nil {
		return nil, err
	}

	if req.Operation != "order_only" && req.Amount == 0 {
		return nil, errors.New("涉及余额或日志操作时，金额不能为0")
	}

	req.Memo = utils.Clean_XSS(req.Memo)
	req.OrderNo = utils.Clean_XSS(req.OrderNo)
	req.TradeNo = utils.Clean_XSS(req.TradeNo)

	balanceReq := &utils.BalanceReq{
		UserID:      userID,
		Amount:      req.Amount,
		Memo:        req.Memo,
		OrderNo:     req.OrderNo,
		TradeNo:     req.TradeNo,
		OrderStatus: req.OrderStatus,
	}

	if !needOrder {
		return utils.ExecuteBalanceOp(balanceReq, opType)
	}
	if req.OrderNo == "" {
		return nil, errors.New("订单号不能为空")
	}

	// 订单已存在：直接走余额事务；订单不存在：建单+余额/订单更新必须同一事务，避免「已支付订单但未到账」
	order, err := models.GetPaymentOrderByOrderNo(req.OrderNo)
	if err == nil {
		if order.UserID != userID {
			return nil, errors.New("订单不属于当前用户")
		}
		return utils.ExecuteBalanceOp(balanceReq, opType)
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, errors.New("开启事务失败: " + err.Error())
	}
	defer tx.Rollback()

	if createErr := createOrderForMoneyOperationTx(tx, userID, req); createErr != nil {
		return nil, createErr
	}
	locked, lockErr := models.GetPaymentOrderForUpdate(tx, req.OrderNo)
	if lockErr != nil {
		return nil, errors.New("订单不存在")
	}
	if locked.UserID != userID {
		return nil, errors.New("订单不属于当前用户")
	}

	result, err := utils.ExecuteBalanceOpTx(tx, balanceReq, opType)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, errors.New("提交事务失败: " + err.Error())
	}
	return result, nil
}

// createOrderForMoneyOperationTx 在事务内补建缺失订单（与余额操作同事务）
func createOrderForMoneyOperationTx(tx *sql.Tx, userID uint64, req MoneyOperationRequest) error {
	amount := req.Amount
	if amount < 0 {
		amount = 0
	}
	subject := "管理员余额操作自动创建订单"
	if req.Memo != "" {
		subject = req.Memo
	}
	err := models.CreatePaymentOrderTx(tx, &models.PaymentOrder{
		OrderNo:        req.OrderNo,
		UserID:         userID,
		GatewayID:      0,
		TradeNo:        req.TradeNo,
		PaymentChannel: "admin",
		PaymentType:    "manual",
		Amount:         amount,
		Fee:            0,
		PayAmount:      amount,
		Subject:        subject,
		Status:         req.OrderStatus,
		NotifyCount:    0,
		PayURL:         "",
		PaidAt:         nil,
		ExpireAt:       0,
		ClientIP:       "admin",
		Extra:          "",
	})
	if err == nil {
		return nil
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique constraint") {
		return nil
	}
	return errors.New("自动创建订单失败: " + err.Error())
}

func mapMoneyOperationType(operation string) (utils.BalanceOpType, bool, error) {
	switch operation {
	case "balance_only":
		return utils.OpChangeOnly, false, nil
	case "log_only":
		return utils.OpLogOnly, false, nil
	case "order_only":
		return utils.OpOrderOnly, true, nil
	case "balance_log":
		return utils.OpChangeAndLog, false, nil
	case "balance_order", "both":
		if operation == "balance_order" {
			return utils.OpChangeAndOrder, true, nil
		}
		return utils.OpFull, true, nil
	case "log_order":
		return utils.OpOrderAndLog, true, nil
	default:
		return 0, false, errors.New("不支持的余额操作类型")
	}
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
	if newScore > 999999999999 {
		return nil, errors.New("积分超出上限")
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

// AddUserScoreLogOnly 仅添加积分变动日志（不修改积分）
// amount 为正数=增加，负数=扣减
func AddUserScoreLogOnly(userID uint64, amount int64, memo string) (*models.UserScoreLog, error) {
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

	logEntry, err := models.CreateUserScoreLogTx(tx, userID, amount, beforeScore, afterScore, memo)
	if err != nil {
		return nil, errors.New("记录积分变动日志失败: " + err.Error())
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.New("提交事务失败: " + err.Error())
	}

	return logEntry, nil
}

