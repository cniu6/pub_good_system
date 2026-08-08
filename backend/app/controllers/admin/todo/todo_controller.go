package todo

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
// @Summary 管理端待办聚合
// @Tags Admin-待办
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/todos [get]
func (ctrl *TodoController) List(c *gin.Context) {
	items := make([]TodoItem, 0, 8)

	var pendingRealname int64
	if err := db.DB.Model(&models.RealnameVerification{}).
		Where("status = 0 AND delete_time IS NULL").
		Count(&pendingRealname).Error; err != nil {
		log.Printf("[ADMIN][TODO] count pending realname: %v", err)
	} else if pendingRealname > 0 {
		items = append(items, TodoItem{
			Type:  "pending_realname",
			Title: "待审核实名认证",
			Count: pendingRealname,
			Link:  "/users/realname",
		})
	}

	var openExceptions int64
	if err := db.DB.Model(&models.PaymentException{}).
		Where("status = ?", models.PaymentExceptionStatusOpen).
		Count(&openExceptions).Error; err != nil {
		log.Printf("[ADMIN][TODO] count open payment exceptions: %v", err)
	} else if openExceptions > 0 {
		items = append(items, TodoItem{
			Type:  "payment_exception",
			Title: "待处理支付异常",
			Count: openExceptions,
			Link:  "/finance/payment-exceptions",
		})
	}

	var failedJobs int64
	if db.CheckTableExists("auto_job_definitions") {
		if err := db.DB.Table("auto_job_definitions").
			Where("last_status = 'failed'").
			Count(&failedJobs).Error; err != nil {
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

	var pendingWithdraw int64
	if db.CheckTableExists("withdraw_requests") {
		if err := db.DB.Model(&models.WithdrawRequest{}).
			Where("status = 0 AND delete_time IS NULL").
			Count(&pendingWithdraw).Error; err != nil {
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
