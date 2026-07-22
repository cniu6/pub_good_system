package services_test

import (
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/testutil"
	"fst/backend/utils"
)

func TestServices_UserMoneyScoreWithdrawSQLite(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "svc-user-1")

	t.Run("余额变更", func(t *testing.T) {
		res, err := utils.ExecuteBalanceOp(&utils.BalanceReq{
			UserID: u.ID, Amount: 5.5, Memo: "服务层余额测试",
		}, utils.OpChangeAndLog)
		if err != nil {
			t.Fatalf("ExecuteBalanceOp: %v", err)
		}
		if res.AfterMoney != 105.5 {
			t.Fatalf("after=%v want 105.5", res.AfterMoney)
		}
	})

	t.Run("积分变更", func(t *testing.T) {
		res, err := services.ChangeUserScore(u.ID, 8, "服务层积分测试")
		if err != nil {
			t.Fatalf("ChangeUserScore: %v", err)
		}
		if res.After != 28 {
			t.Fatalf("after=%d want 28", res.After)
		}
	})

	t.Run("用户列表", func(t *testing.T) {
		out, err := services.NewUserService().GetList(&services.UserListQuery{Page: 1, PageSize: 10, Keyword: "svc-user"})
		if err != nil || out.Total < 1 {
			t.Fatalf("GetList: %+v err=%v", out, err)
		}
	})

	t.Run("设置服务构造", func(t *testing.T) {
		svc := services.NewSettingsService(time.Minute)
		if svc == nil {
			t.Fatal("NewSettingsService nil")
		}
	})
}

// TestOperateUserMoney_DuplicateOrderNoDoubleCredit 验证管理端「统一余额操作」对已存在订单重复提交
// （如重试/重复点击）不会重复加/扣余额：订单已处于目标状态时第二次调用应被拒绝，余额只变更一次。
func TestOperateUserMoney_DuplicateOrderNoDoubleCredit(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "svc-user-dup-order")

	beforeMoney := u.Money

	req := services.MoneyOperationRequest{
		Amount:      20,
		Memo:        "补单测试",
		Operation:   "both",
		OrderNo:     "TEST_DUP_ORDER_001",
		OrderStatus: 2, // models.PaymentStatusPaid
	}

	// 第一次操作：正常建单 + 加款
	first, err := services.OperateUserMoney(u.ID, req)
	if err != nil {
		t.Fatalf("首次操作应成功: %v", err)
	}
	if first.AfterMoney != beforeMoney+20 {
		t.Fatalf("首次加款后余额=%v want %v", first.AfterMoney, beforeMoney+20)
	}

	// 第二次以相同订单号 + 相同目标状态重复提交：应被拒绝，不能重复加款
	_, err = services.OperateUserMoney(u.ID, req)
	if err == nil {
		t.Fatalf("订单已处于目标状态时重复操作应被拒绝，避免重复入账")
	}

	// 余额应仍是只加了一次：核对余额日志中「+20」的记录只有一条
	logs, _, listErr := services.GetUserMoneyLogList(u.ID, 1, 10, "")
	if listErr != nil {
		t.Fatalf("GetUserMoneyLogList: %v", listErr)
	}
	var creditCount int
	for _, l := range logs {
		if l.Money == 20 {
			creditCount++
		}
	}
	if creditCount != 1 {
		t.Fatalf("加款日志应只有 1 条，实际=%d", creditCount)
	}
}

// TestOperateUserMoney_AutoCreatedOrderUsesAbsoluteAmount 验证管理端扣款且订单不存在时，
// 自动补建的订单金额记录为操作幅度的绝对值，而不是被误置为 0（否则审计记录会丢失扣款金额信息）。
func TestOperateUserMoney_AutoCreatedOrderUsesAbsoluteAmount(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "svc-user-negative-order")
	beforeMoney := u.Money

	req := services.MoneyOperationRequest{
		Amount:      -15,
		Memo:        "管理员扣款自动建单测试",
		Operation:   "both",
		OrderNo:     "TEST_NEGATIVE_ORDER_001",
		OrderStatus: 4, // models.PaymentStatusFailed，仅用于测试标记目标状态
	}

	result, err := services.OperateUserMoney(u.ID, req)
	if err != nil {
		t.Fatalf("扣款操作应成功: %v", err)
	}
	if result.AfterMoney != beforeMoney-15 {
		t.Fatalf("扣款后余额=%v want %v", result.AfterMoney, beforeMoney-15)
	}

	order, err := models.GetPaymentOrderByOrderNo(req.OrderNo)
	if err != nil {
		t.Fatalf("GetPaymentOrderByOrderNo: %v", err)
	}
	if order.Amount != 15 || order.PayAmount != 15 {
		t.Fatalf("自动建单金额应为操作幅度的绝对值 15，实际 Amount=%v PayAmount=%v", order.Amount, order.PayAmount)
	}
}
