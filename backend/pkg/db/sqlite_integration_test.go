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
	"fst/backend/utils"
)

// TestSQLiteBusinessIntegration 在真实 SQLite 文件上执行完整迁移后，覆盖主要业务读写路径。
// 每个子测试独立报告错误，避免首个 SQL 兼容问题掩盖后续问题。
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
			_ = db.DB.Close()
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
		t.Fatalf("准备用户失败：models.CreateUser: %v", err)
	}

	t.Run("余额加减和日志锁行", func(t *testing.T) {
		result, err := utils.ExecuteBalanceOp(&utils.BalanceReq{
			UserID: user.ID, Amount: -12.34, Memo: "SQLite 余额扣减",
		}, utils.OpChangeAndLog)
		if err != nil {
			t.Fatalf("utils.ExecuteBalanceOp（SELECT ... FOR UPDATE）失败: %v", err)
		}
		if result.AfterMoney != 87.66 || result.MoneyLog == nil {
			t.Fatalf("余额/日志结果不正确：after=%v log=%+v", result.AfterMoney, result.MoneyLog)
		}
		if _, total, err := models.GetUserMoneyLogList(user.ID, 1, 20, "SQLite"); err != nil || total != 1 {
			t.Fatalf("models.GetUserMoneyLogList（CAST money AS CHAR）失败: total=%d err=%v", total, err)
		}
	})

	t.Run("积分加减和日志锁行", func(t *testing.T) {
		result, err := services.ChangeUserScore(user.ID, -3, "SQLite 积分扣减")
		if err != nil {
			t.Fatalf("services.ChangeUserScore（SELECT ... FOR UPDATE）失败: %v", err)
		}
		if result.After != 7 {
			t.Fatalf("积分结果不正确：after=%d", result.After)
		}
		if _, total, err := models.GetUserScoreLogList(user.ID, 1, 20, "SQLite"); err != nil || total != 1 {
			t.Fatalf("models.GetUserScoreLogList（CAST score AS CHAR）失败: total=%d err=%v", total, err)
		}
	})

	t.Run("支付订单锁行和状态更新", func(t *testing.T) {
		order := &models.PaymentOrder{
			OrderNo: "sqlite-payment-1", UserID: user.ID, PaymentChannel: "test",
			PaymentType: "test", Amount: 10, PayAmount: 10, Subject: "SQLite 测试订单",
			Status: models.PaymentStatusPending, ExpireAt: time.Now().Add(time.Hour).Unix(),
		}
		if err := models.CreatePaymentOrder(order); err != nil {
			t.Fatalf("models.CreatePaymentOrder 失败: %v", err)
		}
		tx, err := db.DB.Begin()
		if err != nil {
			t.Fatalf("开启支付事务失败: %v", err)
		}
		defer tx.Rollback()
		if _, err := models.GetPaymentOrderForUpdate(tx, order.OrderNo); err != nil {
			t.Fatalf("models.GetPaymentOrderForUpdate（FOR UPDATE）失败: %v", err)
		}
		if err := models.UpdatePaymentOrderStatusTx(tx, order.OrderNo, models.PaymentStatusPaid, "sqlite-trade-1"); err != nil {
			t.Fatalf("models.UpdatePaymentOrderStatusTx（FOR UPDATE）失败: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("提交支付事务失败: %v", err)
		}
	})

	t.Run("提现锁行查询", func(t *testing.T) {
		req := &models.WithdrawRequest{
			UserID: user.ID, Amount: 1, AccountType: "bank", AccountName: "测试银行",
			AccountNo: "6222000000000000", RealName: "测试用户",
		}
		if err := models.CreateWithdrawRequest(req); err != nil {
			t.Fatalf("models.CreateWithdrawRequest 失败: %v", err)
		}
		tx, err := db.DB.Begin()
		if err != nil {
			t.Fatalf("开启提现事务失败: %v", err)
		}
		defer tx.Rollback()
		if _, err := models.GetWithdrawRequestByIDForUpdate(tx, req.ID); err != nil {
			t.Fatalf("models.GetWithdrawRequestByIDForUpdate（FOR UPDATE）失败: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("提交提现事务失败: %v", err)
		}
	})

	t.Run("系统设置 NOW 更新", func(t *testing.T) {
		if err := models.UpdateSetting("site_name", "SQLite 集成测试"); err != nil {
			t.Fatalf("models.UpdateSetting（NOW）失败: %v", err)
		}
		if err := models.BatchUpdateSettings(map[string]string{"site_name": "SQLite 批量更新"}); err != nil {
			t.Fatalf("models.BatchUpdateSettings（NOW）失败: %v", err)
		}
	})

	t.Run("验证码时间读写核销清理", func(t *testing.T) {
		contact := "sqlite-code@example.test"
		if err := models.CreateVerificationCode(contact, "123456", "register", time.Now().Add(time.Hour)); err != nil {
			t.Fatalf("models.CreateVerificationCode 失败: %v", err)
		}
		if _, err := models.GetValidVerificationCode(contact, "register"); err != nil {
			t.Fatalf("models.GetValidVerificationCode（NOW/time.Time）失败: %v", err)
		}
		used, err := models.ConsumeVerificationCode(contact, "123456", "register")
		if err != nil || !used {
			t.Fatalf("models.ConsumeVerificationCode（NOW）失败: used=%v err=%v", used, err)
		}
		if _, err := models.CleanupOldVerificationCodes(); err != nil {
			t.Fatalf("models.CleanupOldVerificationCodes（DATE_SUB）失败: %v", err)
		}
	})

	t.Run("API 聚合 ON CONFLICT", func(t *testing.T) {
		now := time.Now().Unix()
		item := &models.APIAccessLog{
			RequestID: "sqlite-api-log-1", UserID: user.ID, Username: user.Username,
			Scene: "admin", Method: "GET", Path: "/api/admin/sqlite", RoutePath: "/api/admin/sqlite",
			IP: "127.0.0.1", StatusCode: 200, Duration: 12, CreateTime: &now,
		}
		if err := models.CreateAPIAccessLog(item); err != nil {
			t.Fatalf("models.CreateAPIAccessLog 失败: %v", err)
		}
		if err := models.RecordAPIAccessLogAggregate(item); err != nil {
			t.Fatalf("models.RecordAPIAccessLogAggregate（ON DUPLICATE KEY）失败: %v", err)
		}
		// 第二次初始化会触发已有日志的汇总回填，覆盖 DATE_FORMAT/FROM_UNIXTIME 方言转换。
		models.InitAPIAccessLogAggregateTables()
		if stats, err := models.GetAPIAccessLogStats(); err != nil || stats.TotalCount < 1 {
			t.Fatalf("models.GetAPIAccessLogStats（SQLite 回填）失败: stats=%+v err=%v", stats, err)
		}
		if _, total, err := models.GetAPIAccessLogList(&models.APIAccessLogQuery{Page: 1, PageSize: 20}); err != nil || total < 1 {
			t.Fatalf("models.GetAPIAccessLogList 失败: total=%d err=%v", total, err)
		}
	})

	t.Run("会话 CRUD 和布尔映射", func(t *testing.T) {
		now := time.Now()
		if err := models.CreateUserSession(user.ID, "user", "token-old", "refresh-old", "127.0.0.1", "sqlite-test", "test", now.Add(time.Hour).Unix(), now.Add(2*time.Hour).Unix()); err != nil {
			t.Fatalf("models.CreateUserSession 失败: %v", err)
		}
		active, err := models.IsUserSessionActive(user.ID, "user", "token-old")
		if err != nil || !active {
			t.Fatalf("models.IsUserSessionActive 失败: active=%v err=%v", active, err)
		}
		sessions, err := models.GetUserSessionsWithGuard(user.ID, "user")
		if err != nil || len(sessions) != 1 || !sessions[0].IsActive {
			t.Fatalf("models.GetUserSessionsWithGuard（TINYINT→bool）失败: sessions=%d err=%v", len(sessions), err)
		}
		if err := models.RevokeUserSessionWithGuard(user.ID, "user", fmt.Sprint(sessions[0].ID)); err != nil {
			t.Fatalf("models.RevokeUserSessionWithGuard 失败: %v", err)
		}
	})

	t.Run("幂等键占坑", func(t *testing.T) {
		tx, err := db.DB.Begin()
		if err != nil {
			t.Fatalf("开启幂等事务失败: %v", err)
		}
		defer tx.Rollback()
		expireAt := time.Now().Add(time.Hour).Unix()
		if err := models.CreateIdempotencyKeyTx(tx, "sqlite-idem", user.ID, "payment", "hash", expireAt); err != nil {
			t.Fatalf("models.CreateIdempotencyKeyTx 失败: %v", err)
		}
		item, err := models.GetActiveIdempotencyKeyTx(tx, "sqlite-idem", user.ID, "payment", time.Now().Unix())
		if err != nil || item.Status != models.IdempotencyStatusProcessing {
			t.Fatalf("models.GetActiveIdempotencyKeyTx 失败: item=%+v err=%v", item, err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("提交幂等事务失败: %v", err)
		}
		if err := models.MarkIdempotencyCompleted("sqlite-idem", user.ID, "payment"); err != nil {
			t.Fatalf("models.MarkIdempotencyCompleted 失败: %v", err)
		}
	})

	t.Run("操作邮件短信支付通道日志", func(t *testing.T) {
		if err := models.CreateOperationLog(&models.OperationLog{UserID: user.ID, Username: user.Username, Module: "sqlite", Action: "test", Method: "GET", Path: "/sqlite", IP: "127.0.0.1"}); err != nil {
			t.Fatalf("models.CreateOperationLog 失败: %v", err)
		}
		if _, total, err := models.GetOperationLogList(&models.OperationLogQuery{Page: 1, PageSize: 20}); err != nil || total < 1 {
			t.Fatalf("models.GetOperationLogList 失败: total=%d err=%v", total, err)
		}
		if err := models.CreateEmailLog("sqlite@example.test", "SQLite", "body", "sqlite", 1, ""); err != nil {
			t.Fatalf("models.CreateEmailLog 失败: %v", err)
		}
		if _, total, err := models.GetEmailLogList(&models.EmailLogQuery{Page: 1, PageSize: 20, Status: -1}); err != nil || total < 1 {
			t.Fatalf("models.GetEmailLogList（time.Time）失败: total=%d err=%v", total, err)
		}
		if err := models.CreateSMSLog(&models.SMSLog{Phone: "13800138000", Provider: "test", Status: 1}); err != nil {
			t.Fatalf("models.CreateSMSLog 失败: %v", err)
		}
		if _, total, err := models.GetSMSLogList(&models.SMSLogQuery{Page: 1, PageSize: 20, Status: -1}); err != nil || total < 1 {
			t.Fatalf("models.GetSMSLogList（time.Time）失败: total=%d err=%v", total, err)
		}
		gateway := &models.PayGateway{Name: "sqlite", Type: "test", PayType: "test", Status: models.PayGatewayStatusEnabled}
		if err := models.CreatePayGateway(gateway); err != nil {
			t.Fatalf("models.CreatePayGateway（反引号 key）失败: %v", err)
		}
		if _, total, err := models.GetPayGatewayList(1, 20, "sqlite", true); err != nil || total != 1 {
			t.Fatalf("models.GetPayGatewayList（反引号 key）失败: total=%d err=%v", total, err)
		}
	})

	t.Run("用户列表分页", func(t *testing.T) {
		result, err := services.NewUserService().GetList(&services.UserListQuery{Page: 1, PageSize: 20, Keyword: "sqlite"})
		if err != nil || result.Total < 1 {
			t.Fatalf("services.UserService.GetList 失败: result=%+v err=%v", result, err)
		}
	})

	t.Run("任务坏UID修复CHAR_LENGTH", func(t *testing.T) {
		if _, err := db.Exec(`INSERT INTO auto_job_runs (run_uid, job_code, status, started_at) VALUES (?, ?, ?, ?)`, "bad", "sqlite-job", "success", time.Now().Unix()); err != nil {
			t.Fatalf("插入坏 run_uid 失败: %v", err)
		}
		n, err := task.RepairBadRunUIDs()
		if err != nil {
			t.Fatalf("task.RepairBadRunUIDs（CHAR_LENGTH→LENGTH）失败: %v", err)
		}
		if n < 1 {
			t.Fatalf("期望至少修复 1 条，实际 %d", n)
		}
	})

	t.Run("仪表盘FROM_UNIXTIME趋势SQL", func(t *testing.T) {
		// dashboard 的 loadDashboardCountMap 会吞错返回空 map，这里必须直接执行并断言 error==nil
		startUnix := time.Now().AddDate(0, 0, -7).Unix()
		type dayRow struct {
			Day   string  `db:"day"`
			Value float64 `db:"value"`
		}
		var rows []dayRow
		qs := []struct {
			name string
			sql  string
			args []any
		}{
			{"newUsers", db.Q("SELECT DATE(FROM_UNIXTIME(create_time)) AS day, COUNT(*) AS value FROM users WHERE create_time >= ? GROUP BY DATE(FROM_UNIXTIME(create_time))"), []any{startUnix}},
			{"activeUsers", db.Q("SELECT DATE(FROM_UNIXTIME(last_login_time)) AS day, COUNT(*) AS value FROM users WHERE last_login_time >= ? GROUP BY DATE(FROM_UNIXTIME(last_login_time))"), []any{startUnix}},
			{"opLogs", db.Q("SELECT DATE(FROM_UNIXTIME(create_time)) AS day, COUNT(*) AS value FROM operation_logs WHERE create_time >= ? GROUP BY DATE(FROM_UNIXTIME(create_time))"), []any{startUnix}},
			{"paidOrders", db.Q(fmt.Sprintf("SELECT DATE(FROM_UNIXTIME(paid_at)) AS day, COUNT(*) AS value FROM payment_orders WHERE status = ? AND paid_at IS NOT NULL AND %s AND paid_at >= ? GROUP BY DATE(FROM_UNIXTIME(paid_at))", models.RealPaidOrderFilterSQL)), []any{models.PaymentStatusPaid, startUnix}},
			{"paidAmount", db.Q(fmt.Sprintf("SELECT DATE(FROM_UNIXTIME(paid_at)) AS day, COALESCE(SUM(pay_amount), 0) AS value FROM payment_orders WHERE status = ? AND paid_at IS NOT NULL AND %s AND paid_at >= ? GROUP BY DATE(FROM_UNIXTIME(paid_at))", models.RealPaidOrderFilterSQL)), []any{models.PaymentStatusPaid, startUnix}},
		}
		for _, item := range qs {
			rows = nil
			if err := db.DB.Select(&rows, item.sql, item.args...); err != nil {
				t.Fatalf("仪表盘趋势 %s 失败: %v\nsql=%s", item.name, err, item.sql)
			}
		}
	})

	t.Run("实名锁行和登录失败锁定", func(t *testing.T) {
		v := &models.RealnameVerification{
			UserID: user.ID, RealName: "测试", CertificateType: 1,
			CertificateNo: "110101199001011234", Status: 0,
		}
		if err := models.CreateRealnameVerification(v); err != nil {
			t.Fatalf("models.CreateRealnameVerification 失败: %v", err)
		}
		tx, err := db.DB.Begin()
		if err != nil {
			t.Fatalf("开启实名事务失败: %v", err)
		}
		defer tx.Rollback()
		if _, err := models.GetRealnameVerificationByUserIDForUpdate(tx, user.ID); err != nil {
			t.Fatalf("GetRealnameVerificationByUserIDForUpdate 失败: %v", err)
		}
		if _, err := models.GetRealnameVerificationByIDForUpdate(tx, v.ID); err != nil {
			t.Fatalf("GetRealnameVerificationByIDForUpdate 失败: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("提交实名事务失败: %v", err)
		}
		if err := services.NewUserService().IncrementLoginFailureWithLock(user.ID, 5, 15); err != nil {
			t.Fatalf("IncrementLoginFailureWithLock（FOR UPDATE）失败: %v", err)
		}
	})

	t.Run("支付统计和验证码软删", func(t *testing.T) {
		if _, err := models.GetPaymentStats(); err != nil {
			t.Fatalf("models.GetPaymentStats 失败: %v", err)
		}
		if _, err := models.SoftDeleteExpiredCodes(); err != nil {
			t.Fatalf("models.SoftDeleteExpiredCodes（NOW）失败: %v", err)
		}
		did, _, err := task.MaybeRenumberRunIDsIfNearLimit()
		if err != nil {
			t.Fatalf("MaybeRenumberRunIDsIfNearLimit 不应报错: %v", err)
		}
		if did {
			t.Fatalf("SQLite 下不应执行重编号")
		}
	})
}
