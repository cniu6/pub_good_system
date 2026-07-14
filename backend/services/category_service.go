package services

import (
	"errors"
	"fst/backend/internal/db"
	"fst/backend/models"
	"fst/backend/utils"
	"time"

	"gorm.io/gorm"
)

// CategoryService 商品分类服务
type CategoryService struct{}

// NewCategoryService 创建分类服务
func NewCategoryService() *CategoryService {
	return &CategoryService{}
}

// ListPublic 公开：仅启用分类
func (s *CategoryService) ListPublic() ([]models.Category, error) {
	var list []models.Category
	err := db.GetDB().Table("categories").
		Where("status", "=", models.CategoryStatusActive).
		OrderBy("sort_order ASC, id ASC").
		Find(&list)
	if err != nil {
		return nil, errors.New("查询分类失败")
	}
	return list, nil
}

// ListAll 管理端：全部分类
func (s *CategoryService) ListAll() ([]models.Category, error) {
	var list []models.Category
	err := db.GetDB().Table("categories").OrderBy("sort_order ASC, id ASC").Find(&list)
	if err != nil {
		return nil, errors.New("查询分类失败")
	}
	return list, nil
}

// GetByID 分类详情；publicOnly 为 true 时要求启用
func (s *CategoryService) GetByID(id uint, publicOnly bool) (*models.Category, error) {
	if id == 0 {
		return nil, errors.New("分类ID不能为空")
	}
	var item models.Category
	q := db.GetDB().Table("categories").Where("id", "=", id)
	if publicOnly {
		q = q.Where("status", "=", models.CategoryStatusActive)
	}
	if err := q.First(&item); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分类不存在")
		}
		return nil, errors.New("查询分类失败")
	}
	return &item, nil
}

// Create 创建分类
func (s *CategoryService) Create(req *models.CategoryCreateRequest) (*models.Category, error) {
	utils.CleanXSSFields(&req.Name, &req.Slug, &req.Description, &req.Icon, &req.Image)

	status := models.CategoryStatusActive
	if req.Status == uint8(models.CategoryStatusInactive) {
		status = models.CategoryStatusInactive
	}
	now := uint64(time.Now().Unix())
	item := &models.Category{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Icon:        req.Icon,
		Image:       req.Image,
		SortOrder:   req.SortOrder,
		Status:      status,
		Level:       1,
		CreateTime:  &now,
		UpdateTime:  &now,
	}
	if err := db.GetDB().Create(item); err != nil {
		return nil, errors.New("创建分类失败: " + err.Error())
	}
	return item, nil
}

// Update 更新分类
func (s *CategoryService) Update(id uint, req *models.CategoryUpdateRequest) error {
	if id == 0 {
		return errors.New("分类ID不能为空")
	}
	if _, err := s.GetByID(id, false); err != nil {
		return err
	}
	utils.CleanXSSFields(&req.Name, &req.Slug, &req.Description, &req.Icon, &req.Image)

	updates := map[string]interface{}{}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Slug != "" {
		updates["slug"] = req.Slug
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.Image != "" {
		updates["image"] = req.Image
	}
	if req.SortOrder != 0 {
		updates["sort_order"] = req.SortOrder
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	now := uint64(time.Now().Unix())
	updates["update_time"] = now

	if len(updates) == 0 {
		return nil
	}
	if err := db.GetDB().Table("categories").Where("id", "=", id).Update(updates); err != nil {
		return errors.New("更新分类失败")
	}
	return nil
}

// Delete 软删除分类
func (s *CategoryService) Delete(id uint) error {
	if id == 0 {
		return errors.New("分类ID不能为空")
	}
	if err := db.GetDB().Table("categories").Where("id", "=", id).Delete(&models.Category{}); err != nil {
		return errors.New("删除分类失败")
	}
	return nil
}
