package models_test

import (
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

func TestModels_UserSettingsSessionsLogsSQLite(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "model-user-1")

	t.Run("用户查询与唯一约束", func(t *testing.T) {
		got, err := models.GetUserByID(u.ID)
		if err != nil || got.Username != u.Username {
			t.Fatalf("GetUserByID: got=%+v err=%v", got, err)
		}
		dup := &models.User{Username: u.Username, Password: "x", Role: "user", Status: 1}
		if err := models.CreateUser(dup); err == nil {
			t.Fatal("同名用户应失败")
		}
	})

	t.Run("设置读写", func(t *testing.T) {
		if err := models.UpdateSetting("site_name", "模型测试站"); err != nil {
			t.Fatalf("UpdateSetting: %v", err)
		}
		s, err := models.GetSettingByKey("site_name")
		if err != nil || s.Value != "模型测试站" {
			t.Fatalf("GetSettingByKey: %+v err=%v", s, err)
		}
		if err := models.BatchUpdateSettings(map[string]string{"site_desc": "desc"}); err != nil {
			t.Fatalf("BatchUpdateSettings: %v", err)
		}
	})

	t.Run("会话", func(t *testing.T) {
		now := time.Now()
		if err := models.CreateUserSession(u.ID, "user", "tok1", "ref1", "127.0.0.1", "ua", "pc", now.Add(time.Hour).Unix(), now.Add(2*time.Hour).Unix()); err != nil {
			t.Fatalf("CreateUserSession: %v", err)
		}
		ok, err := models.IsUserSessionActive(u.ID, "user", "tok1")
		if err != nil || !ok {
			t.Fatalf("IsUserSessionActive: ok=%v err=%v", ok, err)
		}
	})

	t.Run("操作日志与API日志", func(t *testing.T) {
		if err := models.CreateOperationLog(&models.OperationLog{
			UserID: u.ID, Username: u.Username, Module: "test", Action: "list",
			Method: "GET", Path: "/t", IP: "127.0.0.1",
		}); err != nil {
			t.Fatalf("CreateOperationLog: %v", err)
		}
		list, total, err := models.GetOperationLogList(&models.OperationLogQuery{Page: 1, PageSize: 10})
		if err != nil || total < 1 || len(list) < 1 {
			t.Fatalf("GetOperationLogList: total=%d len=%d err=%v", total, len(list), err)
		}
		now := time.Now().Unix()
		if err := models.CreateAPIAccessLog(&models.APIAccessLog{
			RequestID: "req-1", UserID: u.ID, Username: u.Username, Scene: "user",
			Method: "GET", Path: "/api", RoutePath: "/api", IP: "127.0.0.1",
			StatusCode: 200, Duration: 3, CreateTime: &now,
		}); err != nil {
			t.Fatalf("CreateAPIAccessLog: %v", err)
		}
	})

	t.Run("验证码", func(t *testing.T) {
		contact := "model-code@example.test"
		if err := models.CreateVerificationCode(contact, "654321", "register", time.Now().Add(time.Hour)); err != nil {
			t.Fatalf("CreateVerificationCode: %v", err)
		}
		vc, err := models.GetValidVerificationCode(contact, "register")
		if err != nil || vc.Code != "654321" {
			t.Fatalf("GetValidVerificationCode: %+v err=%v", vc, err)
		}
	})

	t.Run("支付网关与订单列表", func(t *testing.T) {
		gw := &models.PayGateway{Name: "gw1", Type: "epay", PayType: "alipay", Status: models.PayGatewayStatusEnabled}
		if err := models.CreatePayGateway(gw); err != nil {
			t.Fatalf("CreatePayGateway: %v", err)
		}
		order := &models.PaymentOrder{
			OrderNo: models.GenerateOrderNo(), UserID: u.ID, GatewayID: gw.ID,
			PaymentChannel: "epay", PaymentType: "alipay", Amount: 1, PayAmount: 1,
			Subject: "t", Status: models.PaymentStatusPending, ExpireAt: time.Now().Add(time.Hour).Unix(),
		}
		if err := models.CreatePaymentOrder(order); err != nil {
			t.Fatalf("CreatePaymentOrder: %v", err)
		}
		orders, total, err := models.GetPaymentOrderList(u.ID, 1, 20, -1, "")
		if err != nil || total < 1 || len(orders) < 1 {
			t.Fatalf("GetPaymentOrderList: total=%d err=%v", total, err)
		}
		if _, err := models.GetPaymentStats(); err != nil {
			t.Fatalf("GetPaymentStats: %v", err)
		}
	})

	t.Run("提现与实名", func(t *testing.T) {
		wr := &models.WithdrawRequest{
			UserID: u.ID, Amount: 1, AccountType: "bank", AccountName: "a",
			AccountNo: "6222", RealName: "张三",
		}
		if err := models.CreateWithdrawRequest(wr); err != nil {
			t.Fatalf("CreateWithdrawRequest: %v", err)
		}
		v := &models.RealnameVerification{
			UserID: u.ID, RealName: "张三", CertificateType: 1, CertificateNo: "110101199001011234",
		}
		if err := models.CreateRealnameVerification(v); err != nil {
			t.Fatalf("CreateRealnameVerification: %v", err)
		}
		got, err := models.GetRealnameVerificationByUserID(u.ID)
		if err != nil || got.RealName != "张三" {
			t.Fatalf("GetRealnameVerificationByUserID: %+v err=%v", got, err)
		}
	})

	t.Run("幂等键", func(t *testing.T) {
		tx, err := db.DB.Begin()
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		exp := time.Now().Add(time.Hour).Unix()
		if err := models.CreateIdempotencyKeyTx(tx, "k1", u.ID, "scope", "h", exp); err != nil {
			t.Fatalf("CreateIdempotencyKeyTx: %v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	})
}
