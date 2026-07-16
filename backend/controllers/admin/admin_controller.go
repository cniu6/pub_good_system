// Package admin 平行草稿栈管理侧控制器（现网请用 app/controllers/admin）。
//
// 【已注释禁用·留档】商品/分类/商城订单相关 handler 代码仍保留在本文件下方，
// 但 internal/ginweb 路由已注释、appinit 电商表迁移已注释，现网入口不会调用。
// 说明见 backend/留档.md →「电商半成品草稿（已注释禁用）」。
package admin

import (
	"fst/backend/models"
	"fst/backend/services"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// GetUserList 管理员获取用户列表
func GetUserList(c *gin.Context) {
	page := utils.ParseInt(c.DefaultQuery("page", "1"))
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	service := services.NewUserService()
	users, total, err := service.GetUserList(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.PageResponse(c, users, total, page, pageSize)
}

// GetUserDetail 管理员获取用户详情
func GetUserDetail(c *gin.Context) {
	userID := utils.ParseUint(c.Param("id"))
	if userID == 0 {
		utils.Fail(c, 400, "用户ID不能为空")
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

// UpdateUser 管理员更新用户（禁止直改余额/积分）
func UpdateUser(c *gin.Context) {
	userID := utils.ParseUint(c.Param("id"))
	if userID == 0 {
		utils.Fail(c, 400, "用户ID不能为空")
		return
	}
	var req models.UserUpdateRequestForAdmin
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	service := services.NewUserService()
	if err := service.AdminUpdateUser(userID, &req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.SuccessMsg(c, "更新成功", nil)
}

// ResetPassword 管理员重置密码
func ResetPassword(c *gin.Context) {
	userID := utils.ParseUint(c.Param("id"))
	if userID == 0 {
		utils.Fail(c, 400, "用户ID不能为空")
		return
	}
	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	service := services.NewUserService()
	if err := service.AdminResetPassword(userID, req.NewPassword); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessMsg(c, "密码重置成功", nil)
}

// GetProductList 管理员产品列表
// 【已注释禁用·留档】路由未挂载
func GetProductList(c *gin.Context) {
	var query models.ProductQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.Fail(c, 400, "请求参数错误")
		return
	}
	service := services.NewProductService()
	products, total, err := service.AdminGetList(&query)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.PageResponse(c, products, total, query.Page, query.PageSize)
}

// GetProductDetail 管理员产品详情
// 【已注释禁用·留档】路由未挂载
func GetProductDetail(c *gin.Context) {
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

// CreateProduct 管理员创建产品
// 【已注释禁用·留档】路由未挂载
func CreateProduct(c *gin.Context) {
	var req models.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	service := services.NewProductService()
	product, err := service.Create(&req, 0)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, product)
}

// UpdateProduct 管理员更新产品
// 【已注释禁用·留档】路由未挂载
func UpdateProduct(c *gin.Context) {
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
	if err := service.Update(productID, 0, &req); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessMsg(c, "更新成功", nil)
}

// DeleteProduct 管理员删除产品
// 【已注释禁用·留档】路由未挂载
func DeleteProduct(c *gin.Context) {
	productID := utils.ParseUint(c.Param("id"))
	if productID == 0 {
		utils.Fail(c, 400, "产品ID不能为空")
		return
	}
	service := services.NewProductService()
	if err := service.Delete(productID, 0); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessMsg(c, "删除成功", nil)
}

// GetCategoryList 分类列表
// 【已注释禁用·留档】路由未挂载
func GetCategoryList(c *gin.Context) {
	service := services.NewCategoryService()
	list, err := service.ListAll()
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, list)
}

// GetCategoryDetail 分类详情
// 【已注释禁用·留档】路由未挂载
func GetCategoryDetail(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	if id == 0 {
		utils.Fail(c, 400, "分类ID不能为空")
		return
	}
	service := services.NewCategoryService()
	item, err := service.GetByID(id, false)
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	utils.Success(c, item)
}

// CreateCategory 创建分类
// 【已注释禁用·留档】路由未挂载
func CreateCategory(c *gin.Context) {
	var req models.CategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	service := services.NewCategoryService()
	item, err := service.Create(&req)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, item)
}

// UpdateCategory 更新分类
// 【已注释禁用·留档】路由未挂载
func UpdateCategory(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	if id == 0 {
		utils.Fail(c, 400, "分类ID不能为空")
		return
	}
	var req models.CategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	service := services.NewCategoryService()
	if err := service.Update(id, &req); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessMsg(c, "更新成功", nil)
}

// DeleteCategory 删除分类
// 【已注释禁用·留档】路由未挂载
func DeleteCategory(c *gin.Context) {
	id := utils.ParseUint(c.Param("id"))
	if id == 0 {
		utils.Fail(c, 400, "分类ID不能为空")
		return
	}
	service := services.NewCategoryService()
	if err := service.Delete(id); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessMsg(c, "删除成功", nil)
}

// GetOrderList 商城订单列表（非充值单）
// 【已注释禁用·留档】路由未挂载
func GetOrderList(c *gin.Context) {
	var query models.OrderQuery
	_ = c.ShouldBindQuery(&query)
	service := services.NewOrderService()
	list, total, err := service.AdminList(&query)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.PageResponse(c, list, total, query.Page, query.PageSize)
}

// GetOrderDetail 商城订单详情
// 【已注释禁用·留档】路由未挂载
func GetOrderDetail(c *gin.Context) {
	orderID := utils.ParseUint(c.Param("id"))
	if orderID == 0 {
		utils.Fail(c, 400, "订单ID不能为空")
		return
	}
	service := services.NewOrderService()
	order, err := service.GetByID(orderID)
	if err != nil {
		utils.Fail(c, 404, err.Error())
		return
	}
	utils.Success(c, order)
}

// UpdateOrderStatus 更新商城订单状态
// 【已注释禁用·留档】路由未挂载
func UpdateOrderStatus(c *gin.Context) {
	orderID := utils.ParseUint(c.Param("id"))
	if orderID == 0 {
		utils.Fail(c, 400, "订单ID不能为空")
		return
	}
	var req models.OrderStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "请求参数错误: "+err.Error())
		return
	}
	service := services.NewOrderService()
	if err := service.AdminUpdateStatus(orderID, &req); err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.SuccessMsg(c, "状态已更新", nil)
}
