package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/db"
	"fst/backend/pkg/payment"
	"fst/backend/utils"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CreatePaymentOrderRequest 创建支付订单请求
type CreatePaymentOrderRequest struct {
	GatewayID uint64  // 支付通道ID
	Amount    float64 // 充值金额
	Subject   string  // 订单标题（可选）
	ClientIP  string  // 客户端IP
}

// CreatePaymentOrderResponse 创建支付订单响应
type CreatePaymentOrderResponse struct {
	OrderNo         string  `json:"order_no"`
	TradeNo         string  `json:"trade_no"`
	PayURL          string  `json:"pay_url"`
	Currency        string  `json:"currency"`
	Amount          float64 `json:"amount"`
	Fee             float64 `json:"fee"`
	PayAmount       float64 `json:"pay_amount"`
	TargetCurrency  string  `json:"target_currency"`
	TargetAmount    float64 `json:"target_amount"`
	TargetFee       float64 `json:"target_fee"`
	TargetPayAmount float64 `json:"target_pay_amount"`
	ExpireAt        int64   `json:"expire_at"`
	GatewayName     string  `json:"gateway_name"`
	PaymentType     string  `json:"payment_type"`
}

// CreatePaymentOrder 创建支付订单并生成支付链接（多通道版本）
func CreatePaymentOrder(userID uint64, req *CreatePaymentOrderRequest, notifyURL, returnURL string) (*CreatePaymentOrderResponse, error) {
	// 1. 检查全局支付开关
	if !GetGlobalPaymentEnabled() {
		return nil, NewClientError("支付功能未启用")
	}
	if req.Amount <= 0 {
		return nil, NewClientError("充值金额必须大于 0")
	}
	// 入参金额先按「分」规范化，后续手续费/落库全部对齐分精度
	normalizedAmount, normErr := utils.NormalizeYuan(req.Amount)
	if normErr != nil || normalizedAmount <= 0 {
		return nil, NewClientError("充值金额非法")
	}
	req.Amount = normalizedAmount

	// 2. 获取支付通道
	gateway, err := models.GetPayGatewayByID(req.GatewayID)
	if err != nil {
		return nil, NewClientError("支付通道不存在")
	}
	if gateway.Status != models.PayGatewayStatusEnabled {
		return nil, NewClientError("该支付通道已禁用")
	}

	// 3. 检查用户是否存在 + 等级校验
	user, err := models.GetUserByID(userID)
	if err != nil || user == nil {
		return nil, NewClientError("用户不存在")
	}
	if gateway.MinLevel > 0 && int(user.Level) < gateway.MinLevel {
		return nil, NewClientError(fmt.Sprintf("该通道要求最低等级 Lv.%d，您当前等级 Lv.%d", gateway.MinLevel, user.Level))
	}

	// 4. 验证金额范围（通道级别）
	if gateway.MinAmount > 0 && req.Amount < gateway.MinAmount {
		return nil, NewClientError(fmt.Sprintf("该通道最低充值金额为 ¥%.2f", gateway.MinAmount))
	}
	if gateway.MaxAmount > 0 && req.Amount > gateway.MaxAmount {
		return nil, NewClientError(fmt.Sprintf("该通道最高充值金额为 ¥%.2f", gateway.MaxAmount))
	}

	// 6. 计算原币种平台手续费（费率配置异常，如包含模式下费率=100%，可能导致到账金额为 0，此时直接拒绝建单）
	fee, payAmount, creditAmount := CalculateFee(req.Amount, gateway.FeeRate, gateway.FeeMode)
	if creditAmount <= 0 {
		return nil, NewClientError("手续费配置异常，到账金额为 0，请联系管理员")
	}

	// 7. 订单源币种始终是系统本位币；通道 currency/target_currency 仅表示上游实际收款币种。
	// 汇率换算必须以用户应付金额为基数：add 模式包含加收手续费，include 模式保持原支付总额。
	// 余额实际到账仍使用 creditAmount，避免把手续费计入用户余额。
	orderCurrency := payment.NormalizeCurrency(GetBaseCurrency())
	if orderCurrency == "" {
		orderCurrency = payment.DefaultBaseCurrency
	}
	targetResult, err := ConvertOrderAmountToTarget(payAmount, orderCurrency, gateway)
	if err != nil {
		return nil, NewClientError("货币转换失败: " + err.Error())
	}

	// 8. 获取订单过期时间
	expireMinutes := getOrderExpireMinutes()
	expireAt := time.Now().Add(time.Duration(expireMinutes) * time.Minute).Unix()

	// 9. 设置订单标题
	subject := req.Subject
	if subject == "" {
		subject = "余额充值"
	}

	// 10. 检查待支付订单数（防刷）+ 创建订单：同一事务内锁定用户行，防止并发请求绕过 10 单限制
	order := &models.PaymentOrder{
		OrderNo:             models.GenerateOrderNo(),
		UserID:              userID,
		GatewayID:           gateway.ID,
		PaymentChannel:      gateway.Type,
		PaymentType:         gateway.PayType,
		Currency:            orderCurrency,
		Amount:              creditAmount,
		Fee:                 fee,
		PayAmount:           payAmount,
		TargetCurrency:      targetResult.TargetCurrency,
		TargetAmount:        targetResult.TargetAmount,
		TargetFee:           targetResult.TargetFee,
		TargetPayAmount:     targetResult.TargetPayAmount,
		TargetCredit:        targetResult.TargetCredit,
		ExchangeRate:        targetResult.ExchangeRate,
		ExchangeFixedAmount: targetResult.ExchangeFixed,
		Subject:             subject,
		Status:              models.PaymentStatusPending,
		ExpireAt:            expireAt,
		ClientIP:            utils.ClampStoredIP(req.ClientIP),
	}
	if err := createPaymentOrderWithPendingLimitTx(userID, order); err != nil {
		return nil, err
	}

	// 10. 根据通道类型发起支付：优先 pkg/payment Provider，兼容旧 PaymentChannel
	gwNotifyURL := notifyURL
	if gateway.NotifyURL != "" {
		gwNotifyURL = gateway.NotifyURL
	}

	payURL, tradeNoFromRemote, createErr := createPayWithProvider(context.Background(), gateway, order, gwNotifyURL, returnURL)
	if createErr != nil {
		log.Printf("[Payment] 创建远程支付失败: type=%s, err=%v", gateway.Type, createErr)
		models.UpdatePaymentOrderStatus(order.OrderNo, models.PaymentStatusFailed, "")
		_ = recordPaymentException(&models.PaymentException{
			OrderNo:       order.OrderNo,
			UserID:        userID,
			GatewayID:     gateway.ID,
			ExceptionType: models.PaymentExceptionPermanentRejected,
			Status:        models.PaymentExceptionStatusOpen,
			Source:        "create",
			Message:       "远程建单失败",
			Detail:        fmt.Sprintf(`{"err":%q}`, truncateForException(createErr.Error(), 200)),
			OrderStatus:   models.PaymentStatusFailed,
		})
		return nil, errors.New("Failed to generate payment link, please check payment configuration")
	}
	order.TradeNo = models.NormalizeTradeNo(tradeNoFromRemote)

	// 11. 保存支付链接到订单。
	// 远程已建单成功时：本地落库失败仍标记 failed 以释放待支付配额，但 failed→paid 允许迟到回调/对账恢复；
	// 同时写异常单便于人工与对账任务跟进（尽量带上 trade_no）。
	order.PayURL = payURL
	if err := models.UpdatePaymentOrderPaymentInfo(order.OrderNo, order.TradeNo, payURL); err != nil {
		log.Printf("[Payment] 保存支付链接失败: order_no=%s, err=%v", order.OrderNo, err)
		if failErr := models.UpdatePaymentOrderStatus(order.OrderNo, models.PaymentStatusFailed, order.TradeNo); failErr != nil {
			log.Printf("[Payment] 标记脏单失败状态也失败: order_no=%s, err=%v", order.OrderNo, failErr)
		}
		_ = recordPaymentException(&models.PaymentException{
			OrderNo:       order.OrderNo,
			UserID:        userID,
			GatewayID:     gateway.ID,
			ExceptionType: models.PaymentExceptionRemoteLocalSaveFail,
			Status:        models.PaymentExceptionStatusOpen,
			Source:        "create",
			Message:       "远程建单成功但本地保存支付信息失败，等待回调或对账恢复",
			Detail:        fmt.Sprintf(`{"err":%q,"has_pay_url":true}`, truncateForException(err.Error(), 200)),
			OrderStatus:   models.PaymentStatusFailed,
			TradeNo:       order.TradeNo,
		})
		return nil, errors.New("Failed to save payment order, please retry later")
	}

	log.Printf("[Payment] 订单创建成功: order_no=%s, user_id=%d, amount=%.2f(%s), fee=%.2f, pay_amount=%.2f, target=%.2f(%s), gateway=%s",
		order.OrderNo, userID, order.Amount, order.Currency, fee, payAmount, order.TargetPayAmount, order.TargetCurrency, gateway.Name)

	tradeNo := models.NormalizeTradeNo(order.TradeNo)

	return &CreatePaymentOrderResponse{
		OrderNo:         order.OrderNo,
		TradeNo:         tradeNo,
		PayURL:          payURL,
		Currency:        order.Currency,
		Amount:          order.Amount,
		Fee:             order.Fee,
		PayAmount:       order.PayAmount,
		TargetCurrency:  order.TargetCurrency,
		TargetAmount:    order.TargetAmount,
		TargetFee:       order.TargetFee,
		TargetPayAmount: order.TargetPayAmount,
		ExpireAt:        order.ExpireAt,
		GatewayName:     gateway.Name,
		PaymentType:     gateway.PayType,
	}, nil
}

