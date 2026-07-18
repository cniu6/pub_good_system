package db_test

import (
	"fmt"
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/migrate"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
	"fst/backend/utils"
)

// TestSQLiteMigrateIndexesAndCRUD 深度覆盖：
// 1) 迁移编排后表/索引是否齐全
// 2) 二次迁移幂等
// 3) 主业务写入→读出→更新→删除全链路
func TestSQLiteMigrateIndexesAndCRUD(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	t.Run("迁移编排_关键表存在", func(t *testing.T) {
		need := []string{
			"users", "email_logs", "email_templates", "verification_codes",
			"user_realname_verifications", "auto_job_definitions", "auto_job_runs",
			"system_settings", "user_settings", "user_sessions",
			"user_money_logs", "user_score_logs", "operation_logs", "api_access_logs",
			"sms_logs", "payment_orders", "withdraw_requests", "idempotency_keys", "pay_gateways",
		}
		for _, name := range need {
			if !db.CheckTableExists(name) {
				t.Fatalf("缺表: %s", name)
			}
		}
	})

	t.Run("迁移编排_关键索引存在", func(t *testing.T) {
		// 这些索引来自 core schema / EnsureIndex；SQLite 适配后应能查到
		indexes := []struct {
			table, index string
		}{
			{"users", "idx_users_username"},
			{"users", "idx_users_email"},
			{"users", "idx_users_status"},
			{"email_logs", "idx_email_logs_to"},
			{"email_logs", "idx_email_logs_created_at"},
			{"verification_codes", "idx_contact_type"},
			{"verification_codes", "idx_expires_at"},
			{"user_realname_verifications", "idx_user_id"},
			{"user_realname_verifications", "idx_status"},
			{"auto_job_definitions", "idx_auto_job_def_enabled"},
			{"auto_job_runs", "idx_auto_job_runs_job_started"},
		}
		missing := make([]string, 0)
		for _, item := range indexes {
			if !db.CheckIndexExists(item.table, item.index) {
				missing = append(missing, item.table+"."+item.index)
			}
		}
		if len(missing) > 0 {
			// 列出 sqlite_master 便于排查适配漏索引
			var names []string
			_ = db.DB.Select(&names, `SELECT name FROM sqlite_master WHERE type='index' ORDER BY name`)
			t.Fatalf("缺索引: %v\n当前索引=%v", missing, names)
		}
	})

	t.Run("二次迁移幂等_不丢表不炸", func(t *testing.T) {
		migrate.RunAutoMigrate()
		if !db.CheckTableExists("users") || !db.CheckIndexExists("users", "idx_users_username") {
			t.Fatal("二次迁移后 users/唯一索引异常")
		}
	})

	t.Run("EnsureIndex_补索引可执行", func(t *testing.T) {
		const table = "operation_logs"
		const idx = "idx_operation_logs_sqlite_test_tmp"
		if db.CheckIndexExists(table, idx) {
			_, _ = db.Exec(`DROP INDEX IF EXISTS ` + idx)
		}
		db.EnsureIndex(table, idx, `ALTER TABLE operation_logs ADD INDEX idx_operation_logs_sqlite_test_tmp (create_time)`)
		if !db.CheckIndexExists(table, idx) {
			t.Fatal("EnsureIndex 后索引仍不存在（DDL 适配/执行失败）")
		}
		// 再调一次应幂等跳过
		db.EnsureIndex(table, idx, `ALTER TABLE operation_logs ADD INDEX idx_operation_logs_sqlite_test_tmp (create_time)`)
	})

	t.Run("用户唯一索引_写入冲突", func(t *testing.T) {
		u1 := testutil.CreateTestUser(t, "crud-unique-1")
		dup := &models.User{Username: u1.Username, Email: "other-" + u1.Email, Password: "x", Role: "user", Status: 1}
		if err := models.CreateUser(dup); err == nil {
			t.Fatal("同 username 应被唯一索引拒绝")
		}
		dup2 := &models.User{Username: "crud-unique-2", Email: u1.Email, Password: "x", Role: "user", Status: 1}
		if err := models.CreateUser(dup2); err == nil {
			t.Fatal("同 email 应被唯一索引拒绝")
		}
	})

	t.Run("CRUD_设置写入读出删除", func(t *testing.T) {
		key := "sqlite_crud_temp_key"
		_ = models.DeleteSetting(key)
		if err := models.CreateSetting(&models.SystemSetting{
			Key: key, Value: "v1", Type: "string", Category: "test", Label: "t", IsEditable: true,
		}); err != nil {
			t.Fatalf("CreateSetting: %v", err)
		}
		got, err := models.GetSettingByKey(key)
		if err != nil || got.Value != "v1" {
			t.Fatalf("读设置失败: %+v err=%v", got, err)
		}
		if err := models.UpdateSetting(key, "v2"); err != nil {
			t.Fatalf("UpdateSetting: %v", err)
		}
		got2, _ := models.GetSettingByKey(key)
		if got2.Value != "v2" {
			t.Fatalf("更新后 value=%q", got2.Value)
		}
		if err := models.DeleteSetting(key); err != nil {
			t.Fatalf("DeleteSetting: %v", err)
		}
		if _, err := models.GetSettingByKey(key); err == nil {
			t.Fatal("删除后仍能读到设置")
		}
	})

	t.Run("CRUD_支付订单写入读出删除", func(t *testing.T) {
		u := testutil.CreateTestUser(t, "crud-pay-user")
		order := &models.PaymentOrder{
			OrderNo: models.GenerateOrderNo(), UserID: u.ID, PaymentChannel: "test",
			PaymentType: "test", Amount: 3.21, PayAmount: 3.21, Subject: "crud",
			Status: models.PaymentStatusPending, ExpireAt: time.Now().Add(time.Hour).Unix(),
		}
		if err := models.CreatePaymentOrder(order); err != nil {
			t.Fatalf("CreatePaymentOrder: %v", err)
		}
		got, err := models.GetPaymentOrderByOrderNo(order.OrderNo)
		if err != nil || got.PayAmount != 3.21 {
			t.Fatalf("读订单失败: %+v err=%v", got, err)
		}
		if err := models.DeletePaymentOrder(got.ID); err != nil {
			t.Fatalf("DeletePaymentOrder: %v", err)
		}
		if _, err := models.GetPaymentOrderByOrderNo(order.OrderNo); err == nil {
			t.Fatal("删除后仍能读到订单")
		}
	})

	t.Run("CRUD_余额日志写入读出删除", func(t *testing.T) {
		u := testutil.CreateTestUser(t, "crud-money-user")
		res, err := utils.ExecuteBalanceOp(&utils.BalanceReq{UserID: u.ID, Amount: 1.11, Memo: "crud-money"}, utils.OpChangeAndLog)
		if err != nil || res.MoneyLog == nil {
			t.Fatalf("写余额日志失败: %+v err=%v", res, err)
		}
		list, total, err := models.GetUserMoneyLogList(u.ID, 1, 10, "crud-money")
		if err != nil || total < 1 || len(list) < 1 {
			t.Fatalf("读余额日志失败: total=%d err=%v", total, err)
		}
		if err := models.DeleteUserMoneyLog(list[0].ID); err != nil {
			t.Fatalf("DeleteUserMoneyLog: %v", err)
		}
		_, total2, err := models.GetUserMoneyLogList(u.ID, 1, 10, "crud-money")
		if err != nil || total2 != 0 {
			t.Fatalf("删除后仍有日志: total=%d err=%v", total2, err)
		}
	})

	t.Run("CRUD_积分日志写入读出删除", func(t *testing.T) {
		u := testutil.CreateTestUser(t, "crud-score-user")
		// 直接插日志（避免强依赖 service）
		now := time.Now().Unix()
		if _, err := db.Exec(
			`INSERT INTO user_score_logs (user_id, score, `+"`before`"+`, `+"`after`"+`, memo, create_time) VALUES (?,?,?,?,?,?)`,
			u.ID, 5, 0, 5, "crud-score", now,
		); err != nil {
			t.Fatalf("插入积分日志: %v", err)
		}
		list, total, err := models.GetUserScoreLogList(u.ID, 1, 10, "crud-score")
		if err != nil || total < 1 {
			t.Fatalf("读积分日志失败: total=%d err=%v", total, err)
		}
		if err := models.DeleteUserScoreLog(list[0].ID); err != nil {
			t.Fatalf("DeleteUserScoreLog: %v", err)
		}
	})

	t.Run("CRUD_验证码写入核销清理删除", func(t *testing.T) {
		contact := fmt.Sprintf("crud-code-%d@example.test", time.Now().UnixNano())
		if err := models.CreateVerificationCode(contact, "888888", "register", time.Now().Add(time.Hour)); err != nil {
			t.Fatalf("CreateVerificationCode: %v", err)
		}
		vc, err := models.GetValidVerificationCode(contact, "register")
		if err != nil {
			t.Fatalf("GetValidVerificationCode: %v", err)
		}
		used, err := models.ConsumeVerificationCode(contact, "888888", "register")
		if err != nil || !used {
			t.Fatalf("ConsumeVerificationCode: used=%v err=%v", used, err)
		}
		_ = vc
		if err := models.DeleteVerificationCodesByContact(contact, "register"); err != nil {
			t.Fatalf("DeleteVerificationCodesByContact: %v", err)
		}
		if _, err := models.GetValidVerificationCode(contact, "register"); err == nil {
			t.Fatal("删除后仍能取到验证码")
		}
	})

	t.Run("CRUD_会话写入撤销", func(t *testing.T) {
		u := testutil.CreateTestUser(t, "crud-session-user")
		now := time.Now()
		if err := models.CreateUserSession(u.ID, "user", "tok-crud", "ref-crud", "127.0.0.1", "ua", "pc", now.Add(time.Hour).Unix(), now.Add(2*time.Hour).Unix()); err != nil {
			t.Fatalf("CreateUserSession: %v", err)
		}
		sessions, err := models.GetUserSessionsWithGuard(u.ID, "user")
		if err != nil || len(sessions) != 1 {
			t.Fatalf("读会话失败: len=%d err=%v", len(sessions), err)
		}
		if err := models.RevokeUserSessionWithGuard(u.ID, "user", fmt.Sprint(sessions[0].ID)); err != nil {
			t.Fatalf("RevokeUserSession: %v", err)
		}
		ok, err := models.IsUserSessionActive(u.ID, "user", "tok-crud")
		if err != nil || ok {
			t.Fatalf("撤销后仍活跃: ok=%v err=%v", ok, err)
		}
	})

	t.Run("CRUD_操作日志写入清理", func(t *testing.T) {
		u := testutil.CreateTestUser(t, "crud-oplog-user")
		if err := models.CreateOperationLog(&models.OperationLog{
			UserID: u.ID, Username: u.Username, Module: "crud", Action: "del",
			Method: "DELETE", Path: "/crud", IP: "127.0.0.1",
		}); err != nil {
			t.Fatalf("CreateOperationLog: %v", err)
		}
		future := time.Now().Add(time.Hour).Unix()
		n, err := models.DeleteOperationLogsBefore(future)
		if err != nil || n < 1 {
			t.Fatalf("DeleteOperationLogsBefore: n=%d err=%v", n, err)
		}
	})

	t.Run("CRUD_支付通道写入读出删除", func(t *testing.T) {
		gw := &models.PayGateway{Name: "crud-gw", Type: "test", PayType: "alipay", Status: models.PayGatewayStatusEnabled}
		if err := models.CreatePayGateway(gw); err != nil {
			t.Fatalf("CreatePayGateway: %v", err)
		}
		list, total, err := models.GetPayGatewayList(1, 20, "crud-gw", true)
		if err != nil || total < 1 || len(list) < 1 {
			t.Fatalf("GetPayGatewayList: total=%d err=%v", total, err)
		}
		if err := models.DeletePayGateway(list[0].ID); err != nil {
			t.Fatalf("DeletePayGateway: %v", err)
		}
		_, total2, err := models.GetPayGatewayList(1, 20, "crud-gw", true)
		if err != nil || total2 != 0 {
			t.Fatalf("删除后仍有通道: total=%d err=%v", total2, err)
		}
	})

	t.Run("CRUD_幂等键写入删除", func(t *testing.T) {
		u := testutil.CreateTestUser(t, "crud-idem-user")
		tx, err := db.DB.Begin()
		if err != nil {
			t.Fatal(err)
		}
		exp := time.Now().Add(time.Hour).Unix()
		if err := models.CreateIdempotencyKeyTx(tx, "crud-idem-key", u.ID, "crud", "hash", exp); err != nil {
			_ = tx.Rollback()
			t.Fatalf("CreateIdempotencyKeyTx: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
		if err := models.DeleteIdempotencyKey("crud-idem-key", u.ID, "crud"); err != nil {
			t.Fatalf("DeleteIdempotencyKey: %v", err)
		}
	})

	t.Run("CRUD_提现与实名写入可读", func(t *testing.T) {
		u := testutil.CreateTestUser(t, "crud-wd-user")
		wr := &models.WithdrawRequest{
			UserID: u.ID, Amount: 2, AccountType: "bank", AccountName: "n",
			AccountNo: "6222", RealName: "测",
		}
		if err := models.CreateWithdrawRequest(wr); err != nil {
			t.Fatalf("CreateWithdrawRequest: %v", err)
		}
		got, err := models.GetWithdrawRequestByID(wr.ID)
		if err != nil || got.Amount != 2 {
			t.Fatalf("读提现失败: %+v err=%v", got, err)
		}
		v := &models.RealnameVerification{UserID: u.ID, RealName: "测", CertificateType: 1, CertificateNo: "110101199001011235"}
		if err := models.CreateRealnameVerification(v); err != nil {
			t.Fatalf("CreateRealnameVerification: %v", err)
		}
		rv, err := models.GetRealnameVerificationByUserID(u.ID)
		if err != nil || rv.RealName != "测" {
			t.Fatalf("读实名失败: %+v err=%v", rv, err)
		}
	})
}
