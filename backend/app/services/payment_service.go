package services

import (
	"context"
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/db"
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
	OrderNo     string  `json:"order_no"`
	TradeNo     string  `json:"trade_no"`
	PayURL      string  `json:"pay_url"`
	Amount      float64 `json:"amount"`
	Fee         float64 `json:"fee"`
	PayAmount   float64 `json:"pay_amount"`
	ExpireAt    int64   `json:"expire_at"`
	GatewayName string  `json:"gateway_name"`
	PaymentType string  `json:"payment_type"`
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

	// 6. 计算手续费（费率配置异常，如包含模式下费率=100%，可能导致到账金额为 0，此时直接拒绝建单）
	fee, payAmount, creditAmount := CalculateFee(req.Amount, gateway.FeeRate, gateway.FeeMode)
	if creditAmount <= 0 {
		return nil, NewClientError("手续费配置异常，到账金额为 0，请联系管理员")
	}

	// 7. 获取订单过期时间
	expireMinutes := getOrderExpireMinutes()
	expireAt := time.Now().Add(time.Duration(expireMinutes) * time.Minute).Unix()

	// 8. 设置订单标题
	subject := req.Subject
	if subject == "" {
		subject = "余额充值"
	}

	// 9. 检查待支付订单数（防刷）+ 创建订单：同一事务内锁定用户行，防止并发请求绕过 10 单限制
	order := &models.PaymentOrder{
		OrderNo:        models.GenerateOrderNo(),
		UserID:         userID,
		GatewayID:      gateway.ID,
		PaymentChannel: gateway.Type,
		PaymentType:    gateway.PayType,
		Amount:         creditAmount,
		Fee:            fee,
		PayAmount:      payAmount,
		Subject:        subject,
		Status:         models.PaymentStatusPending,
		ExpireAt:       expireAt,
		ClientIP:       req.ClientIP,
	}
	if err := createPaymentOrderWithPendingLimitTx(userID, order); err != nil {
		return nil, err
	}

	// 10. 根据通道类型发起支付（由 pay_balance 插件注册的 PaymentChannel 分发）
	channel, ok := GetPaymentChannel(gateway.Type)
	if !ok {
		models.UpdatePaymentOrderStatus(order.OrderNo, models.PaymentStatusFailed, "")
		return nil, NewClientError("不支持的支付通道类型")
	}

	if !channel.ValidatePayType(gateway, order.PaymentType) {
		models.UpdatePaymentOrderStatus(order.OrderNo, models.PaymentStatusFailed, "")
		return nil, NewClientError("支付方式不受该通道支持")
	}

	// 使用回调地址：优先通道自定义，否则用全局
	gwNotifyURL := notifyURL
	if gateway.NotifyURL != "" {
		gwNotifyURL = gateway.NotifyURL
	}

	payURL, tradeNoFromRemote, createErr := channel.CreatePay(gateway, order, gwNotifyURL, returnURL)
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
		return nil, errors.New("生成支付链接失败，请检查支付配置")
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
		return nil, errors.New("支付订单保存失败，请稍后重试")
	}

	log.Printf("[Payment] 订单创建成功: order_no=%s, user_id=%d, amount=%.2f, fee=%.2f, pay_amount=%.2f, gateway=%s",
		order.OrderNo, userID, order.Amount, fee, payAmount, gateway.Name)

	tradeNo := models.NormalizeTradeNo(order.TradeNo)

	return &CreatePaymentOrderResponse{
		OrderNo:     order.OrderNo,
		TradeNo:     tradeNo,
		PayURL:      payURL,
		Amount:      order.Amount,
		Fee:         order.Fee,
		PayAmount:   order.PayAmount,
		ExpireAt:    order.ExpireAt,
		GatewayName: gateway.Name,
		PaymentType: gateway.PayType,
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
			return errors.New("检查待支付订单失败，请稍后重试")
		}
		if pendingCount >= maxPendingOrdersPerUser {
			return NewClientError("您有过多未支付订单，请先支付或等待过期后重试")
		}

		if err := models.CreatePaymentOrderTx(tx, order); err != nil {
			log.Printf("[Payment] 创建订单失败: %v", err)
			return errors.New("创建订单失败，请稍后重试")
		}
		return nil
	})
}

