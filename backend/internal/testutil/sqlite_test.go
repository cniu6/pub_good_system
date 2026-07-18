package testutil_test

import (
	"testing"

	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

func TestSetupSQLite(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	if !db.IsSQLite() {
		t.Fatal("期望 sqlite")
	}
	u := testutil.CreateTestUser(t, "tu1")
	if u.ID == 0 {
		t.Fatal("用户 ID 未回填")
	}
	a := testutil.CreateTestAdmin(t, "ta1")
	if a.Role != "admin" {
		t.Fatalf("role=%q", a.Role)
	}
}