// maxPendingOrdersPerUser 单用户允许同时存在的最大待支付订单数（防刷）
const maxPendingOrdersPerUser = 10

// createPaymentOrderWithPendingLimitTx 在同一事务内锁定用户行、统计待支付订单数并创建订单，
// 避免「先查数量、再建单」两步之间的并发窗口导致限流被绕过（同一用户并发点击多次）。
func createPaymentOrderWithPendingLimitTx(userID uint64, order *models.PaymentOrder) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		// 锁定用户行，串行化同一用户的并发建单请求
		var user models.User
		if err := db.ForUpdate(tx).
			Select("id").
			Where("id = ? AND delete_time IS NULL", userID).
			First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NewClientError("用户不存在")
			}
			return err
		}

		pendingCount, err := models.CountPendingOrdersByUserIDTx(tx, userID)
		if err != nil {
			return errors.New("Failed to check pending payment order, please retry later")
		}
		if pendingCount >= maxPendingOrdersPerUser {
			return NewClientError("您有过多未支付订单，请先支付或等待过期后重试")
		}

		if err := models.CreatePaymentOrderTx(tx, order); err != nil {
			log.Printf("[Payment] 创建订单失败: %v", err)
			return errors.New("Failed to create order, please retry later")
		}
		return nil
	})
}

func settleThirdPartyPaidOrderTx(tx *gorm.DB, orderNo, tradeNo, paymentType, moneyStr, source string) (*models.PaymentOrder, *utils.BalanceResult, bool, error) {
	order, err := models.GetPaymentOrderForUpdate(tx, orderNo)
	if err != nil {
		log.Printf("[Payment] 订单不存在: source=%s, order_no=%s, err=%v", source, orderNo, err)
		return nil, nil, false, errors.New("Order does not exist")
	}

	if order.Status == models.PaymentStatusPaid {
		log.Printf("[Payment] 订单已支付（%s 幂等跳过）: order_no=%s", source, orderNo)
		return order, nil, false, nil
	}
	// 允许 pending / canceled / failed 在验签+金额+绑定通过后恢复到账（迟到回调 / 对账补单）
	if order.Status != models.PaymentStatusPending &&
		order.Status != models.PaymentStatusCanceled &&
		order.Status != models.PaymentStatusFailed {
		log.Printf("[Payment] 订单状态不允许处理: source=%s, order_no=%s, status=%d", source, orderNo, order.Status)
		return order, nil, false, newPaymentNotifyError(true, "订单状态不允许处理回调")
	}
	prevStatus := order.Status

	// 事务内再次校验：支付方式 + 交易号（网关/PID 已在外层 HandlePaymentNotify / 查单完成）
	if err := validatePaymentNotifyBinding(order, nil, "", paymentType, tradeNo); err != nil {
		log.Printf("[Payment] 绑定校验失败: source=%s, order_no=%s, err=%v", source, orderNo, err)
		return order, nil, false, err
	}
	expectedMoney := order.PayAmount
	if order.TargetCurrency != "" && order.TargetCurrency != order.Currency {
		expectedMoney = order.TargetPayAmount
	}
	if err := validateCallbackMoney(expectedMoney, moneyStr); err != nil {
		log.Printf("[Payment] 金额校验失败: source=%s, order_no=%s, err=%v", source, orderNo, err)
		return order, nil, false, err
	}

	balanceResult, err := utils.ExecuteBalanceOpTx(tx, &utils.BalanceReq{
		UserID: order.UserID,
		Amount: order.Amount,
		MemoI18n: map[string]string{
			"zhCN": fmt.Sprintf("在线充值-订单号%s", orderNo),
			"enUS": fmt.Sprintf("Online Recharge - Order#%s", orderNo),
		},
		OrderNo:     orderNo,
		TradeNo:     tradeNo,
		OrderStatus: models.PaymentStatusPaid,
	}, utils.OpFull)
	if err != nil {
		return order, nil, false, fmt.Errorf("Recharge credit failed: %w", err)
	}

	// 迟到恢复：写入审计异常（同事务，已处理）
	if prevStatus == models.PaymentStatusCanceled || prevStatus == models.PaymentStatusFailed {
		exType := models.PaymentExceptionLateCallback
		if source == "主动查单" || source == "reconcile" {
			exType = models.PaymentExceptionReconcilePaid
		}
		_ = models.CreatePaymentExceptionTx(tx, &models.PaymentException{
			OrderNo:       orderNo,
			UserID:        order.UserID,
			GatewayID:     order.GatewayID,
			ExceptionType: exType,
			Status:        models.PaymentExceptionStatusResolved,
			Source:        source,
			Message:       fmt.Sprintf("订单从状态 %d 恢复为已支付并到账", prevStatus),
			Detail:        fmt.Sprintf(`{"prev_status":%d,"source":%q}`, prevStatus, source),
			OrderStatus:   prevStatus,
			TradeNo:       tradeNo,
			ResolvedBy:    0,
			ResolveRemark: "系统自动恢复",
		})
	}

	return order, balanceResult, true, nil
}

