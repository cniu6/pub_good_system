package services

import (
	"testing"

	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"

	"gorm.io/gorm"
)

// TestCreateOrderForMoneyOperationRejectsDuplicate 并发补单重复 order_no 时，
// 后到者必须拿到错误并回滚，不能吞掉唯一键冲突后继续入账。
func TestCreateOrderForMoneyOperationRejectsDuplicate(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	user := testutil.CreateTestUser(t, "money-order-idem")
	req := MoneyOperationRequest{
		Operation:   "balance_order",
		Amount:      10,
		OrderNo:     "IDEM-ORDER-1",
		OrderStatus: 2,
	}

	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := createOrderForMoneyOperationTx(tx, user.ID, req); err != nil {
			t.Fatalf("首次建单应成功: %v", err)
		}
		if err := createOrderForMoneyOperationTx(tx, user.ID, req); err == nil {
			t.Fatal("重复建单必须返回错误，由事务回滚，不能继续入账")
		}
		return nil
	}); err != nil {
		t.Fatalf("事务失败: %v", err)
	}
}
