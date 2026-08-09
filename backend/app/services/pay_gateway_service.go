package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/payment"
	"fst/backend/utils"
	"log"
	"math"
	"strconv"
	"strings"
)

// PayGatewayCreateRequest 创建支付通道请求
// 已迁移到 ext_config 的字段不再单独接收，由前端统一放到 ext_config JSON 中。
type PayGatewayCreateRequest struct {
	Name          string  `json:"name" binding:"required,max=100"`
	Type          string  `json:"type" binding:"required,max=50"`
	PayType       string  `json:"pay_type" binding:"required,max=50"`
	Version       string  `json:"version" binding:"omitempty,max=50"`
	Device        string  `json:"device" binding:"omitempty,max=50"`
	Currency      string  `json:"currency" binding:"omitempty,max=10"`
	Description   string  `json:"description" binding:"omitempty,max=500"`
	Status        int     `json:"status"`
	ApiURL        string  `json:"api_url" binding:"omitempty"`
	PID           string  `json:"pid" binding:"omitempty"`
	ExtConfig     string  `json:"ext_config" binding:"omitempty"`
	LogoURL       string  `json:"logo_url" binding:"omitempty"`
	SortOrder     int     `json:"sort_order"`
	MinAmount     float64 `json:"min_amount"`
	MaxAmount     float64 `json:"max_amount"`
	FeeRate       int     `json:"fee_rate"`
	FeeMode       string  `json:"fee_mode" binding:"omitempty,max=50"`
	MinLevel      int     `json:"min_level"`
	NotifyURL     string  `json:"notify_url" binding:"omitempty"`
	ExpireMinutes int     `json:"expire_minutes"`
}

