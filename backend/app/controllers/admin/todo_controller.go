package admin

import (
	"fst/backend/app/models"
	"fst/backend/pkg/db"
	"fst/backend/utils"
	"log"

	"github.com/gin-gonic/gin"
)

// TodoItem 管理端待办聚合项
type TodoItem struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Count int64  `json:"count"`
	Link  string `json:"link"`
}

// TodoController 管理端待办聚合
type TodoController struct{}

func NewTodoController() *TodoController {
	return &TodoController{}
}

// List GET /todos — 聚合待审实名、开放支付异常、失败自动任务等
func (ctrl *TodoController) List(c *gin.Context) {
	items := make([]TodoItem, 0, 8)

	// 待审实名
	var pendingRealname int64
	if err := db.DB.Get(&pendingRealname, `
		SELECT COUNT(1) FROM user_realname_verifications
		WHERE status = 0 AND delete_time IS NULL`); err != nil {
		log.Printf("[ADMIN][TODO] count pending realname: %v", err)
	} else if pendingRealname > 0 {
		items = append(items, TodoItem{
			Type:  "pending_realname",
			Title: "待审核实名认证",
			Count: pendingRealname,
			Link:  "/users/realname",
		})
	}

	// 开放支付异常
	var openExceptions int64
	if err := db.DB.Get(&openExceptions, `
		SELECT COUNT(1) FROM payment_exceptions WHERE status = ?`, models.PaymentExceptionStatusOpen); err != nil {
		log.Printf("[ADMIN][TODO] count open payment exceptions: %v", err)
	} else if openExceptions > 0 {
		items = append(items, TodoItem{
			Type:  "payment_exception",
			Title: "待处理支付异常",
			Count: openExceptions,
			Link:  "/finance/payment-exceptions",
		})
	}

	// 失败的自动任务定义（最近状态 failed）
	var failedJobs int64
	if db.CheckTableExists("auto_job_definitions") {
		if err := db.DB.Get(&failedJobs, `
			SELECT COUNT(1) FROM auto_job_definitions WHERE last_status = 'failed'`); err != nil {
			log.Printf("[ADMIN][TODO] count failed auto jobs: %v", err)
		} else if failedJobs > 0 {
			items = append(items, TodoItem{
				Type:  "failed_auto_job",
				Title: "失败的自动任务",
				Count: failedJobs,
				Link:  "/settings/auto-jobs",
			})
		}
	}

	// 待审批（双人复核开启时）
	var pendingApprovals int64
	if db.CheckTableExists("approval_requests") {
		if err := db.DB.Get(&pendingApprovals, `
			SELECT COUNT(1) FROM approval_requests WHERE status = ?`, models.ApprovalStatusPending); err != nil {
			log.Printf("[ADMIN][TODO] count pending approvals: %v", err)
		} else if pendingApprovals > 0 {
			items = append(items, TodoItem{
				Type:  "pending_approval",
				Title: "待审批财务操作",
				Count: pendingApprovals,
				Link:  "/finance/approvals",
			})
		}
	}

	// 待审提现
	var pendingWithdraw int64
	if db.CheckTableExists("withdraw_requests") {
		if err := db.DB.Get(&pendingWithdraw, `
			SELECT COUNT(1) FROM withdraw_requests WHERE status = 0`); err != nil {
			log.Printf("[ADMIN][TODO] count pending withdraw: %v", err)
		} else if pendingWithdraw > 0 {
			items = append(items, TodoItem{
				Type:  "pending_withdraw",
				Title: "待审核提现",
				Count: pendingWithdraw,
				Link:  "/finance/withdraw",
			})
		}
	}

	utils.Success(c, gin.H{"list": items})
}

// RegisterRoutes 注册待办路由
func (ctrl *TodoController) RegisterRoutes(adminGroup *gin.RouterGroup) {
	adminGroup.GET("/todos", ctrl.List)
}
