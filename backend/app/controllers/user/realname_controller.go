package user

import (
	"database/sql"
	"fst/backend/app/services"
	"fst/backend/utils"
	"github.com/gin-gonic/gin"
	"log"
)

// RealnameController 实名认证控制器
type RealnameController struct {
	realnameService *services.RealnameService
}

// NewRealnameController 创建实名认证控制器
func NewRealnameController() *RealnameController {
	return &RealnameController{
		realnameService: services.NewRealnameService(),
	}
}

// SubmitRealnameRequest 提交实名认证请求
type SubmitRealnameRequest struct {
	RealName         string `json:"real_name" binding:"required"`
	CertificateType  uint8  `json:"certificate_type" binding:"required,min=1,max=3"`
	CertificateNo    string `json:"certificate_no" binding:"required"`
	CertificateFront string `json:"certificate_front" binding:"required"`
	CertificateBack  string `json:"certificate_back" binding:"required"`
}

// SubmitRealname 提交实名认证
// @Summary 提交实名认证
// @Description 用户提交实名认证申请
// @Tags 实名认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SubmitRealnameRequest true "实名认证信息"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/realname [post]
func (c *RealnameController) SubmitRealname(ctx *gin.Context) {
	user_id, exists := ctx.Get("userID")
	if !exists {
		utils.Fail(ctx, 401, "User not logged in")
		return
	}

	var req SubmitRealnameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, 400, err.Error())
		return
	}

	// XSS 过滤
	req.RealName = utils.Clean_XSS(req.RealName)
	req.CertificateNo = utils.Clean_XSS(req.CertificateNo)
	req.CertificateFront = utils.Clean_XSS(req.CertificateFront)
	req.CertificateBack = utils.Clean_XSS(req.CertificateBack)

	// URL 校验
	if req.CertificateFront != "" && !utils.ValidateURL(req.CertificateFront) {
		utils.Fail(ctx, 400, "证件正面照 URL 仅支持 http/https 协议")
		return
	}
	if req.CertificateBack != "" && !utils.ValidateURL(req.CertificateBack) {
		utils.Fail(ctx, 400, "证件背面照 URL 仅支持 http/https 协议")
		return
	}

	// 长度限制
	if len(req.RealName) > 50 {
		utils.Fail(ctx, 400, "姓名长度不能超过50个字符")
		return
	}
	if len(req.RealName) < 2 {
		utils.Fail(ctx, 400, "姓名长度不能少于2个字符")
		return
	}
	if len(req.CertificateNo) > 50 {
		utils.Fail(ctx, 400, "证件号码长度不能超过50个字符")
		return
	}
	if len(req.CertificateFront) > 500 {
		utils.Fail(ctx, 400, "证件正面照 URL 长度不能超过500个字符")
		return
	}
	if len(req.CertificateBack) > 500 {
		utils.Fail(ctx, 400, "证件背面照 URL 长度不能超过500个字符")
		return
	}

	svc_req := &services.RealnameSubmitRequest{
		RealName:         req.RealName,
		CertificateType:  req.CertificateType,
		CertificateNo:    req.CertificateNo,
		CertificateFront: req.CertificateFront,
		CertificateBack:  req.CertificateBack,
	}

	if err := c.realnameService.Submit(user_id.(uint64), svc_req); err != nil {
		if services.IsClientError(err) {
			utils.Fail(ctx, 400, err.Error())
			return
		}
		log.Printf("[REALNAME] submit failed for user_id=%d: %v", user_id.(uint64), err)
		utils.Fail(ctx, 500, "实名认证提交失败，请稍后重试")
		return
	}

	utils.Success(ctx, gin.H{"message": "实名认证申请已提交，请等待审核"})
}

// GetMyRealnameStatus 获取我的实名认证状态
// @Summary 获取实名认证状态
// @Description 获取当前用户的实名认证状态
// @Tags 实名认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/realname [get]
func (c *RealnameController) GetMyRealnameStatus(ctx *gin.Context) {
	user_id, exists := ctx.Get("userID")
	if !exists {
		utils.Fail(ctx, 401, "User not logged in")
		return
	}

	verification, err := c.realnameService.GetUserVerification(user_id.(uint64))
	if err != nil {
		if err == sql.ErrNoRows {
			utils.Success(ctx, gin.H{
				"hasVerification": false,
				"status":          0,
			})
			return
		}
		log.Printf("[REALNAME] load status failed for user_id=%d: %v", user_id.(uint64), err)
		utils.Fail(ctx, 500, "查询实名状态失败")
		return
	}

	utils.Success(ctx, gin.H{
		"hasVerification": true,
		"id":              verification.ID,
		"status":          verification.Status,
		"realName":        verification.RealName,
		"certificateType": verification.CertificateType,
		// 用户端只回传掩码后的证件号，避免身份证号明文泄露
		"certificateNo":    utils.MaskCertificateNo(verification.CertificateNo),
		"certificateFront": verification.CertificateFront,
		"certificateBack":  verification.CertificateBack,
		"rejectReason":     verification.RejectReason,
		"submittedAt":      verification.SubmittedAt,
		"reviewedAt":       verification.ReviewedAt,
	})
}

// RegisterRoutes 注册实名认证路由
func (c *RealnameController) RegisterRoutes(group *gin.RouterGroup) {
	group.POST("/realname", c.SubmitRealname)
	group.GET("/realname", c.GetMyRealnameStatus)
}
