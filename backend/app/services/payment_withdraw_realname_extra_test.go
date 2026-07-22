package services

import (
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
)

// enablePaymentForTest 注入一份仅用于测试的设置缓存，开启支付功能且不触发真实 DB 刷新
// （cacheTime 设为当前时间 + 足够长 ttl，避免 ensureFreshCache 覆盖手工注入的值）。
func enablePaymentForTest(t *testing.T) {
	t.Helper()
	old := GlobalSettingsService
	t.Cleanup(func() { GlobalSettingsService = old })
	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"payment_enabled": {Key: "payment_enabled", Value: "true"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}
}

func createTestPayGateway(t *testing.T, feeRate int, feeMode string) *models.PayGateway {
	t.Helper()
	gw := &models.PayGateway{
		Name:    "测试通道",
		Type:    "stub",
		PayType: "alipay",
		Status:  models.PayGatewayStatusEnabled,
		FeeRate: feeRate,
		FeeMode: feeMode,
	}
	if err := models.CreatePayGateway(gw); err != nil {
		t.Fatalf("创建支付通道失败: %v", err)
	}
	return gw
}

// TestCreatePaymentOrder_RejectsZeroCreditAmount 手续费配置异常（包含模式下费率=100%）导致到账金额为 0 时，
// 建单必须被拒绝，而不是仍然创建一个「用户付钱、到账 0 元」的订单。
func TestCreatePaymentOrder_RejectsZeroCreditAmount(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	enablePaymentForTest(t)

	ClearPaymentChannels()
	t.Cleanup(ClearPaymentChannels)
	RegisterPaymentChannel(&stubPaymentChannel{typeName: "stub", payURL: "https://pay.example/x", tradeNo: "TN1", verifyOK: true})

	u := testutil.CreateTestUser(t, "pay-zero-credit-user")
	gw := createTestPayGateway(t, 100, models.FeeModInclude)

	_, err := CreatePaymentOrder(u.ID, &CreatePaymentOrderRequest{
		GatewayID: gw.ID,
		Amount:    10,
	}, "https://notify", "https://return")
	if err == nil {
		t.Fatal("费率100%导致到账0元时应拒绝建单")
	}
	if !IsClientError(err) {
		t.Fatalf("期望 ClientError，实际 %T: %v", err, err)
	}
}

// TestCreatePaymentOrder_PendingLimitEnforcedAtomically 验证「待支付订单数量限制」在事务内生效：
// 达到上限后继续建单必须被拒绝，且不会产生第 11 笔挂单。
func TestCreatePaymentOrder_PendingLimitEnforcedAtomically(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	enablePaymentForTest(t)

	ClearPaymentChannels()
	t.Cleanup(ClearPaymentChannels)
	RegisterPaymentChannel(&stubPaymentChannel{typeName: "stub", payURL: "https://pay.example/x", tradeNo: "TN", verifyOK: true})

	u := testutil.CreateTestUser(t, "pay-limit-user")
	gw := createTestPayGateway(t, 0, models.FeeModInclude)

	// 后端允许同金额多笔待支付订单并存；这里用相同金额堆到上限，验证限流本身。
	for i := 0; i < maxPendingOrdersPerUser; i++ {
		if _, err := CreatePaymentOrder(u.ID, &CreatePaymentOrderRequest{
			GatewayID: gw.ID,
			Amount:    10,
		}, "https://notify", "https://return"); err != nil {
			t.Fatalf("第 %d 笔建单应成功: %v", i+1, err)
		}
	}

	_, err := CreatePaymentOrder(u.ID, &CreatePaymentOrderRequest{
		GatewayID: gw.ID,
		Amount:    10,
	}, "https://notify", "https://return")
	if err == nil {
		t.Fatal("超过待支付订单上限后应拒绝继续建单")
	}
	if !IsClientError(err) {
		t.Fatalf("期望 ClientError，实际 %T: %v", err, err)
	}

	pending, _, listErr := models.GetPaymentOrderList(u.ID, 1, 100, models.PaymentStatusPending, "")
	if listErr != nil {
		t.Fatalf("GetPaymentOrderList: %v", listErr)
	}
	if len(pending) != maxPendingOrdersPerUser {
		t.Fatalf("待支付订单数=%d，期望仍为 %d（不应超限）", len(pending), maxPendingOrdersPerUser)
	}
}

// TestCreatePaymentOrder_AllowsSameAmountPendingOrders 后端允许多笔同金额待支付订单并存。
func TestCreatePaymentOrder_AllowsSameAmountPendingOrders(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	enablePaymentForTest(t)

	ClearPaymentChannels()
	t.Cleanup(ClearPaymentChannels)
	RegisterPaymentChannel(&stubPaymentChannel{typeName: "stub", payURL: "https://pay.example/x", tradeNo: "TN", verifyOK: true})

	u := testutil.CreateTestUser(t, "pay-dup-amount-user")
	gw := createTestPayGateway(t, 0, models.FeeModInclude)

	firstOrder, err := CreatePaymentOrder(u.ID, &CreatePaymentOrderRequest{
		GatewayID: gw.ID,
		Amount:    10,
	}, "https://notify", "https://return")
	if err != nil {
		t.Fatalf("创建第一笔 10 元订单应成功: %v", err)
	}

	secondOrder, err := CreatePaymentOrder(u.ID, &CreatePaymentOrderRequest{
		GatewayID: gw.ID,
		Amount:    10,
	}, "https://notify", "https://return")
	if err != nil {
		t.Fatalf("创建第二笔同金额订单应成功: %v", err)
	}

	firstRefreshed, err := models.GetPaymentOrderByOrderNo(firstOrder.OrderNo)
	if err != nil {
		t.Fatalf("查询第一笔订单失败: %v", err)
	}
	if firstRefreshed.Status != models.PaymentStatusPending {
		t.Fatalf("第一笔应仍为待支付，实际状态=%d", firstRefreshed.Status)
	}

	secondRefreshed, err := models.GetPaymentOrderByOrderNo(secondOrder.OrderNo)
	if err != nil {
		t.Fatalf("查询第二笔订单失败: %v", err)
	}
	if secondRefreshed.Status != models.PaymentStatusPending {
		t.Fatalf("第二笔应为待支付，实际状态=%d", secondRefreshed.Status)
	}
}

