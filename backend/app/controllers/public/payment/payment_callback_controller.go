package payment

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// PaymentCallbackController 支付回调控制器（公共接口，无需登录）
type PaymentCallbackController struct{}

// NewPaymentCallbackController 创建支付回调控制器
func NewPaymentCallbackController() *PaymentCallbackController {
	return &PaymentCallbackController{}
}

// Notify 旧统一回调入口，兜底兼容未带通道类型的网关
// 默认按 "epay" 处理；新接入通道应使用 NotifyByChannel
// @Summary 支付异步通知回调（旧入口，兼容 epay）
// @Tags Public-回调
// @Success 200 {string} string
// @Router /v1/public/payment/notify [post]
func (ctrl *PaymentCallbackController) Notify(c *gin.Context) {
	ctrl.handleNotify(c, "epay")
}

// NotifyByChannel 分通道异步通知回调
// 路由中的 :channel_type 用于日志与通道路由，具体验签仍由订单关联的 PayGateway 决定
// @Summary 支付异步通知回调（分通道）
// @Tags Public-回调
// @Param channel_type path string true "通道类型，如 epay"
// @Success 200 {string} string
// @Router /v1/public/payment/notify/:channel_type [post]
func (ctrl *PaymentCallbackController) NotifyByChannel(c *gin.Context) {
	channelType := strings.TrimSpace(c.Param("channel_type"))
	if channelType == "" {
		channelType = "epay"
	}
	ctrl.handleNotify(c, channelType)
}

// handleNotify 统一回调处理
func (ctrl *PaymentCallbackController) handleNotify(c *gin.Context, channelType string) {
	params := extractCallbackParams(c)
	// 回调参数明文落盘（含 sign/金额），便于对账排查；不做日志脱敏
	log.Printf("[Payment Notify] 收到回调 channel=%s: %v", channelType, params)

	ok, err := services.HandlePaymentNotify(params)
	if err != nil {
		log.Printf("[Payment Notify] 处理失败: channel=%s order_no=%s permanent=%v err=%v",
			channelType, params["out_trade_no"], services.IsPermanentPaymentNotifyError(err), err)
		// 永久错误（验签/金额/绑定等）返回 SUCCESS，避免网关无限重试；可重试错误返回 FAIL
		if services.IsPermanentPaymentNotifyError(err) {
			c.String(http.StatusOK, "SUCCESS")
			return
		}
		c.String(http.StatusOK, "FAIL")
		return
	}
	if !ok {
		c.String(http.StatusOK, "FAIL")
		return
	}

	c.String(http.StatusOK, "SUCCESS")
}

// Return 同步跳转回调
// @Summary 支付同步跳转回调
// @Tags Public-回调
// @Success 302 {string} string
// @Router /v1/public/payment/return [get]
func (ctrl *PaymentCallbackController) Return(c *gin.Context) {
	params := extractCallbackParams(c)
	log.Printf("[Payment Return] 收到跳转: %v", params)

	order, err := services.HandlePaymentReturn(params)

	// 构造前端跳转地址
	frontendURL := services.GetGlobalFrontendURL()
	if frontendURL == "" {
		if !config.IsProductionMode() {
			frontendURL = "http://localhost:5173"
		} else {
			log.Printf("[Payment Return] frontend_url missing, cannot redirect order=%v err=%v", order, err)
			c.String(http.StatusOK, "支付结果已处理，请联系管理员检查前端地址配置")
			return
		}
	}

	if err != nil || order == nil {
		// 验签失败或订单不存在，跳转到前端充值页并附加错误提示
		redirectURL := frontendURL + "/user/account/recharge?result=error&msg=invalid_callback"
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// 根据订单状态跳转
	var redirectURL string
	if order.Status == models.PaymentStatusPaid {
		redirectURL = frontendURL + "/user/account/recharge?result=success&order_no=" + order.OrderNo
	} else {
		// 可能异步回调还没到，前端会通过轮询接口再次检查
		redirectURL = frontendURL + "/user/account/recharge?result=pending&order_no=" + order.OrderNo
	}

	c.Redirect(http.StatusFound, redirectURL)
}

// extractCallbackParams 从请求中提取回调参数（支持 GET 和 POST）
func extractCallbackParams(c *gin.Context) map[string]string {
	params := make(map[string]string)

	// 标准回调参数列表（不同通道可能扩展，这里先取通用字段）
	keys := []string{
		"pid", "trade_no", "out_trade_no", "type", "name",
		"money", "trade_status", "sign", "sign_type",
	}

	for _, key := range keys {
		// 优先从 POST form 取值，其次从 URL query 取值
		value := c.PostForm(key)
		if value == "" {
			value = c.Query(key)
		}
		if value != "" {
			params[key] = value
		}
	}

	return params
}

// ========================================
// 注册路由
// ========================================

// RegisterRoutes 注册支付回调路由。
// 说明：这些路径在全局限流中间件里被故意跳过（见 middleware.isGlobalRateLimitExemptPath），
// 避免合法支付回调被 429 丢掉导致漏入账；业务防伪造靠验签与订单状态机，不靠限流。
func (ctrl *PaymentCallbackController) RegisterRoutes(group *gin.RouterGroup) {
	paymentGroup := group.Group("/payment")
	{
		// 异步通知（旧入口，兼容 epay 等未按通道配置回调的网关）
		paymentGroup.POST("/notify", ctrl.Notify)
		paymentGroup.GET("/notify", ctrl.Notify)

		// 异步通知（分通道入口，新接入通道推荐）
		paymentGroup.POST("/notify/:channel_type", ctrl.NotifyByChannel)
		paymentGroup.GET("/notify/:channel_type", ctrl.NotifyByChannel)

		// 同步跳转
		paymentGroup.GET("/return", ctrl.Return)
	}
}
