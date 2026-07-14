package admin

import (
	"fst/models"
	"fst/services"
	"fst/utils"

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
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.PageResponse(c, users, total, page, pageSize)
}

// GetUserDetail 管理员获取用户详情
func GetUserDetail(c *gin.Context) {
	userID := utils.ParseUint(c.Param("id"))
	if userID == 0 {
		utils.Error(c, utils.CodeBadRequest, "用户ID不能为空")
		return
	}

	service := services.NewUserService()
	user, err := service.GetUserByID(userID)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, user)
}

// UpdateUser 管理员更新用户
func UpdateUser(c *gin.Context) {
	userID := utils.ParseUint(c.Param("id"))
	if userID == 0 {
		utils.Error(c, utils.CodeBadRequest, "用户ID不能为空")
		return
	}

	var req models.UserUpdateRequestForAdmin
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	service := services.NewUserService()
	if err := service.AdminUpdateUser(userID, &req); err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// ResetPassword 管理员重置密码
func ResetPassword(c *gin.Context) {
	userID := utils.ParseUint(c.Param("id"))
	if userID == 0 {
		utils.Error(c, utils.CodeBadRequest, "用户ID不能为空")
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	service := services.NewUserService()
	if err := service.AdminResetPassword(userID, req.NewPassword); err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "密码重置成功", nil)
}

// GetProductList 管理员获取产品列表
func GetProductList(c *gin.Context) {
	var query models.ProductQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.Error(c, utils.CodeBadRequest, "请求参数错误")
		return
	}

	service := services.NewProductService()
	products, total, err := service.AdminGetList(&query)
	if err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.PageResponse(c, products, total, query.Page, query.PageSize)
}

// GetProductDetail 管理员获取产品详情
func GetProductDetail(c *gin.Context) {
	productID := utils.ParseUint(c.Param("id"))
	if productID == 0 {
		utils.Error(c, utils.CodeBadRequest, "产品ID不能为空")
		return
	}

	service := services.NewProductService()
	product, err := service.GetByID(productID)
	if err != nil {
		utils.Error(c, utils.CodeNotFound, err.Error())
		return
	}

	utils.Success(c, product)
}

// CreateProduct 管理员创建产品
func CreateProduct(c *gin.Context) {
	var req models.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	service := services.NewProductService()
	product, err := service.Create(&req, 0) // 0表示平台创建
	if err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.Success(c, product)
}

// UpdateProduct 管理员更新产品
func UpdateProduct(c *gin.Context) {
	productID := utils.ParseUint(c.Param("id"))
	if productID == 0 {
		utils.Error(c, utils.CodeBadRequest, "产品ID不能为空")
		return
	}

	var req models.ProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	service := services.NewProductService()
	if err := service.Update(productID, 0, &req); err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteProduct 管理员删除产品
func DeleteProduct(c *gin.Context) {
	productID := utils.ParseUint(c.Param("id"))
	if productID == 0 {
		utils.Error(c, utils.CodeBadRequest, "产品ID不能为空")
		return
	}

	service := services.NewProductService()
	if err := service.Delete(productID, 0); err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}
