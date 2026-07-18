package routes

import "github.com/gin-gonic/gin"

// registerPublicRoutes 公开接口（无需登录）
func registerPublicRoutes(v1 *gin.RouterGroup) {
	publicGroup := v1.Group("/public")
	{
		publicAuthCtrl.RegisterRoutes(publicGroup)            //登录注册
		publicSettingsCtrl.RegisterRoutes(publicGroup)        //系统配置
		publicPaymentCallbackCtrl.RegisterRoutes(publicGroup) //支付回调
	}
}
