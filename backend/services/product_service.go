package services

// 【已注释禁用·留档】本文件为电商半成品商品服务。
// 路由(internal/ginweb)、迁移(appinit.AutoMigrate)、控制器挂载已注释，现网入口不使用。
// 现网「支付订单/充值」见 app/services/payment_service.go。说明见 backend/留档.md。

import (
	"encoding/json"
	"errors"
	"fst/backend/internal/db"
	"fst/backend/models"
	"fst/backend/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 商品列表排序字段白名单，防止 ORDER BY 注入
var productSortColumns = map[string]bool{
	"id": true, "price": true, "original_price": true,
	"stock": true, "sold_count": true, "sort_order": true,
	"create_time": true, "update_time": true,
}

// buildProductOrderBy 校验 sort_by 后生成 ORDER BY 子句
func buildProductOrderBy(sortBy, sortOrder string) string {
	col := strings.ToLower(strings.TrimSpace(sortBy))
	if !productSortColumns[col] {
		return "id DESC"
	}
	dir := "DESC"
	if strings.EqualFold(strings.TrimSpace(sortOrder), "asc") {
		dir = "ASC"
	}
	return col + " " + dir
}

// ProductService 产品服务
type ProductService struct{}

// NewProductService 创建产品服务实例
func NewProductService() *ProductService {
	return &ProductService{}
}

// Create 创建产品
// 参数: req-创建请求, sellerID-卖家ID(0表示平台)
// 返回: 产品对象、错误
func (s *ProductService) Create(req *models.ProductCreateRequest, sellerID uint) (*models.Product, error) {
	// 清理XSS
	utils.CleanXSSFields(&req.Name, &req.Description)

	// 验证分类是否存在
	if req.CategoryID > 0 {
		var category models.Category
		err := db.GetDB().Table("categories").Where("id", "=", req.CategoryID).First(&category)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("分类不存在")
			}
			return nil, errors.New("查询分类失败")
		}
	}

	// 获取当前时间戳
	now := uint64(time.Now().Unix())

	// 创建产品
	product := &models.Product{
		Name:           req.Name,
		Description:    req.Description,
		Type:           req.Type,
		CategoryID:     req.CategoryID,
		Price:          req.Price,
		OriginalPrice:  req.OriginalPrice,
		Currency:       req.Currency,
		Stock:          req.Stock,
		SellerID:       sellerID,
		Attributes:     req.Attributes,
		PluginID:       req.PluginID,
		Images:         req.Images,
		CoverImage:     req.CoverImage,
		SEOKeywords:    req.SEOKeywords,
		SEODescription: req.SEODescription,
		Metadata:       req.Metadata,
		Status:         models.ProductStatusActive,
		CreateTime:     &now,
		UpdateTime:     &now,
	}

	// 保存产品
	if err := db.GetDB().Create(product); err != nil {
		return nil, errors.New("产品创建失败: " + err.Error())
	}

	return product, nil
}

// GetByID 根据ID获取产品
// 参数: productID-产品ID
// 返回: 产品对象、错误
func (s *ProductService) GetByID(productID uint) (*models.Product, error) {
	if productID == 0 {
		return nil, errors.New("产品ID不能为空")
	}

	var product models.Product
	err := db.GetDB().Table("products").Where("id", "=", productID).First(&product)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("产品不存在")
		}
		return nil, errors.New("查询产品失败")
	}

	return &product, nil
}