// TestWithdrawServiceReview_ClearsBalanceDeductedOnReject 验证提现被拒绝退回预扣余额后，
// balance_deducted 字段同步清零，避免字段语义与实际余额状态不一致。
func TestWithdrawServiceReview_ClearsBalanceDeductedOnReject(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	old := GlobalSettingsService
	t.Cleanup(func() { GlobalSettingsService = old })
	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"withdraw_enabled":       {Key: "withdraw_enabled", Value: "true"},
			"withdraw_min_amount":    {Key: "withdraw_min_amount", Value: "1"},
			"withdraw_account_types": {Key: "withdraw_account_types", Value: "[\"bank\"]"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	u := testutil.CreateTestUser(t, "withdraw-reject-user")
	admin := testutil.CreateTestAdmin(t, "withdraw-reject-admin")
	beforeMoney := u.Money

	svc := NewWithdrawService()
	item, err := svc.Create(u.ID, &CreateWithdrawRequest{
		Amount:      20,
		AccountType: "bank",
		AccountName: "张三",
		AccountNo:   "6222000000000000",
		RealName:    "张三",
	})
	if err != nil {
		t.Fatalf("创建提现申请应成功: %v", err)
	}
	if !item.BalanceDeducted {
		t.Fatal("创建后应已预扣余额")
	}

	if err := svc.Review(admin.ID, &ReviewWithdrawRequest{
		ID:           item.ID,
		Status:       models.WithdrawStatusRejected,
		ReviewRemark: "资料有误",
	}); err != nil {
		t.Fatalf("审核拒绝应成功: %v", err)
	}

	refreshed, err := models.GetWithdrawRequestByID(item.ID)
	if err != nil {
		t.Fatalf("GetWithdrawRequestByID: %v", err)
	}
	if refreshed.Status != models.WithdrawStatusRejected {
		t.Fatalf("状态应为已拒绝，实际=%d", refreshed.Status)
	}
	if refreshed.BalanceDeducted {
		t.Fatal("拒绝退款后 balance_deducted 应被清零")
	}

	refreshedUser, err := models.GetUserByID(u.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if refreshedUser.Money != beforeMoney {
		t.Fatalf("拒绝退款后余额应恢复为 %v，实际=%v", beforeMoney, refreshedUser.Money)
	}
}

// TestRealnameSubmit_RejectsDuplicateCertificateNo 验证同一证件号不能被另一账号重复提交实名认证。
func TestRealnameSubmit_RejectsDuplicateCertificateNo(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	old := GlobalSettingsService
	t.Cleanup(func() { GlobalSettingsService = old })
	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"realname_enabled": {Key: "realname_enabled", Value: "true"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	u1 := testutil.CreateTestUser(t, "realname-dup-user-1")
	u2 := testutil.CreateTestUser(t, "realname-dup-user-2")

	const certNo = "110101199001011237" // 校验码合法的测试身份证号
	svc := NewRealnameService()

	if err := svc.Submit(u1.ID, &RealnameSubmitRequest{
		RealName:         "张三",
		CertificateType:  CertificateTypeIDCard,
		CertificateNo:    certNo,
		CertificateFront: "https://cdn.example.com/front.jpg",
		CertificateBack:  "https://cdn.example.com/back.jpg",
	}); err != nil {
		t.Fatalf("首次提交应成功: %v", err)
	}

	err := svc.Submit(u2.ID, &RealnameSubmitRequest{
		RealName:         "李四",
		CertificateType:  CertificateTypeIDCard,
		CertificateNo:    certNo,
		CertificateFront: "https://cdn.example.com/front2.jpg",
		CertificateBack:  "https://cdn.example.com/back2.jpg",
	})
	if err == nil {
		t.Fatal("同一证件号被另一账号提交时应拒绝")
	}
	if !IsClientError(err) {
		t.Fatalf("期望 ClientError，实际 %T: %v", err, err)
	}
}

// TestRealnameSubmit_RejectsNonHTTPCertificateImageURL 校验证件照 URL 协议，拒绝 javascript: 等恶意 scheme。
func TestRealnameSubmit_RejectsNonHTTPCertificateImageURL(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	old := GlobalSettingsService
	t.Cleanup(func() { GlobalSettingsService = old })
	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"realname_enabled": {Key: "realname_enabled", Value: "true"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	u := testutil.CreateTestUser(t, "realname-bad-url-user")
	svc := NewRealnameService()

	err := svc.Submit(u.ID, &RealnameSubmitRequest{
		RealName:         "王五",
		CertificateType:  CertificateTypeIDCard,
		CertificateNo:    "110101199001011237",
		CertificateFront: "javascript:alert(1)",
		CertificateBack:  "https://cdn.example.com/back.jpg",
	})
	if err == nil {
		t.Fatal("非 http(s) 协议的证件照 URL 应被拒绝")
	}
	if !IsClientError(err) {
		t.Fatalf("期望 ClientError，实际 %T: %v", err, err)
	}
}
