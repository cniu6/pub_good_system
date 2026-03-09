package admin

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"
	"strconv"

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

type AdminCompleteOrderRequest struct {
	Memo string `json:"memo"`
}

// ========================================
// 接口方法
// ========================================

// ListOrders 订单列表（管理端，支持筛选）
// @Summary 管理端-支付订单列表
// @Tags 管理端-支付
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
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	userIDStr := c.DefaultQuery("user_id", "0")
	keyword := c.DefaultQuery("keyword", "")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	keyword = utils.Clean_XSS(keyword)

	var userID uint64
	if v, err := strconv.ParseUint(userIDStr, 10, 64); err == nil {
		userID = v
	}

	orders, total, err := models.GetPaymentOrderList(userID, page, pageSize, status, keyword)
	if err != nil {
		utils.Fail(c, 500, "获取订单列表失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"list": orders, "total": total})
}

// OrderDetail 订单详情
// @Summary 管理端-订单详情
// @Tags 管理端-支付
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

	utils.Success(c, order)
}

// CompleteOrder 手动补单
// @Summary 管理端-手动补单
// @Tags 管理端-支付
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
	c.ShouldBindJSON(&req)
	req.Memo = utils.Clean_XSS(req.Memo)

	if err := services.AdminCompleteOrder(orderID, req.Memo); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.SuccessMsg(c, "补单成功", nil)
}

// CancelOrder 取消订单
// @Summary 管理端-取消订单
// @Tags 管理端-支付
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
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.SuccessMsg(c, "订单已取消", nil)
}

// GetStats 支付统计
// @Summary 管理端-支付统计
// @Tags 管理端-支付
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/payment/stats [get]
func (ctrl *PaymentController) GetStats(c *gin.Context) {
	stats, err := models.GetPaymentStats()
	if err != nil {
		utils.Fail(c, 500, "获取统计数据失败: "+err.Error())
		return
	}

	utils.Success(c, stats)
}

// DeleteOrder 删除订单
// @Summary 管理端-删除订单
// @Tags 管理端-支付
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
		utils.Fail(c, 400, err.Error())
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
		utils.Fail(c, 500, "获取支付通道列表失败: "+err.Error())
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

	gw, err := models.GetPayGatewayByID(id)
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
	{
		// 订单管理
		payment.GET("/orders", ctrl.ListOrders)
		payment.GET("/orders/:id", ctrl.OrderDetail)
		payment.POST("/orders/:id/complete", ctrl.CompleteOrder)
		payment.POST("/orders/:id/cancel", ctrl.CancelOrder)
		payment.DELETE("/orders/:id", ctrl.DeleteOrder)
		payment.GET("/stats", ctrl.GetStats)

		// 支付通道管理
		payment.POST("/gateways", ctrl.CreateGateway)
		payment.GET("/gateways", ctrl.ListGateways)
		payment.GET("/gateways/:id", ctrl.GetGateway)
		payment.PUT("/gateways/:id", ctrl.UpdateGateway)
		payment.DELETE("/gateways/:id", ctrl.DeleteGateway)
	}
}
