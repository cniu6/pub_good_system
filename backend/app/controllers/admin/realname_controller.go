package admin

import (
	"database/sql"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RealnameController 实名认证管理控制器
type RealnameController struct {
	realnameService *services.RealnameService
}

// NewRealnameController 创建实名认证管理控制器
func NewRealnameController() *RealnameController {
	return &RealnameController{
		realnameService: services.NewRealnameService(),
	}
}

// ReviewRealnameRequest 审核实名认证请求
type ReviewRealnameRequest struct {
	ID           uint64 `json:"id" binding:"required"`
	Status       uint8  `json:"status" binding:"required,min=1,max=2"`
	RejectReason string `json:"reject_reason"`
}

// List 实名认证列表
// @Summary 获取实名认证列表
// @Description 获取所有实名认证申请列表（分页）
// @Tags Admin-实名认证管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param keyword query string false "搜索关键词（姓名、证件号）"
// @Param status query int false "状态: 0=待审核, 1=通过, 2=拒绝"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/realname [get]
func (c *RealnameController) List(ctx *gin.Context) {
	utils.SanitizeQueryParams(ctx)

	var query models.RealnameVerificationListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		utils.Fail(ctx, 400, "参数错误: "+err.Error())
		return
	}

	result, err := c.realnameService.GetList(&query)
	if err != nil {
		log.Printf("[ADMIN REALNAME] list failed: %v", err)
		utils.Fail(ctx, 500, "查询失败")
		return
	}

	utils.Success(ctx, result)
}

// Detail 实名认证详情
// @Summary 获取实名认证详情
// @Description 根据ID获取实名认证详情
// @Tags Admin-实名认证管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "实名认证记录ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/realname/{id} [get]
func (c *RealnameController) Detail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(ctx, 400, "无效的ID")
		return
	}

	verification, err := c.realnameService.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.Fail(ctx, 404, "实名认证记录不存在")
			return
		}
		log.Printf("[ADMIN REALNAME] get by id failed: id=%d err=%v", id, err)
		utils.Fail(ctx, 500, "查询失败")
		return
	}

	utils.Success(ctx, gin.H{
		"verification": verification,
	})
}

// Review 审核实名认证
// @Summary 审核实名认证
// @Description 管理员审核用户的实名认证申请
// @Tags Admin-实名认证管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ReviewRealnameRequest true "审核信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/realname/review [post]
func (c *RealnameController) Review(ctx *gin.Context) {
	admin_id, exists := ctx.Get("userID")
	if !exists {
		utils.Fail(ctx, 401, "User not logged in")
		return
	}

	var req ReviewRealnameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, 400, "参数错误: "+err.Error())
		return
	}

	// XSS 过滤
	req.RejectReason = utils.Clean_XSS(req.RejectReason)

	svc_req := &services.RealnameReviewRequest{
		ID:           req.ID,
		Status:       req.Status,
		RejectReason: req.RejectReason,
	}

	if err := c.realnameService.Review(admin_id.(uint64), svc_req); err != nil {
		if services.IsClientError(err) {
			utils.Fail(ctx, 400, err.Error())
			return
		}
		log.Printf("[ADMIN REALNAME] review failed: admin_id=%v id=%d err=%v", admin_id, req.ID, err)
		utils.Fail(ctx, 500, "审核操作失败，请稍后重试")
		return
	}

	utils.Success(ctx, gin.H{"message": "审核操作已完成"})
}

// RegisterRoutes 注册实名认证管理路由
func (c *RealnameController) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/realname", c.List)
	group.GET("/realname/:id", c.Detail)
	group.POST("/realname/review", c.Review)
}

