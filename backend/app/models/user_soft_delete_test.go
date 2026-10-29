package models_test

import (
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

// TestSoftDeletedUserBlocksRegistrationLookup 验证软删除用户的 username/email 仍被唯一索引占用：
// 注册查重必须包含软删除记录，否则查重会放行、插入时撞唯一键并变成 500。
func TestSoftDeletedUserBlocksRegistrationLookup(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "soft_deleted_lookup")
	if u.Email == "" {
		t.Fatal("测试用户缺少 email，无法验证邮箱查重")
	}

	now := time.Now().Unix()
	if err := db.DB.Model(&models.User{}).Where("id = ?", u.ID).Updates(map[string]interface{}{
		"delete_time": now,
		"status":      0,
	}).Error; err != nil {
		t.Fatalf("软删除用户失败: %v", err)
	}

	// 旧的查重函数会忽略软删除用户。
	if _, err := models.GetUserByUsername(u.Username); err == nil {
		t.Fatal("GetUserByUsername 不应命中软删除用户")
	}
	if _, err := models.GetUserByEmail(u.Email); err == nil {
		t.Fatal("GetUserByEmail 不应命中软删除用户")
	}

	// 新的注册查重函数必须命中，与数据库唯一索引保持一致。
	if _, err := models.GetUserByUsernameIncludingDeleted(u.Username); err != nil {
		t.Fatalf("GetUserByUsernameIncludingDeleted 应命中软删除用户: %v", err)
	}
	if _, err := models.GetUserByEmailIncludingDeleted(u.Email); err != nil {
		t.Fatalf("GetUserByEmailIncludingDeleted 应命中软删除用户: %v", err)
	}

	// 直接插入同名用户仍会被唯一索引拒绝，证明查重放行后必然失败。
	dup := &models.User{
		Username: u.Username,
		Nickname: "duplicate",
		Email:    "another-" + u.Email,
		Password: "x",
		Role:     "user",
		Status:   1,
	}
	if err := models.CreateUser(dup); err == nil {
		t.Fatal("软删除用户的 username 仍被唯一索引占用，重复插入应失败")
	}
}
