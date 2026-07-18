package admin_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fst/backend/app/controllers/admin"
	"fst/backend/internal/testutil"
	"fst/backend/routes"

	"github.com/gin-gonic/gin"
)

func TestAdminControllers_Construct(t *testing.T) {
	_ = admin.NewUserController()
	_ = admin.NewDashboardController()
	_ = admin.NewDebugController()
	_ = admin.NewSettingsController()
	_ = admin.NewPaymentController()
	_ = admin.NewWithdrawController()
	_ = admin.NewRealnameController()
	_ = admin.NewLogController()
	_ = admin.NewAPILogController()
	_ = admin.NewEmailTemplateController()
	_ = admin.NewEmailLogController()
	_ = admin.NewSMSLogController()
	_ = admin.NewUserMoneyScoreController()
	_ = admin.NewAutoJobController()
}

func TestAdminRoutes_RequireAuthViaSetupRoutes(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	routes.SetupRoutes(r)

	paths := []string{
		"/api/v1/admin/users",
		"/api/v1/admin/dashboard",
		"/api/v1/admin/settings",
		"/api/v1/admin/logs",
		"/api/v1/admin/api-logs",
		"/api/v1/admin/payments",
		"/api/v1/admin/withdraws",
	}
	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			r.ServeHTTP(w, req)
			var resp map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &resp)
			if w.Code == 200 && resp["code"] == float64(0) {
				t.Fatalf("无 token 访问 %s 不应成功: %s", path, w.Body.String())
			}
		})
	}
}
