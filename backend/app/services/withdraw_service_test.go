package services

import (
	"fst/backend/app/models"
	"testing"
	"time"
)

func TestShouldRefundReservedBalance(t *testing.T) {
	tests := []struct {
		name            string
		reviewStatus    uint8
		balanceDeducted bool
		expected        bool
	}{
		{name: "拒绝且已预扣时退回", reviewStatus: models.WithdrawStatusRejected, balanceDeducted: true, expected: true},
		{name: "拒绝但未预扣时不退回", reviewStatus: models.WithdrawStatusRejected, balanceDeducted: false, expected: false},
		{name: "审核通过时不退回", reviewStatus: models.WithdrawStatusApproved, balanceDeducted: true, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRefundReservedBalance(tt.reviewStatus, tt.balanceDeducted); got != tt.expected {
				t.Fatalf("shouldRefundReservedBalance(%d, %v) = %v, want %v", tt.reviewStatus, tt.balanceDeducted, got, tt.expected)
			}
		})
	}
}

func TestShouldDeductBalanceOnWithdrawPay(t *testing.T) {
	if !shouldDeductBalanceOnWithdrawPay(false) {
		t.Fatalf("shouldDeductBalanceOnWithdrawPay(false) = false, want true")
	}
	if shouldDeductBalanceOnWithdrawPay(true) {
		t.Fatalf("shouldDeductBalanceOnWithdrawPay(true) = true, want false")
	}
}

func TestWithdrawServiceCreateRejectsDisabledWithdrawBeforeDB(t *testing.T) {
	oldSettingsService := GlobalSettingsService
	defer func() {
		GlobalSettingsService = oldSettingsService
	}()

	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"withdraw_enabled": {Key: "withdraw_enabled", Value: "false"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	svc := NewWithdrawService()
	_, err := svc.Create(1, &CreateWithdrawRequest{Amount: 10})
	if err == nil {
		t.Fatalf("expected error when withdraw is disabled")
	}
	if !IsClientError(err) {
		t.Fatalf("expected client error, got %T", err)
	}
	if err.Error() != "提现功能暂未开启" {
		t.Fatalf("error = %q, want %q", err.Error(), "提现功能暂未开启")
	}
}

func TestWithdrawServiceCreateRejectsUnsupportedAccountTypeBeforeDB(t *testing.T) {
	oldSettingsService := GlobalSettingsService
	defer func() {
		GlobalSettingsService = oldSettingsService
	}()

	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"withdraw_enabled":      {Key: "withdraw_enabled", Value: "true"},
			"withdraw_min_amount":   {Key: "withdraw_min_amount", Value: "10"},
			"withdraw_account_types": {Key: "withdraw_account_types", Value: "[\"bank\",\"alipay\"]"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	svc := NewWithdrawService()
	_, err := svc.Create(1, &CreateWithdrawRequest{
		Amount:      10,
		AccountType: "wechat",
		AccountName: "test",
		AccountNo:   "123456",
		RealName:    "测试用户",
	})
	if err == nil {
		t.Fatalf("expected error for unsupported account type")
	}
	if !IsClientError(err) {
		t.Fatalf("expected client error, got %T", err)
	}
	if err.Error() != "当前收款方式不可用" {
		t.Fatalf("error = %q, want %q", err.Error(), "当前收款方式不可用")
	}
}

func TestWithdrawServiceCreateRejectsAmountBelowMinBeforeDB(t *testing.T) {
	oldSettingsService := GlobalSettingsService
	defer func() {
		GlobalSettingsService = oldSettingsService
	}()

	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"withdraw_enabled":       {Key: "withdraw_enabled", Value: "true"},
			"withdraw_min_amount":    {Key: "withdraw_min_amount", Value: "20"},
			"withdraw_account_types": {Key: "withdraw_account_types", Value: "[\"bank\"]"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	svc := NewWithdrawService()
	_, err := svc.Create(1, &CreateWithdrawRequest{
		Amount:      10,
		AccountType: "bank",
		AccountName: "test",
		AccountNo:   "123456",
		RealName:    "测试用户",
	})
	if err == nil {
		t.Fatalf("expected error for amount below min")
	}
	if !IsClientError(err) {
		t.Fatalf("expected client error, got %T", err)
	}
	if err.Error() != "提现金额不能低于 20.00" {
		t.Fatalf("error = %q, want %q", err.Error(), "提现金额不能低于 20.00")
	}
}

func TestWithdrawServiceCreateRejectsIncompleteAccountInfoBeforeDB(t *testing.T) {
	oldSettingsService := GlobalSettingsService
	defer func() {
		GlobalSettingsService = oldSettingsService
	}()

	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"withdraw_enabled":       {Key: "withdraw_enabled", Value: "true"},
			"withdraw_min_amount":    {Key: "withdraw_min_amount", Value: "10"},
			"withdraw_account_types": {Key: "withdraw_account_types", Value: "[\"bank\"]"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	svc := NewWithdrawService()
	_, err := svc.Create(1, &CreateWithdrawRequest{
		Amount:      10,
		AccountType: "bank",
		AccountName: "",
		AccountNo:   "123456",
		RealName:    "测试用户",
	})
	if err == nil {
		t.Fatalf("expected error for incomplete account info")
	}
	if !IsClientError(err) {
		t.Fatalf("expected client error, got %T", err)
	}
	if err.Error() != "请完整填写收款信息" {
		t.Fatalf("error = %q, want %q", err.Error(), "请完整填写收款信息")
	}
}

func TestValidateWithdrawTextFieldRejectsTooLongValue(t *testing.T) {
	tooLong := ""
	for i := 0; i < 256; i++ {
		tooLong += "a"
	}

	_, err := validateWithdrawTextField(tooLong, "备注", 255)
	if err == nil {
		t.Fatalf("expected validation error for too long remark")
	}
	if !IsClientError(err) {
		t.Fatalf("expected client error, got %T", err)
	}
	if err.Error() != "备注不能超过255个字符" {
		t.Fatalf("error = %q, want %q", err.Error(), "备注不能超过255个字符")
	}
}
