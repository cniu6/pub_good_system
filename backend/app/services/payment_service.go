package services

import (
	"database/sql"
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/db"
	"fst/backend/utils"
	"log"
	"strconv"
	"strings"
	"time"
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
		return nil, errors.New("生成支付链接失败，请检查支付配置")
	}
	order.TradeNo = models.NormalizeTradeNo(tradeNoFromRemote)

	// 11. 保存支付链接到订单；若落库失败，订单会卡在 pending 但无 pay_url 且无法支付，
	// 因此这里失败时把订单标记为 failed（尽力而为，不影响主错误返回），避免留下无法处理的脏单，
	// 占用用户「待支付订单数」配额；用户可重新发起充值创建新订单。
	order.PayURL = payURL
	if err := models.UpdatePaymentOrderPaymentInfo(order.OrderNo, order.TradeNo, payURL); err != nil {
		log.Printf("[Payment] 保存支付链接失败: order_no=%s, err=%v", order.OrderNo, err)
		if failErr := models.UpdatePaymentOrderStatus(order.OrderNo, models.PaymentStatusFailed, ""); failErr != nil {
			log.Printf("[Payment] 标记脏单失败状态也失败: order_no=%s, err=%v", order.OrderNo, failErr)
		}
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
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 锁定用户行，串行化同一用户的并发建单请求
	var lockedID uint64
	if err := tx.QueryRow(db.Q("SELECT id FROM users WHERE id = ? AND delete_time IS NULL FOR UPDATE"), userID).Scan(&lockedID); err != nil {
		if err == sql.ErrNoRows {
			return NewClientError("用户不存在")
		}
		return err
	}

	// 允许同金额多笔待支付并存（网络重试等）；是否提示用户去付旧单由前端二次确认。
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}
	return nil
}

func settleThirdPartyPaidOrderTx(tx *sql.Tx, orderNo, tradeNo, paymentType, moneyStr, source string) (*models.PaymentOrder, *utils.BalanceResult, bool, error) {
	order, err := models.GetPaymentOrderForUpdate(tx, orderNo)
	if err != nil {
		log.Printf("[Payment] 订单不存在: source=%s, order_no=%s, err=%v", source, orderNo, err)
		return nil, nil, false, errors.New("订单不存在")
	}

	if order.Status == models.PaymentStatusPaid {
		log.Printf("[Payment] 订单已支付（%s 幂等跳过）: order_no=%s", source, orderNo)
		return order, nil, false, nil
	}
	if order.Status != models.PaymentStatusPending {
		log.Printf("[Payment] 订单状态不允许处理: source=%s, order_no=%s, status=%d", source, orderNo, order.Status)
		return order, nil, false, errors.New("订单状态不允许处理回调")
	}
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

	return order, balanceResult, true, nil
}

