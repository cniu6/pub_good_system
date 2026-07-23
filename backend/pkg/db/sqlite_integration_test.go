package db_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/migrate"
	"fst/backend/internal/task"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"

	"gorm.io/gorm"
)

// TestSQLiteBusinessIntegration 只覆盖「跨包 + 锁行/聚合/任务」独特路径。
// 普通 CRUD 不在此重复（见 app/models、app/services 测试）。
func TestSQLiteBusinessIntegration(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_PATH", filepath.Join(t.TempDir(), "sqlite-integration.db"))
	t.Setenv("JWT_SECRET", "sqlite-integration-test-secret")
	t.Setenv("CORS_ORIGINS", "*")
	t.Setenv("APP_ENV", "development")

	config.InitConfig()
	db.InitDB()
	t.Cleanup(func() {
		if db.DB != nil {
			_ = db.Close()
		}
	})
	migrate.RunAutoMigrate()

	user := &models.User{
		Username: "sqlite-admin",
		Nickname: "SQLite 管理员",
		Email:    "sqlite-admin@example.test",
		Password: "not-used-in-test",
		Money:    100,
		Score:    10,
		Role:     "admin",
		Status:   1,
	}
	if err := models.CreateUser(user); err != nil {
		t.Fatalf("准备用户失败: %v", err)
	}

	t.Run("支付订单锁行", func(t *testing.T) {
		order := &models.PaymentOrder{
			OrderNo: "sqlite-payment-1", UserID: user.ID, PaymentChannel: "test",
			PaymentType: "test", Amount: 10, PayAmount: 10, Subject: "SQLite 测试订单",
			Status: models.PaymentStatusPending, ExpireAt: time.Now().Add(time.Hour).Unix(),
		}
		if err := models.CreatePaymentOrder(order); err != nil {
			t.Fatalf("CreatePaymentOrder: %v", err)
		}
		if err := db.DB.Transaction(func(tx *gorm.DB) error {
			if _, err := models.GetPaymentOrderForUpdate(tx, order.OrderNo); err != nil {
				return err
			}
			return models.UpdatePaymentOrderStatusTx(tx, order.OrderNo, models.PaymentStatusPaid, "sqlite-trade-1")
		}); err != nil {
			t.Fatalf("支付订单事务: %v", err)
		}
	})

	t.Run("提现锁行", func(t *testing.T) {
		req := &models.WithdrawRequest{
			UserID: user.ID, Amount: 1, AccountType: "bank", AccountName: "测试银行",
			AccountNo: "6222000000000000", RealName: "测试用户",
		}
		if err := models.CreateWithdrawRequest(req); err != nil {
			t.Fatalf("CreateWithdrawRequest: %v", err)
		}
		if err := db.DB.Transaction(func(tx *gorm.DB) error {
			_, err := models.GetWithdrawRequestByIDForUpdate(tx, req.ID)
			return err
		}); err != nil {
			t.Fatalf("提现锁行: %v", err)
		}
	})

	t.Run("API聚合OnConflict与回填", func(t *testing.T) {
		now := time.Now().Unix()
		item := &models.APIAccessLog{
			RequestID: "sqlite-api-log-1", UserID: user.ID, Username: user.Username,
			Scene: "admin", Method: "GET", Path: "/api/admin/sqlite", RoutePath: "/api/admin/sqlite",
			IP: "127.0.0.1", StatusCode: 200, Duration: 12, CreateTime: &now,
		}
		if err := models.CreateAPIAccessLog(item); err != nil {
			t.Fatalf("CreateAPIAccessLog: %v", err)
		}
		if err := models.RecordAPIAccessLogAggregate(item); err != nil {
			t.Fatalf("RecordAPIAccessLogAggregate: %v", err)
		}
		models.BackfillAPIAccessLogAggregateIfNeeded()
		if stats, err := models.GetAPIAccessLogStats(); err != nil || stats.TotalCount < 1 {
			t.Fatalf("GetAPIAccessLogStats: stats=%+v err=%v", stats, err)
		}
	})

	t.Run("任务坏UID修复", func(t *testing.T) {
		if err := db.DB.Exec(
			`INSERT INTO auto_job_runs (run_uid, job_code, status, started_at) VALUES (?, ?, ?, ?)`,
			"bad", "sqlite-job", "success", time.Now().Unix(),
		).Error; err != nil {
			t.Fatalf("插入坏 run_uid: %v", err)
		}
		n, err := task.RepairBadRunUIDs()
		if err != nil {
			t.Fatalf("RepairBadRunUIDs: %v", err)
		}
		if n < 1 {
			t.Fatalf("期望至少修复 1 条，实际 %d", n)
		}
	})

	t.Run("实名锁行与登录失败锁定", func(t *testing.T) {
		v := &models.RealnameVerification{
			UserID: user.ID, RealName: "测试", CertificateType: 1,
			CertificateNo: fmt.Sprintf("11010119900101%04d", time.Now().Unix()%10000), Status: 0,
		}
		if err := models.CreateRealnameVerification(v); err != nil {
			t.Fatalf("CreateRealnameVerification: %v", err)
		}
		if err := db.DB.Transaction(func(tx *gorm.DB) error {
			if _, err := models.GetRealnameVerificationByUserIDForUpdate(tx, user.ID); err != nil {
				return err
			}
			_, err := models.GetRealnameVerificationByIDForUpdate(tx, v.ID)
			return err
		}); err != nil {
			t.Fatalf("实名锁行: %v", err)
		}
		if err := services.NewUserService().IncrementLoginFailureWithLock(user.ID, 5, 15); err != nil {
			t.Fatalf("IncrementLoginFailureWithLock: %v", err)
		}
	})

	t.Run("支付统计与SQLite跳过重编号", func(t *testing.T) {
		if _, err := models.GetPaymentStats(); err != nil {
			t.Fatalf("GetPaymentStats: %v", err)
		}
		if _, err := models.SoftDeleteExpiredCodes(); err != nil {
			t.Fatalf("SoftDeleteExpiredCodes: %v", err)
		}
		did, _, err := task.MaybeRenumberRunIDsIfNearLimit()
		if err != nil {
			t.Fatalf("MaybeRenumberRunIDsIfNearLimit: %v", err)
		}
		if did {
			t.Fatal("SQLite 下不应执行重编号")
		}
	})
}
