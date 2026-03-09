package admin

import (
	"fst/backend/app/models"
	"fst/backend/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// LogController 日志管理控制器
type LogController struct{}

func NewLogController() *LogController {
	return &LogController{}
}

// List 日志列表
// @Summary 获取操作日志列表
// @Description 获取操作日志列表（分页），仅支持简单分页浏览
// @Tags Admin-操作日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/logs [get]
func (c *LogController) List(ctx *gin.Context) {
	utils.SanitizeQueryParams(ctx)
	var query models.OperationLogQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		utils.Fail(ctx, 400, "参数错误")
		return
	}

	defaultQueryDays := 30
	defaultMaxCount := 500
	settingsMap, err := models.GetSettingsMap([]string{"operation_log_query_days", "operation_log_max_count"})
	if err == nil {
		if v, ok := settingsMap["operation_log_query_days"]; ok {
			if parsed, parseErr := strconv.Atoi(v); parseErr == nil && parsed > 0 {
				defaultQueryDays = parsed
			}
		}
		if v, ok := settingsMap["operation_log_max_count"]; ok {
			if parsed, parseErr := strconv.Atoi(v); parseErr == nil && parsed > 0 {
				defaultMaxCount = parsed
			}
		}
	}

	if defaultQueryDays > 365 {
		defaultQueryDays = 365
	}
	if defaultMaxCount > 10000 {
		defaultMaxCount = 10000
	}

	// 自动清理超出上限的旧日志
	if cleaned, cleanErr := models.CleanExcessOperationLogs(defaultMaxCount); cleanErr == nil && cleaned > 0 {
		// 清理成功，静默处理
		_ = cleaned
	}

	// 设置默认分页
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	now := time.Now().Unix()
	if query.EndTime <= 0 {
		query.EndTime = now
	}
	if query.StartTime <= 0 {
		query.StartTime = query.EndTime - int64(defaultQueryDays*24*60*60)
	}
	if query.StartTime > query.EndTime {
		utils.Fail(ctx, 400, "参数错误")
		return
	}

	logs, total, err := models.GetOperationLogList(&query)
	if err != nil {
		utils.Fail(ctx, 500, "查询失败")
		return
	}

	utils.Success(ctx, gin.H{
		"list":      logs,
		"total":     total,
		"page":      query.Page,
		"page_size": query.PageSize,
	})
}

// Clean 清理日志
// @Summary 清理操作日志
// @Description 清理指定时间之前的操作日志，用于控制日志数量
// @Tags Admin-操作日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body map[string]int64 true "清理参数 {before_time: timestamp}"
// @Success 200 {object} utils.Response
// @Router /api/v1/admin/logs/clean [post]
func (c *LogController) Clean(ctx *gin.Context) {
	var req struct {
		BeforeTime int64 `json:"before_time" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.Fail(ctx, 400, "参数错误")
		return
	}

	affected, err := models.DeleteOperationLogsBefore(req.BeforeTime)
	if err != nil {
		utils.Fail(ctx, 500, "清理失败")
		return
	}

	utils.Success(ctx, gin.H{
		"affected": affected,
	})
}
