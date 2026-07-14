package plugins

// PluginType 插件类型
type PluginType string

const (
	// PluginTypePayment 支付插件
	PluginTypePayment PluginType = "payment"
	// PluginTypeShipping 物流插件
	PluginTypeShipping PluginType = "shipping"
	// PluginTypeNotification 通知插件
	PluginTypeNotification PluginType = "notification"
	// PluginTypeProduct 产品扩展插件
	PluginTypeProduct PluginType = "product"
	// PluginTypeOrder 订单处理插件
	PluginTypeOrder PluginType = "order"
	// PluginTypeUser 用户扩展插件
	PluginTypeUser PluginType = "user"
	// PluginTypeStorage 存储插件
	PluginTypeStorage PluginType = "storage"
	// PluginTypeOther 其他插件
	PluginTypeOther PluginType = "other"
)

// PluginStatus 插件状态
type PluginStatus int

const (
	// PluginStatusDisabled 禁用
	PluginStatusDisabled PluginStatus = 0
	// PluginStatusEnabled 启用
	PluginStatusEnabled PluginStatus = 1
)

// Plugin 插件接口
// 所有插件必须实现此接口
type Plugin interface {
	// GetID 获取插件唯一标识
	GetID() string
	// GetName 获取插件名称
	GetName() string
	// GetVersion 获取插件版本
	GetVersion() string
	// GetDescription 获取插件描述
	GetDescription() string
	// GetType 获取插件类型
	GetType() PluginType
	// GetAuthor 获取插件作者
	GetAuthor() string

	// Install 安装插件
	Install() error
	// Uninstall 卸载插件
	Uninstall() error
	// Enable 启用插件
	Enable() error
	// Disable 禁用插件
	Disable() error

	// Init 初始化插件
	Init(config map[string]interface{}) error
	// GetConfig 获取插件配置
	GetConfig() map[string]interface{}
	// SetConfig 设置插件配置
	SetConfig(config map[string]interface{}) error

	// IsEnabled 检查插件是否启用
	IsEnabled() bool
}

// PaymentPlugin 支付插件接口
type PaymentPlugin interface {
	Plugin
	// CreatePayment 创建支付
	CreatePayment(orderID uint, amount float64, currency string, metadata map[string]interface{}) (paymentURL string, err error)
	// VerifyPayment 验证支付
	VerifyPayment(paymentData map[string]interface{}) (success bool, tradeNo string, err error)
	// Refund 退款
	Refund(orderID uint, amount float64, reason string) (refundID string, err error)
	// GetPaymentStatus 获取支付状态
	GetPaymentStatus(tradeNo string) (status string, err error)
}

// ShippingPlugin 物流插件接口
type ShippingPlugin interface {
	Plugin
	// CreateShipment 创建发货
	CreateShipment(orderID uint, address map[string]interface{}, items []map[string]interface{}) (trackingNo string, err error)
	// GetTrackingInfo 获取物流信息
	GetTrackingInfo(trackingNo string) (info map[string]interface{}, err error)
	// CalculateFee 计算运费
	CalculateFee(address map[string]interface{}, items []map[string]interface{}) (fee float64, err error)
}

// NotificationPlugin 通知插件接口
type NotificationPlugin interface {
	Plugin
	// Send 发送通知
	Send(to string, subject string, content string, template string, data map[string]interface{}) error
	// SendBatch 批量发送
	SendBatch(to []string, subject string, content string, template string, data map[string]interface{}) error
}

// ProductExtension 产品扩展接口
// 用于扩展产品功能和属性
type ProductExtension interface {
	Plugin
	// GetProductFields 获取产品额外字段定义
	GetProductFields() []ProductField
	// ValidateProductData 验证产品数据
	ValidateProductData(data map[string]interface{}) error
	// ProcessProductData 处理产品数据(保存前)
	ProcessProductData(data map[string]interface{}) (map[string]interface{}, error)
	// RenderProductDetail 渲染产品详情
	RenderProductDetail(productID uint) (map[string]interface{}, error)
}

// ProductField 产品字段定义
type ProductField struct {
	Name        string `json:"name"`        // 字段名
	Type        string `json:"type"`        // 字段类型(string/int/float/bool/json)
	Label       string `json:"label"`       // 字段标签
	Required    bool   `json:"required"`    // 是否必填
	Default     any    `json:"default"`     // 默认值
	Options     []any  `json:"options"`     // 选项(用于select类型)
	Description string `json:"description"` // 字段描述
}

// PluginInfo 插件信息结构
type PluginInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Type        PluginType             `json:"type"`
	Author      string                 `json:"author"`
	Status      PluginStatus           `json:"status"`
	Config      map[string]interface{} `json:"config"`
	InstalledAt int64                  `json:"installed_at"`
	UpdatedAt   int64                  `json:"updated_at"`
}
