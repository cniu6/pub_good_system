package services

import (
	"errors"
	"fst/backend/internal/db"
	"fst/backend/models"
	"fst/backend/utils"
	"time"

	"gorm.io/gorm"
)

// OrderService 商品订单服务（平行栈最小实现）
type OrderService struct{}

// NewOrderService 创建订单服务
func NewOrderService() *OrderService {
	return &OrderService{}
}

func normalizeOrderPage(q *models.OrderQuery) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
}

// Create 买家下单：校验商品与库存，写订单并扣库存
func (s *OrderService) Create(buyerID uint, req *models.OrderCreateRequest) (*models.Order, error) {
	if buyerID == 0 {
		return nil, errors.New("未登录")
	}
	if req.ProductID == 0 || req.Quantity < 1 {
		return nil, errors.New("商品或数量无效")
	}
	utils.CleanXSSFields(&req.Remark, &req.ShippingAddress)

	var product models.Product
	if err := db.GetDB().Table("products").Where("id", "=", req.ProductID).First(&product); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商品不存在")
		}
		return nil, errors.New("查询商品失败")
	}
	if product.Status != models.ProductStatusActive {
		return nil, errors.New("商品未上架")
	}
	if product.Stock >= 0 && product.Stock < req.Quantity {
		return nil, errors.New("库存不足")
	}

	now := uint64(time.Now().Unix())
	subtotal := product.Price * float64(req.Quantity)
	order := &models.Order{
		OrderNumber:     utils.GenerateOrderNumber(),
		BuyerID:         buyerID,
		SellerID:        product.SellerID,
		Type:            models.OrderTypeB2C,
		ProductID:       product.ID,
		ProductName:     product.Name,
		Quantity:        req.Quantity,
		UnitPrice:       product.Price,
		Subtotal:        subtotal,
		TotalAmount:     subtotal,
		Currency:        product.Currency,
		Status:          models.OrderStatusPending,
		PaymentStatus:   models.PaymentStatusUnpaid,
		UserRemark:      req.Remark,
		ShippingAddress: req.ShippingAddress,
		CreateTime:      &now,
		UpdateTime:      &now,
	}
	if product.SellerID > 0 {
		order.Type = models.OrderTypeC2C
	}

	// 先扣库存（失败则不下单）
	ps := NewProductService()
	if err := ps.UpdateStock(product.ID, -req.Quantity); err != nil {
		return nil, err
	}
	if err := db.GetDB().Create(order); err != nil {
		// 回滚库存
		_ = ps.UpdateStock(product.ID, req.Quantity)
		return nil, errors.New("创建订单失败: " + err.Error())
	}
	return order, nil
}

// ListByBuyer 买家订单列表
func (s *OrderService) ListByBuyer(buyerID uint, q *models.OrderQuery) ([]models.Order, int64, error) {
	normalizeOrderPage(q)
	inst := db.GetDB().Table("orders").Where("buyer_id", "=", buyerID)
	if q.Status != "" {
		inst = inst.Where("status", "=", q.Status)
	}
	total, err := inst.Count()
	if err != nil {
		return nil, 0, errors.New("统计订单失败")
	}
	var list []models.Order
	if err := inst.Page(q.Page, q.PageSize).OrderBy("id DESC").Find(&list); err != nil {
		return nil, 0, errors.New("查询订单失败")
	}
	return list, total, nil
}

// GetByID 按 ID 查订单
func (s *OrderService) GetByID(id uint) (*models.Order, error) {
	if id == 0 {
		return nil, errors.New("订单ID不能为空")
	}
	var order models.Order
	if err := db.GetDB().Table("orders").Where("id", "=", id).First(&order); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		return nil, errors.New("查询订单失败")
	}
	return &order, nil
}

// GetByIDForBuyer 买家查看自己的订单
func (s *OrderService) GetByIDForBuyer(id, buyerID uint) (*models.Order, error) {
	order, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if order.BuyerID != buyerID {
		return nil, errors.New("订单不存在")
	}
	return order, nil
}

// CancelByBuyer 买家取消待付款订单并回库存
func (s *OrderService) CancelByBuyer(id, buyerID uint, reason string) error {
	order, err := s.GetByIDForBuyer(id, buyerID)
	if err != nil {
		return err
	}
	if order.Status != models.OrderStatusPending {
		return errors.New("仅待付款订单可取消")
	}
	utils.CleanXSSFields(&reason)
	now := uint64(time.Now().Unix())
	if err := db.GetDB().Table("orders").Where("id", "=", id).Update(map[string]interface{}{
		"status":        models.OrderStatusCancelled,
		"cancel_reason": reason,
		"update_time":   now,
	}); err != nil {
		return errors.New("取消订单失败")
	}
	_ = NewProductService().UpdateStock(order.ProductID, order.Quantity)
	return nil
}

// AdminList 管理端订单列表
func (s *OrderService) AdminList(q *models.OrderQuery) ([]models.Order, int64, error) {
	normalizeOrderPage(q)
	inst := db.GetDB().Table("orders")
	if q.BuyerID > 0 {
		inst = inst.Where("buyer_id", "=", q.BuyerID)
	}
	if q.SellerID > 0 {
		inst = inst.Where("seller_id", "=", q.SellerID)
	}
	if q.Status != "" {
		inst = inst.Where("status", "=", q.Status)
	}
	if q.OrderNumber != "" {
		inst = inst.Where("order_number", "=", q.OrderNumber)
	}
	total, err := inst.Count()
	if err != nil {
		return nil, 0, errors.New("统计订单失败")
	}
	var list []models.Order
	if err := inst.Page(q.Page, q.PageSize).OrderBy("id DESC").Find(&list); err != nil {
		return nil, 0, errors.New("查询订单失败")
	}
	return list, total, nil
}

// AdminUpdateStatus 管理端改订单状态
func (s *OrderService) AdminUpdateStatus(id uint, req *models.OrderStatusUpdateRequest) error {
	if _, err := s.GetByID(id); err != nil {
		return err
	}
	utils.CleanXSSFields(&req.Remark, &req.CancelReason)
	now := uint64(time.Now().Unix())
	updates := map[string]interface{}{
		"status":      req.Status,
		"update_time": now,
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.CancelReason != "" {
		updates["cancel_reason"] = req.CancelReason
	}
	if req.Status == models.OrderStatusCompleted {
		updates["completed_time"] = now
	}
	if err := db.GetDB().Table("orders").Where("id", "=", id).Update(updates); err != nil {
		return errors.New("更新订单状态失败")
	}
	return nil
}
