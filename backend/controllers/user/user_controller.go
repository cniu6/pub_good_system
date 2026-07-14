package user

import (
	"fst/models"
	"fst/services"
	"fst/utils"

	"github.com/gin-gonic/gin"
)

// Register 用户注册
func Register(c *gin.Context) {
	var req models.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 获取客户端IP
	req.IP = c.ClientIP()

	// 调用服务层
	service := services.NewUserService()
	user, token, err := service.Register(&req)
	if err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	// 返回数据
	utils.Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}

// Login 用户登录
func Login(c *gin.Context) {
	var req models.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	// 获取客户端IP
	req.IP = c.ClientIP()

	// 调用服务层
	service := services.NewUserService()
	user, token, err := service.Login(&req)
	if err != nil {
		utils.Error(c, utils.CodeUnauthorized, err.Error())
		return
	}

	// 返回数据
	utils.Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}

// GetProfile 获取当前用户信息
func GetProfile(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
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

// UpdateProfile 更新当前用户信息
func UpdateProfile(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}

	var req models.UserUpdateRequestForUser
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	service := services.NewUserService()
	if err := service.UpdateUser(userID, &req); err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// ChangePassword 修改密码
func ChangePassword(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	service := services.NewUserService()
	if err := service.ChangePassword(userID, &req); err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "密码修改成功", nil)
}

// GetList 获取产品列表(公开)
func GetList(c *gin.Context) {
	var query models.ProductQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.Error(c, utils.CodeBadRequest, "请求参数错误")
		return
	}

	service := services.NewProductService()
	products, total, err := service.GetList(&query)
	if err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.PageResponse(c, products, total, query.Page, query.PageSize)
}

// GetDetail 获取产品详情(公开)
func GetDetail(c *gin.Context) {
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

// CreateProduct 创建产品(卖家)
func CreateProduct(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}

	var req models.ProductCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.CodeBadRequest, "请求参数错误: "+err.Error())
		return
	}

	service := services.NewProductService()
	product, err := service.Create(&req, userID)
	if err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.Success(c, product)
}

// UpdateProduct 更新产品(卖家)
func UpdateProduct(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}

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
	if err := service.Update(productID, userID, &req); err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteProduct 删除产品(卖家)
func DeleteProduct(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}

	productID := utils.ParseUint(c.Param("id"))
	if productID == 0 {
		utils.Error(c, utils.CodeBadRequest, "产品ID不能为空")
		return
	}

	service := services.NewProductService()
	if err := service.Delete(productID, userID); err != nil {
		utils.Error(c, utils.CodeServerError, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "删除成功", nil)
}
