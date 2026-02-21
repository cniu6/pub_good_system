package admin

import (
	"fst/backend/app/models"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// LogController 日志管理控制器
type LogController struct{}

func NewLogController() *LogController {
	return &LogController{}
}

// List 日志列表
func (c *LogController) List(ctx *gin.Context) {
	utils.SanitizeQueryParams(ctx)
	var query models.OperationLogQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
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

// Stats 日志统计
func (c *LogController) Stats(ctx *gin.Context) {
	stats, err := models.GetOperationLogStats()
	if err != nil {
		utils.Fail(ctx, 500, "查询失败")
		return
	}

	utils.Success(ctx, stats)
}

// Clean 清理日志
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
