package routes_test

import (
	"net/http"
	"testing"

	"fst/backend/internal/testutil"
	"fst/backend/routes"

	"github.com/gin-gonic/gin"
)

// TestAdminUserDetailNotFound 不存在的用户必须返回 404。
// GORM Scan 查不到行不会报错，若直接用 Scan 会返回 200 + 空用户，管理端无法区分“不存在”和“数据为空”。
func TestAdminUserDetailNotFound(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	_, token := setupAdminSelfProtection(t, "user_detail_admin")

	r := gin.New()
	routes.SetupRoutes(r)
	w := performAdminJSON(t, r, http.MethodGet, "/api/v1/admin/users/999999", token, "")

	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的用户应返回 404，实际 HTTP %d: %s", w.Code, w.Body.String())
	}
}