// PayGatewayUpdateRequest 更新支付通道请求
type PayGatewayUpdateRequest struct {
	Name          *string  `json:"name" binding:"omitempty,max=100"`
	Type          *string  `json:"type" binding:"omitempty,max=50"`
	PayType       *string  `json:"pay_type" binding:"omitempty,max=50"`
	Version       *string  `json:"version" binding:"omitempty,max=50"`
	Device        *string  `json:"device" binding:"omitempty,max=50"`
	Currency      *string  `json:"currency" binding:"omitempty,max=10"`
	Description   *string  `json:"description" binding:"omitempty,max=500"`
	Status        *int     `json:"status"`
	ApiURL        *string  `json:"api_url" binding:"omitempty"`
	PID           *string  `json:"pid" binding:"omitempty"`
	ExtConfig     *string  `json:"ext_config" binding:"omitempty"`
	LogoURL       *string  `json:"logo_url" binding:"omitempty"`
	SortOrder     *int     `json:"sort_order"`
	MinAmount     *float64 `json:"min_amount"`
	MaxAmount     *float64 `json:"max_amount"`
	FeeRate       *int     `json:"fee_rate"`
	FeeMode       *string  `json:"fee_mode" binding:"omitempty,max=50"`
	MinLevel      *int     `json:"min_level"`
	NotifyURL     *string  `json:"notify_url" binding:"omitempty"`
	ExpireMinutes *int     `json:"expire_minutes"`
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

	if req.Currency == "" {
		req.Currency = "CNY"
	}
	if req.ExpireMinutes <= 0 {
		req.ExpireMinutes = getOrderExpireMinutes()
	}

	// 解析前端传入的 ext_config JSON，补充默认值后统一落库
	extMap := payment.ParseExtConfigMap(req.ExtConfig)
	applyPayGatewayExtDefaults(extMap)

	gw := &models.PayGateway{
		Name:          req.Name,
		Type:          req.Type,
		PayType:       req.PayType,
		Version:       req.Version,
		Device:        req.Device,
		Currency:      req.Currency,
		Description:   req.Description,
		Status:        req.Status,
		ApiURL:        req.ApiURL,
		PID:           req.PID,
		ExtConfig:     payment.MarshalExtConfigMap(extMap),
		LogoURL:       req.LogoURL,
		SortOrder:     req.SortOrder,
		MinAmount:     req.MinAmount,
		MaxAmount:     req.MaxAmount,
		FeeRate:       req.FeeRate,
		FeeMode:       req.FeeMode,
		MinLevel:      req.MinLevel,
		NotifyURL:     req.NotifyURL,
		ExpireMinutes: req.ExpireMinutes,
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
		extChanged := req.ExtConfig != nil && *req.ExtConfig != gw.ExtConfig
		if (req.PID != nil && *req.PID != gw.PID) || extChanged {
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
	if req.Version != nil {
		gw.Version = *req.Version
	}
	if req.Device != nil {
		gw.Device = *req.Device
	}
	if req.Currency != nil {
		gw.Currency = *req.Currency
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
	if req.ExtConfig != nil {
		// 合并现有 ext_config 与前端传入的新值，再补充默认值
		extMap := gw.ExtConfigMap()
		newMap := payment.ParseExtConfigMap(*req.ExtConfig)
		for k, v := range newMap {
			extMap[k] = v
		}
		applyPayGatewayExtDefaults(extMap)
		gw.ExtConfig = payment.MarshalExtConfigMap(extMap)
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
	ConfigFields      []payment.ConfigField        `json:"config_fields"` // 网关级动态配置字段
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
		m.ConfigFields = append(m.ConfigFields, defaultPayGatewayConfigFields()...)
		out = append(out, PaymentChannelMetaView{
			Type:              m.Type,
			Name:              m.Name,
			Currency:          m.Currency,
			PayTypes:          m.PayTypes,
			Devices:           m.Devices,
			DefaultNotifyPath: m.DefaultNotifyPath,
			Versions:          m.Versions,
			ConfigFields:      m.ConfigFields,
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

// applyPayGatewayExtDefaults 给 ext_config map 填充已迁移字段的默认值。
// 只处理缺省或空值的字段；显式设置为 0 的开关字段（如 active_query_enabled=0）会被保留。
func applyPayGatewayExtDefaults(m map[string]interface{}) {
	if m == nil {
		return
	}

	if v, ok := m["sign_type"]; !ok || payGatewayExtString(v) == "" {
		m["sign_type"] = "MD5"
	}
	if _, ok := m["key"]; !ok {
		m["key"] = ""
	}
	if _, ok := m["target_currency"]; !ok {
		m["target_currency"] = ""
	}
	if v, ok := m["exchange_rate_mode"]; !ok || payGatewayExtString(v) == "" {
		m["exchange_rate_mode"] = payment.ExchangeRateModeSystem
	}
	if _, ok := m["exchange_rate"]; !ok {
		m["exchange_rate"] = 0
	}
	if _, ok := m["exchange_fixed_amount"]; !ok {
		m["exchange_fixed_amount"] = 0
	}
	if _, ok := m["exchange_rate_source"]; !ok {
		m["exchange_rate_source"] = ""
	}
	if _, ok := m["target_fee_rate"]; !ok {
		m["target_fee_rate"] = 0
	}
	if _, ok := m["target_fee_fixed"]; !ok {
		m["target_fee_fixed"] = 0
	}
	if v, ok := m["target_fee_mode"]; !ok || payGatewayExtString(v) == "" {
		m["target_fee_mode"] = payment.FeeModeAdd
	}
	if _, ok := m["active_query_enabled"]; !ok {
		m["active_query_enabled"] = 1
	}
	if v, ok := m["query_interval_seconds"]; !ok || payGatewayExtInt(v) <= 0 {
		m["query_interval_seconds"] = 120
	}
	if v, ok := m["query_batch_size"]; !ok || payGatewayExtInt(v) <= 0 {
		m["query_batch_size"] = 50
	}
}

// payGatewayExtString 把 ext_config map 中的值转成字符串（辅助 applyPayGatewayExtDefaults）
func payGatewayExtString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case json.Number:
		return val.String()
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int, int64, int32:
		return fmt.Sprintf("%v", v)
	case bool:
		return strconv.FormatBool(val)
	case nil:
		return ""
	default:
		return fmt.Sprint(v)
	}
}

// payGatewayExtInt 把 ext_config map 中的值转成 int（辅助 applyPayGatewayExtDefaults）
func payGatewayExtInt(v interface{}) int {
	switch val := v.(type) {
	case json.Number:
		i, err := val.Int64()
		if err != nil {
			return 0
		}
		return int(i)
	case float64:
		return int(val)
	case float32:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	case int32:
		return int(val)
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}

// defaultPayGatewayConfigFields 返回已迁移到 ext_config 的通用网关配置字段 schema，
// 供管理端动态渲染表单。
func defaultPayGatewayConfigFields() []payment.ConfigField {
	return []payment.ConfigField{
		{
			Name:        "key",
			Label:       "商户密钥",
			Type:        "textarea",
			Secret:      true,
			Placeholder: "通道统一密钥/单密钥，MD5 签名等",
		},
		{
			Name:    "sign_type",
			Label:   "签名算法",
			Type:    "select",
			Options: []payment.ConfigFieldOption{{Value: "MD5", Label: "MD5"}, {Value: "RSA", Label: "RSA"}},
		},
		{
			Name:        "target_currency",
			Label:       "目标币种",
			Type:        "input",
			Placeholder: "如 USD / CNY，空则与币种一致",
		},
		{
			Name:    "exchange_rate_mode",
			Label:   "汇率模式",
			Type:    "select",
			Options: []payment.ConfigFieldOption{{Value: payment.ExchangeRateModeSystem, Label: "系统汇率"}, {Value: payment.ExchangeRateModeFixed, Label: "固定汇率"}, {Value: payment.ExchangeRateModeDynamic, Label: "动态汇率"}},
		},
		{
			Name:        "exchange_rate",
			Label:       "固定汇率",
			Type:        "input",
			Placeholder: "汇率模式为 fixed 时必填",
		},
		{
			Name:        "exchange_fixed_amount",
			Label:       "固定加额",
			Type:        "input",
			Placeholder: "转换后固定加额（元）",
		},
		{
			Name:        "exchange_rate_source",
			Label:       "汇率源标识",
			Type:        "input",
			Placeholder: "动态汇率源标识，如 exchangerate-api",
		},
		{
			Name:        "target_fee_rate",
			Label:       "目标手续费率",
			Type:        "input",
			Placeholder: "百分之 x，如 200 表示 2%",
		},
		{
			Name:        "target_fee_fixed",
			Label:       "目标固定手续费",
			Type:        "input",
			Placeholder: "元",
		},
		{
			Name:    "target_fee_mode",
			Label:   "目标手续费模式",
			Type:    "select",
			Options: []payment.ConfigFieldOption{{Value: payment.FeeModeAdd, Label: "加收（add）"}, {Value: payment.FeeModeInclude, Label: "内扣（include）"}},
		},
		{
			Name:    "active_query_enabled",
			Label:   "主动查单开关",
			Type:    "select",
			Options: []payment.ConfigFieldOption{{Value: "1", Label: "开启"}, {Value: "0", Label: "关闭"}},
		},
		{
			Name:        "query_interval_seconds",
			Label:       "查单间隔",
			Type:        "input",
			Placeholder: "秒，默认 120",
		},
		{
			Name:        "query_batch_size",
			Label:       "查单批次",
			Type:        "input",
			Placeholder: "默认 50",
		},
	}
}
