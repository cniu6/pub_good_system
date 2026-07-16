package ginweb

import (
	"fst/backend/controllers/admin"
	"fst/backend/controllers/user"
	"fst/backend/internal/env"
	"fst/backend/internal/middleware"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InitGin 初始化 Gin Web 服务器（草稿独立栈入口）
//
// 【已注释禁用说明】本包属于「平行草稿栈」，现网入口 main.go / main_embedded.go 不会调用本函数。
// 电商商品/分类/订单路由已整段注释，仅保留用户注册登录与资料相关路由便于对照留档。
// 详见：backend/留档.md →「电商半成品草稿（已注释禁用）」
func InitGin() {
	config := env.GetEnv()
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	setupRoutes(r)

	// 启动 HTTP 服务
	_ = r.Run(":" + strconv.Itoa(config.APIPort))
}

// setupRoutes 配置路由：公开接口与需登录接口拆开，避免注册/登录被鉴权挡住
func setupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	// ---------- 公开路由（无需登录）----------
	public := api.Group("/public")
	{
		// 【已注释禁用】电商半成品：公开商品/分类
		// public.GET("/products", user.GetList)
		// public.GET("/products/:id", user.GetDetail)
		// public.GET("/categories", user.GetCategoryList)
		// public.GET("/categories/:id", user.GetCategoryDetail)
		_ = public // 避免 public 未使用；恢复电商时删除本行并取消上方注释
	}

	// 注册/登录：放在鉴权中间件之外
	authPublic := api.Group("/user")
	{
		authPublic.POST("/register", user.Register)
		authPublic.POST("/login", user.Login)
	}

	// ---------- 用户路由（需要登录）----------
	userRoutes := api.Group("/user")
	userRoutes.Use(middleware.AuthMiddleware())
	{
		userRoutes.GET("/profile", user.GetProfile)
		userRoutes.PUT("/profile", user.UpdateProfile)
		userRoutes.PUT("/password", user.ChangePassword)

		// 【已注释禁用】电商半成品：卖家产品 / 买家订单
		// userRoutes.POST("/products", user.CreateProduct)
		// userRoutes.PUT("/products/:id", user.UpdateProduct)
		// userRoutes.DELETE("/products/:id", user.DeleteProduct)
		// userRoutes.POST("/orders", user.CreateOrder)
		// userRoutes.GET("/orders", user.GetMyOrders)
		// userRoutes.GET("/orders/:id", user.GetOrderDetail)
		// userRoutes.PUT("/orders/:id/cancel", user.CancelOrder)
	}

	// ---------- 管理员路由 ----------
	adminRoutes := api.Group("/admin")
	adminRoutes.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		// 用户管理（草稿栈内对照用，现网请用 app/controllers）
		adminRoutes.GET("/users", admin.GetUserList)
		adminRoutes.GET("/users/:id", admin.GetUserDetail)
		adminRoutes.PUT("/users/:id", admin.UpdateUser)
		adminRoutes.PUT("/users/:id/password", admin.ResetPassword)

		// 【已注释禁用】电商半成品：产品 / 分类 / 订单管理
		// adminRoutes.GET("/products", admin.GetProductList)
		// adminRoutes.GET("/products/:id", admin.GetProductDetail)
		// adminRoutes.POST("/products", admin.CreateProduct)
		// adminRoutes.PUT("/products/:id", admin.UpdateProduct)
		// adminRoutes.DELETE("/products/:id", admin.DeleteProduct)
		// adminRoutes.GET("/categories", admin.GetCategoryList)
		// adminRoutes.GET("/categories/:id", admin.GetCategoryDetail)
		// adminRoutes.POST("/categories", admin.CreateCategory)
		// adminRoutes.PUT("/categories/:id", admin.UpdateCategory)
		// adminRoutes.DELETE("/categories/:id", admin.DeleteCategory)
		// adminRoutes.GET("/orders", admin.GetOrderList)
		// adminRoutes.GET("/orders/:id", admin.GetOrderDetail)
		// adminRoutes.PUT("/orders/:id/status", admin.UpdateOrderStatus)
	}
}
