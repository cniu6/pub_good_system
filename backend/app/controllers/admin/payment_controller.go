package admin

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PaymentController 管理端支付订单控制器
type PaymentController struct{}

// NewPaymentController 创建管理端支付控制器
func NewPaymentController() *PaymentController {
	return &PaymentController{}
}

// ========================================
// 请求结构体
// ========================================

// AdminCompleteOrderRequest 管理端补单请求
type AdminCompleteOrderRequest struct {
	Memo  string `json:"memo"`
	Force bool   `json:"force"` // 强制补单（canceled/failed 高危路径，须填 memo）
}//@name 管理端补单请求

// AdminResolveExceptionRequest 管理端订单异常处理请求
type AdminResolveExceptionRequest struct {
	Action string `json:"action"` // resolve | ignore
	Remark string `json:"remark"`
}//@name 管理端订单异常处理请求

// ========================================
// 接口方法
// ========================================

// ListOrders 订单列表（管理端，支持筛选）
// @Summary 管理端-支付订单列表
// @Tags Admin-支付
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query int false "状态筛选（-1=全部）" default(-1)
// @Param user_id query int false "用户ID筛选"
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/payment/orders [get]
func (ctrl *PaymentController) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = utils.NormalizePagination(page, pageSize)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	userIDStr := c.DefaultQuery("user_id", "0")
	keyword := c.DefaultQuery("keyword", "")

	keyword = utils.Clean_XSS(keyword)

	var userID uint64
	if v, err := strconv.ParseUint(userIDStr, 10, 64); err == nil {
		userID = v
	}

	orders, total, err := models.GetPaymentOrderList(userID, page, pageSize, status, keyword)
	if err != nil {
		log.Printf("[ADMIN][PAYMENT] list orders failed: %v", err)
		utils.Fail(c, 500, "获取订单列表失败")
		return
	}

	utils.Success(c, gin.H{"list": orders, "total": total})
}

// OrderDetail 订单详情
// @Summary 管理端-订单详情
// @Tags Admin-支付
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/payment/orders/{id} [get]
func (ctrl *PaymentController) OrderDetail(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的订单ID")
		return
	}

	order, err := models.GetPaymentOrderByID(orderID)
	if err != nil {
		utils.Fail(c, 404, "订单不存在")
		return
	}

	refreshedOrder, reconciled, reconcileErr := services.ReconcilePaymentOrderByID(orderID)
	if reconcileErr != nil {
		log.Printf("[ADMIN][PAYMENT] reconcile order detail failed order_id=%d: %v", orderID, reconcileErr)
	} else if reconciled && refreshedOrder != nil {
		order = refreshedOrder
	}

	utils.Success(c, order)
}

// CompleteOrder 手动补单
// @Summary 管理端-手动补单
// @Tags Admin-支付
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Param body body AdminCompleteOrderRequest false "补单备注"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/payment/orders/{id}/complete [post]
func (ctrl *PaymentController) CompleteOrder(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的订单ID")
		return
	}

	var req AdminCompleteOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	req.Memo = utils.Clean_XSS(req.Memo)

	if err := services.AdminCompleteOrder(orderID, req.Memo, req.Force); err != nil {
		if services.IsClientError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		log.Printf("[ADMIN][PAYMENT] complete order failed order_id=%d: %v", orderID, err)
		utils.Fail(c, 500, "补单失败，请稍后重试")
		return
	}

	utils.SuccessMsg(c, "补单成功", nil)
}

// ReconcileOrder 单笔主动对账
// @Summary 管理端-支付订单对账
// @Tags Admin-支付
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/payment/orders/{id}/reconcile [post]
func (ctrl *PaymentController) ReconcileOrder(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的订单ID")
		return
	}
	order, changed, err := services.ReconcilePaymentOrderByID(orderID)
	if err != nil {
		if services.IsClientError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		log.Printf("[ADMIN][PAYMENT] reconcile failed order_id=%d: %v", orderID, err)
		utils.Fail(c, 500, "对账失败，请稍后重试")
		return
	}
	utils.Success(c, gin.H{"changed": changed, "order": order})
}

// ListExceptions 支付异常列表
// @Summary 管理端-支付异常列表
// @Tags Admin-支付
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/payment/exceptions [get]
func (ctrl *PaymentController) ListExceptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = utils.NormalizePagination(page, pageSize)

	var statusPtr *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			statusPtr = &v
		}
	}
	exceptionType := c.Query("exception_type")
	orderNo := c.Query("order_no")
	var userID uint64
	if u := c.Query("user_id"); u != "" {
		userID, _ = strconv.ParseUint(u, 10, 64)
	}

	list, total, err := models.ListPaymentExceptions(page, pageSize, statusPtr, exceptionType, orderNo, userID)
	if err != nil {
		log.Printf("[ADMIN][PAYMENT] list exceptions failed: %v", err)
		utils.Fail(c, 500, "查询异常列表失败")
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

// ResolveException 处理/忽略支付异常
// @Summary 管理端-处理支付异常
// @Tags Admin-支付
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "异常ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/payment/exceptions/{id}/resolve [post]
func (ctrl *PaymentController) ResolveException(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的异常ID")
		return
	}
	var req AdminResolveExceptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	req.Remark = utils.Clean_XSS(req.Remark)
	if err := utils.ValidateRuneLen(req.Remark, "处理备注", utils.MaxResolveRemarkLength); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	status := models.PaymentExceptionStatusResolved
	switch req.Action {
	case "resolve":
		status = models.PaymentExceptionStatusResolved
	case "ignore":
		status = models.PaymentExceptionStatusIgnored
	default:
		utils.Fail(c, 400, "action 仅支持 resolve/ignore")
		return
	}
	adminID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "用户未登录")
		return
	}
	if err := models.ResolvePaymentException(id, adminID.(uint64), status, req.Remark); err != nil {
		utils.Fail(c, 400, "处理失败: "+err.Error())
		return
	}
	utils.SuccessMsg(c, "已处理", nil)
}

