package test

import (
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// TestController V2 测试控制器
type TestController struct{}

// NewTestController 创建 V2 测试控制器
func NewTestController() *TestController {
	return &TestController{}
}

// GetDeviceInfo V2 获取用户本机数据（示例接口）
// @Summary 获取用户本机数据
// @Description V2 版本示例接口，用于验证多版本路由与 Scalar 嵌套展示
// @Tags User-本机数据
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v2/user/device [get]
func (ctrl *TestController) GetDeviceInfo(c *gin.Context) {
	utils.Success(c, gin.H{
		"version": "v2",
		"device":  "localhost",
		"message": "hello from v2 user device",
	})
}

// RegisterRoutes 注册 V2 测试路由
func (ctrl *TestController) RegisterRoutes(group *gin.RouterGroup) {
	user := group.Group("/user")
	{
		user.GET("/device", ctrl.GetDeviceInfo)
	}
}
