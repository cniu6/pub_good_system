package utils

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/db"
	"log"
	"math"
)

// ========================================
// Sentinel 错误
// ========================================

// ErrInsufficientBalance 扣款金额超出用户余额（调用方可用 errors.Is 判断，避免脆弱的字符串匹配）
var ErrInsufficientBalance = errors.New("扣款金额超出用户余额")

// ErrCreditLimitExceeded 充值金额超出上限
var ErrCreditLimitExceeded = errors.New("充值金额超出上限")

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
	// OpOrderOnly 只更新订单状态（不修改余额，不写余额日志）
	OpOrderOnly
	// OpChangeAndLog 修改余额 + 添加余额变动记录
	OpChangeAndLog
	// OpChangeAndOrder 修改余额 + 更新订单状态（不写余额日志）
	OpChangeAndOrder
	// OpOrderAndLog 更新订单状态 + 添加余额变动记录（不修改余额）
	OpOrderAndLog
	// OpFull 修改余额 + 更新订单状态 + 添加余额变动记录
	OpFull
)

// ========================================
// 请求 / 结果 结构体
// ========================================

// BalanceReq 统一余额操作请求
// Amount 入参仍为「元」（API 兼容）；内部一律转「分」int64 再加减。
type BalanceReq struct {
	UserID   uint64            // 用户ID（必填）
	Amount   float64           // 变动金额（元）：正数=加款，负数=扣款
	Memo     string            // 单语言备注（当 MemoI18n 为空时使用）
	MemoI18n map[string]string // 多语言备注，如 {"zhCN":"在线充值","enUS":"Online Recharge"}

	// 订单相关字段（仅 OpOrderAndLog / OpFull 模式使用）
	OrderNo     string // 要更新的订单号
	TradeNo     string // 第三方交易号
	OrderStatus int    // 目标订单状态（如 models.PaymentStatusPaid）
}

// BalanceResult 余额操作结果（对外仍返回「元」，已按分规范化）
type BalanceResult struct {
	MoneyLog    *models.UserMoneyLog // 创建的余额变动记录（如有）
	BeforeMoney float64              // 变动前余额（元）
	AfterMoney  float64              // 变动后余额（元）
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
//   - OpOrderOnly:   只更新订单状态
//   - OpChangeAndLog: 修改余额 + 添加变动记录
//   - OpChangeAndOrder: 修改余额 + 更新订单状态（不写日志）
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
// 内部计算全部按「分」整数进行，落库前再转回「元」，兼容现有 DECIMAL(10,2) 字段。
func ExecuteBalanceOpTx(tx *sql.Tx, req *BalanceReq, opType BalanceOpType) (*BalanceResult, error) {
	if req.UserID == 0 {
		return nil, errors.New("用户ID不能为空")
	}
	// 拒绝 NaN/Inf，避免脏浮点写入余额字段或绕过上下限判断
	if math.IsNaN(req.Amount) || math.IsInf(req.Amount, 0) {
		return nil, errors.New("金额非法")
	}

	// 变动金额先规范到「分」，后续加减只用 int64
	amountFen, err := YuanToFen(req.Amount)
	if err != nil {
		return nil, err
	}

	memo := BuildMemo(req.Memo, req.MemoI18n)
	result := &BalanceResult{}

	needBalance := opType == OpChangeOnly || opType == OpChangeAndLog || opType == OpFull
	needBalance = needBalance || opType == OpChangeAndOrder
	needLog := opType == OpLogOnly || opType == OpChangeAndLog || opType == OpOrderAndLog || opType == OpFull
	needOrder := opType == OpOrderOnly || opType == OpChangeAndOrder || opType == OpOrderAndLog || opType == OpFull

	// ---- 1. 锁定用户余额行（读出元 → 转分） ----
	if needBalance || needLog {
		beforeMoneyYuan, err := models.GetUserMoneyForUpdate(tx, req.UserID)
		if err != nil {
			return nil, errors.New("用户不存在")
		}
		beforeFen, err := YuanToFen(beforeMoneyYuan)
		if err != nil {
			return nil, fmt.Errorf("用户余额非法: %w", err)
		}
		afterFen := beforeFen + amountFen

		// 边界校验（按分）
		if amountFen < 0 && afterFen < 0 {
			return nil, ErrInsufficientBalance
		}
		if amountFen > 0 && afterFen > MoneyMaxFen {
			return nil, ErrCreditLimitExceeded
		}

		beforeYuan := FenToYuan(beforeFen)
		afterYuan := FenToYuan(afterFen)
		amountYuan := FenToYuan(amountFen)
		result.BeforeMoney = beforeYuan

		// ---- 2. 修改余额（写回规范化后的元） ----
		if needBalance {
			if err := models.UpdateUserMoneyTx(tx, req.UserID, afterYuan); err != nil {
				return nil, fmt.Errorf("更新用户余额失败: %w", err)
			}
			result.AfterMoney = afterYuan
		} else {
			// 不实际修改余额，仅在日志中记录计算值
			result.AfterMoney = afterYuan
		}

		// ---- 3. 创建余额变动记录（金额字段也按分规范化） ----
		if needLog {
			logEntry, err := models.CreateUserMoneyLogTx(tx, req.UserID, amountYuan, result.BeforeMoney, result.AfterMoney, memo)
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

	log.Printf("[BalanceOp] op=%d user=%d amountFen=%d amount=%.2f before=%.2f after=%.2f order=%s memo=%s",
		opType, req.UserID, amountFen, FenToYuan(amountFen), result.BeforeMoney, result.AfterMoney, req.OrderNo, memo)

	return result, nil
}
