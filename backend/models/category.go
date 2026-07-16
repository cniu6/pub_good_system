package models

// 【已注释禁用·留档】电商半成品 Category 模型。AutoMigrate 与路由已注释，现网不建表/不挂接口。
// 说明见 backend/留档.md。

import "gorm.io/gorm"

// CategoryStatus 分类状态
type CategoryStatus uint8

const (
	CategoryStatusInactive CategoryStatus = 0 // 禁用
	CategoryStatusActive   CategoryStatus = 1 // 启用
)

// Category 产品分类模型
// 支持多级分类
type Category struct {
	ID          uint           `json:"id" gorm:"primaryKey;type:bigint(20) unsigned;autoIncrement;comment:分类ID"`
	ParentID    uint           `json:"parent_id" gorm:"index;default:0;not null;comment:父分类ID(0=顶级分类)"`
	Name        string         `json:"name" gorm:"index;size:100;not null;comment:分类名称"`
	Slug        string         `json:"slug" gorm:"uniqueIndex;size:100;not null;comment:分类别名(URL用)"`
	Description string         `json:"description" gorm:"size:255;default:'';comment:分类描述"`
	Icon        string         `json:"icon" gorm:"size:100;default:'';comment:分类图标"`
	Image       string         `json:"image" gorm:"size:255;default:'';comment:分类图片"`
	SortOrder   int            `json:"sort_order" gorm:"index;default:0;not null;comment:排序权重"`
	Status      CategoryStatus `json:"status" gorm:"index;default:1;not null;comment:状态:0=禁用,1=启用"`
	Level       uint8          `json:"level" gorm:"default:1;not null;comment:分类层级"`
	Path        string         `json:"path" gorm:"size:255;default:'';comment:分类路径(如:1,2,3)"`

	CreateTime *uint64        `json:"create_time" gorm:"type:bigint(16) unsigned;comment:创建时间戳"`
	UpdateTime *uint64        `json:"update_time" gorm:"type:bigint(16) unsigned;comment:更新时间戳"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}

// CategoryCreateRequest 创建分类请求
type CategoryCreateRequest struct {
	ParentID    uint   `json:"parent_id" binding:"omitempty"`           // 父分类ID
	Name        string `json:"name" binding:"required,max=100"`         // 分类名称
	Slug        string `json:"slug" binding:"required,max=100"`         // 分类别名
	Description string `json:"description" binding:"omitempty,max=255"` // 分类描述
	Icon        string `json:"icon" binding:"omitempty,max=100"`        // 图标
	Image       string `json:"image" binding:"omitempty,max=255"`       // 图片
	SortOrder   int    `json:"sort_order" binding:"omitempty"`          // 排序
	Status      uint8  `json:"status" binding:"omitempty,min=0,max=1"`  // 状态
}

// CategoryUpdateRequest 更新分类请求
type CategoryUpdateRequest struct {
	ParentID    *uint  `json:"parent_id" binding:"omitempty"`           // 父分类ID
	Name        string `json:"name" binding:"omitempty,max=100"`        // 分类名称
	Slug        string `json:"slug" binding:"omitempty,max=100"`        // 分类别名
	Description string `json:"description" binding:"omitempty,max=255"` // 分类描述
	Icon        string `json:"icon" binding:"omitempty,max=100"`        // 图标
	Image       string `json:"image" binding:"omitempty,max=255"`       // 图片
	SortOrder   int    `json:"sort_order" binding:"omitempty"`          // 排序
	Status      *uint8 `json:"status" binding:"omitempty,min=0,max=1"`  // 状态
}
