package public

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// getFrontendURL 从系统设置获取前端地址
func getFrontendURL() string {
	setting, err := models.GetSettingByKey("frontend_url")
	if err == nil && setting.Value != "" {
		return strings.TrimRight(setting.Value, "/")
	}
	return ""
}

// PaymentCallbackController 支付回调控制器（公共接口，无需登录）
type PaymentCallbackController struct{}

// NewPaymentCallbackController 创建支付回调控制器
func NewPaymentCallbackController() *PaymentCallbackController {
	return &PaymentCallbackController{}
}

// Notify 易支付异步通知回调
// 易支付服务器会以 GET 或 POST 方式发送回调
// 成功处理后必须返回纯文本 "SUCCESS"，否则易支付会重复通知
func (ctrl *PaymentCallbackController) Notify(c *gin.Context) {
	// 从 GET 或 POST 参数中提取回调数据
	params := extractCallbackParams(c)

	log.Printf("[Payment Notify] 收到回调: %v", params)

	ok, err := services.HandlePaymentNotify(params)
	if !ok || err != nil {
		log.Printf("[Payment Notify] 处理失败: %v", err)
		c.String(http.StatusOK, "FAIL")
		return
	}

	// 必须返回 "SUCCESS" 告知易支付已成功处理
	c.String(http.StatusOK, "SUCCESS")
}

// Return 易支付同步跳转回调
// 用户支付完成后浏览器跳转回来，仅做页面跳转，不做到账处理
func (ctrl *PaymentCallbackController) Return(c *gin.Context) {
	params := extractCallbackParams(c)

	log.Printf("[Payment Return] 收到跳转: %v", params)

	order, err := services.HandlePaymentReturn(params)

	// 构造前端跳转地址
	frontendURL := getFrontendURL()

	if err != nil || order == nil {
		// 验签失败或订单不存在，跳转到前端充值页并附加错误提示
		redirectURL := frontendURL + "/#/user/recharge?result=error&msg=invalid_callback"
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// 根据订单状态跳转
	var redirectURL string
	if order.Status == models.PaymentStatusPaid {
		redirectURL = frontendURL + "/#/user/recharge?result=success&order_no=" + order.OrderNo
	} else {
		// 可能异步回调还没到，前端会通过轮询接口再次检查
		redirectURL = frontendURL + "/#/user/recharge?result=pending&order_no=" + order.OrderNo
	}

	c.Redirect(http.StatusFound, redirectURL)
}

// extractCallbackParams 从请求中提取回调参数（支持 GET 和 POST）
func extractCallbackParams(c *gin.Context) map[string]string {
	params := make(map[string]string)

	// 易支付标准回调参数列表
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

// RegisterRoutes 注册支付回调路由
func (ctrl *PaymentCallbackController) RegisterRoutes(group *gin.RouterGroup) {
	payment := group.Group("/payment")
	{
		// 异步通知（支持 GET 和 POST，因为不同易支付网关可能使用不同方式）
		payment.POST("/notify", ctrl.Notify)
		payment.GET("/notify", ctrl.Notify)
		// 同步跳转
		payment.GET("/return", ctrl.Return)
	}
}
