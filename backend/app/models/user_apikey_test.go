package models_test

import (
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

// TestResetUserApiKeyStoresPlaintext 验证重置后：库内明文可回显，鉴权可用，管理端掩码正常。
func TestResetUserApiKeyStoresPlaintext(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "apikey_plain_user")

	plain, err := models.ResetUserApiKey(u.ID)
	if err != nil {
		t.Fatalf("ResetUserApiKey: %v", err)
	}
	if plain == "" {
		t.Fatalf("ResetUserApiKey 应返回非空明文")
	}

	var storedApikey, storedHint string
	if err := db.DB.Raw("SELECT apikey FROM users WHERE id = ?", u.ID).Scan(&storedApikey).Error; err != nil {
		t.Fatalf("查询 apikey 列失败: %v", err)
	}
	if storedApikey != plain {
		t.Fatalf("apikey 列应存明文，期望=%q 实际=%q", plain, storedApikey)
	}
	if err := db.DB.Raw("SELECT apikey_hint FROM users WHERE id = ?", u.ID).Scan(&storedHint).Error; err != nil {
		t.Fatalf("查询 apikey_hint 失败: %v", err)
	}
	if storedHint != plain[len(plain)-4:] {
		t.Fatalf("apikey_hint 期望=%q 实际=%q", plain[len(plain)-4:], storedHint)
	}

	found, err := models.GetUserByApiKey(plain)
	if err != nil {
		t.Fatalf("GetUserByApiKey(明文) 应命中: %v", err)
	}
	if found.ID != u.ID {
		t.Fatalf("命中用户ID=%d，期望=%d", found.ID, u.ID)
	}
	if got := found.PlainApikeyForOwner(); got != plain {
		t.Fatalf("PlainApikeyForOwner=%q，期望=%q", got, plain)
	}
	wantMasked := "********" + plain[len(plain)-4:]
	if got := found.MaskedApikey(); got != wantMasked {
		t.Fatalf("MaskedApikey=%q，期望=%q", got, wantMasked)
	}
}

// TestRepairHashedApiKeysRotatesLegacyHash 启动补丁会把 64 位 hex 旧哈希重置为新明文。
func TestRepairHashedApiKeysRotatesLegacyHash(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "apikey_hash_legacy")
	oldHash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err := db.DB.Exec("UPDATE users SET apikey = ?, apikey_hint = ? WHERE id = ?", oldHash, "cdef", u.ID).Error; err != nil {
		t.Fatalf("写入旧哈希失败: %v", err)
	}

	models.RepairHashedApiKeys()

	var stored string
	if err := db.DB.Raw("SELECT apikey FROM users WHERE id = ?", u.ID).Scan(&stored).Error; err != nil {
		t.Fatalf("查询 apikey 失败: %v", err)
	}
	if stored == "" || stored == oldHash || len(stored) != 40 {
		t.Fatalf("应重置为 40 位明文，实际=%q", stored)
	}
	found, err := models.GetUserByApiKey(stored)
	if err != nil || found.ID != u.ID {
		t.Fatalf("新明文应可鉴权: err=%v", err)
	}
}
