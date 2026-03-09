package services

import (
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/internal/db"
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
	settingsMap, _ := models.GetSettingsMap([]string{"payment_enabled"})
	paymentEnabled := settingsMap["payment_enabled"] == "true" || settingsMap["payment_enabled"] == "1"
	if !paymentEnabled {
		return nil, errors.New("支付功能未启用")
	}
	if req.Amount <= 0 {
		return nil, errors.New("充值金额必须大于 0")
	}

	// 2. 获取支付通道
	gateway, err := models.GetPayGatewayByID(req.GatewayID)
	if err != nil {
		return nil, errors.New("支付通道不存在")
	}
	if gateway.Status != models.PayGatewayStatusEnabled {
		return nil, errors.New("该支付通道已禁用")
	}

	// 3. 检查用户是否存在 + 等级校验
	user, err := models.GetUserByID(userID)
	if err != nil || user == nil {
		return nil, errors.New("用户不存在")
	}
	if gateway.MinLevel > 0 && int(user.Level) < gateway.MinLevel {
		return nil, fmt.Errorf("该通道要求最低等级 Lv.%d，您当前等级 Lv.%d", gateway.MinLevel, user.Level)
	}

	// 4. 验证金额范围（通道级别）
	if gateway.MinAmount > 0 && req.Amount < gateway.MinAmount {
		return nil, fmt.Errorf("该通道最低充值金额为 ¥%.2f", gateway.MinAmount)
	}
	if gateway.MaxAmount > 0 && req.Amount > gateway.MaxAmount {
		return nil, fmt.Errorf("该通道最高充值金额为 ¥%.2f", gateway.MaxAmount)
	}

	// 5. 检查用户是否有过多未支付订单（防刷）
	pendingOrders, _, err := models.GetPaymentOrderList(userID, 1, 100, models.PaymentStatusPending, "")
	if err == nil && len(pendingOrders) >= 10 {
		return nil, errors.New("您有过多未支付订单，请先支付或等待过期后重试")
	}

	// 6. 计算手续费
	fee, payAmount, _ := CalculateFee(req.Amount, gateway.FeeRate, gateway.FeeMode)

	// 7. 获取订单过期时间
	expireMinutes := getOrderExpireMinutes()
	expireAt := time.Now().Add(time.Duration(expireMinutes) * time.Minute).Unix()

	// 8. 设置订单标题
	subject := req.Subject
	if subject == "" {
		subject = "余额充值"
	}

	// 9. 创建订单
	order := &models.PaymentOrder{
		OrderNo:        models.GenerateOrderNo(),
		UserID:         userID,
		GatewayID:      gateway.ID,
		PaymentChannel: gateway.Type,
		PaymentType:    gateway.PayType,
		Amount:         req.Amount,
		Fee:            fee,
		PayAmount:      payAmount,
		Subject:        subject,
		Status:         models.PaymentStatusPending,
		ExpireAt:       expireAt,
		ClientIP:       req.ClientIP,
	}

	if err := models.CreatePaymentOrder(order); err != nil {
		log.Printf("[Payment] 创建订单失败: %v", err)
		return nil, errors.New("创建订单失败，请稍后重试")
	}

	// 10. 根据通道类型发起支付
	var payURL string
	if gateway.Type == "epay" {
		// 使用通道自身的配置构建易支付请求
		epayConfig := &EpayConfig{
			Enabled:      true,
			ApiURL:       strings.TrimRight(gateway.ApiURL, "/"),
			PID:          gateway.PID,
			Key:          gateway.Key,
			PaymentTypes: []string{gateway.PayType},
		}

		// 使用回调地址：优先通道自定义，否则用全局
		gwNotifyURL := notifyURL
		if gateway.NotifyURL != "" {
			gwNotifyURL = gateway.NotifyURL
		}

		// 先尝试 API 支付（mapi）
		apiPayURL, apiErr := EpayAPIPay(epayConfig, order, gwNotifyURL, returnURL)
		if apiErr != nil {
			log.Printf("[Payment] API支付失败，回退到跳转支付: %v", apiErr)
			// 回退到跳转模式
			payURL, err = BuildEpaySubmitURL(epayConfig, order, gwNotifyURL, returnURL)
			if err != nil {
				log.Printf("[Payment] 构造支付链接失败: %v", err)
				models.UpdatePaymentOrderStatus(order.OrderNo, models.PaymentStatusFailed, "")
				return nil, errors.New("生成支付链接失败，请检查支付配置")
			}
		} else {
			payURL = apiPayURL
		}
	} else {
		models.UpdatePaymentOrderStatus(order.OrderNo, models.PaymentStatusFailed, "")
		return nil, fmt.Errorf("不支持的支付通道类型: %s", gateway.Type)
	}

	// 11. 保存支付链接到订单
	order.PayURL = payURL
	db.DB.Exec("UPDATE payment_orders SET pay_url = ? WHERE id = ?", payURL, order.ID)

	log.Printf("[Payment] 订单创建成功: order_no=%s, user_id=%d, amount=%.2f, fee=%.2f, pay_amount=%.2f, gateway=%s",
		order.OrderNo, userID, req.Amount, fee, payAmount, gateway.Name)

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
	if !VerifyEpaySign(params, gateway.Key) {
		log.Printf("[Payment] 回调签名验证失败: params=%v", params)
		return false, errors.New("签名验证失败")
	}
	if err := validatePaymentNotifyBinding(nil, gateway, pid, "", tradeNo); err != nil {
		log.Printf("[Payment] 回调绑定校验失败: order_no=%s, err=%v", outTradeNo, err)
		return false, err
	}

	// 5. 只处理 TRADE_SUCCESS 状态
	if tradeStatus != "TRADE_SUCCESS" {
		log.Printf("[Payment] 非成功状态回调: order_no=%s, status=%s", outTradeNo, tradeStatus)
		models.IncrementNotifyCount(outTradeNo)
		return true, nil
	}

	// 6. 在事务中处理到账（保证原子性+幂等性）
	tx, err := db.DB.Begin()
	if err != nil {
		return false, fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	// 7. 锁定订单行
	order, err := models.GetPaymentOrderForUpdate(tx, outTradeNo)
	if err != nil {
		log.Printf("[Payment] 订单不存在: order_no=%s, err=%v", outTradeNo, err)
		return false, errors.New("订单不存在")
	}

	// 8. 幂等检查
	if order.Status != models.PaymentStatusPending {
		log.Printf("[Payment] 订单已处理（幂等跳过）: order_no=%s, status=%d", outTradeNo, order.Status)
		return true, nil
	}
	if err := validatePaymentNotifyBinding(order, nil, "", callbackType, tradeNo); err != nil {
		log.Printf("[Payment] 回调绑定校验失败: order_no=%s, err=%v", outTradeNo, err)
		return false, err
	}

	// 9. 金额校验：回调金额应匹配实际支付金额（pay_amount）
	if err := validateCallbackMoney(order.PayAmount, moneyStr); err != nil {
		log.Printf("[Payment] 回调金额校验失败: order_no=%s, err=%v", outTradeNo, err)
		return false, err
	}

	// 10. 通过统一余额工具完成：修改余额 + 更新订单状态 + 添加余额变动记录
	balanceResult, err := utils.ExecuteBalanceOpTx(tx, &utils.BalanceReq{
		UserID: order.UserID,
		Amount: order.Amount,
		MemoI18n: map[string]string{
			"zhCN": fmt.Sprintf("在线充值-订单号%s", outTradeNo),
			"enUS": fmt.Sprintf("Online Recharge - Order#%s", outTradeNo),
		},
		OrderNo:     outTradeNo,
		TradeNo:     tradeNo,
		OrderStatus: models.PaymentStatusPaid,
	}, utils.OpFull)
	if err != nil {
		return false, fmt.Errorf("充值到账失败: %w", err)
	}

	// 11. 提交事务
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("提交事务失败: %w", err)
	}

	log.Printf("[Payment] 充值到账成功: order_no=%s, user_id=%d, amount=%.2f, fee=%.2f, pay_amount=%.2f, before=%.2f, after=%.2f",
		outTradeNo, order.UserID, order.Amount, order.Fee, order.PayAmount, balanceResult.BeforeMoney, balanceResult.AfterMoney)

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

	if !VerifyEpaySign(params, gateway.Key) {
		return nil, errors.New("签名验证失败")
	}

	return order, nil
}

