package models

import (
	"gorm.io/gorm"
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"    // 待付款
	OrderStatusPaid       OrderStatus = "paid"       // 已付款
	OrderStatusProcessing OrderStatus = "processing" // 处理中
	OrderStatusShipped    OrderStatus = "shipped"    // 已发货(实体商品)
	OrderStatusCompleted  OrderStatus = "completed"  // 已完成
	OrderStatusCancelled  OrderStatus = "cancelled"  // 已取消
	OrderStatusRefunded   OrderStatus = "refunded"   // 已退款
)

// OrderType 订单类型
type OrderType string

const (
	OrderTypeC2C  OrderType = "c2c"  // C2C交易
	OrderTypeB2C  OrderType = "b2c"  // B2C交易
	OrderTypeSaaS OrderType = "saas" // SaaS订阅
)

// PaymentStatus 支付状态
type PaymentStatus string

const (
	PaymentStatusUnpaid   PaymentStatus = "unpaid"   // 未支付
	PaymentStatusPaid     PaymentStatus = "paid"     // 已支付
	PaymentStatusPartial  PaymentStatus = "partial"  // 部分支付
	PaymentStatusRefunded PaymentStatus = "refunded" // 已退款
	PaymentStatusFailed   PaymentStatus = "failed"   // 支付失败
)

