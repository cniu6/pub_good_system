package routes

// routes.go 是路由文件，负责将各个控制器注册到路由中。
import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fst/backend/app/controllers"
	"fst/backend/app/controllers/admin"
	"fst/backend/app/controllers/public"
	"fst/backend/app/controllers/user"
	"fst/backend/app/models"
	_ "fst/backend/docs"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// 进程启动时间（供 /metrics uptime）
var processStartedAt = time.Now()

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

	// Swagger：注解仍写 /api/v1/admin/*；返回 doc.json 时按 ADMIN_API_PATH 改写
	if cfg := config.GlobalConfig; cfg != nil && cfg.EnableSwagger {
		router.GET(
			"/swagger/*any",
			middleware.SwaggerAdminPathRewriteMiddleware(),
			ginSwagger.WrapHandler(swaggerFiles.Handler),
		)
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

	// 就绪检查：Ping 数据库；失败返回 HTTP 503 + 统一 {code,message,data}
	// router.GET("/ready", handleReady)

	// 简易 Prometheus 文本指标（无额外依赖）
	// router.GET("/metrics", handleMetrics)
}

// handleReady 就绪探针：DB Ping 成功才算 ready。
func handleReady(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	database, err := db.SQLDB()
	if err != nil || database == nil {
		utils.Fail(c, 503, "database not initialized")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		utils.Fail(c, 503, "database ping failed: "+err.Error())
		return
	}
	utils.Success(c, gin.H{
		"status":     "ready",
		"db_driver":  db.DriverName(),
		"request_id": requestID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

// handleMetrics 输出 text/plain Prometheus 风格指标。
func handleMetrics(c *gin.Context) {
	var b strings.Builder
	b.WriteString(middleware.FormatPrometheusHTTPCounters())

	uptime := time.Since(processStartedAt).Seconds()
	b.WriteString("# HELP fst_uptime_seconds Process uptime in seconds\n")
	b.WriteString("# TYPE fst_uptime_seconds gauge\n")
	b.WriteString("fst_uptime_seconds ")
	b.WriteString(strconv.FormatFloat(uptime, 'f', 3, 64))
	b.WriteByte('\n')

	openConns := 0
	inUse := 0
	idle := 0
	if database, err := db.SQLDB(); err == nil && database != nil {
		st := database.Stats()
		openConns = st.OpenConnections
		inUse = st.InUse
		idle = st.Idle
	}
	b.WriteString("# HELP fst_db_open_connections Database open connections\n")
	b.WriteString("# TYPE fst_db_open_connections gauge\n")
	b.WriteString(fmt.Sprintf("fst_db_open_connections %d\n", openConns))
	b.WriteString("# HELP fst_db_in_use Database connections in use\n")
	b.WriteString("# TYPE fst_db_in_use gauge\n")
	b.WriteString(fmt.Sprintf("fst_db_in_use %d\n", inUse))
	b.WriteString("# HELP fst_db_idle Database idle connections\n")
	b.WriteString("# TYPE fst_db_idle gauge\n")
	b.WriteString(fmt.Sprintf("fst_db_idle %d\n", idle))

	openExceptions := countOpenPaymentExceptions()
	b.WriteString("# HELP fst_payment_exceptions_open Open payment exceptions\n")
	b.WriteString("# TYPE fst_payment_exceptions_open gauge\n")
	b.WriteString(fmt.Sprintf("fst_payment_exceptions_open %d\n", openExceptions))

	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.String(200, b.String())
}

// countOpenPaymentExceptions 统计待处理支付异常数；表未就绪或查询失败时返回 0。
func countOpenPaymentExceptions() int64 {
	if db.GetDB() == nil {
		return 0
	}
	var n int64
	err := db.DB.Raw("SELECT COUNT(*) FROM payment_exceptions WHERE status = ?", models.PaymentExceptionStatusOpen).Scan(&n).Error
	if err != nil {
		return 0
	}
	return n
}
