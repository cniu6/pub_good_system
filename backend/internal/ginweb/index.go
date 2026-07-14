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
		public.GET("/products", user.GetList)
		public.GET("/products/:id", user.GetDetail)
		// 分类公开列表
		public.GET("/categories", user.GetCategoryList)
		public.GET("/categories/:id", user.GetCategoryDetail)
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

		// 产品管理（C2C 卖家）
		userRoutes.POST("/products", user.CreateProduct)
		userRoutes.PUT("/products/:id", user.UpdateProduct)
		userRoutes.DELETE("/products/:id", user.DeleteProduct)

		// 订单：买家最小可用 CRUD
		userRoutes.POST("/orders", user.CreateOrder)
		userRoutes.GET("/orders", user.GetMyOrders)
		userRoutes.GET("/orders/:id", user.GetOrderDetail)
		userRoutes.PUT("/orders/:id/cancel", user.CancelOrder)
	}

	// ---------- 管理员路由 ----------
	adminRoutes := api.Group("/admin")
	adminRoutes.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		// 用户管理
		adminRoutes.GET("/users", admin.GetUserList)
		adminRoutes.GET("/users/:id", admin.GetUserDetail)
		adminRoutes.PUT("/users/:id", admin.UpdateUser)
		adminRoutes.PUT("/users/:id/password", admin.ResetPassword)

		// 产品管理
		adminRoutes.GET("/products", admin.GetProductList)
		adminRoutes.GET("/products/:id", admin.GetProductDetail)
		adminRoutes.POST("/products", admin.CreateProduct)
		adminRoutes.PUT("/products/:id", admin.UpdateProduct)
		adminRoutes.DELETE("/products/:id", admin.DeleteProduct)

		// 分类管理
		adminRoutes.GET("/categories", admin.GetCategoryList)
		adminRoutes.GET("/categories/:id", admin.GetCategoryDetail)
		adminRoutes.POST("/categories", admin.CreateCategory)
		adminRoutes.PUT("/categories/:id", admin.UpdateCategory)
		adminRoutes.DELETE("/categories/:id", admin.DeleteCategory)

		// 订单管理
		adminRoutes.GET("/orders", admin.GetOrderList)
		adminRoutes.GET("/orders/:id", admin.GetOrderDetail)
		adminRoutes.PUT("/orders/:id/status", admin.UpdateOrderStatus)
	}
}