// Order 订单模型
// 支持C2C/B2C/SaaS等多种交易模式
type Order struct {
	ID          uint   `json:"id" gorm:"primaryKey;type:bigint(20) unsigned;autoIncrement;comment:订单ID"`
	OrderNumber string `json:"order_number" gorm:"uniqueIndex;size:50;not null;comment:订单编号"`

	// 交易双方
	BuyerID  uint  `json:"buyer_id" gorm:"index;not null;comment:买家ID"`
	Buyer    *User `json:"buyer,omitempty" gorm:"foreignKey:BuyerID"`
	SellerID uint  `json:"seller_id" gorm:"index;not null;comment:卖家ID(0=平台)"`
	Seller   *User `json:"seller,omitempty" gorm:"foreignKey:SellerID"`

	// 订单类型
	Type OrderType `json:"type" gorm:"index;size:20;not null;default:'b2c';comment:订单类型:c2c/b2c/saas"`

	// 商品信息
	ProductID   uint     `json:"product_id" gorm:"index;not null;comment:产品ID"`
	Product     *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	ProductName string   `json:"product_name" gorm:"size:255;not null;comment:商品名称(快照)"`
	Quantity    int64    `json:"quantity" gorm:"default:1;not null;comment:购买数量"`
	UnitPrice   float64  `json:"unit_price" gorm:"type:decimal(12,2);not null;comment:单价(快照)"`

	// 金额信息
	Subtotal    float64 `json:"subtotal" gorm:"type:decimal(12,2);not null;comment:小计金额"`
	Discount    float64 `json:"discount" gorm:"type:decimal(12,2);default:0.00;not null;comment:折扣金额"`
	Tax         float64 `json:"tax" gorm:"type:decimal(12,2);default:0.00;not null;comment:税费"`
	ShippingFee float64 `json:"shipping_fee" gorm:"type:decimal(12,2);default:0.00;not null;comment:运费"`
	TotalAmount float64 `json:"total_amount" gorm:"type:decimal(12,2);index;not null;comment:订单总金额"`
	Currency    string  `json:"currency" gorm:"size:10;not null;default:'CNY';comment:货币代码"`

	// 订单状态
	Status OrderStatus `json:"status" gorm:"index;size:20;not null;default:'pending';comment:订单状态"`

	// 支付信息
	PaymentStatus  PaymentStatus `json:"payment_status" gorm:"index;size:20;not null;default:'unpaid';comment:支付状态"`
	PaymentMethod  string        `json:"payment_method" gorm:"size:50;default:'';comment:支付方式"`
	PaymentTime    *uint64       `json:"payment_time" gorm:"type:bigint(16) unsigned;comment:支付时间戳"`
	PaymentTradeNo string        `json:"payment_trade_no" gorm:"size:100;default:'';comment:第三方支付流水号"`
	PaymentData    string        `json:"payment_data" gorm:"type:json;comment:支付详情数据"`

	// 配送信息(实体商品)
	ShippingAddress string  `json:"shipping_address" gorm:"type:json;comment:配送地址"`
	ShippingCompany string  `json:"shipping_company" gorm:"size:100;default:'';comment:物流公司"`
	TrackingNumber  string  `json:"tracking_number" gorm:"size:100;default:'';comment:物流单号"`
	ShippedTime     *uint64 `json:"shipped_time" gorm:"type:bigint(16) unsigned;comment:发货时间戳"`
	ReceivedTime    *uint64 `json:"received_time" gorm:"type:bigint(16) unsigned;comment:签收时间戳"`

	// SaaS订阅信息
	SubscriptionStart *uint64 `json:"subscription_start" gorm:"type:bigint(16) unsigned;comment:订阅开始时间"`
	SubscriptionEnd   *uint64 `json:"subscription_end" gorm:"type:bigint(16) unsigned;comment:订阅结束时间"`
	BillingCycle      string  `json:"billing_cycle" gorm:"size:20;default:'';comment:计费周期:monthly/yearly"`

	// 其他信息
	Remark       string  `json:"remark" gorm:"size:500;default:'';comment:订单备注"`
	UserRemark   string  `json:"user_remark" gorm:"size:500;default:'';comment:用户备注"`
	CancelReason string  `json:"cancel_reason" gorm:"size:255;default:'';comment:取消原因"`
	RefundReason string  `json:"refund_reason" gorm:"size:255;default:'';comment:退款原因"`
	RefundAmount float64 `json:"refund_amount" gorm:"type:decimal(12,2);default:0.00;comment:退款金额"`

	// 时间戳
	CreateTime    *uint64        `json:"create_time" gorm:"index;type:bigint(16) unsigned;comment:创建时间戳"`
	UpdateTime    *uint64        `json:"update_time" gorm:"type:bigint(16) unsigned;comment:更新时间戳"`
	CompletedTime *uint64        `json:"completed_time" gorm:"type:bigint(16) unsigned;comment:完成时间戳"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// OrderCreateRequest 创建订单请求
type OrderCreateRequest struct {
	ProductID       uint   `json:"product_id" binding:"required"`        // 产品ID
	Quantity        int64  `json:"quantity" binding:"required,min=1"`    // 数量
	Remark          string `json:"remark" binding:"omitempty,max=500"`   // 备注
	ShippingAddress string `json:"shipping_address" binding:"omitempty"` // 配送地址(实体商品)
}

// OrderQuery 订单查询参数
type OrderQuery struct {
	OrderNumber   string        `form:"order_number"`   // 订单编号
	Keyword       string        `form:"keyword"`        // 关键词
	BuyerID       uint          `form:"buyer_id"`       // 买家ID
	SellerID      uint          `form:"seller_id"`      // 卖家ID
	ProductID     uint          `form:"product_id"`     // 产品ID
	Type          OrderType     `form:"type"`           // 订单类型
	Status        OrderStatus   `form:"status"`         // 订单状态
	PaymentStatus PaymentStatus `form:"payment_status"` // 支付状态
	StartTime     uint64        `form:"start_time"`     // 开始时间
	EndTime       uint64        `form:"end_time"`       // 结束时间
	MinAmount     float64       `form:"min_amount"`     // 最小金额
	MaxAmount     float64       `form:"max_amount"`     // 最大金额
	SortBy        string        `form:"sort_by"`        // 排序字段
	SortOrder     string        `form:"sort_order"`     // 排序方向
	Page          int           `form:"page"`           // 页码
	PageSize      int           `form:"page_size"`      // 每页数量
}

// OrderStatusUpdateRequest 订单状态更新请求
type OrderStatusUpdateRequest struct {
	Status       OrderStatus `json:"status" binding:"required"`                 // 新状态
	Remark       string      `json:"remark" binding:"omitempty,max=500"`        // 备注
	CancelReason string      `json:"cancel_reason" binding:"omitempty,max=255"` // 取消原因
}

// OrderPaymentRequest 订单支付请求
type OrderPaymentRequest struct {
	PaymentMethod string `json:"payment_method" binding:"required,max=50"` // 支付方式
}
