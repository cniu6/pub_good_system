package routes

import (
	"fst/backend/pkg/presence"

	"github.com/gin-gonic/gin"
)

// registerWSRoutes 注册独立鉴权的 WebSocket 路由。
func registerWSRoutes(v1 *gin.RouterGroup) {
	v1.GET("/ws/presence", presence.HandlePresence)
}
