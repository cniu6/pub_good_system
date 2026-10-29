package services

import (
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

// TestUpdateCannotDemoteLastAdmin 唯一启用管理员不能被降级，
// 否则系统将失去最后一个可管理账号。
func TestUpdateCannotDemoteLastAdmin(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	admin := testutil.CreateTestAdmin(t, "only-admin")
	role := "user"

	err := NewUserService().Update(&UserUpdateRequest{ID: admin.ID, Role: &role})
	if err == nil {
		t.Fatal("降级最后一个启用管理员必须被拒绝")
	}

	var got string
	if err := db.DB.Model(&models.User{}).Where("id = ?", admin.ID).Pluck("role", &got).Error; err != nil {
		t.Fatalf("读取角色失败: %v", err)
	}
	if got != "admin" {
		t.Fatalf("拒绝后角色不应变化，got=%q", got)
	}
}

// TestUpdateCanDemoteAdminWhenAnotherAdminExists 还有其它启用管理员时，
// 降级应当放行，避免保护逻辑误伤正常操作。
func TestUpdateCanDemoteAdminWhenAnotherAdminExists(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	adminA := testutil.CreateTestAdmin(t, "admin-a")
	_ = testutil.CreateTestAdmin(t, "admin-b")
	role := "user"

	if err := NewUserService().Update(&UserUpdateRequest{ID: adminA.ID, Role: &role}); err != nil {
		t.Fatalf("存在其他启用管理员时应允许降级: %v", err)
	}
}