// PaymentNotifyError 回调处理错误：Permanent=true 时网关应停止重试（返回 SUCCESS）
type PaymentNotifyError struct {
	Permanent bool
	Msg       string
}

func (e *PaymentNotifyError) Error() string {
	if e == nil {
		return ""
	}
	return e.Msg
}

func newPaymentNotifyError(permanent bool, msg string) error {
	return &PaymentNotifyError{Permanent: permanent, Msg: msg}
}

// IsPermanentPaymentNotifyError 判断是否为永久回调错误（验签/金额/绑定等）
func IsPermanentPaymentNotifyError(err error) bool {
	var pe *PaymentNotifyError
	if errors.As(err, &pe) {
		return pe.Permanent
	}
	// 历史字符串错误：视为永久（避免网关无限重试）
	if err == nil {
		return false
	}
	msg := err.Error()
	permanentMsgs := []string{
		"Signature verification failed", "Merchant ID mismatch", "Transaction number mismatch", "Payment gateway mismatch", "Payment method mismatch",
		"Callback payment type mismatch", "Callback amount does not match order amount", "Callback amount cannot be empty", "Invalid callback amount format",
		"Invalid callback amount", "Invalid order amount", "Order does not exist", "Payment gateway does not exist", "Unsupported payment gateway type",
		"Incomplete callback parameters", "Order status does not allow callback processing",
	}
	for _, m := range permanentMsgs {
		if strings.Contains(msg, m) {
			return true
		}
	}
	return false
}

func recordPaymentException(ex *models.PaymentException) error {
	if ex == nil {
		return nil
	}
	if err := models.CreatePaymentException(ex); err != nil {
		log.Printf("[Payment] 写入异常记录失败: type=%s order_no=%s err=%v", ex.ExceptionType, ex.OrderNo, err)
		return err
	}
	return nil
}

func truncateForException(s string, max int) string {
	return utils.ClampBytes(s, max)
}

func classifyNotifyExceptionType(err error) string {
	if err == nil {
		return models.PaymentExceptionPermanentRejected
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "signature") || strings.Contains(msg, "Signature"):
		return models.PaymentExceptionSignFailed
	case strings.Contains(msg, "amount") || strings.Contains(msg, "Amount"):
		return models.PaymentExceptionAmountMismatch
	case strings.Contains(msg, "mismatch") || strings.Contains(msg, "Mismatch") || strings.Contains(msg, "binding") || strings.Contains(msg, "Binding"):
		return models.PaymentExceptionBindingMismatch
	case strings.Contains(msg, "Order does not exist"):
		return models.PaymentExceptionOrderMissing
	default:
		return models.PaymentExceptionPermanentRejected
	}
}

