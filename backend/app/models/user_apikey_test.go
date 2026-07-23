package models_test

import (
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

// TestResetUserApiKeyStoresHashNotPlaintext 验证重置 API Key 后：
// 1. 返回值是明文（一次性可见）；
// 2. 数据库 apikey 列存的是哈希，不等于明文；
// 3. 用返回的明文可以通过 GetUserByApiKey 查回同一用户。
func TestResetUserApiKeyStoresHashNotPlaintext(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "apikey_hash_user")

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
	if storedApikey == plain {
		t.Fatalf("apikey 列不应存明文，实际=%q", storedApikey)
	}
	if len(storedApikey) != 64 {
		t.Fatalf("apikey 列应为 SHA256 十六进制(64位)，实际长度=%d", len(storedApikey))
	}

	if err := db.DB.Raw("SELECT apikey_hint FROM users WHERE id = ?", u.ID).Scan(&storedHint).Error; err != nil {
		t.Fatalf("查询 apikey_hint 列失败: %v", err)
	}
	if storedHint != plain[len(plain)-4:] {
		t.Fatalf("apikey_hint 期望=%q 实际=%q", plain[len(plain)-4:], storedHint)
	}

	found, err := models.GetUserByApiKey(plain)
	if err != nil {
		t.Fatalf("GetUserByApiKey(明文) 应命中: %v", err)
	}
	if found.ID != u.ID {
		t.Fatalf("GetUserByApiKey 命中用户ID=%d，期望=%d", found.ID, u.ID)
	}

	masked := found.MaskedApikey()
	want := "********" + plain[len(plain)-4:]
	if masked != want {
		t.Fatalf("MaskedApikey()=%q，期望=%q", masked, want)
	}
}

// TestGetUserByApiKeyLegacyPlaintextCompat 验证升级前遗留的明文 API Key：
// 首次查询仍能命中（回退按明文匹配），并会被自动回写为哈希，后续该库里不再是明文。
func TestGetUserByApiKeyLegacyPlaintextCompat(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "apikey_legacy_user")

	legacyPlain := "legacyplaintextapikey1234567890abcdef01"
	if err := db.DB.Exec("UPDATE users SET apikey = ?, apikey_hint = NULL WHERE id = ?", legacyPlain, u.ID).Error; err != nil {
		t.Fatalf("模拟写入历史明文 apikey 失败: %v", err)
	}

	found, err := models.GetUserByApiKey(legacyPlain)
	if err != nil {
		t.Fatalf("GetUserByApiKey(历史明文) 应命中: %v", err)
	}
	if found.ID != u.ID {
		t.Fatalf("命中用户ID=%d，期望=%d", found.ID, u.ID)
	}

	var storedApikey string
	if err := db.DB.Raw("SELECT apikey FROM users WHERE id = ?", u.ID).Scan(&storedApikey).Error; err != nil {
		t.Fatalf("查询 apikey 列失败: %v", err)
	}
	if storedApikey == legacyPlain {
		t.Fatalf("命中一次后应已回写为哈希，实际仍是明文")
	}

	found2, err := models.GetUserByApiKey(legacyPlain)
	if err != nil {
		t.Fatalf("回写后 GetUserByApiKey 应仍能命中: %v", err)
	}
	if found2.ID != u.ID {
		t.Fatalf("回写后命中用户ID=%d，期望=%d", found2.ID, u.ID)
	}
}
