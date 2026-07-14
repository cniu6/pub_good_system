package models

import (
	"gorm.io/gorm"
)

// ProductType 产品类型
type ProductType string

const (
	ProductTypePhysical     ProductType = "physical"     // 实体商品
	ProductTypeDigital      ProductType = "digital"      // 数字商品
	ProductTypeService      ProductType = "service"      // 服务
	ProductTypeSubscription ProductType = "subscription" // 订阅服务
)

// ProductStatus 产品状态
type ProductStatus uint8

const (
	ProductStatusInactive ProductStatus = 0 // 下架
	ProductStatusActive   ProductStatus = 1 // 上架
	ProductStatusReview   ProductStatus = 2 // 审核中
)

// Product 通用产品模型
// 支持C2C/B2C/SaaS等多种模式，通过插件扩展功能
type Product struct {
	ID uint `json:"id" gorm:"primaryKey;type:bigint(20) unsigned;autoIncrement;comment:产品ID"`

	// 基础信息
	Name        string        `json:"name" gorm:"index;size:255;not null;comment:产品名称"`
	Description string        `json:"description" gorm:"type:text;comment:产品描述"`
	Type        ProductType   `json:"type" gorm:"index;size:50;not null;default:'physical';comment:产品类型:physical=实体,digital=数字,service=服务,subscription=订阅"`
	Status      ProductStatus `json:"status" gorm:"index;default:1;not null;comment:状态:0=下架,1=上架,2=审核中"`

	// 分类信息
	CategoryID uint      `json:"category_id" gorm:"index;default:0;not null;comment:分类ID"`
	Category   *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`

	// 价格信息
	Price         float64 `json:"price" gorm:"type:decimal(12,2);index;default:0.00;not null;comment:售价"`
	OriginalPrice float64 `json:"original_price" gorm:"type:decimal(12,2);default:0.00;not null;comment:原价"`
	Currency      string  `json:"currency" gorm:"size:10;not null;default:'CNY';comment:货币代码"`

	// 库存信息
	Stock     int64 `json:"stock" gorm:"default:0;not null;comment:库存数量(-1=无限)"`
	SoldCount int64 `json:"sold_count" gorm:"default:0;not null;comment:已售数量"`

	// 卖家信息(C2C模式)
	SellerID uint  `json:"seller_id" gorm:"index;default:0;not null;comment:卖家ID(0=平台自营B2C)"`
	Seller   *User `json:"seller,omitempty" gorm:"foreignKey:SellerID"`

	// 扩展属性(JSON存储，灵活定义)
	Attributes string `json:"attributes" gorm:"type:json;comment:产品属性(颜色、尺寸等)"`

	// 插件集成
	PluginID   string `json:"plugin_id" gorm:"size:100;default:'';comment:关联插件ID(用于扩展功能)"`
	PluginData string `json:"plugin_data" gorm:"type:json;comment:插件数据"`

	// 多媒体
	Images     string `json:"images" gorm:"type:json;comment:产品图片列表(JSON数组)"`
	CoverImage string `json:"cover_image" gorm:"size:255;default:'';comment:封面图"`

	// 营销信息
	IsRecommend bool `json:"is_recommend" gorm:"index;default:false;not null;comment:是否推荐"`
	IsHot       bool `json:"is_hot" gorm:"index;default:false;not null;comment:是否热门"`
	SortOrder   int  `json:"sort_order" gorm:"index;default:0;not null;comment:排序权重"`

	// SEO信息
	SEOKeywords    string `json:"seo_keywords" gorm:"size:255;default:'';comment:SEO关键词"`
	SEODescription string `json:"seo_description" gorm:"size:500;default:'';comment:SEO描述"`

	// 元数据
	Metadata string `json:"metadata" gorm:"type:json;comment:扩展元数据"`

	CreateTime *uint64        `json:"create_time" gorm:"type:bigint(16) unsigned;comment:创建时间戳"`
	UpdateTime *uint64        `json:"update_time" gorm:"type:bigint(16) unsigned;comment:更新时间戳"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}