// HandlePaymentNotify 处理多通道异步回调
// body / headers 供 WeChat V3 / Stripe / PayPal 等需要 raw body + headers 的通道使用。
// 返回: 是否处理成功, 错误信息。永久错误也会返回 err，由控制器决定回 SUCCESS 停重试。
func HandlePaymentNotify(params map[string]string, body []byte, headers map[string]string) (bool, error) {
	// 1. 提取系统订单号：form/query 优先，其次尝试 JSON body
	outTradeNo := strings.TrimSpace(params["out_trade_no"])
	if outTradeNo == "" && len(body) > 0 {
		outTradeNo = extractOutTradeNoFromJSONBody(body)
	}

	var tradeNo, tradeStatus, moneyStr, callbackType, pid string
	if outTradeNo == "" {
		return false, newPaymentNotifyError(true, "回调参数不完整")
	}

	// 2. 查找订单获取对应通道
	orderForGateway, err := models.GetPaymentOrderByOrderNo(outTradeNo)
	if err != nil {
		_ = recordPaymentException(&models.PaymentException{
			OrderNo:       outTradeNo,
			ExceptionType: models.PaymentExceptionOrderMissing,
			Status:        models.PaymentExceptionStatusOpen,
			Source:        "notify",
			Message:       "回调订单不存在",
			TradeNo:       tradeNo,
			OrderStatus:   -1,
		})
		return false, newPaymentNotifyError(true, "订单不存在")
	}

	// 3. 获取通道配置以验签
	gateway, err := models.GetPayGatewayByID(orderForGateway.GatewayID)
	if err != nil {
		return false, newPaymentNotifyError(true, "支付通道不存在")
	}

	// 4. 验证签名（防篡改）：优先 pkg/payment Provider
	verified, payload, usedPayload := verifyNotifyWithProvider(gateway, params, body, headers)
	if !verified {
		log.Printf("[Payment] 回调签名验证失败: order_no=%s", outTradeNo)
		_ = recordPaymentException(&models.PaymentException{
			OrderNo:       outTradeNo,
			UserID:        orderForGateway.UserID,
			GatewayID:     orderForGateway.GatewayID,
			ExceptionType: models.PaymentExceptionSignFailed,
			Status:        models.PaymentExceptionStatusOpen,
			Source:        "notify",
			Message:       "回调签名验证失败",
			OrderStatus:   orderForGateway.Status,
			TradeNo:       tradeNo,
		})
		return false, newPaymentNotifyError(true, "签名验证失败")
	}

	// 4a. 若通道返回了归一化 payload，用它覆盖 form/query 中的字段
	if usedPayload && payload != nil {
		if payload.TradeNo != "" {
			tradeNo = models.NormalizeTradeNo(payload.TradeNo)
		}
		if payload.TradeStatus != "" {
			tradeStatus = payload.TradeStatus
		}
		if payload.Money != "" {
			moneyStr = payload.Money
		}
		if payload.PayType != "" {
			callbackType = payload.PayType
		}
	} else {
		// 旧 form/query 通道，从 params 提取
		tradeNo = models.NormalizeTradeNo(params["trade_no"])
		tradeStatus = params["trade_status"]
		moneyStr = params["money"]
		callbackType = strings.TrimSpace(params["type"])
	}

	pid = strings.TrimSpace(params["pid"])

	if tradeNo == "" {
		return false, newPaymentNotifyError(true, "回调缺少第三方交易号")
	}
	// 5. 订单与通道/商户/支付方式绑定校验（防串单）
	if err := validatePaymentNotifyBinding(orderForGateway, gateway, pid, callbackType, tradeNo); err != nil {
		log.Printf("[Payment] 回调绑定校验失败: order_no=%s, gateway_id=%d, err=%v", outTradeNo, orderForGateway.GatewayID, err)
		_ = recordPaymentException(&models.PaymentException{
			OrderNo:       outTradeNo,
			UserID:        orderForGateway.UserID,
			GatewayID:     orderForGateway.GatewayID,
			ExceptionType: models.PaymentExceptionBindingMismatch,
			Status:        models.PaymentExceptionStatusOpen,
			Source:        "notify",
			Message:       err.Error(),
			OrderStatus:   orderForGateway.Status,
			TradeNo:       tradeNo,
		})
		return false, newPaymentNotifyError(true, err.Error())
	}

	// 5. 只处理 TRADE_SUCCESS 状态
	if tradeStatus != "TRADE_SUCCESS" {
		log.Printf("[Payment] 非成功状态回调: order_no=%s, status=%s", outTradeNo, tradeStatus)
		if countErr := models.IncrementNotifyCount(outTradeNo); countErr != nil {
			log.Printf("[Payment] 更新回调通知次数失败: order_no=%s, err=%v", outTradeNo, countErr)
		}
		return true, nil
	}

	// 金额预检（入事务前也记异常，便于审计）
	expectedMoney := orderForGateway.PayAmount
	if orderForGateway.TargetCurrency != "" && orderForGateway.TargetCurrency != orderForGateway.Currency {
		expectedMoney = orderForGateway.TargetPayAmount
	}
	if err := validateCallbackMoney(expectedMoney, moneyStr); err != nil {
		log.Printf("[Payment] 回调金额不符: order_no=%s, err=%v", outTradeNo, err)
		_ = recordPaymentException(&models.PaymentException{
			OrderNo:       outTradeNo,
			UserID:        orderForGateway.UserID,
			GatewayID:     orderForGateway.GatewayID,
			ExceptionType: models.PaymentExceptionAmountMismatch,
			Status:        models.PaymentExceptionStatusOpen,
			Source:        "notify",
			Message:       err.Error(),
			Detail:        fmt.Sprintf(`{"expected_pay_amount":%.2f}`, orderForGateway.PayAmount),
			OrderStatus:   orderForGateway.Status,
			TradeNo:       tradeNo,
		})
		return false, newPaymentNotifyError(true, err.Error())
	}

	// 6. 在事务中处理到账（保证原子性+幂等性）
	var order *models.PaymentOrder
	var balanceResult *utils.BalanceResult
	var changed bool
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var innerErr error
		order, balanceResult, changed, innerErr = settleThirdPartyPaidOrderTx(tx, outTradeNo, tradeNo, callbackType, moneyStr, "notify")
		return innerErr
	})
	if err != nil {
		// 失败也记通知次数（事务外，避免回滚吞掉计数）
		if countErr := models.IncrementNotifyCount(outTradeNo); countErr != nil {
			log.Printf("[Payment] 更新回调通知次数失败: order_no=%s, err=%v", outTradeNo, countErr)
		}
		// 可重试的数据库类错误保持非永久；业务永久错误包装
		if IsPermanentPaymentNotifyError(err) {
			return false, err
		}
		if _, ok := err.(*PaymentNotifyError); ok {
			return false, err
		}
		return false, err
	}

	if changed {
		log.Printf("[Payment] 充值到账成功: order_no=%s, user_id=%d, amount=%.2f, fee=%.2f, pay_amount=%.2f, before=%.2f, after=%.2f",
			outTradeNo, order.UserID, order.Amount, order.Fee, order.PayAmount, balanceResult.BeforeMoney, balanceResult.AfterMoney)
	}

	return true, nil
}

// HandlePaymentReturn 处理同步跳转回调（仅验签+查询状态，不做到账）
func HandlePaymentReturn(params map[string]string) (*models.PaymentOrder, error) {
	outTradeNo := params["out_trade_no"]
	if outTradeNo == "" {
		return nil, errors.New("Missing order number parameter")
	}

	order, err := models.GetPaymentOrderByOrderNo(outTradeNo)
	if err != nil {
		return nil, errors.New("Order does not exist")
	}

	// 获取通道密钥验签：优先 pkg/payment Provider
	gateway, err := models.GetPayGatewayByID(order.GatewayID)
	if err != nil {
		return nil, errors.New("Payment gateway does not exist")
	}

	verified, _, _ := verifyNotifyWithProvider(gateway, params, nil, nil)
	if !verified {
		return nil, errors.New("Signature verification failed")
	}

	return order, nil
}

// extractOutTradeNoFromJSONBody 尝试从 JSON 回调体中找系统订单号
func extractOutTradeNoFromJSONBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	// 先按通用 object 解析
	var root map[string]interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		return ""
	}

	// 按常见路径查找：
	// 1. 顶层 out_trade_no（微信 V3 等）
	if v, ok := root["out_trade_no"].(string); ok && v != "" {
		return v
	}

	// 2. Stripe / PayPal 等嵌套在 resource / data.object 里
	paths := []string{"reference_id", "custom_id"}

	if resource, ok := root["resource"].(map[string]interface{}); ok {
		if ref, ok := resource["reference_id"].(string); ok && ref != "" {
			return ref
		}
		if custom, ok := resource["custom_id"].(string); ok && custom != "" {
			return custom
		}
		if units, ok := resource["purchase_units"].([]interface{}); ok && len(units) > 0 {
			if unit, ok := units[0].(map[string]interface{}); ok {
				for _, k := range paths {
					if v, ok := unit[k].(string); ok && v != "" {
						return v
					}
				}
			}
		}
	}

	if data, ok := root["data"].(map[string]interface{}); ok {
		if obj, ok := data["object"].(map[string]interface{}); ok {
			if meta, ok := obj["metadata"].(map[string]interface{}); ok {
				if v, ok := meta["order_no"].(string); ok && v != "" {
					return v
				}
			}
		}
	}

	return ""
}

