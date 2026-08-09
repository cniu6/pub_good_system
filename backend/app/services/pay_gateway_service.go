package services

import (
	"context"
	"errors"
	"fst/backend/app/models"
	"fst/backend/pkg/payment"
	"fst/backend/utils"
	"log"
	"math"
)

// PayGatewayCreateRequest 创建支付通道请求
type PayGatewayCreateRequest struct {
	Name                 string  `json:"name" binding:"required,max=100"`
	Type                 string  `json:"type" binding:"required,max=50"`
	PayType              string  `json:"pay_type" binding:"required,max=50"`
	SignType             string  `json:"sign_type" binding:"omitempty,max=50"`
	Version              string  `json:"version" binding:"omitempty,max=50"`
	Device               string  `json:"device" binding:"omitempty,max=50"`
	Currency             string  `json:"currency" binding:"omitempty,max=10"`
	TargetCurrency       string  `json:"target_currency" binding:"omitempty,max=10"`
	ExchangeRateMode     string  `json:"exchange_rate_mode" binding:"omitempty,max=20"`
	ExchangeRate         float64 `json:"exchange_rate"`
	ExchangeFixedAmount  float64 `json:"exchange_fixed_amount"`
	ExchangeRateSource   string  `json:"exchange_rate_source" binding:"omitempty,max=255"`
	TargetFeeRate        int     `json:"target_fee_rate"`
	TargetFeeFixed       float64 `json:"target_fee_fixed"`
	TargetFeeMode        string  `json:"target_fee_mode" binding:"omitempty,max=20"`
	Description          string  `json:"description" binding:"omitempty,max=500"`
	Status               int     `json:"status"`
	ApiURL               string  `json:"api_url" binding:"omitempty"`
	PID                  string  `json:"pid" binding:"omitempty"`
	Key                  string  `json:"key" binding:"omitempty"`
	ExtConfig            string  `json:"ext_config" binding:"omitempty"`
	LogoURL              string  `json:"logo_url" binding:"omitempty"`
	SortOrder            int     `json:"sort_order"`
	MinAmount            float64 `json:"min_amount"`
	MaxAmount            float64 `json:"max_amount"`
	FeeRate              int     `json:"fee_rate"`
	FeeMode              string  `json:"fee_mode" binding:"omitempty,max=50"`
	MinLevel             int     `json:"min_level"`
	NotifyURL            string  `json:"notify_url" binding:"omitempty"`
	ExpireMinutes        int     `json:"expire_minutes"`
	ActiveQueryEnabled   int     `json:"active_query_enabled"`
	QueryIntervalSeconds int     `json:"query_interval_seconds"`
	QueryBatchSize       int     `json:"query_batch_size"`
}

// PayGatewayUpdateRequest 更新支付通道请求
type PayGatewayUpdateRequest struct {
	Name                 *string  `json:"name" binding:"omitempty,max=100"`
	Type                 *string  `json:"type" binding:"omitempty,max=50"`
	PayType              *string  `json:"pay_type" binding:"omitempty,max=50"`
	SignType             *string  `json:"sign_type" binding:"omitempty,max=50"`
	Version              *string  `json:"version" binding:"omitempty,max=50"`
	Device               *string  `json:"device" binding:"omitempty,max=50"`
	Currency             *string  `json:"currency" binding:"omitempty,max=10"`
	TargetCurrency       *string  `json:"target_currency" binding:"omitempty,max=10"`
	ExchangeRateMode     *string  `json:"exchange_rate_mode" binding:"omitempty,max=20"`
	ExchangeRate         *float64 `json:"exchange_rate"`
	ExchangeFixedAmount  *float64 `json:"exchange_fixed_amount"`
	ExchangeRateSource   *string  `json:"exchange_rate_source" binding:"omitempty,max=255"`
	TargetFeeRate        *int     `json:"target_fee_rate"`
	TargetFeeFixed       *float64 `json:"target_fee_fixed"`
	TargetFeeMode        *string  `json:"target_fee_mode" binding:"omitempty,max=20"`
	Description          *string  `json:"description" binding:"omitempty,max=500"`
	Status               *int     `json:"status"`
	ApiURL               *string  `json:"api_url" binding:"omitempty"`
	PID                  *string  `json:"pid" binding:"omitempty"`
	Key                  *string  `json:"key" binding:"omitempty"`
	ExtConfig            *string  `json:"ext_config" binding:"omitempty"`
	LogoURL              *string  `json:"logo_url" binding:"omitempty"`
	SortOrder            *int     `json:"sort_order"`
	MinAmount            *float64 `json:"min_amount"`
	MaxAmount            *float64 `json:"max_amount"`
	FeeRate              *int     `json:"fee_rate"`
	FeeMode              *string  `json:"fee_mode" binding:"omitempty,max=50"`
	MinLevel             *int     `json:"min_level"`
	NotifyURL            *string  `json:"notify_url" binding:"omitempty"`
	ExpireMinutes        *int     `json:"expire_minutes"`
	ActiveQueryEnabled   *int     `json:"active_query_enabled"`
	QueryIntervalSeconds *int     `json:"query_interval_seconds"`
	QueryBatchSize       *int     `json:"query_batch_size"`
}

