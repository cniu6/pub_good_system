package user

import (
	"fst/backend/models"
	"fst/backend/services"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// Register 用户注册
func Register(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	req.IP = c.ClientIP()

	service := services.NewUserService()
	user, token, err := service.Register(&req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{"user": user, "token": token})
}

// Login 用户登录
func Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	req.IP = c.ClientIP()

	service := services.NewUserService()
	user, token, err := service.Login(&req)
	if err != nil {
		utils.Fail(c, 401, err.Error())
		return
	}
	utils.Success(c, gin.H{"user": user, "token": token})
}

// GetProfile 获取当前用户信息
func GetProfile(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Fail(c, 401, "未登录")
		return
	}
	service := services.NewUserService()
	user, err := service.GetUserByID(userID)
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	utils.Success(c, user)
}

// UpdateProfile 更新当前用户信息
func UpdateProfile(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Fail(c, 401, "未登录")
		return
	}
	var req models.UserUpdateRequestForUser
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	service := services.NewUserService()
	if err := service.UpdateUser(userID, &req); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessMsg(c, "更新成功", nil)
}

// ChangePassword 修改密码
func ChangePassword(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Fail(c, 401, "未登录")
		return
	}
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	service := services.NewUserService()
	if err := service.ChangePassword(userID, &req); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessMsg(c, "密码修改成功", nil)
}

// GetList 公开产品列表
func GetList(c *gin.Context) {
	var query models.ProductQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.Fail(c, 400, "请求参数错误")
		return
	}
	service := services.NewProductService()
	products, total, err := service.GetList(&query)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.PageResponse(c, products, total, query.Page, query.PageSize)
}

// GetDetail 公开产品详情
func GetDetail(c *gin.Context) {
	productID := utils.ParseUint(c.Param("id"))
	if productID == 0 {
		utils.Fail(c, 400, "产品ID不能为空")
		return
	}
	service := services.NewProductService()
	product, err := service.GetByID(productID)
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	utils.Success(c, product)
}

// CreateProduct 卖家创建产品
func CreateProduct(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Fail(c, 401, "未登录")
		return
	}
	var req models.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	service := services.NewProductService()
	product, err := service.Create(&req, userID)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, product)
}

// UpdateProduct 卖家更新产品
func UpdateProduct(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Fail(c, 401, "未登录")
		return
	}
	productID := utils.ParseUint(c.Param("id"))
	if productID == 0 {
		utils.Fail(c, 400, "产品ID不能为空")
		return
	}
	var req models.ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	service := services.NewProductService()
	if err := service.Update(productID, userID, &req); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessMsg(c, "更新成功", nil)
}

// DeleteProduct 卖家删除产品
func DeleteProduct(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Fail(c, 401, "未登录")
		return
	}
	productID := utils.ParseUint(c.Param("id"))
	if productID == 0 {
		utils.Fail(c, 400, "产品ID不能为空")
		return
	}
	service := services.NewProductService()
	if err := service.Delete(productID, userID); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessMsg(c, "删除成功", nil)
}

// GetCategoryList 公开分类列表
func GetCategoryList(c *gin.Context) {
	service := services.NewCategoryService()
	list, err := service.ListPublic()
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, list)
}

// GetCategoryDetail 公开分类详情
func GetCategoryDetail(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	if id == 0 {
		utils.Fail(c, 400, "分类ID不能为空")
		return
	}
	service := services.NewCategoryService()
	item, err := service.GetByID(id, true)
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	utils.Success(c, item)
}

// CreateOrder 创建订单
func CreateOrder(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Fail(c, 401, "未登录")
		return
	}
	var req models.OrderCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	service := services.NewOrderService()
	order, err := service.Create(userID, &req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, order)
}

// GetMyOrders 我的订单列表
func GetMyOrders(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Fail(c, 401, "未登录")
		return
	}
	var query models.OrderQuery
	_ = c.ShouldBindQuery(&query)
	service := services.NewOrderService()
	list, total, err := service.ListByBuyer(userID, &query)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.PageResponse(c, list, total, query.Page, query.PageSize)
}

// GetOrderDetail 订单详情（买家）
func GetOrderDetail(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Fail(c, 401, "未登录")
		return
	}
	orderID := utils.ParseUint(c.Param("id"))
	if orderID == 0 {
		utils.Fail(c, 400, "订单ID不能为空")
		return
	}
	service := services.NewOrderService()
	order, err := service.GetByIDForBuyer(orderID, userID)
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	utils.Success(c, order)
}

// CancelOrder 取消订单
func CancelOrder(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Fail(c, 401, "未登录")
		return
	}
	orderID := utils.ParseUint(c.Param("id"))
	if orderID == 0 {
		utils.Fail(c, 400, "订单ID不能为空")
		return
	}
	var req models.OrderStatusUpdateRequest
	_ = c.ShouldBindJSON(&req)
	service := services.NewOrderService()
	if err := service.CancelByBuyer(orderID, userID, req.CancelReason); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessMsg(c, "订单已取消", nil)
}
