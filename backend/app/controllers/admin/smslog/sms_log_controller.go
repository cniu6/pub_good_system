package smslog

import (
	"fst/backend/app/models"
	"fst/backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SMSLogController 短信发送记录管理控制器
type SMSLogController struct{}

func NewSMSLogController() *SMSLogController {
	return &SMSLogController{}
}

// List 短信日志列表
// @Summary 获取短信发送记录列表
// @Description 分页获取短信发送记录，支持按手机号、服务商、模板名、语言、状态筛选
// @Tags Admin-短信日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param phone query string false "手机号（模糊）"
// @Param provider query string false "服务商: aliyun, tencent, custom, console"
// @Param template_name query string false "模板名称"
// @Param lang query string false "语言: zh-CN, en-US"
// @Param status query int false "状态: -1=全部, 0=失败, 1=成功" default(-1)
// @Param start_time query string false "开始时间 (YYYY-MM-DD HH:MM:SS)"
// @Param end_time query string false "结束时间 (YYYY-MM-DD HH:MM:SS)"
// @Success 200 {object} utils.Response
// @Router /v1/admin/sms-logs [get]
func (ctrl *SMSLogController) List(c *gin.Context) {
	utils.SanitizeQueryParams(c)

	var q models.SMSLogQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	if q.Status == 0 && c.Query("status") == "" {
		q.Status = -1
	}

	q.Page, q.PageSize = utils.NormalizePagination(q.Page, q.PageSize)

	logs, total, err := models.GetSMSLogList(&q)
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

// Detail 短信日志详情
// @Summary 获取短信发送记录详情
// @Description 根据 ID 获取短信日志详情，包含完整响应内容
// @Tags Admin-短信日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "日志ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/sms-logs/{id} [get]
func (ctrl *SMSLogController) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的 ID")
		return
	}

	log, err := models.GetSMSLogByID(id)
	if err != nil {
		utils.Fail(c, 404, "记录不存在")
		return
	}

	utils.Success(c, log)
}

// Clean 清理短信日志
// @Summary 清理短信发送记录
// @Description 删除指定日期之前的短信日志
// @Tags Admin-短信日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body map[string]string true "清理参数 {before: '2025-01-01 00:00:00'}"
// @Success 200 {object} utils.Response
// @Router /v1/admin/sms-logs/clean [post]
func (ctrl *SMSLogController) Clean(c *gin.Context) {
	var req struct {
		Before string `json:"before" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误：请提供 before 日期")
		return
	}

	affected, err := models.DeleteSMSLogsBefore(req.Before)
	if err != nil {
		utils.Fail(c, 500, "清理失败")
		return
	}

	utils.Success(c, gin.H{
		"affected": affected,
	})
}

// Stats 短信日志统计
// @Summary 短信发送统计
// @Description 获取短信发送总数、今日、成功数、失败数、热门模板（读独立聚合表）
// @Tags Admin-短信日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/sms-logs/stats [get]
func (ctrl *SMSLogController) Stats(c *gin.Context) {
	stats, err := models.GetSMSLogStatsDetail()
	if err != nil {
		utils.Fail(c, 500, "统计失败")
		return
	}
	utils.Success(c, stats)
}

// TemplateNames 获取模板名列表（用于筛选下拉）
// @Summary 获取短信模板名列表
// @Description 获取短信日志中出现的所有模板名，用于筛选
// @Tags Admin-短信日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/sms-logs/template-names [get]
func (ctrl *SMSLogController) TemplateNames(c *gin.Context) {
	names, err := models.GetSMSTemplateNames()
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	utils.Success(c, names)
}

