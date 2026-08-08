package autojob

import (
	"log"
	"strconv"
	"strings"

	"fst/backend/internal/task"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// AutoJobController 自动任务管理器
type AutoJobController struct{}

func NewAutoJobController() *AutoJobController {
	return &AutoJobController{}
}

// RegisterRoutes 挂到 /api/v1/admin/auto-jobs
func (ctrl *AutoJobController) RegisterRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/auto-jobs")
	{
		g.GET("/overview", ctrl.Overview)
		g.GET("/config", ctrl.GetConfig)
		g.PUT("/config", ctrl.PutConfig)
		g.POST("/presets/import", ctrl.ImportPresets)
		g.GET("/handlers", ctrl.ListHandlers)
		g.GET("/running", ctrl.ListRunning)
		g.GET("/runs", ctrl.ListRuns)
		g.GET("/runs/:id", ctrl.RunDetail)
		g.POST("/runs/clean", ctrl.CleanRuns)
		g.POST("/runs/mark-keep", ctrl.MarkKeep)
		g.GET("", ctrl.ListJobs)
		g.GET("/:job_code", ctrl.JobDetail)
		g.PUT("/:job_code", ctrl.UpdateJob)
		g.POST("/:job_code/run", ctrl.RunJob)
		g.POST("/:job_code/enable", ctrl.EnableJob)
		g.POST("/:job_code/disable", ctrl.DisableJob)
	}
}

// Overview 获取自动任务概览
// @Summary 自动任务概览
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/overview [get]
func (ctrl *AutoJobController) Overview(c *gin.Context) {
	ov, err := task.BuildOverview()
	if err != nil {
		utils.Fail(c, 500, "查询失败: "+err.Error())
		return
	}
	utils.Success(c, ov)
}

// GetConfig 获取自动任务配置
// @Summary 获取自动任务配置
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/config [get]
func (ctrl *AutoJobController) GetConfig(c *gin.Context) {
	utils.Success(c, task.LoadGlobalConfig())
}

// PutConfig 更新自动任务配置
// @Summary 更新自动任务配置
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/config [put]
func (ctrl *AutoJobController) PutConfig(c *gin.Context) {
	var req task.GlobalConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "Invalid parameters")
		return
	}
	if req.RunMaxCount <= 0 {
		req.RunMaxCount = 10000
	}
	if err := task.SaveGlobalConfig(req); err != nil {
		utils.Fail(c, 500, "保存失败: "+err.Error())
		return
	}
	utils.Success(c, task.LoadGlobalConfig())
}

// ImportPresets 导入自动任务预设
// @Summary 导入自动任务预设
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/presets/import [post]
func (ctrl *AutoJobController) ImportPresets(c *gin.Context) {
	var req task.ImportPresetsRequest
	_ = c.ShouldBindJSON(&req)
	if req.Mode == "" {
		req.Mode = "skip"
	}
	result, err := task.ImportPresets(req.Mode)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, result)
}

// ListHandlers 列出自动任务处理器
// @Summary 列出自动任务处理器
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/handlers [get]
func (ctrl *AutoJobController) ListHandlers(c *gin.Context) {
	utils.Success(c, gin.H{"handlers": task.ListHandlerKeys()})
}

// ListJobs 列出自动任务定义
// @Summary 自动任务定义列表
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs [get]
func (ctrl *AutoJobController) ListJobs(c *gin.Context) {
	utils.SanitizeQueryParams(c)
	keyword := strings.TrimSpace(c.Query("keyword"))
	category := strings.TrimSpace(c.Query("category"))
	var enabledPtr *bool
	if e := c.Query("enabled"); e != "" {
		v := e == "1" || e == "true"
		enabledPtr = &v
	}
	list, err := task.ListDefinitions(keyword, category, enabledPtr)
	if err != nil {
		utils.Fail(c, 500, "Query failed")
		return
	}
	utils.Success(c, gin.H{"list": list, "total": len(list)})
}

// JobDetail 自动任务定义详情
// @Summary 自动任务定义详情
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/{job_code} [get]
func (ctrl *AutoJobController) JobDetail(c *gin.Context) {
	code := strings.TrimSpace(c.Param("job_code"))
	def, err := task.GetDefinition(code)
	if err != nil {
		utils.Fail(c, 500, "Query failed")
		return
	}
	if def == nil {
		utils.Fail(c, 404, "Task does not exist")
		return
	}
	utils.Success(c, def)
}

