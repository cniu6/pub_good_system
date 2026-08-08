package routes

import (
	v2test "fst/backend/app/controllers/v2/test"

	"github.com/gin-gonic/gin"
)

// registerV2Routes 注册 V2 版本路由（用于多版本 API 验证）
func registerV2Routes(group *gin.RouterGroup) {
	v2TestCtrl := v2test.NewTestController()
	v2TestCtrl.RegisterRoutes(group)
}