// AdminCompleteOrder 管理员手动补单
func AdminCompleteOrder(orderID uint64, memo string) error {
	order, err := models.GetPaymentOrderByID(orderID)
	if err != nil {
		return errors.New("订单不存在")
	}

	if order.Status == models.PaymentStatusPaid {
		return errors.New("订单已支付，无需重复操作")
	}
	if order.Status != models.PaymentStatusPending {
		return errors.New("只能对待支付的订单进行补单操作")
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
		return errors.New("锁定订单失败")
	}
	if lockedOrder.Status != models.PaymentStatusPending {
		return errors.New("订单状态已变更")
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
		return errors.New("订单不存在")
	}

	if order.Status != models.PaymentStatusPending {
		return errors.New("只能取消待支付的订单")
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
func CancelExpiredOrders() {
	affected, err := models.CancelExpiredOrders()
	if err != nil {
		log.Printf("[Payment] 取消过期订单失败: %v", err)
		return
	}
	if affected > 0 {
		log.Printf("[Payment] 已取消 %d 个过期订单", affected)
	}
}

func AdminDeleteOrder(orderID uint64) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer tx.Rollback()

	order, err := models.GetPaymentOrderByIDForUpdate(tx, orderID)
	if err != nil {
		return errors.New("订单不存在")
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

func validatePaymentNotifyBinding(order *models.PaymentOrder, gateway *models.PayGateway, pid, callbackType, tradeNo string) error {
	if gateway != nil && gateway.PID != "" && pid != gateway.PID {
		return errors.New("商户号不匹配")
	}
	if order != nil {
		if order.TradeNo != "" && order.TradeNo != tradeNo {
			return errors.New("交易号不匹配")
		}
		if order.PaymentType != "" && callbackType != "" && callbackType != order.PaymentType {
			return errors.New("支付类型不匹配")
		}
	}
	return nil
}

func validateCallbackMoney(expected float64, moneyStr string) error {
	if moneyStr == "" {
		return nil
	}
	callbackMoney, err := strconv.ParseFloat(moneyStr, 64)
	if err != nil {
		return errors.New("回调金额格式非法")
	}
	if abs(callbackMoney-expected) > 0.01 {
		return errors.New("回调金额与订单金额不一致")
	}
	return nil
}

func validatePaymentOrderDeletion(status int) error {
	if status != models.PaymentStatusCanceled && status != models.PaymentStatusFailed {
		return errors.New("仅允许删除已取消或失败的订单")
	}
	return nil
}

// getOrderExpireMinutes 从系统配置获取订单过期时间（分钟）
func getOrderExpireMinutes() int {
	settingsMap, err := models.GetSettingsMap([]string{"payment_order_expire_minutes"})
	if err != nil {
		return 30
	}
	if v, err := strconv.Atoi(settingsMap["payment_order_expire_minutes"]); err == nil && v > 0 {
		return v
	}
	return 30
}

// abs 浮点数绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
