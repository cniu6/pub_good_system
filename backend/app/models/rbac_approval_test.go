package models_test

import (
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
)

func TestRBAC_SeedAndAssign(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	roles, err := models.ListRoles()
	if err != nil {
		t.Fatalf("ListRoles: %v", err)
	}
	if len(roles) < 3 {
		t.Fatalf("期望至少 3 个种子角色，实际 %d", len(roles))
	}

	perms, err := models.ListPermissions()
	if err != nil {
		t.Fatalf("ListPermissions: %v", err)
	}
	if len(perms) < 5 {
		t.Fatalf("期望至少 5 个权限点，实际 %d", len(perms))
	}

	adminRole, err := models.GetRoleByCode("admin")
	if err != nil || adminRole == nil {
		t.Fatalf("GetRoleByCode admin: %v", err)
	}
	viewerRole, err := models.GetRoleByCode("viewer")
	if err != nil || viewerRole == nil {
		t.Fatalf("GetRoleByCode viewer: %v", err)
	}

	u := testutil.CreateTestUser(t, "rbac-user-1")
	if err := models.AssignUserRole(u.ID, viewerRole.ID); err != nil {
		t.Fatalf("AssignUserRole viewer: %v", err)
	}

	ok, err := models.UserHasPermissionCode(u.ID, "user:read")
	if err != nil || !ok {
		t.Fatalf("viewer 应有 user:read: ok=%v err=%v", ok, err)
	}
	ok, err = models.UserHasPermissionCode(u.ID, "finance:write")
	if err != nil {
		t.Fatalf("UserHasPermissionCode: %v", err)
	}
	if ok {
		t.Fatal("viewer 不应有 finance:write")
	}

	// 替换为 operator
	op, err := models.GetRoleByCode("operator")
	if err != nil {
		t.Fatalf("GetRoleByCode operator: %v", err)
	}
	if err := models.AssignUserRole(u.ID, op.ID); err != nil {
		t.Fatalf("AssignUserRole operator: %v", err)
	}
	ok, err = models.UserHasPermissionCode(u.ID, "payment:write")
	if err != nil || !ok {
		t.Fatalf("operator 应有 payment:write: ok=%v err=%v", ok, err)
	}

	list, err := models.ListUserRoles(u.ID)
	if err != nil || len(list) != 1 || list[0].Code != "operator" {
		t.Fatalf("ListUserRoles: %+v err=%v", list, err)
	}
}

func TestApprovalRequest_CreateAndApprove(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	payload, err := models.MarshalApprovalPayload(map[string]interface{}{
		"order_id": uint64(42),
		"memo":     "test",
	})
	if err != nil {
		t.Fatalf("MarshalApprovalPayload: %v", err)
	}
	req := &models.ApprovalRequest{
		Type:        models.ApprovalTypeForcePaymentComplete,
		PayloadJSON: payload,
		Status:      models.ApprovalStatusPending,
		RequesterID: 1,
		Comment:     "force",
	}
	if err := models.CreateApprovalRequest(req); err != nil {
		t.Fatalf("CreateApprovalRequest: %v", err)
	}
	if req.ID == 0 {
		t.Fatal("期望生成审批 ID")
	}

	got, err := models.GetApprovalRequestByID(req.ID)
	if err != nil || got.Status != models.ApprovalStatusPending {
		t.Fatalf("GetApprovalRequestByID: %+v err=%v", got, err)
	}

	if err := models.ApproveApprovalRequest(req.ID, 2, "ok"); err != nil {
		t.Fatalf("ApproveApprovalRequest: %v", err)
	}
	got, err = models.GetApprovalRequestByID(req.ID)
	if err != nil || got.Status != models.ApprovalStatusApproved || got.ReviewerID == nil || *got.ReviewerID != 2 {
		t.Fatalf("approved state: %+v err=%v", got, err)
	}
}
