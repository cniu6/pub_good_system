package services

import (
	"fst/backend/app/models"
	"fst/backend/internal/testutil"
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

// TestWithdrawServiceCreateRejectsWithoutApprovedRealnameWhenRequired 开启 withdraw_require_realname 后，
// 用户尚未完成实名认证（或未通过审核）时，提现申请应被拒绝。
func TestWithdrawServiceCreateRejectsWithoutApprovedRealnameWhenRequired(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	oldSettingsService := GlobalSettingsService
	defer func() {
		GlobalSettingsService = oldSettingsService
	}()
	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"withdraw_enabled":          {Key: "withdraw_enabled", Value: "true"},
			"withdraw_require_realname": {Key: "withdraw_require_realname", Value: "true"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	user := testutil.CreateTestUser(t, "withdraw-no-realname")

	svc := NewWithdrawService()
	_, err := svc.Create(user.ID, &CreateWithdrawRequest{
		Amount:      10,
		AccountType: "bank",
		AccountName: "test",
		AccountNo:   "123456",
		RealName:    "测试用户",
	})
	if err == nil {
		t.Fatalf("期望未实名认证时提现被拒绝")
	}
	if !IsClientError(err) {
		t.Fatalf("期望 client error，实际=%T", err)
	}
	if err.Error() != "请先完成实名认证并通过审核后再提现" {
		t.Fatalf("error = %q, want %q", err.Error(), "请先完成实名认证并通过审核后再提现")
	}
}

// TestWithdrawServiceCreateAllowsWithApprovedRealnameWhenRequired 开启 withdraw_require_realname 后，
// 用户已有「已通过」的实名认证记录时，不应再被实名检查拦截（后续能正常走到提现创建流程）。
func TestWithdrawServiceCreateAllowsWithApprovedRealnameWhenRequired(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	oldSettingsService := GlobalSettingsService
	defer func() {
		GlobalSettingsService = oldSettingsService
	}()
	GlobalSettingsService = &SettingsService{
		cache: map[string]*models.SystemSetting{
			"withdraw_enabled":          {Key: "withdraw_enabled", Value: "true"},
			"withdraw_require_realname": {Key: "withdraw_require_realname", Value: "true"},
		},
		cacheTime: time.Now(),
		ttl:       time.Hour,
	}

	user := testutil.CreateTestUser(t, "withdraw-approved-realname")
	if err := models.CreateRealnameVerification(&models.RealnameVerification{
		UserID:          user.ID,
		RealName:        "测试用户",
		CertificateType: 1,
		CertificateNo:   "110101199001010011",
		Status:          RealnameStatusApproved,
	}); err != nil {
		t.Fatalf("创建实名认证记录失败: %v", err)
	}

	svc := NewWithdrawService()
	result, err := svc.Create(user.ID, &CreateWithdrawRequest{
		Amount:      10,
		AccountType: "bank",
		AccountName: "test",
		AccountNo:   "123456",
		RealName:    "测试用户",
	})
	if err != nil {
		t.Fatalf("期望已实名认证通过时提现成功创建，实际报错: %v", err)
	}
	if result == nil {
		t.Fatalf("期望返回提现记录")
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
