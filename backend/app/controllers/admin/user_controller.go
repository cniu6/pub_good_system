package admin

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	"fst/backend/utils"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UserController 用户管理控制器
type UserController struct {
	userService *services.UserService
}

type AdminUserRealnameSummary struct {
	HasVerification bool    `json:"has_verification"`
	ID              uint64  `json:"id,omitempty"`
	Status          uint8   `json:"status,omitempty"`
	RealName        string  `json:"real_name,omitempty"`
	CertificateType uint8   `json:"certificate_type,omitempty"`
	CertificateNo   string  `json:"certificate_no,omitempty"`
	SubmittedAt     *int64  `json:"submitted_at,omitempty"`
	ReviewedAt      *int64  `json:"reviewed_at,omitempty"`
	RejectReason    string  `json:"reject_reason,omitempty"`
}

type AdminUserDetailResponse struct {
	User     *services.AdminUserListItem `json:"user"`
	Realname AdminUserRealnameSummary    `json:"realname"`
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
// @Param realname_status query int false "实名认证状态: 0=待审核, 1=通过, 2=拒绝"
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
		log.Printf("[ADMIN USER] list users failed: %v", err)
		utils.Fail(ctx, 500, "查询失败")
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
// @Success 200 {object} utils.Response{data=AdminUserDetailResponse}
// @Router /api/v1/admin/users/{id} [get]
func (c *UserController) Detail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(ctx, 400, "无效的用户ID")
		return
	}

	user, err := c.userService.GetAdminDetail(id)
	if err != nil {
		utils.Fail(ctx, 404, "用户不存在")
		return
	}

	realnameSummary := AdminUserRealnameSummary{
		HasVerification: false,
	}
	if verification, err := models.GetRealnameVerificationByUserID(id); err == nil && verification != nil {
		realnameSummary = AdminUserRealnameSummary{
			HasVerification: true,
			ID:              verification.ID,
			Status:          verification.Status,
			RealName:        verification.RealName,
			CertificateType: verification.CertificateType,
			CertificateNo:   verification.CertificateNo,
			SubmittedAt:     verification.SubmittedAt,
			ReviewedAt:      verification.ReviewedAt,
			RejectReason:    verification.RejectReason,
		}
	}

	utils.Success(ctx, AdminUserDetailResponse{
		User:     user,
		Realname: realnameSummary,
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
		if utils.IsPasswordValidationError(err) {
			utils.Fail(ctx, 400, err.Error())
			return
		}
		log.Printf("[ADMIN USER] hash password failed on create: %v", err)
		utils.Fail(ctx, 500, "创建用户失败")
		return
	}
	req.Password = hashed

	user, err := c.userService.Create(&req)
	if err != nil {
		if services.IsClientError(err) {
			utils.Fail(ctx, 400, err.Error())
			return
		}
		log.Printf("[ADMIN USER] create user failed: %v", err)
		utils.Fail(ctx, 500, "创建用户失败")
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

	req := services.UserUpdateRequest{ID: id}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, 400, "参数错误: "+err.Error())
		return
	}

	if err := c.userService.Update(&req); err != nil {
		if services.IsClientError(err) {
			utils.Fail(ctx, 400, err.Error())
			return
		}
		log.Printf("[ADMIN USER] update user failed for user_id=%d: %v", id, err)
		utils.Fail(ctx, 500, "更新用户失败")
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
		log.Printf("[ADMIN USER] delete user failed for user_id=%d: %v", id, err)
		utils.Fail(ctx, 500, "删除用户失败")
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
		log.Printf("[ADMIN USER] update status failed for user_id=%d: %v", id, err)
		utils.Fail(ctx, 500, "更新用户状态失败")
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
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, 400, "参数错误: 密码至少8位")
		return
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		if utils.IsPasswordValidationError(err) {
			utils.Fail(ctx, 400, err.Error())
			return
		}
		log.Printf("[ADMIN USER] hash password failed on reset for user_id=%d: %v", id, err)
		utils.Fail(ctx, 500, "重置密码失败")
		return
	}

	if err := c.userService.UpdatePassword(id, hashed); err != nil {
		log.Printf("[ADMIN USER] reset password failed for user_id=%d: %v", id, err)
		utils.Fail(ctx, 500, "重置密码失败")
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
	refreshTTL := time.Duration(config.GlobalConfig.JWTRefreshExpire) * time.Second
	token, err := utils.GenerateTokenForGuardWithTTL(user.ID, user.Role, utils.UserAuthGuard, accessTTL)
	if err != nil {
		utils.Fail(ctx, 500, "生成 token 失败")
		return
	}
	refreshToken, err := utils.GenerateRefreshTokenForGuardWithTTL(user.ID, utils.UserAuthGuard, refreshTTL)
	if err != nil {
		utils.Fail(ctx, 500, "生成 refresh token 失败")
		return
	}

	clientIP := utils.GetClientIP(ctx)
	userAgent := ctx.GetHeader("User-Agent")
	expiresAt := time.Now().Add(accessTTL).Unix()
	refreshExpiresAt := time.Now().Add(refreshTTL).Unix()
	if err := models.CreateUserSession(user.ID, utils.UserAuthGuard, utils.HashToken(token), utils.HashToken(refreshToken), clientIP, userAgent, "Admin Impersonation", expiresAt, refreshExpiresAt); err != nil {
		utils.Fail(ctx, 500, "创建登录会话失败")
		return
	}

	utils.Success(ctx, gin.H{
		"user": gin.H{
			"id":              user.ID,
			"group_id":        user.GroupId,
			"username":        user.Username,
			"nickname":        user.Nickname,
			"email":           user.Email,
			"mobile":          user.Mobile,
			"avatar":          user.Avatar,
			"back_ground":     user.BackGround,
			"gender":          user.Gender,
			"birthday":        user.Birthday,
			"money":           user.Money,
			"score":           user.Score,
			"level":           user.Level,
			"role":            user.Role,
			"last_login_time": user.LastLoginTime,
			"last_login_ip":   user.LastLoginIp,
			"login_failure":   user.LoginFailure,
			"join_ip":         user.JoinIp,
			"join_time":       user.JoinTime,
			"motto":           user.Motto,
			"status":          user.Status,
			"apikey":          user.Apikey,
			"update_time":     user.UpdateTime,
			"create_time":     user.CreateTime,
			"language":        user.Language,
			"country":         user.Country,
			"token":           user.Token,
			"realname":        services.BuildLoginRealnameSummaryForAPI(user.ID),
		},
		"token":            token,
		"refreshToken":     refreshToken,
		"expiresAt":        expiresAt,
		"refreshExpiresAt": refreshExpiresAt,
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
		log.Printf("[ADMIN USER] reset api key failed for user_id=%d: %v", id, err)
		utils.Fail(ctx, 500, "重置 API Key 失败")
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