// ReconcilePaymentOrderByID 主动查单并对账。
// 只查一次订单，reconcile 成功后直接返回事务内最新的订单对象，避免二次查询。
func ReconcilePaymentOrderByID(orderID uint64) (*models.PaymentOrder, bool, error) {
	order, err := models.GetPaymentOrderByID(orderID)
	if err != nil {
		return nil, false, err
	}

	updated, reconciled, err := reconcilePaymentOrder(order)
	if err != nil {
		return order, false, err
	}
	if !reconciled {
		return order, false, nil
	}
	return updated, true, nil
}

// PaymentReconcileBatchResult 批量对账结果
type PaymentReconcileBatchResult struct {
	Scanned    int
	Recovered  int
	Exceptions int
	Skipped    int
	Errors     []string
}

// ReconcilePaymentOrdersBatch 扫描待支付与近期取消/失败订单并向网关查单补账
// 新增：按通道开关、每通道批次、冷却间隔控制，避免高频空扫。
func ReconcilePaymentOrdersBatch(ctx context.Context, limit int) (*PaymentReconcileBatchResult, error) {
	result := &PaymentReconcileBatchResult{}
	orders, err := models.ListPaymentOrdersForReconcile(limit*3, 7*24*3600)
	if err != nil {
		return nil, err
	}

	// 预加载通道配置，按 gateway_id 分组
	gateways, err := models.GetEnabledPayGateways()
	if err != nil {
		return nil, err
	}
	gatewayMap := make(map[uint64]*models.PayGateway)
	for i := range gateways {
		gatewayMap[gateways[i].ID] = &gateways[i]
	}

	type channelPolicy struct {
		enabled     bool
		cooldownSec int64
		batchSize   int
		processed   int
	}
	channelPolicies := make(map[uint64]*channelPolicy)

	now := time.Now().Unix()
	for _, order := range orders {
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return result, err
			}
		}

		// 扫描总量达到上限即停止
		if result.Scanned >= limit {
			break
		}

		gw := gatewayMap[order.GatewayID]
		if gw == nil {
			result.Skipped++
			continue
		}

		policy, exists := channelPolicies[order.GatewayID]
		if !exists {
			policy = &channelPolicy{
				enabled:     gw.GetActiveQueryEnabled() != 0,
				cooldownSec: int64(gw.GetQueryIntervalSeconds()),
				batchSize:   gw.GetQueryBatchSize(),
			}
			if policy.batchSize <= 0 {
				policy.batchSize = 50
			}
			if policy.cooldownSec <= 0 {
				policy.cooldownSec = 120
			}
			channelPolicies[order.GatewayID] = policy
		}

		if !policy.enabled {
			result.Skipped++
			log.Printf("[Payment Reconcile] skipped order_no=%s reason=active_query_disabled gateway_id=%d", order.OrderNo, order.GatewayID)
			continue
		}
		if policy.processed >= policy.batchSize {
			result.Skipped++
			continue
		}
		// 冷却检查：last_query_at 与当前时间间隔小于通道查询间隔则跳过
		if order.LastQueryAt != nil && *order.LastQueryAt > 0 && now-*order.LastQueryAt < policy.cooldownSec {
			result.Skipped++
			continue
		}

		result.Scanned++
		policy.processed++
		o := order

		// 提前标记查询尝试，减少并发重复查单（不阻塞对账结果）
		_ = models.UpdatePaymentOrderQueryAttempt(o.OrderNo)

		_, changed, err := reconcilePaymentOrder(&o)
		if err != nil {
			result.Exceptions++
			msg := fmt.Sprintf("%s: %v", o.OrderNo, err)
			if len(result.Errors) < 20 {
				result.Errors = append(result.Errors, msg)
			}
			_ = recordPaymentException(&models.PaymentException{
				OrderNo:       o.OrderNo,
				UserID:        o.UserID,
				GatewayID:     o.GatewayID,
				ExceptionType: classifyNotifyExceptionType(err),
				Status:        models.PaymentExceptionStatusOpen,
				Source:        "reconcile",
				Message:       truncateForException(err.Error(), 400),
				OrderStatus:   o.Status,
				TradeNo:       o.TradeNo,
			})
			continue
		}
		if changed {
			result.Recovered++
		} else {
			result.Skipped++
		}
	}
	return result, nil
}