// ProductCreateRequest 创建产品请求
type ProductCreateRequest struct {
	Name           string      `json:"name" binding:"required,max=255"`                                     // 产品名称
	Description    string      `json:"description" binding:"omitempty"`                                     // 产品描述
	Type           ProductType `json:"type" binding:"required,oneof=physical digital service subscription"` // 产品类型
	CategoryID     uint        `json:"category_id" binding:"omitempty"`                                     // 分类ID
	Price          float64     `json:"price" binding:"required,gt=0"`                                       // 售价
	OriginalPrice  float64     `json:"original_price" binding:"omitempty,gt=0"`                             // 原价
	Currency       string      `json:"currency" binding:"omitempty,max=10"`                                 // 货币
	Stock          int64       `json:"stock" binding:"omitempty,gte=-1"`                                    // 库存(-1=无限)
	Attributes     string      `json:"attributes" binding:"omitempty"`                                      // 属性JSON
	PluginID       string      `json:"plugin_id" binding:"omitempty,max=100"`                               // 插件ID
	Images         string      `json:"images" binding:"omitempty"`                                          // 图片列表JSON
	CoverImage     string      `json:"cover_image" binding:"omitempty,max=255"`                             // 封面图
	SEOKeywords    string      `json:"seo_keywords" binding:"omitempty,max=255"`                            // SEO关键词
	SEODescription string      `json:"seo_description" binding:"omitempty,max=500"`                         // SEO描述
	Metadata       string      `json:"metadata" binding:"omitempty"`                                        // 元数据JSON
}

// ProductUpdateRequest 更新产品请求
type ProductUpdateRequest struct {
	Name           string         `json:"name" binding:"omitempty,max=255"`                                     // 产品名称
	Description    string         `json:"description" binding:"omitempty"`                                      // 产品描述
	Type           ProductType    `json:"type" binding:"omitempty,oneof=physical digital service subscription"` // 产品类型
	CategoryID     uint           `json:"category_id" binding:"omitempty"`                                      // 分类ID
	Price          float64        `json:"price" binding:"omitempty,gt=0"`                                       // 售价
	OriginalPrice  float64        `json:"original_price" binding:"omitempty,gt=0"`                              // 原价
	Currency       string         `json:"currency" binding:"omitempty,max=10"`                                  // 货币
	Stock          int64          `json:"stock" binding:"omitempty,gte=-1"`                                     // 库存
	Status         *ProductStatus `json:"status" binding:"omitempty"`                                           // 状态
	Attributes     string         `json:"attributes" binding:"omitempty"`                                       // 属性JSON
	PluginID       string         `json:"plugin_id" binding:"omitempty,max=100"`                                // 插件ID
	Images         string         `json:"images" binding:"omitempty"`                                           // 图片列表JSON
	CoverImage     string         `json:"cover_image" binding:"omitempty,max=255"`                              // 封面图
	IsRecommend    *bool          `json:"is_recommend" binding:"omitempty"`                                     // 是否推荐
	IsHot          *bool          `json:"is_hot" binding:"omitempty"`                                           // 是否热门
	SortOrder      int            `json:"sort_order" binding:"omitempty"`                                       // 排序
	SEOKeywords    string         `json:"seo_keywords" binding:"omitempty,max=255"`                             // SEO关键词
	SEODescription string         `json:"seo_description" binding:"omitempty,max=500"`                          // SEO描述
	Metadata       string         `json:"metadata" binding:"omitempty"`                                         // 元数据JSON
}

// ProductQuery 产品查询参数
type ProductQuery struct {
	Keyword     string      `form:"keyword"`      // 关键词搜索
	CategoryID  uint        `form:"category_id"`  // 分类筛选
	Type        ProductType `form:"type"`         // 类型筛选
	Status      uint8       `form:"status"`       // 状态筛选
	SellerID    uint        `form:"seller_id"`    // 卖家筛选
	MinPrice    float64     `form:"min_price"`    // 最低价格
	MaxPrice    float64     `form:"max_price"`    // 最高价格
	IsRecommend *bool       `form:"is_recommend"` // 是否推荐
	IsHot       *bool       `form:"is_hot"`       // 是否热门
	SortBy      string      `form:"sort_by"`      // 排序字段
	SortOrder   string      `form:"sort_order"`   // 排序方向(asc/desc)
	Page        int         `form:"page"`         // 页码
	PageSize    int         `form:"page_size"`    // 每页数量
}
