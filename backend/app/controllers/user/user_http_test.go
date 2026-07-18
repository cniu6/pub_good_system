package user_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fst/backend/internal/testutil"
	"fst/backend/routes"

	"github.com/gin-gonic/gin"
)

func TestUserRoutes_RequireAuthViaSetupRoutes(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	routes.SetupRoutes(r)

	paths := []string{
		"/api/v1/user/profile",
		"/api/v1/user/payment/orders",
		"/api/v1/user/payment/gateways",
		"/api/v1/user/withdraw",
		"/api/v1/user/realname",
		"/api/v1/system/cleanup-status",
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