// reconcilePaymentOrder 对单笔订单向网关查单；支持 pending/canceled/failed 恢复。
// 返回事务内最新的订单对象、是否对账成功、错误。
func reconcilePaymentOrder(order *models.PaymentOrder) (*models.PaymentOrder, bool, error) {
	if order == nil {
		return nil, false, errors.New("Order does not exist")
	}
	switch order.Status {
	case models.PaymentStatusPending, models.PaymentStatusCanceled, models.PaymentStatusFailed:
		// ok
	default:
		return order, false, nil
	}
	if order.GatewayID == 0 {
		return order, false, nil
	}
	orderNo := strings.TrimSpace(order.OrderNo)
	tradeNo := models.NormalizeTradeNo(order.TradeNo)
	if orderNo == "" && tradeNo == "" {
		return order, false, nil
	}

	gateway, err := models.GetPayGatewayByID(order.GatewayID)
	if err != nil {
		return order, false, fmt.Errorf("Failed to get payment gateway: %w", err)
	}

	queryResult, err := queryOrderWithProvider(context.Background(), gateway, orderNo, tradeNo)
	if err != nil {
		return order, false, err
	}
	if queryResult == nil {
		return order, false, errors.New("Query result is empty")
	}
	if queryResult.Code != 1 {
		msg := strings.TrimSpace(queryResult.Msg)
		status := strings.TrimSpace(queryResult.TradeStatus)
		log.Printf("[Payment] 主动查单未返回成功结果: order_no=%s, trade_no=%s, code=%d, msg=%s, status=%s", orderNo, tradeNo, queryResult.Code, msg, status)
		// 永久失败（如订单号不存在）：落异常 + pending 标 failed，避免后台反复空扫网关
		if isPermanentGatewayQueryFailure(queryResult.Code, msg, status) {
			markPermanentReconcileFailure(order, queryResult.Code, msg, status)
		}
		return order, false, nil
	}
	if queryResult.OutTradeNo != "" && queryResult.OutTradeNo != orderNo {
		return order, false, newPaymentNotifyError(true, "云端订单号不匹配")
	}

	queryTradeNo := models.NormalizeTradeNo(queryResult.TradeNo)
	if queryTradeNo == "" {
		queryTradeNo = tradeNo
	}
	if err := validatePaymentNotifyBinding(order, gateway, strings.TrimSpace(gateway.PID), strings.TrimSpace(queryResult.Type), queryTradeNo); err != nil {
		return order, false, newPaymentNotifyError(true, err.Error())
	}
	expectedMoney := order.PayAmount
	if order.TargetCurrency != "" && order.TargetCurrency != order.Currency {
		expectedMoney = order.TargetPayAmount
	}
	if err := validateCallbackMoney(expectedMoney, queryResult.Money); err != nil {
		return order, false, newPaymentNotifyError(true, err.Error())
	}
	if queryResult.TradeStatus != "TRADE_SUCCESS" {
		log.Printf("[Payment] 主动查单未确认支付成功: order_no=%s, trade_no=%s, code=%d, msg=%s, status=%s", orderNo, tradeNo, queryResult.Code, strings.TrimSpace(queryResult.Msg), strings.TrimSpace(queryResult.TradeStatus))
		return order, false, nil
	}

	var changed bool
	var finalOrder *models.PaymentOrder
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var balanceResult *utils.BalanceResult
		var innerErr error
		finalOrder, balanceResult, changed, innerErr = settleThirdPartyPaidOrderTx(tx, order.OrderNo, queryTradeNo, strings.TrimSpace(queryResult.Type), queryResult.Money, "reconcile")
		if innerErr != nil {
			return innerErr
		}
		if changed {
			log.Printf("[Payment] 主动查单补账成功: order_no=%s, user_id=%d, amount=%.2f, fee=%.2f, pay_amount=%.2f, before=%.2f, after=%.2f",
				finalOrder.OrderNo, finalOrder.UserID, finalOrder.Amount, finalOrder.Fee, finalOrder.PayAmount, balanceResult.BeforeMoney, balanceResult.AfterMoney)
		}
		return nil
	})
	if err != nil {
		return order, false, err
	}
	if finalOrder != nil {
		*order = *finalOrder
	}

	return order, changed, nil
}

// isPermanentGatewayQueryFailure 判断网关查单是否为不可恢复错误（继续轮询只会浪费）
func isPermanentGatewayQueryFailure(_ int, msg, tradeStatus string) bool {
	text := strings.ToLower(strings.TrimSpace(msg + " " + tradeStatus))
	if text == "" {
		return false
	}
	permanentHints := []string{
		"订单号不存在",
		"订单不存在",
		"order not found",
		"order does not exist",
		"no such order",
		"invalid order",
	}
	for _, h := range permanentHints {
		if strings.Contains(text, strings.ToLower(h)) {
			return true
		}
	}
	return false
}

// markPermanentReconcileFailure 永久查单失败：写异常并尽量把 pending 标为 failed，后续扫描会排除
func markPermanentReconcileFailure(order *models.PaymentOrder, code int, msg, tradeStatus string) {
	if order == nil {
		return
	}
	exType := models.PaymentExceptionPermanentRejected
	lower := strings.ToLower(msg + " " + tradeStatus)
	if strings.Contains(lower, "订单号不存在") || strings.Contains(lower, "订单不存在") ||
		strings.Contains(lower, "order not found") || strings.Contains(lower, "order does not exist") {
		exType = models.PaymentExceptionOrderMissing
	}
	// 先标 failed，确保即使异常写入因死锁失败，批量扫描也不会再捞到该单
	if order.Status == models.PaymentStatusPending {
		if err := models.UpdatePaymentOrderStatus(order.OrderNo, models.PaymentStatusFailed, ""); err != nil {
			log.Printf("[Payment] 永久失败标 failed 失败: order_no=%s err=%v", order.OrderNo, err)
		} else {
			order.Status = models.PaymentStatusFailed
		}
	}
	detail := fmt.Sprintf(`{"code":%d,"msg":%q,"trade_status":%q}`, code, msg, tradeStatus)
	ex := &models.PaymentException{
		OrderNo:       order.OrderNo,
		UserID:        order.UserID,
		GatewayID:     order.GatewayID,
		ExceptionType: exType,
		Status:        models.PaymentExceptionStatusOpen,
		Source:        "reconcile",
		Message:       msg,
		Detail:        detail,
		OrderStatus:   order.Status,
		TradeNo:       models.NormalizeTradeNo(order.TradeNo),
	}
	if err := models.CreatePaymentException(ex); err != nil {
		log.Printf("[Payment] 写入永久查单失败异常失败: order_no=%s err=%v", order.OrderNo, err)
	}
	log.Printf("[Payment] 永久查单失败已停扫: order_no=%s type=%s msg=%s", order.OrderNo, exType, msg)
}

// AdminCompleteOrder 管理员手动补单。
// force=true 时允许对 canceled/failed 强制补单（高危）；默认仅 pending，且优先应走网关对账。
func AdminCompleteOrder(orderID uint64, memo string, force bool) error {
	memo = strings.TrimSpace(memo)
	if err := validateClientRuneLen(memo, "备注", utils.MaxCommentLength); err != nil {
		return err
	}
	order, err := models.GetPaymentOrderByID(orderID)
	if err != nil {
		return NewClientError("订单不存在")
	}

	if order.Status == models.PaymentStatusPaid {
		return NewClientError("订单已支付，无需重复操作")
	}
	if order.Status == models.PaymentStatusPending {
		// 默认：先尝试网关查单补账
		if _, changed, rerr := reconcilePaymentOrder(order); rerr == nil && changed {
			return nil
		}
	} else if order.Status == models.PaymentStatusCanceled || order.Status == models.PaymentStatusFailed {
		if !force {
			// 非强制：仅允许对账路径
			if _, changed, rerr := reconcilePaymentOrder(order); rerr == nil && changed {
				return nil
			}
			return NewClientError("订单非待支付状态；请先对账，或确认后强制补单")
		}
		if strings.TrimSpace(memo) == "" {
			return NewClientError("强制补单必须填写原因")
		}
	} else {
		return NewClientError("当前订单状态不允许补单")
	}

	return adminCompleteOrderExec(order, memo, force)
}

