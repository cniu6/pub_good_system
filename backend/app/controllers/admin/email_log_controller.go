package admin

import (
	"fst/backend/app/models"
	"fst/backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// EmailLogController 邮件发送记录管理控制器
type EmailLogController struct{}

func NewEmailLogController() *EmailLogController {
	return &EmailLogController{}
}

// List 邮件日志列表
// @Summary 获取邮件发送记录列表
// @Description 分页获取邮件发送记录，支持按收件人、模板名、状态筛选
// @Tags Admin-邮件日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param to_email query string false "收件人邮箱（模糊）"
// @Param template_name query string false "模板名称"
// @Param status query int false "状态: -1=全部, 0=失败, 1=成功" default(-1)
// @Param start_time query string false "开始时间 (YYYY-MM-DD HH:MM:SS)"
// @Param end_time query string false "结束时间 (YYYY-MM-DD HH:MM:SS)"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/email-logs [get]
func (ctrl *EmailLogController) List(c *gin.Context) {
	utils.SanitizeQueryParams(c)

	var q models.EmailLogQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	if q.Status == 0 && c.Query("status") == "" {
		q.Status = -1
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}

	logs, total, err := models.GetEmailLogList(&q)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":      logs,
		"total":     total,
		"page":      q.Page,
		"page_size": q.PageSize,
	})
}

// Detail 邮件日志详情
// @Summary 获取邮件发送记录详情
// @Description 根据 ID 获取邮件日志详情，包含邮件内容
// @Tags Admin-邮件日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "日志ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/email-logs/{id} [get]
func (ctrl *EmailLogController) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的 ID")
		return
	}

	log, err := models.GetEmailLogByID(id)
	if err != nil {
		utils.Fail(c, 404, "记录不存在")
		return
	}

	utils.Success(c, log)
}

// Clean 清理邮件日志
// @Summary 清理邮件发送记录
// @Description 删除指定日期之前的邮件日志
// @Tags Admin-邮件日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body map[string]string true "清理参数 {before: '2025-01-01 00:00:00'}"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/email-logs/clean [post]
func (ctrl *EmailLogController) Clean(c *gin.Context) {
	var req struct {
		Before string `json:"before" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误：请提供 before 日期")
		return
	}

	affected, err := models.DeleteEmailLogsBefore(req.Before)
	if err != nil {
		utils.Fail(c, 500, "清理失败")
		return
	}

	utils.Success(c, gin.H{
		"affected": affected,
	})
}

// Stats 邮件日志统计
// @Summary 邮件发送统计
// @Description 获取邮件发送总数、成功数、失败数
// @Tags Admin-邮件日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/email-logs/stats [get]
func (ctrl *EmailLogController) Stats(c *gin.Context) {
	total, success, fail, err := models.GetEmailLogStats()
	if err != nil {
		utils.Fail(c, 500, "统计失败")
		return
	}

	utils.Success(c, gin.H{
		"total":   total,
		"success": success,
		"fail":    fail,
	})
}

// TemplateNames 获取模板名列表（用于筛选下拉）
// @Summary 获取邮件模板名列表
// @Description 获取邮件日志中出现的所有模板名，用于筛选
// @Tags Admin-邮件日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/email-logs/template-names [get]
func (ctrl *EmailLogController) TemplateNames(c *gin.Context) {
	names, err := models.GetEmailTemplateNames()
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	utils.Success(c, names)
}
