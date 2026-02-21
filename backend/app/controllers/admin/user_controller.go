package admin

import (
	"fst/backend/app/services"
	"fst/backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserController 用户管理控制器
type UserController struct {
	userService *services.UserService
}

func NewUserController() *UserController {
	return &UserController{
		userService: services.NewUserService(),
	}
}

// List 用户列表
// @Summary 获取用户列表
// @Tags Admin-User
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param keyword query string false "搜索关键词"
// @Param status query int false "状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/users [get]
func (c *UserController) List(ctx *gin.Context) {
	utils.SanitizeQueryParams(ctx)
	var query services.UserListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		utils.Fail(ctx, 400, "参数错误: "+err.Error())
		return
	}

	result, err := c.userService.GetList(&query)
	if err != nil {
		utils.Fail(ctx, 500, "查询失败: "+err.Error())
		return
	}

	utils.Success(ctx, result)
}

// Detail 用户详情
// @Summary 获取用户详情
// @Tags Admin-User
// @Param id path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/users/{id} [get]
func (c *UserController) Detail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(ctx, 400, "无效的用户ID")
		return
	}

	user, err := c.userService.GetByID(id)
	if err != nil {
		utils.Fail(ctx, 404, "用户不存在")
		return
	}

	utils.Success(ctx, gin.H{
		"user": user,
	})
}

// Create 创建用户
// @Summary 创建用户
// @Tags Admin-User
// @Param body body services.UserCreateRequest true "用户信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/users [post]
func (c *UserController) Create(ctx *gin.Context) {
	var req services.UserCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, 400, "参数错误: "+err.Error())
		return
	}

	// 加密密码
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Fail(ctx, 500, "密码加密失败")
		return
	}
	req.Password = hashed

	user, err := c.userService.Create(&req)
	if err != nil {
		utils.Fail(ctx, 400, err.Error())
		return
	}

	utils.Success(ctx, user)
}

// Update 更新用户
// @Summary 更新用户
// @Tags Admin-User
// @Param id path int true "用户ID"
// @Param body body services.UserUpdateRequest true "用户信息"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/users/{id} [put]
func (c *UserController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(ctx, 400, "无效的用户ID")
		return
	}

	var req services.UserUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, 400, "参数错误: "+err.Error())
		return
	}
	req.ID = id

	if err := c.userService.Update(&req); err != nil {
		utils.Fail(ctx, 400, err.Error())
		return
	}

	utils.Success(ctx, nil)
}

// Delete 删除用户
// @Summary 删除用户
// @Tags Admin-User
// @Param id path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/users/{id} [delete]
func (c *UserController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(ctx, 400, "无效的用户ID")
		return
	}

	if err := c.userService.Delete(id); err != nil {
		utils.Fail(ctx, 400, err.Error())
		return
	}

	utils.Success(ctx, nil)
}

// UpdateStatus 更新用户状态
// @Summary 更新用户状态
// @Tags Admin-User
// @Param id path int true "用户ID"
// @Param body body map[string]int true "状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/users/{id}/status [put]
func (c *UserController) UpdateStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(ctx, 400, "无效的用户ID")
		return
	}

	var req struct {
		Status uint8 `json:"status" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, 400, "参数错误")
		return
	}

	if err := c.userService.UpdateStatus(id, req.Status); err != nil {
		utils.Fail(ctx, 400, err.Error())
		return
	}

	utils.Success(ctx, nil)
}

// ResetPassword 重置用户密码
// @Summary 重置用户密码
// @Tags Admin-User
// @Param id path int true "用户ID"
// @Param body body map[string]string true "新密码"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/users/{id}/password [put]
func (c *UserController) ResetPassword(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(ctx, 400, "无效的用户ID")
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, 400, "参数错误: 密码至少6位")
		return
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Fail(ctx, 500, "密码加密失败")
		return
	}

	if err := c.userService.UpdatePassword(id, hashed); err != nil {
		utils.Fail(ctx, 400, err.Error())
		return
	}

	utils.Success(ctx, nil)
}