// adminCompleteOrderExec 实际入账逻辑（原 AdminCompleteOrder 事务体）
func adminCompleteOrderExec(order *models.PaymentOrder, memo string, force bool) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		lockedOrder, err := models.GetPaymentOrderForUpdate(tx, order.OrderNo)
		if err != nil {
			return fmt.Errorf("Failed to lock order: %w", err)
		}
		if lockedOrder.Status == models.PaymentStatusPaid {
			return NewClientError("订单已支付，无需重复操作")
		}
		if lockedOrder.Status != models.PaymentStatusPending &&
			!(force && (lockedOrder.Status == models.PaymentStatusCanceled || lockedOrder.Status == models.PaymentStatusFailed)) {
			return NewClientError("订单状态已变更或不允许补单")
		}

		prevStatus := lockedOrder.Status
		memoZh := fmt.Sprintf("管理员手动补单-订单号%s", order.OrderNo)
		memoEn := fmt.Sprintf("Admin Manual - Order#%s", order.OrderNo)
		if memo != "" {
			memoZh += " (" + memo + ")"
			memoEn += " (" + memo + ")"
		}
		if _, err = utils.ExecuteBalanceOpTx(tx, &utils.BalanceReq{
			UserID: order.UserID,
			Amount: order.Amount,
			MemoI18n: map[string]string{
				"zhCN": memoZh,
				"enUS": memoEn,
			},
			OrderNo:     order.OrderNo,
			TradeNo:     "MANUAL",
			OrderStatus: models.PaymentStatusPaid,
		}, utils.OpFull); err != nil {
			return fmt.Errorf("Reconciliation failed: %w", err)
		}

		_ = models.CreatePaymentExceptionTx(tx, &models.PaymentException{
			OrderNo:       order.OrderNo,
			UserID:        order.UserID,
			GatewayID:     order.GatewayID,
			ExceptionType: models.PaymentExceptionManualResolve,
			Status:        models.PaymentExceptionStatusResolved,
			Source:        "admin",
			Message:       fmt.Sprintf("管理员手动补单(prev=%d, force=%v)", prevStatus, force),
			Detail:        fmt.Sprintf(`{"memo":%q,"force":%v,"prev_status":%d}`, truncateForException(memo, 200), force, prevStatus),
			OrderStatus:   prevStatus,
			TradeNo:       "MANUAL",
			ResolveRemark: memo,
		})

		log.Printf("[Payment] 管理员手动补单成功: order_no=%s, user_id=%d, amount=%.2f, force=%v",
			order.OrderNo, order.UserID, order.Amount, force)
		return nil
	})
}

// AdminCancelOrder 管理员取消订单
func AdminCancelOrder(orderID uint64) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		order, err := models.GetPaymentOrderByIDForUpdate(tx, orderID)
		if err != nil {
			return NewClientError("订单不存在")
		}
		if order.Status != models.PaymentStatusPending {
			return NewClientError("只能取消待支付的订单")
		}
		return models.UpdatePaymentOrderStatusTx(tx, order.OrderNo, models.PaymentStatusCanceled, "")
	})
}

// CancelExpiredOrders 取消过期未支付订单（定时任务调用）
func CancelExpiredOrders() (int64, error) {
	affected, err := models.CancelExpiredOrders()
	if err != nil {
		log.Printf("[Payment] 取消过期订单失败: %v", err)
		return 0, err
	}
	if affected > 0 {
		log.Printf("[Payment] 已取消 %d 个过期订单", affected)
	}
	return affected, nil
}

func AdminDeleteOrder(orderID uint64) error {
	_ = orderID
	// 财务治理：支付订单禁止物理删除，保留完整审计链路；仅允许取消待支付单
	return NewClientError("支付订单禁止删除，请使用取消或保留审计记录")
}

// isNonStandardEpayCallbackType 判断易支付回调 type 是否为非标准值。
// 标准值（alipay/wxpay 等）必须与订单一致；空串、数字、厂商自定义值不强校验。
func isNonStandardEpayCallbackType(callbackType string) bool {
	callbackType = strings.TrimSpace(callbackType)
	if callbackType == "" {
		return true
	}
	switch strings.ToLower(callbackType) {
	case "alipay", "wxpay", "qqpay", "bank", "jdpay", "paypal", "usdt":
		return false
	default:
		return true
	}
}

// validatePaymentNotifyBinding 校验支付回调/查单结果与本地订单、通道绑定关系，防止串单。
// 校验项：商户号 PID、订单归属网关 ID、通道类型、支付方式、交易号、标准回调 type。
func validatePaymentNotifyBinding(order *models.PaymentOrder, gateway *models.PayGateway, pid, callbackType, tradeNo string) error {
	pid = strings.TrimSpace(pid)
	callbackType = strings.TrimSpace(callbackType)
	tradeNo = models.NormalizeTradeNo(tradeNo)

	if gateway != nil && strings.TrimSpace(gateway.PID) != "" && pid != "" && pid != strings.TrimSpace(gateway.PID) {
		return errors.New("Merchant ID mismatch")
	}
	// 网关已配置商户号时，回调必须带上且一致（避免空 pid 绕过）
	if gateway != nil && strings.TrimSpace(gateway.PID) != "" && pid == "" {
		return errors.New("Merchant ID mismatch")
	}

	if order != nil {
		if order.TradeNo != "" && tradeNo != "" && order.TradeNo != tradeNo {
			return errors.New("Transaction number mismatch")
		}
		if gateway != nil {
			if order.GatewayID != 0 && gateway.ID != 0 && order.GatewayID != gateway.ID {
				return errors.New("Payment gateway mismatch")
			}
			if order.PaymentChannel != "" && gateway.Type != "" &&
				!strings.EqualFold(order.PaymentChannel, gateway.Type) {
				return errors.New("Payment gateway type mismatch")
			}
			if order.PaymentType != "" && gateway.PayType != "" &&
				!strings.EqualFold(order.PaymentType, gateway.PayType) {
				return errors.New("Payment method mismatch")
			}
		}
		// 标准支付类型（alipay/wxpay 等）必须与订单一致；数字型自定义 type 放行
		if order.PaymentType != "" && callbackType != "" && !isNonStandardEpayCallbackType(callbackType) {
			if !strings.EqualFold(callbackType, order.PaymentType) {
				return errors.New("Callback payment type mismatch")
			}
		}
	}
	return nil
}

// validateCallbackMoney 支付回调/查单金额必须提供且与订单应付金额一致。
// 按「分」整数比较，避免 float 容差误放行或误拒。
func validateCallbackMoney(expected float64, moneyStr string) error {
	moneyStr = strings.TrimSpace(moneyStr)
	if moneyStr == "" {
		return errors.New("Callback amount cannot be empty")
	}
	callbackMoney, err := strconv.ParseFloat(moneyStr, 64)
	if err != nil {
		return errors.New("Invalid callback amount format")
	}
	expectedFen, err := utils.YuanToFen(expected)
	if err != nil {
		return errors.New("Invalid order amount")
	}
	callbackFen, err := utils.YuanToFen(callbackMoney)
	if err != nil {
		return errors.New("Invalid callback amount")
	}
	if expectedFen != callbackFen {
		return errors.New("Callback amount does not match order amount")
	}
	return nil
}

