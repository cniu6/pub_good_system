package withdraw

import (
	"database/sql"
	"errors"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type WithdrawController struct {
	withdrawService *services.WithdrawService
}

func NewWithdrawController() *WithdrawController {
	return &WithdrawController{
		withdrawService: services.NewWithdrawService(),
	}
}

type ReviewWithdrawBody struct {
	Status       uint8  `json:"status" binding:"required"`
	ReviewRemark string `json:"review_remark"`
}

type PayWithdrawBody struct {
	TransferRemark string `json:"transfer_remark"`
}

// List 提现列表
// @Summary 提现列表
// @Tags Admin-提现
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/withdraw [get]
func (ctrl *WithdrawController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = utils.NormalizePagination(page, pageSize)
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)
	keyword := utils.Clean_XSS(c.DefaultQuery("keyword", ""))

	var status *uint8
	statusStr := c.Query("status")
	if statusStr != "" {
		if v, err := strconv.ParseUint(statusStr, 10, 8); err == nil {
			val := uint8(v)
			status = &val
		}
	}

	result, err := ctrl.withdrawService.GetList(&models.WithdrawListQuery{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
		Keyword:  keyword,
		Status:   status,
	})
	if err != nil {
		log.Printf("[ADMIN][WITHDRAW] list failed: %v", err)
		utils.Fail(c, 500, "获取提现列表失败")
		return
	}
	utils.Success(c, result)
}

// Stats 提现统计
// @Summary 提现统计
// @Tags Admin-提现
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/withdraw/stats [get]
func (ctrl *WithdrawController) Stats(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)
	keyword := utils.Clean_XSS(c.DefaultQuery("keyword", ""))

	var status *uint8
	statusStr := c.Query("status")
	if statusStr != "" {
		if v, err := strconv.ParseUint(statusStr, 10, 8); err == nil {
			val := uint8(v)
			status = &val
		}
	}

	result, err := ctrl.withdrawService.GetStats(&models.WithdrawListQuery{
		Page:     page,
		PageSize: 20,
		UserID:   userID,
		Keyword:  keyword,
		Status:   status,
	})
	if err != nil {
		log.Printf("[ADMIN][WITHDRAW] stats failed: %v", err)
		utils.Fail(c, 500, "获取提现统计失败")
		return
	}
	utils.Success(c, result)
}

// Detail 提现详情
// @Summary 提现详情
// @Tags Admin-提现
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/withdraw/{id} [get]
func (ctrl *WithdrawController) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的ID")
		return
	}
	item, err := ctrl.withdrawService.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.Fail(c, 404, "提现记录不存在")
			return
		}
		log.Printf("[ADMIN][WITHDRAW] detail failed id=%d: %v", id, err)
		utils.Fail(c, 500, "获取提现详情失败")
		return
	}
	utils.Success(c, item)
}

// LegacyRisk 只读检测：已审核通过/已打款但 balance_deducted=false 的历史风险单（防重复扣款排查）
// @Summary 提现历史风险单检测
// @Tags Admin-提现
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/withdraw/legacy-risk [get]
func (ctrl *WithdrawController) LegacyRisk(c *gin.Context) {
	list, err := models.ListWithdrawLegacyBalanceRisk(50)
	if err != nil {
		log.Printf("[ADMIN][WITHDRAW] legacy risk query failed: %v", err)
		utils.Fail(c, 500, "查询历史风险单失败")
		return
	}
	utils.Success(c, gin.H{
		"list":    list,
		"total":   len(list),
		"message": "只读检测：status 已通过/已打款且 balance_deducted=false 的记录，处理前请人工核对余额流水",
	})
}

// Review 提现审核
// @Summary 提现审核
// @Tags Admin-提现
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/withdraw/{id}/review [post]
func (ctrl *WithdrawController) Review(c *gin.Context) {
	adminID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "用户未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的ID")
		return
	}

	var req ReviewWithdrawBody
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	req.ReviewRemark = utils.Clean_XSS(req.ReviewRemark)

	if err := ctrl.withdrawService.Review(adminID.(uint64), &services.ReviewWithdrawRequest{
		ID:           id,
		Status:       req.Status,
		ReviewRemark: req.ReviewRemark,
	}); err != nil {
		if services.IsClientError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		log.Printf("[ADMIN][WITHDRAW] review failed admin_id=%d request_id=%d: %v", adminID.(uint64), id, err)
		utils.Fail(c, 500, "提现审核失败，请稍后重试")
		return
	}
	utils.SuccessMsg(c, "审核完成", nil)
}

// MarkPaid 标记提现已人工打款
// @Summary 标记提现已人工打款
// @Tags Admin-提现
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/withdraw/{id}/pay [post]
func (ctrl *WithdrawController) MarkPaid(c *gin.Context) {
	adminID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "用户未登录")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的ID")
		return
	}

	var req PayWithdrawBody
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	req.TransferRemark = utils.Clean_XSS(req.TransferRemark)

	if err := ctrl.withdrawService.MarkPaid(adminID.(uint64), &services.PayWithdrawRequest{
		ID:             id,
		TransferRemark: req.TransferRemark,
	}); err != nil {
		if services.IsClientError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		log.Printf("[ADMIN][WITHDRAW] mark paid failed admin_id=%d request_id=%d: %v", adminID.(uint64), id, err)
		utils.Fail(c, 500, "提现打款处理失败，请稍后重试")
		return
	}
	utils.SuccessMsg(c, "已标记为人工打款完成", nil)
}

func (ctrl *WithdrawController) RegisterRoutes(group *gin.RouterGroup) {
	withdraw := group.Group("/withdraw")
	withdraw.Use(middleware.SimpleLogMiddleware("提现管理"))
	{
		withdraw.GET("", ctrl.List)
		withdraw.GET("/stats", ctrl.Stats)
		withdraw.GET("/legacy-risk", ctrl.LegacyRisk)
		withdraw.GET("/:id", ctrl.Detail)
		withdraw.POST("/:id/review", middleware.RequireIdempotency("admin_withdraw_review", 10*time.Minute), ctrl.Review)
		withdraw.POST("/:id/pay", middleware.RequireIdempotency("admin_withdraw_pay", 10*time.Minute), ctrl.MarkPaid)
	}
}