// Update 更新产品
// 参数: productID-产品ID, sellerID-卖家ID(用于权限检查), req-更新请求
// 返回: 错误
func (s *ProductService) Update(productID, sellerID uint, req *models.ProductUpdateRequest) error {
	if productID == 0 {
		return errors.New("产品ID不能为空")
	}

	// 获取产品信息
	var product models.Product
	err := db.GetDB().Table("products").Where("id", "=", productID).First(&product)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("产品不存在")
		}
		return errors.New("查询产品失败")
	}

	// 检查权限(非管理员只能修改自己的产品)
	if sellerID > 0 && product.SellerID != sellerID {
		return errors.New("无权修改此产品")
	}

	// 清理XSS
	utils.CleanXSSFields(&req.Name, &req.Description)

	// 构建更新数据
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.CategoryID > 0 {
		updates["category_id"] = req.CategoryID
	}
	if req.Price > 0 {
		updates["price"] = req.Price
	}
	if req.OriginalPrice > 0 {
		updates["original_price"] = req.OriginalPrice
	}
	if req.Currency != "" {
		updates["currency"] = req.Currency
	}
	if req.Stock >= -1 {
		updates["stock"] = req.Stock
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Attributes != "" {
		updates["attributes"] = req.Attributes
	}
	if req.PluginID != "" {
		updates["plugin_id"] = req.PluginID
	}
	if req.Images != "" {
		updates["images"] = req.Images
	}
	if req.CoverImage != "" {
		updates["cover_image"] = req.CoverImage
	}
	if req.IsRecommend != nil {
		updates["is_recommend"] = *req.IsRecommend
	}
	if req.IsHot != nil {
		updates["is_hot"] = *req.IsHot
	}
	if req.SortOrder != 0 || req.SortOrder == 0 {
		updates["sort_order"] = req.SortOrder
	}
	if req.SEOKeywords != "" {
		updates["seo_keywords"] = req.SEOKeywords
	}
	if req.SEODescription != "" {
		updates["seo_description"] = req.SEODescription
	}
	if req.Metadata != "" {
		updates["metadata"] = req.Metadata
	}

	// 更新时间戳
	now := uint64(time.Now().Unix())
	updates["update_time"] = now

	// 执行更新
	if len(updates) > 0 {
		err = db.GetDB().Table("products").Where("id", "=", productID).Update(updates)
		if err != nil {
			return errors.New("产品更新失败")
		}
	}

	return nil
}

// Delete 删除产品(软删除)
// 参数: productID-产品ID, sellerID-卖家ID(用于权限检查)
// 返回: 错误
func (s *ProductService) Delete(productID, sellerID uint) error {
	if productID == 0 {
		return errors.New("产品ID不能为空")
	}

	// 获取产品信息
	var product models.Product
	err := db.GetDB().Table("products").Where("id", "=", productID).First(&product)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("产品不存在")
		}
		return errors.New("查询产品失败")
	}

	// 检查权限
	if sellerID > 0 && product.SellerID != sellerID {
		return errors.New("无权删除此产品")
	}

	// 软删除
	err = db.GetDB().Table("products").Where("id", "=", productID).Delete(&models.Product{})
	if err != nil {
		return errors.New("产品删除失败")
	}

	return nil
}

// GetList 获取产品列表
// 参数: query-查询参数
// 返回: 产品列表、总数、错误
func (s *ProductService) GetList(query *models.ProductQuery) ([]models.Product, int64, error) {
	dbInstance := db.GetDB().Table("products")

	// 只查询上架产品(前端展示)
	dbInstance = dbInstance.Where("status", "=", models.ProductStatusActive)

	// 关键词搜索
	if query.Keyword != "" {
		dbInstance = dbInstance.WhereLike("name", query.Keyword)
	}

	// 分类筛选
	if query.CategoryID > 0 {
		dbInstance = dbInstance.Where("category_id", "=", query.CategoryID)
	}

	// 类型筛选
	if query.Type != "" {
		dbInstance = dbInstance.Where("type", "=", query.Type)
	}

	// 卖家筛选(C2C)
	if query.SellerID > 0 {
		dbInstance = dbInstance.Where("seller_id", "=", query.SellerID)
	}

	// 价格范围
	if query.MinPrice > 0 {
		dbInstance = dbInstance.Where("price", ">=", query.MinPrice)
	}
	if query.MaxPrice > 0 {
		dbInstance = dbInstance.Where("price", "<=", query.MaxPrice)
	}

	// 推荐/热门筛选
	if query.IsRecommend != nil && *query.IsRecommend {
		dbInstance = dbInstance.Where("is_recommend", "=", true)
	}
	if query.IsHot != nil && *query.IsHot {
		dbInstance = dbInstance.Where("is_hot", "=", true)
	}

	// 统计总数
	total, err := dbInstance.Count()
	if err != nil {
		return nil, 0, errors.New("统计产品数量失败")
	}

	// 设置分页参数默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	orderBy := buildProductOrderBy(query.SortBy, query.SortOrder)
	dbInstance = dbInstance.OrderBy(orderBy)

	// 查询列表
	var products []models.Product
	err = dbInstance.Page(query.Page, query.PageSize).Find(&products)
	if err != nil {
		return nil, 0, errors.New("查询产品列表失败")
	}

	return products, total, nil
}

