package services

import (
	"errors"
	"testing"
)

func TestIsPermanentPaymentNotifyError(t *testing.T) {
	t.Run("PaymentNotifyError permanent", func(t *testing.T) {
		err := newPaymentNotifyError(true, "Signature verification failed")
		if !IsPermanentPaymentNotifyError(err) {
			t.Fatal("expected permanent")
		}
	})
	t.Run("PaymentNotifyError retryable", func(t *testing.T) {
		err := newPaymentNotifyError(false, "Temporary failure")
		if IsPermanentPaymentNotifyError(err) {
			t.Fatal("expected retryable")
		}
	})
	t.Run("wrapped permanent string", func(t *testing.T) {
		err := errors.New("Callback amount does not match order amount")
		if !IsPermanentPaymentNotifyError(err) {
			t.Fatal("expected amount mismatch permanent")
		}
	})
	t.Run("db error retryable", func(t *testing.T) {
		err := errors.New("Failed to start transaction: connection reset")
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
		{"Signature verification failed", "sign_failed"},
		{"Callback amount does not match order amount", "amount_mismatch"},
		{"Merchant ID mismatch", "binding_mismatch"},
		{"Order does not exist", "order_missing"},
		{"Other", "permanent_rejected"},
	}
	for _, c := range cases {
		got := classifyNotifyExceptionType(errors.New(c.msg))
		if got != c.want {
			t.Fatalf("msg=%q got %s want %s", c.msg, got, c.want)
		}
	}
}
