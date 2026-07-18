package appinit_test

import (
	"testing"

	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
	"fst/backend/pkg/middleware"
	"fst/backend/routes"

	"github.com/gin-gonic/gin"
)

// Bootstrap 会 Listen 相关之外的全量初始化；这里验证「迁移+路由」组装链在测试环境可工作，
// 不调用 Bootstrap() 以免启动定时任务/占用全局。
func TestAppInit_ChainPieces(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	if !db.IsSQLite() {
		t.Fatal("期望 sqlite")
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CorsMiddleware())
	routes.SetupRoutes(r)
	if r.Routes() == nil || len(r.Routes()) < 5 {
		t.Fatalf("路由过少: %d", len(r.Routes()))
	}
}
