package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fst/backend/app/controllers"
	"fst/backend/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestSystemController_Exists(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()
	gin.SetMode(gin.TestMode)

	ctrl := &controllers.SystemController{}
	_ = ctrl

	// SystemController 若有公开路由，在 routes 里挂；这里验证类型可实例化
	r := gin.New()
	r.GET("/ping-system", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": 0, "message": "ok"})
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping-system", nil)
	r.ServeHTTP(w, req)
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != float64(0) {
		t.Fatal(resp)
	}
}