// CreatePayGateway 创建支付通道
func CreatePayGateway(req *PayGatewayCreateRequest) (*models.PayGateway, error) {
	if req.Name == "" {
		return nil, errors.New("Gateway name cannot be empty")
	}
	if req.PayType == "" {
		return nil, errors.New("Payment type cannot be empty")
	}
	if req.MaxAmount > 0 && req.MinAmount > req.MaxAmount {
		return nil, errors.New("Minimum amount cannot exceed maximum amount")
	}
	if req.FeeRate < 0 || req.FeeRate > 100 {
		return nil, errors.New("Fee rate must be between 0 and 100")
	}

	if req.SignType == "" {
		req.SignType = "MD5"
	}
	if req.Currency == "" {
		req.Currency = "CNY"
	}
	if req.ExchangeRateMode == "" {
		req.ExchangeRateMode = payment.ExchangeRateModeSystem
	}
	if req.TargetFeeMode == "" {
		req.TargetFeeMode = payment.FeeModeAdd
	}
	if req.ExpireMinutes <= 0 {
		req.ExpireMinutes = getOrderExpireMinutes()
	}
	if req.QueryIntervalSeconds <= 0 {
		req.QueryIntervalSeconds = 120
	}
	if req.QueryBatchSize <= 0 {
		req.QueryBatchSize = 50
	}

	gw := &models.PayGateway{
		Name:                 req.Name,
		Type:                 req.Type,
		PayType:              req.PayType,
		SignType:             req.SignType,
		Version:              req.Version,
		Device:               req.Device,
		Currency:             req.Currency,
		TargetCurrency:       req.TargetCurrency,
		ExchangeRateMode:     req.ExchangeRateMode,
		ExchangeRate:         req.ExchangeRate,
		ExchangeFixedAmount:  req.ExchangeFixedAmount,
		ExchangeRateSource:   req.ExchangeRateSource,
		TargetFeeRate:        req.TargetFeeRate,
		TargetFeeFixed:       req.TargetFeeFixed,
		TargetFeeMode:        req.TargetFeeMode,
		Description:          req.Description,
		Status:               req.Status,
		ApiURL:               req.ApiURL,
		PID:                  req.PID,
		Key:                  req.Key,
		ExtConfig:            req.ExtConfig,
		LogoURL:              req.LogoURL,
		SortOrder:            req.SortOrder,
		MinAmount:            req.MinAmount,
		MaxAmount:            req.MaxAmount,
		FeeRate:              req.FeeRate,
		FeeMode:              req.FeeMode,
		MinLevel:             req.MinLevel,
		NotifyURL:            req.NotifyURL,
		ExpireMinutes:        req.ExpireMinutes,
		ActiveQueryEnabled:   req.ActiveQueryEnabled,
		QueryIntervalSeconds: req.QueryIntervalSeconds,
		QueryBatchSize:       req.QueryBatchSize,
	}

	if err := models.CreatePayGateway(gw); err != nil {
		return nil, errors.New("创建支付通道失败: " + err.Error())
	}

	return gw, nil
}

