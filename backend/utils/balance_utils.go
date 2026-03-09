package utils

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/internal/db"
	"log"
)

// ========================================
// 余额操作模式
// ========================================

// BalanceOpType 余额操作模式
type BalanceOpType int

const (
	// OpChangeOnly 只修改用户余额（不产生任何记录）
	OpChangeOnly BalanceOpType = iota + 1
	// OpLogOnly 只添加余额变动记录（不修改余额）
	OpLogOnly
	// OpChangeAndLog 修改余额 + 添加余额变动记录
	OpChangeAndLog
	// OpOrderAndLog 更新订单状态 + 添加余额变动记录（不修改余额）
	OpOrderAndLog
	// OpFull 修改余额 + 更新订单状态 + 添加余额变动记录
	OpFull
)

// ========================================
// 请求 / 结果 结构体
// ========================================

// BalanceReq 统一余额操作请求
type BalanceReq struct {
	UserID   uint64            // 用户ID（必填）
	Amount   float64           // 变动金额：正数=加款，负数=扣款
	Memo     string            // 单语言备注（当 MemoI18n 为空时使用）
	MemoI18n map[string]string // 多语言备注，如 {"zhCN":"在线充值","enUS":"Online Recharge"}

	// 订单相关字段（仅 OpOrderAndLog / OpFull 模式使用）
	OrderNo     string // 要更新的订单号
	TradeNo     string // 第三方交易号
	OrderStatus int    // 目标订单状态（如 models.PaymentStatusPaid）
}

// BalanceResult 余额操作结果
type BalanceResult struct {
	MoneyLog    *models.UserMoneyLog // 创建的余额变动记录（如有）
	BeforeMoney float64              // 变动前余额
	AfterMoney  float64              // 变动后余额
}

// ========================================
// 多语言备注工具
// ========================================

// BuildMemo 构建备注字符串
// 如果有多语言备注，序列化为 JSON 字符串存储；否则使用单语言备注
func BuildMemo(memo string, memoI18n map[string]string) string {
	if len(memoI18n) > 0 {
		data, err := json.Marshal(memoI18n)
		if err == nil {
			return string(data)
		}
	}
	return memo
}

// ParseMemo 解析备注字符串，返回指定语言的文本
// 如果 memo 是 JSON 格式的多语言对象，返回对应语言版本
// 否则原样返回纯文本
func ParseMemo(memo string, lang string) string {
	if memo == "" || memo[0] != '{' {
		return memo
	}
	var i18n map[string]string
	if err := json.Unmarshal([]byte(memo), &i18n); err != nil {
		return memo
	}
	// 精确匹配
	if text, ok := i18n[lang]; ok {
		return text
	}
	// 回退到中文
	if text, ok := i18n["zhCN"]; ok {
		return text
	}
	// 回退到第一个可用语言
	for _, text := range i18n {
		return text
	}
	return memo
}

// ========================================
// 统一余额操作入口
// ========================================

// ExecuteBalanceOp 执行余额操作（自动管理事务）
//
// 操作模式:
//   - OpChangeOnly:  只修改余额
//   - OpLogOnly:     只添加余额变动记录
//   - OpChangeAndLog: 修改余额 + 添加变动记录
//   - OpOrderAndLog:  更新订单状态 + 添加变动记录（不修改余额）
//   - OpFull:         修改余额 + 更新订单状态 + 添加变动记录
func ExecuteBalanceOp(req *BalanceReq, opType BalanceOpType) (*BalanceResult, error) {
	if req.UserID == 0 {
		return nil, errors.New("用户ID不能为空")
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	result, err := ExecuteBalanceOpTx(tx, req, opType)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	return result, nil
}

// ExecuteBalanceOpTx 在已有事务中执行余额操作
// 用于嵌入到更大的事务流程中（如支付回调）
func ExecuteBalanceOpTx(tx *sql.Tx, req *BalanceReq, opType BalanceOpType) (*BalanceResult, error) {
	if req.UserID == 0 {
		return nil, errors.New("用户ID不能为空")
	}

	memo := BuildMemo(req.Memo, req.MemoI18n)
	result := &BalanceResult{}

	needBalance := opType == OpChangeOnly || opType == OpChangeAndLog || opType == OpFull
	needLog := opType == OpLogOnly || opType == OpChangeAndLog || opType == OpOrderAndLog || opType == OpFull
	needOrder := opType == OpOrderAndLog || opType == OpFull

	// ---- 1. 锁定用户余额行 ----
	if needBalance || needLog {
		beforeMoney, err := models.GetUserMoneyForUpdate(tx, req.UserID)
		if err != nil {
			return nil, errors.New("用户不存在")
		}
		result.BeforeMoney = beforeMoney
		afterMoney := beforeMoney + req.Amount

		// 边界校验
		if req.Amount < 0 && afterMoney < 0 {
			return nil, errors.New("扣款金额超出用户余额")
		}
		if req.Amount > 0 && afterMoney > 999999999999 {
			return nil, errors.New("充值金额超出上限")
		}

		// ---- 2. 修改余额 ----
		if needBalance {
			if err := models.UpdateUserMoneyTx(tx, req.UserID, afterMoney); err != nil {
				return nil, fmt.Errorf("更新用户余额失败: %w", err)
			}
			result.AfterMoney = afterMoney
		} else {
			// 不实际修改余额，仅在日志中记录计算值
			result.AfterMoney = afterMoney
		}

		// ---- 3. 创建余额变动记录 ----
		if needLog {
			logEntry, err := models.CreateUserMoneyLogTx(tx, req.UserID, req.Amount, result.BeforeMoney, result.AfterMoney, memo)
			if err != nil {
				return nil, fmt.Errorf("创建余额变动记录失败: %w", err)
			}
			result.MoneyLog = logEntry
		}
	}

	// ---- 4. 更新订单状态 ----
	if needOrder {
		if req.OrderNo == "" {
			return nil, errors.New("订单号不能为空")
		}
		if err := models.UpdatePaymentOrderStatusTx(tx, req.OrderNo, req.OrderStatus, req.TradeNo); err != nil {
			return nil, fmt.Errorf("更新订单状态失败: %w", err)
		}
	}

	log.Printf("[BalanceOp] op=%d user=%d amount=%.2f before=%.2f after=%.2f order=%s memo=%s",
		opType, req.UserID, req.Amount, result.BeforeMoney, result.AfterMoney, req.OrderNo, memo)

	return result, nil
}
