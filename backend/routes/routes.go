package routes

// routes.go 是路由文件，负责将各个控制器注册到路由中。
import (
	"net/http"
	"time"

	adminannouncement "fst/backend/app/controllers/admin/announcement"
	adminapilog "fst/backend/app/controllers/admin/apilog"
	adminautojob "fst/backend/app/controllers/admin/autojob"
	admindashboard "fst/backend/app/controllers/admin/dashboard"
	admindbconsole "fst/backend/app/controllers/admin/dbconsole"
	admindebug "fst/backend/app/controllers/admin/debug"
	adminemaillog "fst/backend/app/controllers/admin/emaillog"
	adminemailtemplate "fst/backend/app/controllers/admin/emailtemplate"
	adminlog "fst/backend/app/controllers/admin/log"
	adminonline "fst/backend/app/controllers/admin/online"
	adminpayment "fst/backend/app/controllers/admin/payment"
	adminprofile "fst/backend/app/controllers/admin/profile"
	adminrealname "fst/backend/app/controllers/admin/realname"
	adminsettings "fst/backend/app/controllers/admin/settings"
	adminsmslog "fst/backend/app/controllers/admin/smslog"
	adminsmstemplate "fst/backend/app/controllers/admin/smstemplate"
	adminterminal "fst/backend/app/controllers/admin/terminal"
	admintodo "fst/backend/app/controllers/admin/todo"
	adminuser "fst/backend/app/controllers/admin/user"
	adminuserlevel "fst/backend/app/controllers/admin/userlevel"
	adminusermoney "fst/backend/app/controllers/admin/usermoney"
	adminwithdraw "fst/backend/app/controllers/admin/withdraw"
	publicauth "fst/backend/app/controllers/public/auth"
	publicgeo "fst/backend/app/controllers/public/geo"
	publicpayment "fst/backend/app/controllers/public/payment"
	publicsession "fst/backend/app/controllers/public/session"
	publicsettings "fst/backend/app/controllers/public/settings"
	systemctrl "fst/backend/app/controllers/system/system"
	userannouncement "fst/backend/app/controllers/user/announcement"
	userpayment "fst/backend/app/controllers/user/payment"
	userprofile "fst/backend/app/controllers/user/profile"
	userrealname "fst/backend/app/controllers/user/realname"
	userwithdraw "fst/backend/app/controllers/user/withdraw"
	_ "fst/backend/docs"
	"fst/backend/pkg/config"
	"fst/backend/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// 控制器实例（延迟初始化，避免包加载期循环依赖）
var (
	publicAuthCtrl            *publicauth.AuthController               //登录注册
	publicSettingsCtrl        *publicsettings.SettingsController       //系统配置
	publicGeoCtrl             *publicgeo.GeoController                 //地理/区号探测
	publicSessionCtrl         *publicsession.SessionController         //会话强退清理
	publicPaymentCallbackCtrl *publicpayment.PaymentCallbackController //支付回调
	userProfileCtrl           *userprofile.ProfileController           //用户资料
	userPaymentCtrl           *userpayment.PaymentController           //用户支付
	userRealnameCtrl          *userrealname.RealnameController         //实名认证
	userWithdrawCtrl          *userwithdraw.WithdrawController         //用户提现
	systemCtrl                *systemctrl.SystemController
	adminUserCtrl             *adminuser.UserController     //用户管理
	adminLogCtrl              *adminlog.LogController       //日志
	adminAPILogCtrl           *adminapilog.APILogController //API日志
	adminEmailTplCtrl         *adminemailtemplate.EmailTemplateController
	adminSMSTplCtrl           *adminsmstemplate.SMSTemplateController //短信模板
	adminEmailLogCtrl         *adminemaillog.EmailLogController       //邮件日志
	adminSettingsCtrl         *adminsettings.SettingsController       //系统设置
	adminDebugCtrl            *admindebug.DebugController             //调试
	adminMoneyScoreCtrl       *adminusermoney.UserMoneyScoreController
	adminPaymentCtrl          *adminpayment.PaymentController           //支付
	adminRealnameCtrl         *adminrealname.RealnameController         //实名认证
	adminWithdrawCtrl         *adminwithdraw.WithdrawController         //提现
	adminSMSLogCtrl           *adminsmslog.SMSLogController             //短信日志
	adminDashboardCtrl        *admindashboard.DashboardController       //仪表盘
	adminAutoJobCtrl          *adminautojob.AutoJobController           //自动任务
	adminOnlineCtrl           *adminonline.OnlineController             //在线用户
	adminAnnouncementCtrl     *adminannouncement.AnnouncementController //站内公告
	userAnnouncementCtrl      *userannouncement.AnnouncementController  //用户公告
	adminProfileCtrl          *adminprofile.ProfileController
	adminTodoCtrl             *admintodo.TodoController
	adminDBConsoleCtrl        *admindbconsole.DBConsoleController
	adminTerminalCtrl         *adminterminal.TerminalController
	adminUserLevelCtrl        *adminuserlevel.UserLevelController
)

func initControllers() {
	publicAuthCtrl = publicauth.NewAuthController()
	publicSettingsCtrl = publicsettings.NewSettingsController()
	publicGeoCtrl = publicgeo.NewGeoController()
	publicSessionCtrl = publicsession.NewSessionController()
	publicPaymentCallbackCtrl = publicpayment.NewPaymentCallbackController()
	userProfileCtrl = userprofile.NewProfileController()
	userPaymentCtrl = userpayment.NewPaymentController()
	userRealnameCtrl = userrealname.NewRealnameController()
	userWithdrawCtrl = userwithdraw.NewWithdrawController()
	systemCtrl = &systemctrl.SystemController{}
	adminUserCtrl = adminuser.NewUserController()
	adminLogCtrl = adminlog.NewLogController()
	adminAPILogCtrl = adminapilog.NewAPILogController()
	adminEmailTplCtrl = adminemailtemplate.NewEmailTemplateController()
	adminSMSTplCtrl = adminsmstemplate.NewSMSTemplateController()
	adminEmailLogCtrl = adminemaillog.NewEmailLogController()
	adminSettingsCtrl = adminsettings.NewSettingsController()
	adminDebugCtrl = admindebug.NewDebugController()
	adminMoneyScoreCtrl = adminusermoney.NewUserMoneyScoreController()
	adminPaymentCtrl = adminpayment.NewPaymentController()
	adminRealnameCtrl = adminrealname.NewRealnameController()
	adminWithdrawCtrl = adminwithdraw.NewWithdrawController()
	adminDashboardCtrl = admindashboard.NewDashboardController()
	adminSMSLogCtrl = adminsmslog.NewSMSLogController()
	adminAutoJobCtrl = adminautojob.NewAutoJobController()
	adminOnlineCtrl = adminonline.NewOnlineController()
	adminAnnouncementCtrl = adminannouncement.NewAnnouncementController()
	userAnnouncementCtrl = userannouncement.NewAnnouncementController()
	adminProfileCtrl = adminprofile.NewProfileController()
	adminTodoCtrl = admintodo.NewTodoController()
	adminDBConsoleCtrl = admindbconsole.NewDBConsoleController()
	adminTerminalCtrl = adminterminal.NewTerminalController()
	adminUserLevelCtrl = adminuserlevel.NewUserLevelController()
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
	v2 := api.Group("/v2")

	registerPublicRoutes(v1) // 公开接口
	registerUserRoutes(v1)   // 用户接口
	registerSystemRoutes(v1) // 系统接口
	registerAdminRoutes(v1)  // 管理员接口
	registerWSRoutes(v1)     // WebSocket（自行鉴权，不挂 HTTP AuthMiddleware）
	registerV2Routes(v2)     // V2 测试接口

	// 健康检查：轻量、不碰库（负载均衡探活）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    0,
			"message": "ok",
			"data": gin.H{
				"status":    "ok",
				"message":   "Health check passed",
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