// UpdatePayGateway 更新支付通道
func UpdatePayGateway(id uint64, req *PayGatewayUpdateRequest) (*models.PayGateway, error) {
	gw, err := models.GetPayGatewayByID(id)
	if err != nil {
		return nil, errors.New("Payment gateway does not exist")
	}

	pendingCount, err := models.CountPendingOrdersByGatewayID(id)
	if err != nil {
		return nil, errors.New("检查在途订单失败: " + err.Error())
	}
	// 允许在存在待支付订单时修改 PID/密钥：运维纠错（密钥填错）时不能被卡死。
	// 风险：旧待支付单的回调验签可能失败，需管理员自行处理（取消旧单或补单）。
	if pendingCount > 0 {
		keyChanged := req.Key != nil && *req.Key != gw.Key
		extChanged := req.ExtConfig != nil && *req.ExtConfig != gw.ExtConfig
		if (req.PID != nil && *req.PID != gw.PID) || keyChanged || extChanged {
			log.Printf("[PayGateway] 存在 %d 笔待支付订单，仍修改通道敏感配置: gateway_id=%d", pendingCount, id)
		}
	}

	if req.Name != nil {
		gw.Name = *req.Name
	}
	if req.Type != nil {
		gw.Type = *req.Type
	}
	if req.PayType != nil {
		gw.PayType = *req.PayType
	}
	if req.SignType != nil {
		gw.SignType = *req.SignType
	}
	if req.Version != nil {
		gw.Version = *req.Version
	}
	if req.Device != nil {
		gw.Device = *req.Device
	}
	if req.Currency != nil {
		gw.Currency = *req.Currency
	}
	if req.TargetCurrency != nil {
		gw.TargetCurrency = *req.TargetCurrency
	}
	if req.ExchangeRateMode != nil {
		gw.ExchangeRateMode = *req.ExchangeRateMode
	}
	if req.ExchangeRate != nil {
		gw.ExchangeRate = *req.ExchangeRate
	}
	if req.ExchangeFixedAmount != nil {
		gw.ExchangeFixedAmount = *req.ExchangeFixedAmount
	}
	if req.ExchangeRateSource != nil {
		gw.ExchangeRateSource = *req.ExchangeRateSource
	}
	if req.TargetFeeRate != nil {
		if *req.TargetFeeRate < 0 || *req.TargetFeeRate > 100 {
			return nil, errors.New("Target fee rate must be between 0 and 100")
		}
		gw.TargetFeeRate = *req.TargetFeeRate
	}
	if req.TargetFeeFixed != nil {
		gw.TargetFeeFixed = *req.TargetFeeFixed
	}
	if req.TargetFeeMode != nil {
		gw.TargetFeeMode = *req.TargetFeeMode
	}
	if req.Description != nil {
		gw.Description = *req.Description
	}
	if req.Status != nil {
		gw.Status = *req.Status
	}
	if req.ApiURL != nil {
		gw.ApiURL = *req.ApiURL
	}
	if req.PID != nil {
		gw.PID = *req.PID
	}
	if req.Key != nil {
		gw.Key = *req.Key
	}
	if req.ExtConfig != nil {
		gw.ExtConfig = *req.ExtConfig
	}
	if req.LogoURL != nil {
		gw.LogoURL = *req.LogoURL
	}
	if req.SortOrder != nil {
		gw.SortOrder = *req.SortOrder
	}
	if req.MinAmount != nil {
		gw.MinAmount = *req.MinAmount
	}
	if req.MaxAmount != nil {
		gw.MaxAmount = *req.MaxAmount
	}
	if req.FeeRate != nil {
		if *req.FeeRate < 0 || *req.FeeRate > 100 {
			return nil, errors.New("Fee rate must be between 0 and 100")
		}
		gw.FeeRate = *req.FeeRate
	}
	if req.FeeMode != nil {
		gw.FeeMode = *req.FeeMode
	}
	if req.MinLevel != nil {
		gw.MinLevel = *req.MinLevel
	}
	if req.NotifyURL != nil {
		gw.NotifyURL = *req.NotifyURL
	}
	if req.ExpireMinutes != nil {
		gw.ExpireMinutes = *req.ExpireMinutes
	}
	if req.ActiveQueryEnabled != nil {
		gw.ActiveQueryEnabled = *req.ActiveQueryEnabled
	}
	if req.QueryIntervalSeconds != nil {
		gw.QueryIntervalSeconds = *req.QueryIntervalSeconds
	}
	if req.QueryBatchSize != nil {
		gw.QueryBatchSize = *req.QueryBatchSize
	}

	// 验证金额
	if gw.MaxAmount > 0 && gw.MinAmount > gw.MaxAmount {
		return nil, errors.New("Minimum amount cannot exceed maximum amount")
	}

	if err := models.UpdatePayGateway(gw); err != nil {
		return nil, errors.New("更新支付通道失败: " + err.Error())
	}

	return gw, nil
}

// DeletePayGateway 删除支付通道
func DeletePayGateway(id uint64) error {
	_, err := models.GetPayGatewayByID(id)
	if err != nil {
		return errors.New("Payment gateway does not exist")
	}
	pendingCount, err := models.CountPendingOrdersByGatewayID(id)
	if err != nil {
		return errors.New("检查在途订单失败: " + err.Error())
	}
	if pendingCount > 0 {
		return errors.New("There are pending orders, cannot delete this payment gateway")
	}
	return models.DeletePayGateway(id)
}

