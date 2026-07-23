package services

import (
	"errors"
	"testing"
)

func TestIsPermanentPaymentNotifyError(t *testing.T) {
	t.Run("PaymentNotifyError permanent", func(t *testing.T) {
		err := newPaymentNotifyError(true, "签名验证失败")
		if !IsPermanentPaymentNotifyError(err) {
			t.Fatal("expected permanent")
		}
	})
	t.Run("PaymentNotifyError retryable", func(t *testing.T) {
		err := newPaymentNotifyError(false, "临时故障")
		if IsPermanentPaymentNotifyError(err) {
			t.Fatal("expected retryable")
		}
	})
	t.Run("wrapped permanent string", func(t *testing.T) {
		err := errors.New("回调金额与订单金额不一致")
		if !IsPermanentPaymentNotifyError(err) {
			t.Fatal("expected amount mismatch permanent")
		}
	})
	t.Run("db error retryable", func(t *testing.T) {
		err := errors.New("开启事务失败: connection reset")
		if IsPermanentPaymentNotifyError(err) {
			t.Fatal("db errors should be retryable")
		}
	})
	t.Run("nil", func(t *testing.T) {
		if IsPermanentPaymentNotifyError(nil) {
			t.Fatal("nil should not be permanent")
		}
	})
}

func TestClassifyNotifyExceptionType(t *testing.T) {
	cases := []struct {
		msg  string
		want string
	}{
		{"签名验证失败", "sign_failed"},
		{"回调金额与订单金额不一致", "amount_mismatch"},
		{"商户号不匹配", "binding_mismatch"},
		{"订单不存在", "order_missing"},
		{"其他", "permanent_rejected"},
	}
	for _, c := range cases {
		got := classifyNotifyExceptionType(errors.New(c.msg))
		if got != c.want {
			t.Fatalf("msg=%q got %s want %s", c.msg, got, c.want)
		}
	}
}
