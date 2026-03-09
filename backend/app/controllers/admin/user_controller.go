package admin

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/config"
	"fst/backend/utils"
	"strconv"
	"time"

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
// @Description 获取所有用户列表（分页）
// @Tags Admin-用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param keyword query string false "搜索关键词"
// @Param status query int false "状态"
// @Success 200 {object} utils.Response
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
// @Description 根据ID获取用户详情
// @Tags Admin-用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response
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
// @Description 管理员创建新用户
// @Tags Admin-用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body services.UserCreateRequest true "用户信息"
// @Success 200 {object} utils.Response
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
// @Description 更新用户信息
// @Tags Admin-用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param body body services.UserUpdateRequest true "用户信息"
// @Success 200 {object} utils.Response
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
// @Description 删除指定用户
// @Tags Admin-用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response
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
// @Description 启用/禁用用户
// @Tags Admin-用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param body body map[string]int true "状态 {status: 0|1}"
// @Success 200 {object} utils.Response
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
// @Description 管理员重置用户密码
// @Tags Admin-用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param body body map[string]string true "新密码 {password: \"xxx\"}"
// @Success 200 {object} utils.Response
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

// BatchGetSimpleInfo 批量获取用户简要信息
// @Summary 批量获取用户简要信息
// @Description 根据用户ID列表批量获取用户简要信息（用户名、昵称等），用于日志等场景显示用户名
// @Tags Admin-用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body map[string][]uint64 true "用户ID列表 {ids: [1, 2, 3]}"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/users/batch-simple [post]
func (c *UserController) BatchGetSimpleInfo(ctx *gin.Context) {
	var req struct {
		IDs []uint64 `json:"ids"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, 400, "参数错误")
		return
	}

	// 去重
	uniqueIDs := make(map[uint64]bool)
	var deduplicatedIDs []uint64
	for _, id := range req.IDs {
		if id > 0 && !uniqueIDs[id] {
			uniqueIDs[id] = true
			deduplicatedIDs = append(deduplicatedIDs, id)
		}
	}

	users, err := c.userService.BatchGetUserSimpleInfo(deduplicatedIDs)
	if err != nil {
		utils.Fail(ctx, 500, "查询失败")
		return
	}

	utils.Success(ctx, gin.H{
		"users": users,
	})
}

// LoginToUser 管理员登录指定用户（生成该用户的 JWT token）
// @Summary 管理员登录指定用户
// @Description 管理员可以生成任意用户的 JWT token 进行调试
// @Tags Admin-用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/users/{id}/login-as [post]
func (c *UserController) LoginToUser(ctx *gin.Context) {
	if config.IsProductionMode() {
		utils.Fail(ctx, 403, "生产环境已禁用该功能")
		return
	}

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

	accessTTL := time.Duration(config.GlobalConfig.JWTAccessExpire) * time.Second
	token, err := utils.GenerateTokenWithTTL(user.ID, user.Role, accessTTL)
	if err != nil {
		utils.Fail(ctx, 500, "生成 token 失败")
		return
	}

	clientIP := ctx.ClientIP()
	if clientIP == "" {
		clientIP = "unknown"
	}
	userAgent := ctx.GetHeader("User-Agent")
	expiresAt := time.Now().Add(accessTTL).Unix()
	_ = models.CreateUserSession(user.ID, utils.HashToken(token), "", clientIP, userAgent, "Admin Impersonation", expiresAt, 0)

	utils.Success(ctx, gin.H{
		"user":  user,
		"token": token,
	})
}

// ResetApiKey 管理员重置指定用户的 API Key
// @Summary 重置用户 API Key
// @Description 管理员重置指定用户的 API 密钥
// @Tags Admin-用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/users/{id}/reset-apikey [post]
func (c *UserController) ResetApiKey(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(ctx, 400, "无效的用户ID")
		return
	}

	newKey, err := models.ResetUserApiKey(id)
	if err != nil {
		utils.Fail(ctx, 500, "重置 API Key 失败: "+err.Error())
		return
	}

	utils.Success(ctx, gin.H{
		"apikey": newKey,
	})
}

// LookupUser 按标识查找用户（ID/用户名/邮箱）
// @Summary 按标识查找用户
// @Description 通过 ID、用户名或邮箱查找用户
// @Tags Admin-用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param keyword query string true "用户标识（ID/用户名/邮箱）"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/users/lookup [get]
func (c *UserController) LookupUser(ctx *gin.Context) {
	keyword := utils.Clean_XSS(ctx.DefaultQuery("keyword", ""))
	if keyword == "" {
		utils.Fail(ctx, 400, "用户标识不能为空")
		return
	}

	// 先尝试按 ID 查找
	if id, err := strconv.ParseUint(keyword, 10, 64); err == nil {
		user, err := c.userService.GetByID(id)
		if err == nil {
			utils.Success(ctx, gin.H{"user": user})
			return
		}
	}

	// 按用户名查找
	user, err := models.GetUserByUsername(keyword)
	if err == nil && user != nil {
		utils.Success(ctx, gin.H{"user": user})
		return
	}

	// 按邮箱查找
	user, err = models.GetUserByEmail(keyword)
	if err == nil && user != nil {
		utils.Success(ctx, gin.H{"user": user})
		return
	}

	utils.Fail(ctx, 404, "用户不存在")
}
