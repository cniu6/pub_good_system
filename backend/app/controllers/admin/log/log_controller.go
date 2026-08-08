package log

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"
	"strconv"

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
// @Router /v1/admin/logs [get]
func (c *LogController) List(ctx *gin.Context) {
	utils.SanitizeQueryParams(ctx)
	var query models.OperationLogQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		utils.Fail(ctx, 400, "参数错误")
		return
	}

	// 注意：保留清理不在这里做。写日志时已经异步节流触发过 scheduleOperationLogRetentionCleanup
	// （见 models/operation_log.go），管理端浏览列表只是查询，不应该有「顺手清理」的副作用——
	// 之前这里同步调用清理，高并发刷列表页会重复触发 DELETE，且没有节流。
	defaultQueryDays := 30
	opCfg := services.GetGlobalOperationLogRuntimeConfig()
	if opCfg.QueryDays > 0 {
		defaultQueryDays = opCfg.QueryDays
	}

	query.Page, query.PageSize = utils.NormalizePagination(query.Page, query.PageSize)

	var rangeErr error
	query.StartTime, query.EndTime, rangeErr = utils.NormalizeTimeRange(query.StartTime, query.EndTime, defaultQueryDays, 365)
	if rangeErr != nil {
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

// Detail 日志详情
// @Summary 获取操作日志详情
// @Description 获取操作日志详情
// @Tags Admin-操作日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint64 true "日志 ID"
// @Success 200 {object} utils.Response
// @Router /v1/admin/logs/{id} [get]
func (c *LogController) Detail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(ctx, 400, "无效的 ID")
		return
	}

	logItem, err := models.GetOperationLogByID(id)
	if err != nil {
		utils.Fail(ctx, 404, "记录不存在")
		return
	}

	utils.Success(ctx, logItem)
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
// @Router /v1/admin/logs/clean [post]
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

// Stats 操作日志统计（读独立聚合表，清理明细不影响累计）
// @Summary 操作日志统计
// @Description 获取操作日志总请求数、今日、4xx、5xx、热门模块等
// @Tags Admin-操作日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/logs/stats [get]
func (c *LogController) Stats(ctx *gin.Context) {
	stats, err := models.GetOperationLogStatsDetail()
	if err != nil {
		utils.Fail(ctx, 500, "统计失败")
		return
	}
	utils.Success(ctx, stats)
}

