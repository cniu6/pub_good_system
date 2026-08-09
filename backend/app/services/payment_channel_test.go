package services

import (
	"errors"
	"fst/backend/app/models"
	"testing"
)

// stubPaymentChannel 测试用假通道
type stubPaymentChannel struct {
	typeName      string
	payURL        string
	tradeNo       string
	createErr     error
	verifyOK      bool
	queryResult   *PaymentQueryResult
	queryErr      error
	validPayTypes map[string]bool
	createCalled  bool
	verifyCalled  bool
	queryCalled   bool
	lastNotifyURL string
	lastReturnURL string
}

func (s *stubPaymentChannel) Type() string { return s.typeName }

func (s *stubPaymentChannel) CreatePay(gateway *models.PayGateway, order *models.PaymentOrder, notifyURL, returnURL string) (string, string, error) {
	s.createCalled = true
	s.lastNotifyURL = notifyURL
	s.lastReturnURL = returnURL
	return s.payURL, s.tradeNo, s.createErr
}

func (s *stubPaymentChannel) VerifyNotify(params map[string]string, key string) bool {
	s.verifyCalled = true
	return s.verifyOK
}

func (s *stubPaymentChannel) QueryOrder(gateway *models.PayGateway, orderNo, tradeNo string) (*PaymentQueryResult, error) {
	s.queryCalled = true
	return s.queryResult, s.queryErr
}

func (s *stubPaymentChannel) ValidatePayType(gateway *models.PayGateway, payType string) bool {
	if s.validPayTypes == nil {
		return true
	}
	return s.validPayTypes[payType]
}

func TestRegisterAndGetPaymentChannel(t *testing.T) {
	ClearPaymentChannels()
	t.Cleanup(ClearPaymentChannels)

	if _, ok := GetPaymentChannel("epay"); ok {
		t.Fatal("empty registry should not find epay")
	}

	RegisterPaymentChannel(nil) // 不应 panic
	if len(ListPaymentChannelTypes()) != 0 {
		t.Fatal("nil register should not add channel")
	}

	stub := &stubPaymentChannel{typeName: "epay", payURL: "https://pay.example/x"}
	RegisterPaymentChannel(stub)

	got, ok := GetPaymentChannel("epay")
	if !ok || got == nil {
		t.Fatal("expected epay channel after register")
	}
	if got.Type() != "epay" {
		t.Fatalf("Type() = %q", got.Type())
	}

	types := ListPaymentChannelTypes()
	if len(types) != 1 || types[0] != "epay" {
		t.Fatalf("ListPaymentChannelTypes = %v", types)
	}

	// 同类型覆盖注册
	stub2 := &stubPaymentChannel{typeName: "epay", payURL: "https://pay.example/y"}
	RegisterPaymentChannel(stub2)
	got2, _ := GetPaymentChannel("epay")
	if got2.(*stubPaymentChannel).payURL != "https://pay.example/y" {
		t.Fatal("re-register should replace channel")
	}
}

func TestGetPaymentChannel_Unknown(t *testing.T) {
	ClearPaymentChannels()
	t.Cleanup(ClearPaymentChannels)

	RegisterPaymentChannel(&stubPaymentChannel{typeName: "epay"})
	if _, ok := GetPaymentChannel("alipay"); ok {
		t.Fatal("unknown type should not be found")
	}
	if _, ok := GetPaymentChannel(""); ok {
		t.Fatal("empty type should not be found")
	}
}

func TestStubPaymentChannel_CreatePayAndVerify(t *testing.T) {
	ClearPaymentChannels()
	t.Cleanup(ClearPaymentChannels)

	stub := &stubPaymentChannel{
		typeName:      "epay",
		payURL:        "https://pay.example/cashier",
		tradeNo:       "TN001",
		verifyOK:      true,
		validPayTypes: map[string]bool{"alipay": true},
		queryResult: &PaymentQueryResult{
			Code:        1,
			TradeStatus: "TRADE_SUCCESS",
			Money:       "10.00",
		},
	}
	RegisterPaymentChannel(stub)

	ch, ok := GetPaymentChannel("epay")
	if !ok {
		t.Fatal("channel not registered")
	}

	gw := &models.PayGateway{Type: "epay", PayType: "alipay", ExtConfig: `{"key":"k"}`}
	order := &models.PaymentOrder{OrderNo: "P1", PaymentType: "alipay", PayAmount: 10}

	if !ch.ValidatePayType(gw, "alipay") {
		t.Fatal("alipay should be valid")
	}
	if ch.ValidatePayType(gw, "wxpay") {
		t.Fatal("wxpay should be invalid")
	}

	url, tradeNo, err := ch.CreatePay(gw, order, "https://n", "https://r")
	if err != nil || url != stub.payURL || tradeNo != "TN001" {
		t.Fatalf("CreatePay got url=%q tradeNo=%q err=%v", url, tradeNo, err)
	}
	if !stub.createCalled || stub.lastNotifyURL != "https://n" || stub.lastReturnURL != "https://r" {
		t.Fatal("CreatePay should record notify/return urls")
	}

	if !ch.VerifyNotify(map[string]string{"sign": "x"}, "k") || !stub.verifyCalled {
		t.Fatal("VerifyNotify should succeed")
	}

	qr, err := ch.QueryOrder(gw, "P1", "TN001")
	if err != nil || qr == nil || qr.Code != 1 || qr.TradeStatus != "TRADE_SUCCESS" {
		t.Fatalf("QueryOrder unexpected: %+v err=%v", qr, err)
	}
}

func TestStubPaymentChannel_CreatePayError(t *testing.T) {
	ClearPaymentChannels()
	t.Cleanup(ClearPaymentChannels)

	stub := &stubPaymentChannel{
		typeName:  "epay",
		createErr: errors.New("remote down"),
	}
	RegisterPaymentChannel(stub)

	ch, _ := GetPaymentChannel("epay")
	_, _, err := ch.CreatePay(&models.PayGateway{}, &models.PaymentOrder{}, "", "")
	if err == nil || err.Error() != "remote down" {
		t.Fatalf("expected create error, got %v", err)
	}
}

func TestPaymentServiceDispatchRequiresRegisteredChannel(t *testing.T) {
	ClearPaymentChannels()
	t.Cleanup(ClearPaymentChannels)

	// 未注册通道时，按类型查找应失败（CreatePaymentOrder / 回调验签的前置条件）
	if _, ok := GetPaymentChannel("epay"); ok {
		t.Fatal("registry should be empty")
	}

	// 注册后可按 gateway.Type 分发
	RegisterPaymentChannel(&stubPaymentChannel{
		typeName: "epay",
		payURL:   "https://x",
		tradeNo:  "T1",
		verifyOK: true,
	})
	ch, ok := GetPaymentChannel("epay")
	if !ok {
		t.Fatal("expected registered epay")
	}

	url, tradeNo, err := ch.CreatePay(
		&models.PayGateway{Type: "epay", NotifyURL: "https://custom-notify"},
		&models.PaymentOrder{OrderNo: "P1"},
		"https://global-notify",
		"https://return",
	)
	if err != nil || url != "https://x" || tradeNo != "T1" {
		t.Fatalf("dispatch CreatePay failed: url=%q tradeNo=%q err=%v", url, tradeNo, err)
	}
	if !ch.VerifyNotify(nil, "k") {
		t.Fatal("dispatch VerifyNotify failed")
	}
}
