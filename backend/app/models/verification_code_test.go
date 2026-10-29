package models_test

import (
	"fmt"
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

// 与 models.maxVerificationAttempts 保持一致（未导出常量，测试侧用同名语义常量对齐）。
const testMaxVerificationAttempts = 5

func TestConsumeVerificationCode_Success(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	contact := fmt.Sprintf("ok-%d@example.test", time.Now().UnixNano())
	if err := models.CreateVerificationCode(contact, "123456", "register", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateVerificationCode: %v", err)
	}

	used, err := models.ConsumeVerificationCode(contact, "123456", "register")
	if err != nil || !used {
		t.Fatalf("正确验证码应消费成功: used=%v err=%v", used, err)
	}

	// 已消费后不可再消费（幂等防重放）
	used, err = models.ConsumeVerificationCode(contact, "123456", "register")
	if err != nil {
		t.Fatalf("二次消费不应报错: %v", err)
	}
	if used {
		t.Fatal("已使用的验证码不应再次消费成功")
	}
}

func TestConsumeVerificationCode_WrongCodeIncrementsAttempts(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	contact := fmt.Sprintf("wrong-%d@example.test", time.Now().UnixNano())
	if err := models.CreateVerificationCode(contact, "654321", "register", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateVerificationCode: %v", err)
	}

	used, err := models.ConsumeVerificationCode(contact, "000000", "register")
	if err != nil {
		t.Fatalf("错误码不应报错: %v", err)
	}
	if used {
		t.Fatal("错误码不应消费成功")
	}

	vc, err := models.GetValidVerificationCode(contact, "register")
	if err != nil {
		t.Fatalf("失败 1 次后验证码仍应有效: %v", err)
	}
	if vc.Attempts != 1 {
		t.Fatalf("attempts=%d, want 1", vc.Attempts)
	}

	// 未达上限前，正确码仍可消费
	used, err = models.ConsumeVerificationCode(contact, "654321", "register")
	if err != nil || !used {
		t.Fatalf("失败未达上限时正确码应成功: used=%v err=%v", used, err)
	}
}

func TestConsumeVerificationCode_BurnAfterMaxFailures(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	contact := fmt.Sprintf("burn-%d@example.test", time.Now().UnixNano())
	const realCode = "112233"
	if err := models.CreateVerificationCode(contact, realCode, "register", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateVerificationCode: %v", err)
	}

	// 连续猜错达到上限：码应被软删作废
	for i := 1; i <= testMaxVerificationAttempts; i++ {
		used, err := models.ConsumeVerificationCode(contact, "999999", "register")
		if err != nil {
			t.Fatalf("第 %d 次错误码不应报错: %v", i, err)
		}
		if used {
			t.Fatalf("第 %d 次错误码不应消费成功", i)
		}
	}

	// 作废后不应再能取到有效码
	if _, err := models.GetValidVerificationCode(contact, "register"); err == nil {
		t.Fatal("达到失败上限后仍能取到有效验证码")
	}

	var attempts, isDeleted, isUsed int
	if err := db.DB.Raw(`SELECT attempts FROM verification_codes WHERE contact = ? AND code = ?`, contact, realCode).Scan(&attempts).Error; err != nil {
		t.Fatalf("读 attempts 失败: %v", err)
	}
	if err := db.DB.Raw(`SELECT is_deleted FROM verification_codes WHERE contact = ? AND code = ?`, contact, realCode).Scan(&isDeleted).Error; err != nil {
		t.Fatalf("读 is_deleted 失败: %v", err)
	}
	if err := db.DB.Raw(`SELECT is_used FROM verification_codes WHERE contact = ? AND code = ?`, contact, realCode).Scan(&isUsed).Error; err != nil {
		t.Fatalf("读 is_used 失败: %v", err)
	}
	if attempts != testMaxVerificationAttempts {
		t.Fatalf("attempts=%d, want %d", attempts, testMaxVerificationAttempts)
	}
	if isDeleted != 1 {
		t.Fatalf("is_deleted=%d, want 1（应被作废）", isDeleted)
	}
	if isUsed != 0 {
		t.Fatalf("is_used=%d, want 0（失败作废不是标记已使用）", isUsed)
	}

	// 即便事后猜对，也必须失败——攻击者不能靠「最后一把蒙对」绕过上限
	used, err := models.ConsumeVerificationCode(contact, realCode, "register")
	if err != nil {
		t.Fatalf("作废后正确码不应报错: %v", err)
	}
	if used {
		t.Fatal("作废后正确码不应再被消费成功")
	}
}

func TestConsumeVerificationCode_FailJustBelowLimitStillWorks(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	contact := fmt.Sprintf("below-%d@example.test", time.Now().UnixNano())
	const realCode = "445566"
	if err := models.CreateVerificationCode(contact, realCode, "reset_password", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateVerificationCode: %v", err)
	}

	// 失败次数 = 上限 - 1，码仍有效，正确码应成功
	for i := 1; i < testMaxVerificationAttempts; i++ {
		used, err := models.ConsumeVerificationCode(contact, "000000", "reset_password")
		if err != nil || used {
			t.Fatalf("第 %d 次错误码异常: used=%v err=%v", i, used, err)
		}
	}

	vc, err := models.GetValidVerificationCode(contact, "reset_password")
	if err != nil {
		t.Fatalf("未达上限应仍有效: %v", err)
	}
	if vc.Attempts != testMaxVerificationAttempts-1 {
		t.Fatalf("attempts=%d, want %d", vc.Attempts, testMaxVerificationAttempts-1)
	}

	used, err := models.ConsumeVerificationCode(contact, realCode, "reset_password")
	if err != nil || !used {
		t.Fatalf("失败未达上限时正确码应成功: used=%v err=%v", used, err)
	}
}

func TestConsumeVerificationCode_NoValidCodeDoesNotError(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	contact := fmt.Sprintf("none-%d@example.test", time.Now().UnixNano())
	used, err := models.ConsumeVerificationCode(contact, "123456", "register")
	if err != nil {
		t.Fatalf("无有效码时不应报错: %v", err)
	}
	if used {
		t.Fatal("无有效码时不应消费成功")
	}
}

func TestConsumeVerificationCode_NewCodeResetsAttempts(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	contact := fmt.Sprintf("reset-%d@example.test", time.Now().UnixNano())
	if err := models.CreateVerificationCode(contact, "111111", "register", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateVerificationCode 旧码: %v", err)
	}
	// 先失败几次
	for i := 0; i < 3; i++ {
		_, _ = models.ConsumeVerificationCode(contact, "000000", "register")
	}

	// 重新发码会软删旧码，新码 attempts 从 0 开始
	if err := models.CreateVerificationCode(contact, "222222", "register", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateVerificationCode 新码: %v", err)
	}
	vc, err := models.GetValidVerificationCode(contact, "register")
	if err != nil {
		t.Fatalf("GetValidVerificationCode: %v", err)
	}
	if vc.Code != "222222" {
		t.Fatalf("应拿到新码, got=%s", vc.Code)
	}
	if vc.Attempts != 0 {
		t.Fatalf("新码 attempts 应为 0, got=%d", vc.Attempts)
	}

	used, err := models.ConsumeVerificationCode(contact, "222222", "register")
	if err != nil || !used {
		t.Fatalf("新码应可消费: used=%v err=%v", used, err)
	}
}