func validatePaymentOrderDeletion(status int) error {
	if status != models.PaymentStatusCanceled && status != models.PaymentStatusFailed {
		return NewClientError("仅允许删除已取消或失败的订单")
	}
	return nil
}

// getOrderExpireMinutes 从系统配置获取订单过期时间（分钟）
func getOrderExpireMinutes() int {
	return GetGlobalPaymentOrderExpireMinutes()
}

// gatewayExtConfig 解析支付通道的 ext_config，并注入部分基础字段，方便 Provider.TestConnection 读取。
func gatewayExtConfig(gateway *models.PayGateway) map[string]string {
	if gateway == nil {
		return map[string]string{}
	}
	extConfig := payment.ParseExtConfig(gateway.ExtConfig)
	// 兼容旧通道：没有 ext_config 时，从模型 getter 取密钥兜底
	if len(extConfig) == 0 && gateway.GetKey() != "" {
		extConfig = map[string]string{"key": gateway.GetKey()}
	}
	if extConfig["key"] == "" && gateway.GetKey() != "" {
		extConfig["key"] = gateway.GetKey()
	}
	if extConfig["sign_type"] == "" {
		extConfig["sign_type"] = gateway.GetSignType()
	}
	if extConfig["version"] == "" {
		extConfig["version"] = gateway.Version
	}
	if extConfig["api_url"] == "" {
		extConfig["api_url"] = gateway.ApiURL
	}
	if extConfig["pid"] == "" {
		extConfig["pid"] = gateway.PID
	}
	return extConfig
}

// gatewaySignType 返回通道签名算法，空值兜底 MD5
func gatewaySignType(gateway *models.PayGateway) string {
	if gateway == nil {
		return "MD5"
	}
	return gateway.GetSignType()
}

// createPayWithProvider 优先使用 pkg/payment Provider 创建支付订单，未注册则回退到旧 PaymentChannel
func createPayWithProvider(ctx context.Context, gateway *models.PayGateway, order *models.PaymentOrder, notifyURL, returnURL string) (string, string, error) {
	if provider := payment.GetProvider(gateway.Type); provider != nil {
		targetCurrency := payment.NormalizeCurrency(order.TargetCurrency)
		if targetCurrency == "" {
			targetCurrency = payment.NormalizeCurrency(gateway.Currency)
		}
		targetMoney := order.TargetPayAmount
		if targetMoney <= 0 {
			targetMoney = order.PayAmount
		}

		req := &payment.CreatePayRequest{
			PID:            gateway.PID,
			ExtConfig:      gatewayExtConfig(gateway),
			ApiURL:         gateway.ApiURL,
			PayType:        order.PaymentType,
			SignType:       gatewaySignType(gateway),
			Version:        gateway.Version,
			Device:         gateway.Device,
			OrderNo:        order.OrderNo,
			Subject:        order.Subject,
			Money:          fmt.Sprintf("%.2f", order.PayAmount),
			Currency:       strings.ToUpper(strings.TrimSpace(gateway.Currency)),
			TargetMoney:    fmt.Sprintf("%.2f", targetMoney),
			TargetCurrency: targetCurrency,
			NotifyURL:      notifyURL,
			ReturnURL:      returnURL,
			ClientIP:       order.ClientIP,
		}
		if !provider.ValidatePayType(req.PayType, req.ExtConfig) {
			return "", "", NewClientError("支付方式不受该通道支持")
		}
		resp, err := provider.CreatePay(ctx, req)
		if err != nil {
			return "", "", err
		}
		if resp == nil {
			return "", "", errors.New("payment provider returned empty response")
		}
		return resp.PayURL, resp.TradeNo, nil
	}

	// 兼容旧注册表
	channel, ok := GetPaymentChannel(gateway.Type)
	if !ok {
		return "", "", errors.New("unsupported payment channel type")
	}
	if !channel.ValidatePayType(gateway, order.PaymentType) {
		return "", "", NewClientError("支付方式不受该通道支持")
	}
	return channel.CreatePay(gateway, order, notifyURL, returnURL)
}

// verifyNotifyWithProvider 优先使用 pkg/payment Provider 验签
// 返回 (是否通过, 归一化 payload, 是否使用了 PayloadVerifier)
func verifyNotifyWithProvider(gateway *models.PayGateway, params map[string]string, body []byte, headers map[string]string) (bool, *payment.NotifyPayload, bool) {
	provider := payment.GetProvider(gateway.Type)
	if provider == nil {
		channel, ok := GetPaymentChannel(gateway.Type)
		if !ok {
			return false, nil, false
		}
		return channel.VerifyNotify(params, gateway.GetKey()), nil, false
	}

	// 若 Provider 实现了 PayloadVerifier 且有 body/headers，优先走 raw body 验签
	if pv, ok := provider.(payment.PayloadVerifier); ok && (len(body) > 0 || len(headers) > 0) {
		verified, payload, err := pv.VerifyNotifyWithPayload(context.Background(), body, headers, gatewaySignType(gateway), gatewayExtConfig(gateway))
		if err != nil {
			log.Printf("[Payment] %s PayloadVerifier error: %v", gateway.Type, err)
			return false, nil, true
		}
		return verified, payload, true
	}

	// 旧 form/query 回调
	return provider.VerifyNotify(params, gatewaySignType(gateway), gatewayExtConfig(gateway)), nil, false
}

// queryOrderWithProvider 优先使用 pkg/payment Provider 查单
func queryOrderWithProvider(ctx context.Context, gateway *models.PayGateway, orderNo, tradeNo string) (*payment.QueryOrderResponse, error) {
	if provider := payment.GetProvider(gateway.Type); provider != nil {
		return provider.QueryOrder(ctx, &payment.QueryOrderRequest{
			PID:       gateway.PID,
			ExtConfig: gatewayExtConfig(gateway),
			SignType:  gatewaySignType(gateway),
			Version:   gateway.Version,
			ApiURL:    gateway.ApiURL,
			OrderNo:   orderNo,
			TradeNo:   tradeNo,
		})
	}
	channel, ok := GetPaymentChannel(gateway.Type)
	if !ok {
		return nil, errors.New("unsupported payment channel type")
	}
	result, err := channel.QueryOrder(gateway, orderNo, tradeNo)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return &payment.QueryOrderResponse{
		Code:        result.Code,
		Msg:         result.Msg,
		TradeNo:     result.TradeNo,
		OutTradeNo:  result.OutTradeNo,
		Type:        result.Type,
		Name:        result.Name,
		Money:       result.Money,
		TradeStatus: result.TradeStatus,
	}, nil
}
