package payment

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/middleware"
	"fst/backend/utils"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PaymentController 用户支付控制器（需要登录）
type PaymentController struct{}

// NewPaymentController 创建支付控制器
func NewPaymentController() *PaymentController {
	return &PaymentController{}
}

// ========================================
// 请求结构体
// ========================================

type CreateOrderRequest struct {
	GatewayID uint64  `json:"gateway_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	Subject   string  `json:"subject"`
}

//@name 创建订单请求

// ========================================
// 接口方法
// ========================================

// CreateOrder 创建充值订单
// @Summary 创建充值订单
// @Tags User-支付
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateOrderRequest true "订单信息"
// @Success 200 {object} utils.Response
// @Router /v1/user/payment/create [post]
func (ctrl *PaymentController) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "Unauthorized")
		return
	}
	uid := userID.(uint64)
	// 用户等级能力：充值开关
	if ok, msg := models.CheckUserLevelAllows(uid, "recharge"); !ok {
		utils.Fail(c, 403, msg)
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "Invalid parameter: "+err.Error())
		return
	}

	// XSS 防护
	req.Subject = utils.Clean_XSS(req.Subject)

	// 从系统设置读取后端API地址（用于异步回调和同步跳转）
	frontendURL := services.GetGlobalFrontendURL()
	backendAPIURL := services.GetGlobalBackendAPIURL()
	if frontendURL == "" {
		utils.Fail(c, 500, "Frontend URL not configured")
		return
	}
	if backendAPIURL == "" {
		utils.Fail(c, 500, "Backend API URL not configured")
		return
	}

	notifyURL := fmt.Sprintf("%s/api/v1/public/payment/notify", backendAPIURL)
	returnURL := fmt.Sprintf("%s/api/v1/public/payment/return", backendAPIURL)

	clientIP := utils.GetClientIP(c)

	result, err := services.CreatePaymentOrder(uid, &services.CreatePaymentOrderRequest{
		GatewayID: req.GatewayID,
		Amount:    req.Amount,
		Subject:   req.Subject,
		ClientIP:  clientIP,
	}, notifyURL, returnURL)
	if err != nil {
		if services.IsClientError(err) {
			utils.Fail(c, 400, err.Error())
			return
		}
		log.Printf("[PAYMENT] create order failed for user_id=%d: %v", uid, err)
		utils.Fail(c, 500, "Failed to create order, please try again later")
		return
	}

	utils.Success(c, result)
}

// GetOrders 获取当前用户的订单列表
// @Summary 获取我的充值订单列表
// @Tags User-支付
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query int false "状态筛选（-1=全部）" default(-1)
// @Success 200 {object} utils.Response
// @Router /v1/user/payment/orders [get]
func (ctrl *PaymentController) GetOrders(c *gin.Context) {
	uid, ok := utils.GetUserID(c)
	if !ok {
		utils.Fail(c, 401, "Unauthorized")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = utils.NormalizePagination(page, pageSize)
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	orders, total, err := models.GetPaymentOrderList(uid, page, pageSize, status, "")
	if err != nil {
		utils.Fail(c, 500, "Failed to get order list")
		return
	}

	utils.Success(c, gin.H{"list": orders, "total": total})
}

// GetOrderDetail 获取订单详情（仅限自己的订单）
// @Summary 获取订单详情
// @Tags User-支付
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Success 200 {object} utils.Response
// @Router /v1/user/payment/orders/{id} [get]
func (ctrl *PaymentController) GetOrderDetail(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "Unauthorized")
		return
	}
	uid := userID.(uint64)

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "Invalid order ID")
		return
	}

	order, err := models.GetPaymentOrderByID(orderID)
	if err != nil {
		utils.Fail(c, 404, "Order not found")
		return
	}

	// 只能查看自己的订单
	if order.UserID != uid {
		utils.Fail(c, 403, "No permission to view this order")
		return
	}

	refreshedOrder, reconciled, reconcileErr := services.ReconcilePaymentOrderByID(orderID)
	if reconcileErr != nil {
		log.Printf("[PAYMENT] reconcile order detail failed order_id=%d user_id=%d: %v", orderID, uid, reconcileErr)
	} else if reconciled && refreshedOrder != nil {
		order = refreshedOrder
	}

	utils.Success(c, order)
}

// CheckOrderStatus 轮询订单支付状态
// @Summary 检查订单状态
// @Tags User-支付
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Success 200 {object} utils.Response
// @Router /v1/user/payment/orders/{id}/status [get]
func (ctrl *PaymentController) CheckOrderStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "Unauthorized")
		return
	}
	uid := userID.(uint64)

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "Invalid order ID")
		return
	}

	order, err := models.GetPaymentOrderByID(orderID)
	if err != nil {
		utils.Fail(c, 404, "Order not found")
		return
	}

	if order.UserID != uid {
		utils.Fail(c, 403, "No permission to view this order")
		return
	}

	reconciled := false
	refreshedOrder, didReconcile, reconcileErr := services.ReconcilePaymentOrderByID(orderID)
	if reconcileErr != nil {
		log.Printf("[PAYMENT] reconcile order status failed order_id=%d user_id=%d: %v", orderID, uid, reconcileErr)
	} else {
		reconciled = didReconcile
		if refreshedOrder != nil {
			order = refreshedOrder
		}
	}

	utils.Success(c, gin.H{
		"order_no":   order.OrderNo,
		"status":     order.Status,
		"paid_at":    order.PaidAt,
		"reconciled": reconciled,
	})
}

// GetPayGateways 获取可用支付通道列表
// @Summary 获取可用支付通道列表
// @Tags User-支付
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/user/payment/gateways [get]
func (ctrl *PaymentController) GetPayGateways(c *gin.Context) {
	gateways, err := services.GetPayGatewayListForUser()
	if err != nil {
		utils.Fail(c, 500, "Failed to get payment gateways")
		return
	}

	utils.Success(c, gin.H{
		"list": gateways,
	})
}

// ========================================
// 注册路由
// ========================================

// RegisterRoutes 注册用户支付路由
func (ctrl *PaymentController) RegisterRoutes(group *gin.RouterGroup) {
	payment := group.Group("/payment")
	{
		payment.POST("/create", middleware.RequireIdempotency("user_payment_create", 10*time.Minute), ctrl.CreateOrder)
		payment.GET("/orders", ctrl.GetOrders)
		payment.GET("/orders/:id", ctrl.GetOrderDetail)
		payment.GET("/orders/:id/status", ctrl.CheckOrderStatus)
		payment.GET("/gateways", ctrl.GetPayGateways)
	}
}