// CancelOrder 取消订单
// @Summary 管理端-取消订单
// @Tags Admin-支付
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/payment/orders/{id}/cancel [post]
func (ctrl *PaymentController) CancelOrder(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的订单ID")
		return
	}

	if err := services.AdminCancelOrder(orderID); err != nil {
		if services.IsClientError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		log.Printf("[ADMIN][PAYMENT] cancel order failed order_id=%d: %v", orderID, err)
		utils.Fail(c, 500, "取消订单失败，请稍后重试")
		return
	}

	utils.SuccessMsg(c, "订单已取消", nil)
}

// GetStats 支付统计
// @Summary 管理端-支付统计
// @Tags Admin-支付
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/payment/stats [get]
func (ctrl *PaymentController) GetStats(c *gin.Context) {
	stats, err := models.GetPaymentStats()
	if err != nil {
		log.Printf("[ADMIN][PAYMENT] get stats failed: %v", err)
		utils.Fail(c, 500, "获取统计数据失败")
		return
	}

	utils.Success(c, stats)
}

// DeleteOrder 删除订单
// @Summary 管理端-删除订单
// @Tags Admin-支付
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/payment/orders/{id} [delete]
func (ctrl *PaymentController) DeleteOrder(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的订单ID")
		return
	}

	if err := services.AdminDeleteOrder(orderID); err != nil {
		if services.IsClientError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		log.Printf("[ADMIN][PAYMENT] delete order failed order_id=%d: %v", orderID, err)
		utils.Fail(c, 500, "删除订单失败，请稍后重试")
		return
	}

	utils.SuccessMsg(c, "订单已删除", nil)
}

// ========================================
// 支付通道管理
// ========================================

// CreateGateway 创建支付通道
func (ctrl *PaymentController) CreateGateway(c *gin.Context) {
	var req services.PayGatewayCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "JSON格式错误: "+err.Error())
		return
	}

	gw, err := services.CreatePayGateway(&req)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.SuccessMsg(c, "支付通道创建成功", gw)
}

// ListGateways 获取支付通道列表
func (ctrl *PaymentController) ListGateways(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))
	keyword := c.DefaultQuery("keyword", "")

	if page < 1 {
		page = 1
	}
	keyword = utils.Clean_XSS(keyword)

	gateways, total, err := services.GetPayGatewayListForAdmin(page, pageSize, keyword)
	if err != nil {
		log.Printf("[ADMIN][PAYMENT] list gateways failed: %v", err)
		utils.Fail(c, 500, "获取支付通道列表失败")
		return
	}

	utils.Success(c, gin.H{"list": gateways, "total": total})
}

// GetGateway 获取支付通道详情
func (ctrl *PaymentController) GetGateway(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的通道ID")
		return
	}

	gw, err := services.GetPayGatewayDetailForAdmin(id)
	if err != nil {
		utils.Fail(c, 404, "支付通道不存在")
		return
	}

	utils.Success(c, gw)
}

// UpdateGateway 更新支付通道
func (ctrl *PaymentController) UpdateGateway(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的通道ID")
		return
	}

	var req services.PayGatewayUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "JSON格式错误: "+err.Error())
		return
	}

	gw, err := services.UpdatePayGateway(id, &req)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.SuccessMsg(c, "支付通道更新成功", gw)
}

// DeleteGateway 删除支付通道
func (ctrl *PaymentController) DeleteGateway(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的通道ID")
		return
	}

	if err := services.DeletePayGateway(id); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.SuccessMsg(c, "支付通道删除成功", nil)
}

// ========================================
// 注册路由
// ========================================

// RegisterPaymentRoutes 注册管理端支付路由
func (ctrl *PaymentController) RegisterPaymentRoutes(group *gin.RouterGroup) {
	payment := group.Group("/payment")
	payment.Use(middleware.SimpleLogMiddleware("支付管理"))
	{
		// 订单管理（补单/取消属于资金相关写操作，强制幂等键防双击）
		payment.GET("/orders", ctrl.ListOrders)
		payment.GET("/orders/:id", ctrl.OrderDetail)
		payment.POST("/orders/:id/complete", middleware.RequireIdempotency("admin_payment_complete", 10*time.Minute), ctrl.CompleteOrder)
		payment.POST("/orders/:id/cancel", middleware.RequireIdempotency("admin_payment_cancel", 10*time.Minute), ctrl.CancelOrder)
		payment.POST("/orders/:id/reconcile", middleware.RequireIdempotency("admin_payment_reconcile", 10*time.Minute), ctrl.ReconcileOrder)
		payment.DELETE("/orders/:id", ctrl.DeleteOrder)
		payment.GET("/stats", ctrl.GetStats)

		// 支付异常工作台
		payment.GET("/exceptions", ctrl.ListExceptions)
		payment.POST("/exceptions/:id/resolve", middleware.RequireIdempotency("admin_payment_exception_resolve", 10*time.Minute), ctrl.ResolveException)

		// 支付通道管理
		payment.POST("/gateways", ctrl.CreateGateway)
		payment.GET("/gateways", ctrl.ListGateways)
		payment.GET("/gateways/:id", ctrl.GetGateway)
		payment.PUT("/gateways/:id", ctrl.UpdateGateway)
		payment.DELETE("/gateways/:id", ctrl.DeleteGateway)
	}
}