// HandlePaymentNotify 处理易支付异步回调（多通道版本）
// 返回: 是否处理成功, 错误信息
func HandlePaymentNotify(params map[string]string) (bool, error) {
	// 1. 提取回调参数
	outTradeNo := params["out_trade_no"]
	tradeNo := models.NormalizeTradeNo(params["trade_no"])
	tradeStatus := params["trade_status"]
	moneyStr := params["money"]
	pid := strings.TrimSpace(params["pid"])
	callbackType := strings.TrimSpace(params["type"])

	if outTradeNo == "" || tradeNo == "" {
		return false, errors.New("回调参数不完整")
	}

	// 2. 查找订单获取对应通道
	orderForGateway, err := models.GetPaymentOrderByOrderNo(outTradeNo)
	if err != nil {
		return false, errors.New("订单不存在")
	}

	// 3. 获取通道配置以验签
	gateway, err := models.GetPayGatewayByID(orderForGateway.GatewayID)
	if err != nil {
		return false, errors.New("支付通道不存在")
	}

	// 4. 验证签名（防篡改）
	notifyChannel, ok := GetPaymentChannel(gateway.Type)
	if !ok {
		return false, errors.New("不支持的支付通道类型")
	}
	if !notifyChannel.VerifyNotify(params, gateway.Key) {
		log.Printf("[Payment] 回调签名验证失败: order_no=%s, trade_status=%s", outTradeNo, tradeStatus)
		return false, errors.New("签名验证失败")
	}
	// 5. 订单与通道/商户/支付方式绑定校验（防串单）
	if err := validatePaymentNotifyBinding(orderForGateway, gateway, pid, callbackType, tradeNo); err != nil {
		log.Printf("[Payment] 回调绑定校验失败: order_no=%s, gateway_id=%d, err=%v", outTradeNo, orderForGateway.GatewayID, err)
		return false, err
	}

	// 5. 只处理 TRADE_SUCCESS 状态
	if tradeStatus != "TRADE_SUCCESS" {
		log.Printf("[Payment] 非成功状态回调: order_no=%s, status=%s", outTradeNo, tradeStatus)
		if countErr := models.IncrementNotifyCount(outTradeNo); countErr != nil {
			log.Printf("[Payment] 更新回调通知次数失败: order_no=%s, err=%v", outTradeNo, countErr)
		}
		return true, nil
	}

	// 6. 在事务中处理到账（保证原子性+幂等性）
	tx, err := db.DB.Begin()
	if err != nil {
		return false, fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	order, balanceResult, changed, err := settleThirdPartyPaidOrderTx(tx, outTradeNo, tradeNo, callbackType, moneyStr, "支付回调")
	if err != nil {
		if countErr := models.IncrementNotifyCount(outTradeNo); countErr != nil {
			log.Printf("[Payment] 更新回调通知次数失败: order_no=%s, err=%v", outTradeNo, countErr)
		}
		return false, err
	}

	// 7. 提交事务
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("提交事务失败: %w", err)
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

	reconciled, err := reconcilePendingPaymentOrder(order)
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

func reconcilePendingPaymentOrder(order *models.PaymentOrder) (bool, error) {
	if order == nil {
		return false, errors.New("订单不存在")
	}
	if order.Status != models.PaymentStatusPending {
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
		return false, errors.New("云端订单号不匹配")
	}

	queryTradeNo := models.NormalizeTradeNo(queryResult.TradeNo)
	if queryTradeNo == "" {
		queryTradeNo = tradeNo
	}
	if err := validatePaymentNotifyBinding(order, gateway, strings.TrimSpace(gateway.PID), strings.TrimSpace(queryResult.Type), queryTradeNo); err != nil {
		return false, err
	}
	if err := validateCallbackMoney(order.PayAmount, queryResult.Money); err != nil {
		return false, err
	}
	if queryResult.TradeStatus != "TRADE_SUCCESS" {
		log.Printf("[Payment] 主动查单未确认支付成功: order_no=%s, trade_no=%s, code=%d, msg=%s, status=%s", orderNo, tradeNo, queryResult.Code, strings.TrimSpace(queryResult.Msg), strings.TrimSpace(queryResult.TradeStatus))
		return false, nil
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return false, fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	lockedOrder, balanceResult, changed, err := settleThirdPartyPaidOrderTx(tx, order.OrderNo, queryTradeNo, strings.TrimSpace(queryResult.Type), queryResult.Money, "主动查单")
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("提交事务失败: %w", err)
	}

	if changed {
		log.Printf("[Payment] 主动查单补账成功: order_no=%s, user_id=%d, amount=%.2f, fee=%.2f, pay_amount=%.2f, before=%.2f, after=%.2f",
			lockedOrder.OrderNo, lockedOrder.UserID, lockedOrder.Amount, lockedOrder.Fee, lockedOrder.PayAmount, balanceResult.BeforeMoney, balanceResult.AfterMoney)
	}

	return changed, nil
}

// AdminCompleteOrder 管理员手动补单
func AdminCompleteOrder(orderID uint64, memo string) error {
	order, err := models.GetPaymentOrderByID(orderID)
	if err != nil {
		return NewClientError("订单不存在")
	}

	if order.Status == models.PaymentStatusPaid {
		return NewClientError("订单已支付，无需重复操作")
	}
	if order.Status != models.PaymentStatusPending {
		return NewClientError("只能对待支付的订单进行补单操作")
	}

	// 事务处理
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 锁定订单
	lockedOrder, err := models.GetPaymentOrderForUpdate(tx, order.OrderNo)
	if err != nil {
		return fmt.Errorf("锁定订单失败: %w", err)
	}
	if lockedOrder.Status != models.PaymentStatusPending {
		return NewClientError("订单状态已变更")
	}

	// 通过统一余额工具完成：修改余额 + 更新订单状态 + 添加余额变动记录
	memoZh := fmt.Sprintf("管理员手动补单-订单号%s", order.OrderNo)
	memoEn := fmt.Sprintf("Admin Manual - Order#%s", order.OrderNo)
	if memo != "" {
		memoZh += " (" + memo + ")"
		memoEn += " (" + memo + ")"
	}
	_, err = utils.ExecuteBalanceOpTx(tx, &utils.BalanceReq{
		UserID: order.UserID,
		Amount: order.Amount,
		MemoI18n: map[string]string{
			"zhCN": memoZh,
			"enUS": memoEn,
		},
		OrderNo:     order.OrderNo,
		TradeNo:     "MANUAL",
		OrderStatus: models.PaymentStatusPaid,
	}, utils.OpFull)
	if err != nil {
		return fmt.Errorf("补单失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	log.Printf("[Payment] 管理员手动补单成功: order_no=%s, user_id=%d, amount=%.2f",
		order.OrderNo, order.UserID, order.Amount)
	return nil
}

// AdminCancelOrder 管理员取消订单
func AdminCancelOrder(orderID uint64) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	order, err := models.GetPaymentOrderByIDForUpdate(tx, orderID)
	if err != nil {
		return NewClientError("订单不存在")
	}

	if order.Status != models.PaymentStatusPending {
		return NewClientError("只能取消待支付的订单")
	}

	if err := models.UpdatePaymentOrderStatusTx(tx, order.OrderNo, models.PaymentStatusCanceled, ""); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
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
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	order, err := models.GetPaymentOrderByIDForUpdate(tx, orderID)
	if err != nil {
		return NewClientError("订单不存在")
	}
	if err := validatePaymentOrderDeletion(order.Status); err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM payment_orders WHERE id = ?", orderID); err != nil {
		return fmt.Errorf("删除订单失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
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