// GetPayGatewayListForAdmin 管理端获取支付通道列表（运营密钥明文返回，不掩码）
func GetPayGatewayListForAdmin(page, pageSize int, keyword string) ([]models.PayGateway, int64, error) {
	return models.GetPayGatewayList(page, pageSize, keyword, false)
}

// GetPayGatewayDetailForAdmin 管理端获取支付通道详情（运营密钥明文返回）
func GetPayGatewayDetailForAdmin(id uint64) (*models.PayGateway, error) {
	return models.GetPayGatewayByID(id)
}

// GetPayGatewayListForUser 用户端获取支付通道列表（隐藏敏感信息）
func GetPayGatewayListForUser() ([]models.PayGateway, error) {
	if !GetGlobalPaymentEnabled() {
		return []models.PayGateway{}, nil
	}

	gateways, err := models.GetEnabledPayGateways()
	if err != nil {
		return nil, err
	}

	// 隐藏敏感信息
	for i := range gateways {
		gateways[i].ApiURL = ""
		gateways[i].Key = ""
		gateways[i].ExtConfig = ""
		gateways[i].PID = ""
		gateways[i].NotifyURL = ""
	}

	return gateways, nil
}

// PaymentChannelMetaView 返回已注册支付通道元数据，供管理端动态渲染表单
type PaymentChannelMetaView struct {
	Type              string                       `json:"type"`
	Name              string                       `json:"name"`
	Currency          string                       `json:"currency"`
	PayTypes          []payment.PayTypeMeta        `json:"pay_types"`
	Devices           []payment.DeviceMeta         `json:"devices"`
	DefaultNotifyPath string                       `json:"default_notify_path"`
	Versions          []payment.ChannelVersionMeta `json:"versions"`
}

// TestGatewayConnection 测试支付通道配置是否可用
// 返回 (是否可用, 提示信息)
func TestGatewayConnection(gatewayID uint64) (bool, string, error) {
	gateway, err := models.GetPayGatewayByID(gatewayID)
	if err != nil {
		return false, "", err
	}

	provider := payment.GetProvider(gateway.Type)
	if provider == nil {
		// 未注册新 Provider 时，只检查基本配置是否存在
		if gateway.ApiURL == "" || gateway.PID == "" {
			return false, "API 地址或商户 ID 为空", nil
		}
		return true, "旧通道配置存在", nil
	}

	extConfig := gatewayExtConfig(gateway)
	ok, msg := provider.TestConnection(context.Background(), extConfig)
	return ok, msg, nil
}

// ListPaymentChannelMetas 列出已注册通道类型及其版本/配置字段
func ListPaymentChannelMetas() []PaymentChannelMetaView {
	metas := payment.ListChannelMetas()
	out := make([]PaymentChannelMetaView, 0, len(metas))
	for _, m := range metas {
		out = append(out, PaymentChannelMetaView{
			Type:              m.Type,
			Name:              m.Name,
			Currency:          m.Currency,
			PayTypes:          m.PayTypes,
			Devices:           m.Devices,
			DefaultNotifyPath: m.DefaultNotifyPath,
			Versions:          m.Versions,
		})
	}
	return out
}

// CalculateFee 计算手续费与到账金额
// feeRate: 百分比（1 = 1%），内部按「分」整数算避免 float 误差
// feeMode: add 用户多付 / include 手续费从金额中扣除
func CalculateFee(amount float64, feeRate int, feeMode string) (fee float64, payAmount float64, creditAmount float64) {
	amountFen, err := utils.YuanToFen(amount)
	if err != nil || amountFen <= 0 {
		return 0, amount, amount
	}
	if feeRate <= 0 {
		yuan := utils.FenToYuan(amountFen)
		return 0, yuan, yuan
	}

	feeFen := int64(math.Round(float64(amountFen) * float64(feeRate) / 100.0))
	if feeFen < 0 {
		feeFen = 0
	}

	var payFen, creditFen int64
	if feeMode == models.FeeModAdd {
		payFen = amountFen + feeFen
		creditFen = amountFen
	} else {
		payFen = amountFen
		creditFen = amountFen - feeFen
		if creditFen < 0 {
			creditFen = 0
		}
	}

	return utils.FenToYuan(feeFen), utils.FenToYuan(payFen), utils.FenToYuan(creditFen)
}
