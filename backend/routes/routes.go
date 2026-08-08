package routes

// routes.go 是路由文件，负责将各个控制器注册到路由中。
import (
	"net/http"
	"time"

	"fst/backend/app/controllers"
	"fst/backend/app/controllers/admin"
	"fst/backend/app/controllers/public"
	"fst/backend/app/controllers/user"
	_ "fst/backend/docs"
	"fst/backend/pkg/config"
	"fst/backend/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// 控制器实例（延迟初始化，避免包加载期循环依赖）
var (
	publicAuthCtrl            *public.AuthController            //登录注册
	publicSettingsCtrl        *public.SettingsController        //系统配置
	publicGeoCtrl             *public.GeoController             //地理/区号探测
	publicSessionCtrl         *public.SessionController         //会话强退清理
	publicPaymentCallbackCtrl *public.PaymentCallbackController //支付回调
	userProfileCtrl           *user.ProfileController           //用户资料
	userPaymentCtrl           *user.PaymentController           //用户支付
	userRealnameCtrl          *user.RealnameController          //实名认证
	userWithdrawCtrl          *user.WithdrawController          //用户提现
	systemCtrl                *controllers.SystemController
	adminUserCtrl             *admin.UserController   //用户管理
	adminLogCtrl              *admin.LogController    //日志
	adminAPILogCtrl           *admin.APILogController //API日志
	adminEmailTplCtrl         *admin.EmailTemplateController
	adminSMSTplCtrl           *admin.SMSTemplateController //短信模板
	adminEmailLogCtrl         *admin.EmailLogController    //邮件日志
	adminSettingsCtrl         *admin.SettingsController    //系统设置
	adminDebugCtrl            *admin.DebugController       //调试
	adminMoneyScoreCtrl       *admin.UserMoneyScoreController
	adminPaymentCtrl          *admin.PaymentController      //支付
	adminRealnameCtrl         *admin.RealnameController     //实名认证
	adminWithdrawCtrl         *admin.WithdrawController     //提现
	adminSMSLogCtrl           *admin.SMSLogController       //短信日志
	adminDashboardCtrl        *admin.DashboardController    //仪表盘
	adminAutoJobCtrl          *admin.AutoJobController      //自动任务
	adminOnlineCtrl           *admin.OnlineController       //在线用户
	adminAnnouncementCtrl     *admin.AnnouncementController //站内公告
	userAnnouncementCtrl      *user.AnnouncementController  //用户公告
	adminProfileCtrl          *admin.ProfileController
	adminTodoCtrl             *admin.TodoController
	adminDBConsoleCtrl        *admin.DBConsoleController
	adminTerminalCtrl         *admin.TerminalController
	adminUserLevelCtrl        *admin.UserLevelController
)

func initControllers() {
	publicAuthCtrl = public.NewAuthController()
	publicSettingsCtrl = public.NewSettingsController()
	publicGeoCtrl = public.NewGeoController()
	publicSessionCtrl = public.NewSessionController()
	publicPaymentCallbackCtrl = public.NewPaymentCallbackController()
	userProfileCtrl = user.NewProfileController()
	userPaymentCtrl = user.NewPaymentController()
	userRealnameCtrl = user.NewRealnameController()
	userWithdrawCtrl = user.NewWithdrawController()
	systemCtrl = &controllers.SystemController{}
	adminUserCtrl = admin.NewUserController()
	adminLogCtrl = admin.NewLogController()
	adminAPILogCtrl = admin.NewAPILogController()
	adminEmailTplCtrl = admin.NewEmailTemplateController()
	adminSMSTplCtrl = admin.NewSMSTemplateController()
	adminEmailLogCtrl = admin.NewEmailLogController()
	adminSettingsCtrl = admin.NewSettingsController()
	adminDebugCtrl = admin.NewDebugController()
	adminMoneyScoreCtrl = admin.NewUserMoneyScoreController()
	adminPaymentCtrl = admin.NewPaymentController()
	adminRealnameCtrl = admin.NewRealnameController()
	adminWithdrawCtrl = admin.NewWithdrawController()
	adminDashboardCtrl = admin.NewDashboardController()
	adminSMSLogCtrl = admin.NewSMSLogController()
	adminAutoJobCtrl = admin.NewAutoJobController()
	adminOnlineCtrl = admin.NewOnlineController()
	adminAnnouncementCtrl = admin.NewAnnouncementController()
	userAnnouncementCtrl = user.NewAnnouncementController()
	adminProfileCtrl = admin.NewProfileController()
	adminTodoCtrl = admin.NewTodoController()
	adminDBConsoleCtrl = admin.NewDBConsoleController()
	adminTerminalCtrl = admin.NewTerminalController()
	adminUserLevelCtrl = admin.NewUserLevelController()
}

// SetupRoutes 汇总注册全部 HTTP 路由（详情拆在 public/user/admin 文件）。
func SetupRoutes(router *gin.Engine) {
	initControllers()

	// Scalar API 文档：注解仍写 /api/v1/admin/*；返回 openapi.json 时按 ADMIN_API_PATH 改写
	if cfg := config.GlobalConfig; cfg != nil && cfg.EnableSwagger {
		mountScalar(router)
	}

	api := router.Group("/api")
	v1 := api.Group("/v1")

	registerPublicRoutes(v1) //公开接口
	registerUserRoutes(v1)   //用户接口
	registerAdminRoutes(v1)  //管理员接口
	registerWSRoutes(v1)     //WebSocket（自行鉴权，不挂 HTTP AuthMiddleware）

	// 健康检查：轻量、不碰库（负载均衡探活）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"status":    "ok",
				"message":   "健康检查成功",
				"timestamp": time.Now().Format("2006-01-02 15:04:05"),
			},
		})
	})
}

// mountScalar 挂载 Scalar API 文档入口和管理端路径自适应的 openapi.json。
// 入口：GET /scalar；OpenAPI 文档：GET /scalar/openapi.json。
func mountScalar(router *gin.Engine) {
	// 管理端路径自适应的中间件：仅对 /scalar/openapi.json 返回改写后的 doc.json
	router.GET("/scalar/openapi.json", middleware.ScalarAdminPathRewriteMiddleware())

	// 托管 Scalar 本地前端静态资源
	router.Static("/scalar-static", "./backend/static/scalar/dist/browser")

	// Scalar API 文档主页面：直接返回内嵌 Scalar 的 HTML，指定 spec URL 为同源 openapi.json
	router.GET("/scalar", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, scalarReferenceHTML())
	})
}

// scalarReferenceHTML 生成 Scalar API Reference 的 HTML 页面。
// 使用本地 @scalar/api-reference 静态资源，并指定 data-url 加载后端 OpenAPI 文档。
func scalarReferenceHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FST Platform API</title>
    <style>
        body { margin: 0; padding: 0; }
        #app { height: 100vh; }
    </style>
</head>
<body>
    <script
      id="api-reference"
      data-url="/scalar/openapi.json"
      data-configuration='{"theme":"default","layout":"modern","darkMode":true,"showSidebar":true,"agent":{"disabled":true}}'
      src="/scalar-static/standalone.js"></script>
</body>
</html>`
}
