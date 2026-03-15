package user

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"
	"strconv"
	"strings"

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

// ========================================
// 接口方法
// ========================================

// CreateOrder 创建充值订单
// @Summary 创建充值订单
// @Tags 支付
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateOrderRequest true "订单信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/payment/create [post]
func (ctrl *PaymentController) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "用户未登录")
		return
	}
	uid := userID.(uint64)

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	// XSS 防护
	req.Subject = utils.Clean_XSS(req.Subject)

	// 从系统设置读取后端API地址（用于异步回调和同步跳转）
	urlSettings, err := models.GetSettingsMap([]string{"frontend_url", "backend_api_url"})
	if err != nil {
		utils.Fail(c, 500, "读取系统配置失败")
		return
	}
	frontendURL := strings.TrimRight(urlSettings["frontend_url"], "/")
	backendAPIURL := strings.TrimRight(urlSettings["backend_api_url"], "/")
	if frontendURL == "" {
		utils.Fail(c, 500, "系统未配置前端地址，请在管理后台「基本设置」中配置")
		return
	}
	if backendAPIURL == "" {
		utils.Fail(c, 500, "系统未配置后端API地址，请在管理后台「基本设置」中配置")
		return
	}

	notifyURL := fmt.Sprintf("%s/api/v1/public/payment/notify", backendAPIURL)
	returnURL := fmt.Sprintf("%s/api/v1/public/payment/return", backendAPIURL)

	clientIP := c.ClientIP()

	result, err := services.CreatePaymentOrder(uid, &services.CreatePaymentOrderRequest{
		GatewayID: req.GatewayID,
		Amount:    req.Amount,
		Subject:   req.Subject,
		ClientIP:  clientIP,
	}, notifyURL, returnURL)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.Success(c, result)
}

// GetOrders 获取当前用户的订单列表
// @Summary 获取我的充值订单列表
// @Tags 支付
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query int false "状态筛选（-1=全部）" default(-1)
// @Success 200 {object} utils.Response
// @Router /api/v1/user/payment/orders [get]
func (ctrl *PaymentController) GetOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "用户未登录")
		return
	}
	uid := userID.(uint64)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orders, total, err := models.GetPaymentOrderList(uid, page, pageSize, status, "")
	if err != nil {
		utils.Fail(c, 500, "获取订单列表失败")
		return
	}

	utils.Success(c, gin.H{"list": orders, "total": total})
}

// GetOrderDetail 获取订单详情（仅限自己的订单）
// @Summary 获取订单详情
// @Tags 支付
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/payment/orders/{id} [get]
func (ctrl *PaymentController) GetOrderDetail(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "用户未登录")
		return
	}
	uid := userID.(uint64)

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

	// 只能查看自己的订单
	if order.UserID != uid {
		utils.Fail(c, 403, "无权查看该订单")
		return
	}

	utils.Success(c, order)
}

// CheckOrderStatus 轮询订单支付状态
// @Summary 检查订单状态
// @Tags 支付
// @Produce json
// @Security BearerAuth
// @Param id path int true "订单ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/payment/orders/{id}/status [get]
func (ctrl *PaymentController) CheckOrderStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "用户未登录")
		return
	}
	uid := userID.(uint64)

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

	if order.UserID != uid {
		utils.Fail(c, 403, "无权查看该订单")
		return
	}

	utils.Success(c, gin.H{
		"order_no": order.OrderNo,
		"status":   order.Status,
		"paid_at":  order.PaidAt,
	})
}

// GetPayGateways 获取可用支付通道列表
// @Summary 获取可用支付通道列表
// @Tags 支付
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/payment/gateways [get]
func (ctrl *PaymentController) GetPayGateways(c *gin.Context) {
	gateways, err := services.GetPayGatewayListForUser()
	if err != nil {
		utils.Fail(c, 500, "获取支付通道失败")
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
		payment.POST("/create", ctrl.CreateOrder)
		payment.GET("/orders", ctrl.GetOrders)
		payment.GET("/orders/:id", ctrl.GetOrderDetail)
		payment.GET("/orders/:id/status", ctrl.CheckOrderStatus)
		payment.GET("/gateways", ctrl.GetPayGateways)
	}
}
