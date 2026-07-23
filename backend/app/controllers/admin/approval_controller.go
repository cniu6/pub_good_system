package admin

import (
	"database/sql"
	"encoding/json"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ApprovalController 高危财务审批（双人复核）
type ApprovalController struct{}

func NewApprovalController() *ApprovalController {
	return &ApprovalController{}
}

// ListPending GET /approvals/pending
func (ctrl *ApprovalController) ListPending(c *gin.Context) {
	list, err := models.ListPendingApprovals(50)
	if err != nil {
		log.Printf("[ADMIN][APPROVAL] list failed: %v", err)
		utils.Fail(c, 500, "获取待审批失败")
		return
	}
	utils.Success(c, gin.H{"list": list})
}

// Approve POST /approvals/:id/approve — 另一管理员批准后执行强制补单
func (ctrl *ApprovalController) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的审批ID")
		return
	}
	reviewerID, ok := adminUID(c)
	if !ok {
		utils.Fail(c, 401, "Unauthorized")
		return
	}
	var req struct {
		Comment string `json:"comment"`
	}
	_ = c.ShouldBindJSON(&req)

	item, err := models.GetApprovalRequestByID(id)
	if err != nil {
		utils.Fail(c, 404, "审批不存在")
		return
	}
	if item.Status != models.ApprovalStatusPending {
		utils.Fail(c, 400, "该审批已处理")
		return
	}
	if item.RequesterID == reviewerID {
		utils.Fail(c, 403, "不能审批自己发起的请求")
		return
	}

	if err := models.ApproveApprovalRequest(id, reviewerID, strings.TrimSpace(req.Comment)); err != nil {
		if err == sql.ErrNoRows {
			utils.Fail(c, 400, "审批状态已变更")
			return
		}
		utils.Fail(c, 500, "审批失败")
		return
	}

	if item.Type == models.ApprovalTypeForcePaymentComplete {
		var payload struct {
			OrderID uint64 `json:"order_id"`
			Memo    string `json:"memo"`
		}
		if pErr := json.Unmarshal([]byte(item.PayloadJSON), &payload); pErr != nil || payload.OrderID == 0 {
			utils.Fail(c, 500, "审批载荷无效")
			return
		}
		// skipDualCheck=true：审批通过后直接入账，避免再次进入待审
		if err := services.AdminCompleteOrder(payload.OrderID, payload.Memo, true, reviewerID, true); err != nil {
			if services.IsClientError(err) {
				utils.Fail(c, 400, err.Error())
				return
			}
			log.Printf("[ADMIN][APPROVAL] execute force complete failed id=%d: %v", id, err)
			utils.Fail(c, 500, "执行补单失败")
			return
		}
	}

	utils.SuccessMsg(c, "已批准并执行", nil)
}

// Reject POST /approvals/:id/reject
func (ctrl *ApprovalController) Reject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的审批ID")
		return
	}
	reviewerID, ok := adminUID(c)
	if !ok {
		utils.Fail(c, 401, "Unauthorized")
		return
	}
	var req struct {
		Comment string `json:"comment"`
	}
	_ = c.ShouldBindJSON(&req)

	item, err := models.GetApprovalRequestByID(id)
	if err != nil {
		utils.Fail(c, 404, "审批不存在")
		return
	}
	if item.Status != models.ApprovalStatusPending {
		utils.Fail(c, 400, "该审批已处理")
		return
	}
	if item.RequesterID == reviewerID {
		utils.Fail(c, 403, "不能审批自己发起的请求")
		return
	}
	if err := models.RejectApprovalRequest(id, reviewerID, strings.TrimSpace(req.Comment)); err != nil {
		utils.Fail(c, 500, "拒绝失败")
		return
	}
	utils.SuccessMsg(c, "已拒绝", nil)
}

// RegisterRoutes 注册审批路由
func (ctrl *ApprovalController) RegisterRoutes(adminGroup *gin.RouterGroup) {
	g := adminGroup.Group("/approvals")
	g.Use(middleware.SimpleLogMiddleware("财务审批"))
	g.Use(middleware.RequirePermission("finance:write"))
	{
		g.GET("/pending", ctrl.ListPending)
		g.POST("/:id/approve", middleware.RequireIdempotency("admin_approval_approve", 10*time.Minute), ctrl.Approve)
		g.POST("/:id/reject", ctrl.Reject)
	}
}