// AdminGetList 管理员获取产品列表(包含所有状态)
// 参数: query-查询参数
// 返回: 产品列表、总数、错误
func (s *ProductService) AdminGetList(query *models.ProductQuery) ([]models.Product, int64, error) {
	dbInstance := db.GetDB().Table("products")

	// 关键词搜索
	if query.Keyword != "" {
		dbInstance = dbInstance.WhereLike("name", query.Keyword)
	}

	// 分类筛选
	if query.CategoryID > 0 {
		dbInstance = dbInstance.Where("category_id", "=", query.CategoryID)
	}

	// 类型筛选
	if query.Type != "" {
		dbInstance = dbInstance.Where("type", "=", query.Type)
	}

	// 状态筛选
	if query.Status > 0 {
		dbInstance = dbInstance.Where("status", "=", query.Status)
	}

	// 卖家筛选
	if query.SellerID > 0 {
		dbInstance = dbInstance.Where("seller_id", "=", query.SellerID)
	}

	// 统计总数
	total, err := dbInstance.Count()
	if err != nil {
		return nil, 0, errors.New("统计产品数量失败")
	}

	// 设置分页参数
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	orderBy := buildProductOrderBy(query.SortBy, query.SortOrder)

	// 查询列表
	var products []models.Product
	err = dbInstance.Page(query.Page, query.PageSize).OrderBy(orderBy).Find(&products)
	if err != nil {
		return nil, 0, errors.New("查询产品列表失败")
	}

	return products, total, nil
}

// UpdateStock 更新库存
// 参数: productID-产品ID, quantity-变化数量(正数增加，负数减少)
// 返回: 错误
func (s *ProductService) UpdateStock(productID uint, quantity int64) error {
	if productID == 0 {
		return errors.New("产品ID不能为空")
	}

	// 获取产品信息
	var product models.Product
	err := db.GetDB().Table("products").Where("id", "=", productID).First(&product)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("产品不存在")
		}
		return errors.New("查询产品失败")
	}

	// 检查库存(无限库存不检查)
	if product.Stock >= 0 {
		newStock := product.Stock + quantity
		if newStock < 0 {
			return errors.New("库存不足")
		}
	}

	// 更新库存和销售量
	now := uint64(time.Now().Unix())
	updates := map[string]interface{}{
		"update_time": now,
	}

	if quantity > 0 {
		// 增加库存
		if product.Stock >= 0 {
			updates["stock"] = product.Stock + quantity
		}
	} else {
		// 减少库存，增加销量
		if product.Stock >= 0 {
			updates["stock"] = product.Stock + quantity
		}
		updates["sold_count"] = product.SoldCount + (-quantity)
	}

	err = db.GetDB().Table("products").Where("id", "=", productID).Update(updates)
	if err != nil {
		return errors.New("库存更新失败")
	}

	return nil
}

// GetAttributes 获取产品属性(解析JSON)
// 参数: productID-产品ID
// 返回: 属性映射、错误
func (s *ProductService) GetAttributes(productID uint) (map[string]interface{}, error) {
	product, err := s.GetByID(productID)
	if err != nil {
		return nil, err
	}

	var attributes map[string]interface{}
	if product.Attributes != "" {
		if err := json.Unmarshal([]byte(product.Attributes), &attributes); err != nil {
			return nil, errors.New("产品属性解析失败")
		}
	}

	return attributes, nil
}
