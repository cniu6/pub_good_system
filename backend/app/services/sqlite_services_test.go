package services_test

import (
	"testing"
	"time"

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
