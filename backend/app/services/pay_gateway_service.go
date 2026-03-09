package services

import (
	"errors"
	"fst/backend/app/models"
)

// PayGatewayCreateRequest 创建支付通道请求
type PayGatewayCreateRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Type        string  `json:"type" binding:"required,max=50"`
	PayType     string  `json:"pay_type" binding:"required,max=50"`
	Description string  `json:"description" binding:"omitempty,max=500"`
	Status      int     `json:"status"`
	ApiURL      string  `json:"api_url" binding:"omitempty"`
	PID         string  `json:"pid" binding:"omitempty"`
	Key         string  `json:"key" binding:"omitempty"`
	LogoURL     string  `json:"logo_url" binding:"omitempty"`
	SortOrder   int     `json:"sort_order"`
	MinAmount   float64 `json:"min_amount"`
	MaxAmount   float64 `json:"max_amount"`
	FeeRate     int     `json:"fee_rate"`
	FeeMode     string  `json:"fee_mode" binding:"omitempty,max=50"`
	MinLevel    int     `json:"min_level"`
	NotifyURL   string  `json:"notify_url" binding:"omitempty"`
}

// PayGatewayUpdateRequest 更新支付通道请求
type PayGatewayUpdateRequest struct {
	Name        *string  `json:"name" binding:"omitempty,max=100"`
	Type        *string  `json:"type" binding:"omitempty,max=50"`
	PayType     *string  `json:"pay_type" binding:"omitempty,max=50"`
	Description *string  `json:"description" binding:"omitempty,max=500"`
	Status      *int     `json:"status"`
	ApiURL      *string  `json:"api_url" binding:"omitempty"`
	PID         *string  `json:"pid" binding:"omitempty"`
	Key         *string  `json:"key" binding:"omitempty"`
	LogoURL     *string  `json:"logo_url" binding:"omitempty"`
	SortOrder   *int     `json:"sort_order"`
	MinAmount   *float64 `json:"min_amount"`
	MaxAmount   *float64 `json:"max_amount"`
	FeeRate     *int     `json:"fee_rate"`
	FeeMode     *string  `json:"fee_mode" binding:"omitempty,max=50"`
	MinLevel    *int     `json:"min_level"`
	NotifyURL   *string  `json:"notify_url" binding:"omitempty"`
}

// CreatePayGateway 创建支付通道
func CreatePayGateway(req *PayGatewayCreateRequest) (*models.PayGateway, error) {
	if req.Name == "" {
		return nil, errors.New("通道名称不能为空")
	}
	if req.PayType == "" {
		return nil, errors.New("支付类型不能为空")
	}
	if req.MaxAmount > 0 && req.MinAmount > req.MaxAmount {
		return nil, errors.New("最小金额不能大于最大金额")
	}
	if req.FeeRate < 0 || req.FeeRate > 100 {
		return nil, errors.New("手续费率必须在 0-100 之间")
	}

	gw := &models.PayGateway{
		Name:        req.Name,
		Type:        req.Type,
		PayType:     req.PayType,
		Description: req.Description,
		Status:      req.Status,
		ApiURL:      req.ApiURL,
		PID:         req.PID,
		Key:         req.Key,
		LogoURL:     req.LogoURL,
		SortOrder:   req.SortOrder,
		MinAmount:   req.MinAmount,
		MaxAmount:   req.MaxAmount,
		FeeRate:     req.FeeRate,
		FeeMode:     req.FeeMode,
		MinLevel:    req.MinLevel,
		NotifyURL:   req.NotifyURL,
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
		return nil, errors.New("支付通道不存在")
	}

	pendingCount, err := models.CountPendingOrdersByGatewayID(id)
	if err != nil {
		return nil, errors.New("检查在途订单失败: " + err.Error())
	}
	hasPendingOrders := pendingCount > 0
	if hasPendingOrders {
		if req.PID != nil && *req.PID != gw.PID {
			return nil, errors.New("存在待支付订单时不允许修改商户PID")
		}
		if req.Key != nil && *req.Key != gw.Key {
			return nil, errors.New("存在待支付订单时不允许修改商户密钥")
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
			return nil, errors.New("手续费率必须在 0-100 之间")
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

	// 验证金额
	if gw.MaxAmount > 0 && gw.MinAmount > gw.MaxAmount {
		return nil, errors.New("最小金额不能大于最大金额")
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
		return errors.New("支付通道不存在")
	}
	pendingCount, err := models.CountPendingOrdersByGatewayID(id)
	if err != nil {
		return errors.New("检查在途订单失败: " + err.Error())
	}
	if pendingCount > 0 {
		return errors.New("存在待支付订单，不能删除该支付通道")
	}
	return models.DeletePayGateway(id)
}

// GetPayGatewayListForAdmin 管理端获取支付通道列表（包含全部信息）
func GetPayGatewayListForAdmin(page, pageSize int, keyword string) ([]models.PayGateway, int64, error) {
	return models.GetPayGatewayList(page, pageSize, keyword, false)
}

// GetPayGatewayListForUser 用户端获取支付通道列表（隐藏敏感信息）
func GetPayGatewayListForUser() ([]models.PayGateway, error) {
	settingsMap, err := models.GetSettingsMap([]string{"payment_enabled"})
	if err == nil {
		paymentEnabled := settingsMap["payment_enabled"] == "true" || settingsMap["payment_enabled"] == "1"
		if !paymentEnabled {
			return []models.PayGateway{}, nil
		}
	}

	gateways, err := models.GetEnabledPayGateways()
	if err != nil {
		return nil, err
	}

	// 隐藏敏感信息
	for i := range gateways {
		gateways[i].ApiURL = ""
		gateways[i].Key = ""
		gateways[i].PID = ""
		gateways[i].NotifyURL = ""
	}

	return gateways, nil
}

// CalculateFee 计算手续费
// 返回: 手续费金额, 实际支付金额（用户掏的钱）, 到账金额
func CalculateFee(amount float64, feeRate int, feeMode string) (fee float64, payAmount float64, creditAmount float64) {
	if feeRate <= 0 {
		return 0, amount, amount
	}

	rate := float64(feeRate) / 100.0

	if feeMode == models.FeeModAdd {
		// 加收模式：用户多付手续费，到账金额 = 充值金额
		fee = amount * rate
		payAmount = amount + fee
		creditAmount = amount
	} else {
		// 包含模式（默认）：到账金额 = 充值金额 - 手续费
		fee = amount * rate
		payAmount = amount
		creditAmount = amount - fee
	}

	return fee, payAmount, creditAmount
}