// UpdateJob 更新自动任务定义
// @Summary 更新自动任务定义
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/{job_code} [put]
func (ctrl *AutoJobController) UpdateJob(c *gin.Context) {
	code := strings.TrimSpace(c.Param("job_code"))
	var req task.UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "Invalid parameters")
		return
	}
	if err := task.UpdateDefinitionFields(code, req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	def, _ := task.GetDefinition(code)
	utils.Success(c, def)
}

// RunJob 立即执行自动任务
// @Summary 立即执行自动任务
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/{job_code}/run [post]
func (ctrl *AutoJobController) RunJob(c *gin.Context) {
	code := strings.TrimSpace(c.Param("job_code"))
	operator, _ := c.Get("username")
	op, _ := operator.(string)
	run, err := task.Trigger(code, task.RunOptions{
		Trigger:  task.TriggerManual,
		Operator: op,
		Force:    true,
	})
	if err != nil && run == nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	// 失败也返回 run 详情；对外用固定文案，内部错误写入 run.ErrorText / 日志
	if err != nil {
		log.Printf("[ADMIN][AUTOJOB] run failed job=%s: %v", code, err)
		utils.SuccessMsg(c, "Task execution failed", run)
		return
	}
	utils.Success(c, run)
}

// EnableJob 启用自动任务
// @Summary 启用自动任务
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/{job_code}/enable [post]
func (ctrl *AutoJobController) EnableJob(c *gin.Context) {
	code := strings.TrimSpace(c.Param("job_code"))
	if err := task.SetEnabled(code, true); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{"job_code": code, "enabled": true})
}

// DisableJob 禁用自动任务
// @Summary 禁用自动任务
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/{job_code}/disable [post]
func (ctrl *AutoJobController) DisableJob(c *gin.Context) {
	code := strings.TrimSpace(c.Param("job_code"))
	if err := task.SetEnabled(code, false); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{"job_code": code, "enabled": false})
}

// ListRunning 列出运行中的自动任务
// @Summary 列出运行中的自动任务
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/running [get]
func (ctrl *AutoJobController) ListRunning(c *gin.Context) {
	list, err := task.ListRunningDefinitions()
	if err != nil {
		utils.Fail(c, 500, "Query failed")
		return
	}
	utils.Success(c, gin.H{"list": list, "total": len(list)})
}

// ListRuns 获取自动任务运行记录
// @Summary 自动任务运行记录列表
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/runs [get]
func (ctrl *AutoJobController) ListRuns(c *gin.Context) {
	utils.SanitizeQueryParams(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = utils.NormalizePagination(page, pageSize)
	keyword := strings.TrimSpace(c.Query("keyword"))
	status := strings.TrimSpace(c.Query("status"))
	category := strings.TrimSpace(c.Query("category"))
	jobCode := strings.TrimSpace(c.Query("job_code"))
	startAt, _ := strconv.ParseInt(c.Query("start_time"), 10, 64)
	endAt, _ := strconv.ParseInt(c.Query("end_time"), 10, 64)
	list, total, err := task.ListRuns(page, pageSize, keyword, status, category, jobCode, startAt, endAt)
	if err != nil {
		utils.Fail(c, 500, "Query failed")
		return
	}
	utils.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RunDetail 自动任务运行详情
// @Summary 自动任务运行详情
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/runs/{id} [get]
func (ctrl *AutoJobController) RunDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "Invalid ID")
		return
	}
	run, err := task.GetRun(id)
	if err != nil {
		utils.Fail(c, 500, "Query failed")
		return
	}
	if run == nil {
		utils.Fail(c, 404, "Record does not exist")
		return
	}
	utils.Success(c, run)
}

// CleanRuns 清理自动任务运行记录
// @Summary 清理自动任务运行记录
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/runs/clean [post]
func (ctrl *AutoJobController) CleanRuns(c *gin.Context) {
	var req task.CleanRunsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "Invalid parameters")
		return
	}
	aff, err := task.CleanRuns(req)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{"affected": aff})
}

// MarkKeep 标记运行记录永久保留
// @Summary 标记运行记录永久保留
// @Tags Admin-自动任务
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/auto-jobs/runs/mark-keep [post]
func (ctrl *AutoJobController) MarkKeep(c *gin.Context) {
	var req task.MarkKeepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "Invalid parameters")
		return
	}
	aff, err := task.MarkKeepForever(req.IDs, req.KeepForever)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{"affected": aff})
}