func settleThirdPartyPaidOrderTx(tx *gorm.DB, orderNo, tradeNo, paymentType, moneyStr, source string) (*models.PaymentOrder, *utils.BalanceResult, bool, error) {
	order, err := models.GetPaymentOrderForUpdate(tx, orderNo)
	if err != nil {
		log.Printf("[Payment] 订单不存在: source=%s, order_no=%s, err=%v", source, orderNo, err)
		return nil, nil, false, errors.New("订单不存在")
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
	if err := validateCallbackMoney(order.PayAmount, moneyStr); err != nil {
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
		return order, nil, false, fmt.Errorf("充值到账失败: %w", err)
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
		"签名验证失败", "商户号不匹配", "交易号不匹配", "支付通道不匹配", "支付方式不匹配",
		"回调支付类型不匹配", "回调金额与订单金额不一致", "回调金额不能为空", "回调金额格式非法",
		"回调金额非法", "订单金额非法", "订单不存在", "支付通道不存在", "不支持的支付通道类型",
		"回调参数不完整", "订单状态不允许处理回调",
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
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func classifyNotifyExceptionType(err error) string {
	if err == nil {
		return models.PaymentExceptionPermanentRejected
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "签名"):
		return models.PaymentExceptionSignFailed
	case strings.Contains(msg, "金额"):
		return models.PaymentExceptionAmountMismatch
	case strings.Contains(msg, "不匹配") || strings.Contains(msg, "绑定"):
		return models.PaymentExceptionBindingMismatch
	case strings.Contains(msg, "订单不存在"):
		return models.PaymentExceptionOrderMissing
	default:
		return models.PaymentExceptionPermanentRejected
	}
}

// HandlePaymentNotify 处理易支付异步回调（多通道版本）
// 返回: 是否处理成功, 错误信息。永久错误也会返回 err，由控制器决定回 SUCCESS 停重试。
func HandlePaymentNotify(params map[string]string) (bool, error) {
	// 1. 提取回调参数
	outTradeNo := params["out_trade_no"]
	tradeNo := models.NormalizeTradeNo(params["trade_no"])
	tradeStatus := params["trade_status"]
	moneyStr := params["money"]
	pid := strings.TrimSpace(params["pid"])
	callbackType := strings.TrimSpace(params["type"])

	if outTradeNo == "" || tradeNo == "" {
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

	// 4. 验证签名（防篡改）
	notifyChannel, ok := GetPaymentChannel(gateway.Type)
	if !ok {
		return false, newPaymentNotifyError(true, "不支持的支付通道类型")
	}
	if !notifyChannel.VerifyNotify(params, gateway.Key) {
		log.Printf("[Payment] 回调签名验证失败: order_no=%s, trade_status=%s", outTradeNo, tradeStatus)
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
	if err := validateCallbackMoney(orderForGateway.PayAmount, moneyStr); err != nil {
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
		return nil, errors.New("缺少订单号参数")
	}

	order, err := models.GetPaymentOrderByOrderNo(outTradeNo)
	if err != nil {
		return nil, errors.New("订单不存在")
	}

	// 获取通道密钥验签
	gateway, err := models.GetPayGatewayByID(order.GatewayID)
	if err != nil {
		return nil, errors.New("支付通道不存在")
	}

	returnChannel, ok := GetPaymentChannel(gateway.Type)
	if !ok {
		return nil, errors.New("不支持的支付通道类型")
	}
	if !returnChannel.VerifyNotify(params, gateway.Key) {
		return nil, errors.New("签名验证失败")
	}

	return order, nil
}

func ReconcilePaymentOrderByID(orderID uint64) (*models.PaymentOrder, bool, error) {
	order, err := models.GetPaymentOrderByID(orderID)
	if err != nil {
		return nil, false, err
	}

	reconciled, err := reconcilePaymentOrder(order)
	if err != nil {
		return order, false, err
	}
	if !reconciled {
		return order, false, nil
	}

	refreshedOrder, err := models.GetPaymentOrderByID(orderID)
	if err != nil {
		return nil, true, err
	}
	return refreshedOrder, true, nil
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
func ReconcilePaymentOrdersBatch(ctx context.Context, limit int) (*PaymentReconcileBatchResult, error) {
	result := &PaymentReconcileBatchResult{}
	orders, err := models.ListPaymentOrdersForReconcile(limit, 7*24*3600)
	if err != nil {
		return nil, err
	}
	result.Scanned = len(orders)
	for _, order := range orders {
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return result, err
			}
		}
		o := order
		changed, err := reconcilePaymentOrder(&o)
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

// reconcilePaymentOrder 对单笔订单向网关查单；支持 pending/canceled/failed 恢复
func reconcilePaymentOrder(order *models.PaymentOrder) (bool, error) {
	if order == nil {
		return false, errors.New("订单不存在")
	}
	switch order.Status {
	case models.PaymentStatusPending, models.PaymentStatusCanceled, models.PaymentStatusFailed:
		// ok
	default:
		return false, nil
	}
	if order.GatewayID == 0 {
		return false, nil
	}
	channel, ok := GetPaymentChannel(order.PaymentChannel)
	if !ok {
		return false, nil
	}

	orderNo := strings.TrimSpace(order.OrderNo)
	tradeNo := models.NormalizeTradeNo(order.TradeNo)
	if orderNo == "" && tradeNo == "" {
		return false, nil
	}

	gateway, err := models.GetPayGatewayByID(order.GatewayID)
	if err != nil {
		return false, fmt.Errorf("获取支付通道失败: %w", err)
	}

	queryResult, err := channel.QueryOrder(gateway, orderNo, tradeNo)
	if err != nil {
		return false, err
	}
	if queryResult == nil {
		return false, errors.New("查询结果为空")
	}
	if queryResult.Code != 1 {
		log.Printf("[Payment] 主动查单未返回成功结果: order_no=%s, trade_no=%s, code=%d, msg=%s, status=%s", orderNo, tradeNo, queryResult.Code, strings.TrimSpace(queryResult.Msg), strings.TrimSpace(queryResult.TradeStatus))
		return false, nil
	}
	if queryResult.OutTradeNo != "" && queryResult.OutTradeNo != orderNo {
		return false, newPaymentNotifyError(true, "云端订单号不匹配")
	}

	queryTradeNo := models.NormalizeTradeNo(queryResult.TradeNo)
	if queryTradeNo == "" {
		queryTradeNo = tradeNo
	}
	if err := validatePaymentNotifyBinding(order, gateway, strings.TrimSpace(gateway.PID), strings.TrimSpace(queryResult.Type), queryTradeNo); err != nil {
		return false, newPaymentNotifyError(true, err.Error())
	}
	if err := validateCallbackMoney(order.PayAmount, queryResult.Money); err != nil {
		return false, newPaymentNotifyError(true, err.Error())
	}
	if queryResult.TradeStatus != "TRADE_SUCCESS" {
		log.Printf("[Payment] 主动查单未确认支付成功: order_no=%s, trade_no=%s, code=%d, msg=%s, status=%s", orderNo, tradeNo, queryResult.Code, strings.TrimSpace(queryResult.Msg), strings.TrimSpace(queryResult.TradeStatus))
		return false, nil
	}

	var changed bool
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		var lockedOrder *models.PaymentOrder
		var balanceResult *utils.BalanceResult
		var innerErr error
		lockedOrder, balanceResult, changed, innerErr = settleThirdPartyPaidOrderTx(tx, order.OrderNo, queryTradeNo, strings.TrimSpace(queryResult.Type), queryResult.Money, "reconcile")
		if innerErr != nil {
			return innerErr
		}
		if changed {
			log.Printf("[Payment] 主动查单补账成功: order_no=%s, user_id=%d, amount=%.2f, fee=%.2f, pay_amount=%.2f, before=%.2f, after=%.2f",
				lockedOrder.OrderNo, lockedOrder.UserID, lockedOrder.Amount, lockedOrder.Fee, lockedOrder.PayAmount, balanceResult.BeforeMoney, balanceResult.AfterMoney)
		}
		return nil
	})
	if err != nil {
		return false, err
	}

	return changed, nil
}

// AdminCompleteOrder 管理员手动补单。
// force=true 时允许对 canceled/failed 强制补单（高危）；默认仅 pending，且优先应走网关对账。
// adminUserID：操作人；skipDualCheck=true 表示审批通过后执行，不再二次进审。
// 当 finance_dual_approval=true 且 force 且未 skip 时：只创建 pending 审批，不立即入账。
// 当 finance_dual_approval=false 且 force 时：立即入账，并写一条 self-approved 审计记录（便于日后开双人复核）。
func AdminCompleteOrder(orderID uint64, memo string, force bool, adminUserID uint64, skipDualCheck bool) error {
	order, err := models.GetPaymentOrderByID(orderID)
	if err != nil {
		return NewClientError("订单不存在")
	}

	if order.Status == models.PaymentStatusPaid {
		return NewClientError("订单已支付，无需重复操作")
	}
	if order.Status == models.PaymentStatusPending {
		// 默认：先尝试网关查单补账
		if changed, rerr := reconcilePaymentOrder(order); rerr == nil && changed {
			return nil
		}
	} else if order.Status == models.PaymentStatusCanceled || order.Status == models.PaymentStatusFailed {
		if !force {
			// 非强制：仅允许对账路径
			if changed, rerr := reconcilePaymentOrder(order); rerr == nil && changed {
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

	// 双人复核：强制补单先挂起审批
	dualOn := false
	if GlobalSettingsService != nil {
		dualOn = GlobalSettingsService.GetBoolWithDefault("finance_dual_approval", false)
	}
	if force && dualOn && !skipDualCheck {
		payload, _ := models.MarshalApprovalPayload(map[string]interface{}{
			"order_id": orderID,
			"order_no": order.OrderNo,
			"memo":     memo,
			"force":    true,
		})
		req := &models.ApprovalRequest{
			Type:        models.ApprovalTypeForcePaymentComplete,
			PayloadJSON: payload,
			Status:      models.ApprovalStatusPending,
			RequesterID: adminUserID,
			Comment:     memo,
		}
		if err := models.CreateApprovalRequest(req); err != nil {
			return fmt.Errorf("创建审批失败: %w", err)
		}
		return NewClientError(fmt.Sprintf("强制补单已提交审批(#%d)，等待另一管理员复核", req.ID))
	}

	if err := adminCompleteOrderExec(order, memo, force); err != nil {
		return err
	}

	// 强制补单审计：未开双人复核时记一条 self-approved；审批通过路径不再重复写。
	if force && adminUserID > 0 && !skipDualCheck {
		payload, _ := models.MarshalApprovalPayload(map[string]interface{}{
			"order_id": orderID,
			"order_no": order.OrderNo,
			"memo":     memo,
			"force":    true,
		})
		now := time.Now().Unix()
		rid := adminUserID
		_ = models.CreateApprovalRequest(&models.ApprovalRequest{
			Type:        models.ApprovalTypeForcePaymentComplete,
			PayloadJSON: payload,
			Status:      models.ApprovalStatusApproved,
			RequesterID: adminUserID,
			ReviewerID:  &rid,
			Comment:     memo,
			ReviewTime:  &now,
		})
	}
	return nil
}

// adminCompleteOrderExec 实际入账逻辑（原 AdminCompleteOrder 事务体）
func adminCompleteOrderExec(order *models.PaymentOrder, memo string, force bool) error {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		lockedOrder, err := models.GetPaymentOrderForUpdate(tx, order.OrderNo)
		if err != nil {
			return fmt.Errorf("锁定订单失败: %w", err)
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
			return fmt.Errorf("补单失败: %w", err)
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
		return errors.New("商户号不匹配")
	}
	// 网关已配置商户号时，回调必须带上且一致（避免空 pid 绕过）
	if gateway != nil && strings.TrimSpace(gateway.PID) != "" && pid == "" {
		return errors.New("商户号不匹配")
	}

	if order != nil {
		if order.TradeNo != "" && tradeNo != "" && order.TradeNo != tradeNo {
			return errors.New("交易号不匹配")
		}
		if gateway != nil {
			if order.GatewayID != 0 && gateway.ID != 0 && order.GatewayID != gateway.ID {
				return errors.New("支付通道不匹配")
			}
			if order.PaymentChannel != "" && gateway.Type != "" &&
				!strings.EqualFold(order.PaymentChannel, gateway.Type) {
				return errors.New("支付通道类型不匹配")
			}
			if order.PaymentType != "" && gateway.PayType != "" &&
				!strings.EqualFold(order.PaymentType, gateway.PayType) {
				return errors.New("支付方式不匹配")
			}
		}
		// 标准支付类型（alipay/wxpay 等）必须与订单一致；数字型自定义 type 放行
		if order.PaymentType != "" && callbackType != "" && !isNonStandardEpayCallbackType(callbackType) {
			if !strings.EqualFold(callbackType, order.PaymentType) {
				return errors.New("回调支付类型不匹配")
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
		return errors.New("回调金额不能为空")
	}
	callbackMoney, err := strconv.ParseFloat(moneyStr, 64)
	if err != nil {
		return errors.New("回调金额格式非法")
	}
	expectedFen, err := utils.YuanToFen(expected)
	if err != nil {
		return errors.New("订单金额非法")
	}
	callbackFen, err := utils.YuanToFen(callbackMoney)
	if err != nil {
		return errors.New("回调金额非法")
	}
	if expectedFen != callbackFen {
		return errors.New("回调金额与订单金额不一致")
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
