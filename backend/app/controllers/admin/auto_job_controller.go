package admin

import (
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

func (ctrl *AutoJobController) Overview(c *gin.Context) {
	ov, err := task.BuildOverview()
	if err != nil {
		utils.Fail(c, 500, "查询失败: "+err.Error())
		return
	}
	utils.Success(c, ov)
}

func (ctrl *AutoJobController) GetConfig(c *gin.Context) {
	utils.Success(c, task.LoadGlobalConfig())
}

func (ctrl *AutoJobController) PutConfig(c *gin.Context) {
	var req task.GlobalConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
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

func (ctrl *AutoJobController) ListHandlers(c *gin.Context) {
	utils.Success(c, gin.H{"handlers": task.ListHandlerKeys()})
}

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
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.Success(c, gin.H{"list": list, "total": len(list)})
}

func (ctrl *AutoJobController) JobDetail(c *gin.Context) {
	code := strings.TrimSpace(c.Param("job_code"))
	def, err := task.GetDefinition(code)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	if def == nil {
		utils.Fail(c, 404, "任务不存在")
		return
	}
	utils.Success(c, def)
}

func (ctrl *AutoJobController) UpdateJob(c *gin.Context) {
	code := strings.TrimSpace(c.Param("job_code"))
	var req task.UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	if err := task.UpdateDefinitionFields(code, req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	def, _ := task.GetDefinition(code)
	utils.Success(c, def)
}

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
	// 失败也返回 run 详情
	if err != nil {
		utils.SuccessMsg(c, err.Error(), run)
		return
	}
	utils.Success(c, run)
}

func (ctrl *AutoJobController) EnableJob(c *gin.Context) {
	code := strings.TrimSpace(c.Param("job_code"))
	if err := task.SetEnabled(code, true); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{"job_code": code, "enabled": true})
}

func (ctrl *AutoJobController) DisableJob(c *gin.Context) {
	code := strings.TrimSpace(c.Param("job_code"))
	if err := task.SetEnabled(code, false); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{"job_code": code, "enabled": false})
}

func (ctrl *AutoJobController) ListRunning(c *gin.Context) {
	list, err := task.ListRunningDefinitions()
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.Success(c, gin.H{"list": list, "total": len(list)})
}

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
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (ctrl *AutoJobController) RunDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效 ID")
		return
	}
	run, err := task.GetRun(id)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	if run == nil {
		utils.Fail(c, 404, "记录不存在")
		return
	}
	utils.Success(c, run)
}

func (ctrl *AutoJobController) CleanRuns(c *gin.Context) {
	var req task.CleanRunsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	aff, err := task.CleanRuns(req)
	if err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{"affected": aff})
}

func (ctrl *AutoJobController) MarkKeep(c *gin.Context) {
	var req task.MarkKeepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}
	aff, err := task.MarkKeepForever(req.IDs, req.KeepForever)
	if err != nil {
		utils.Fail(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{"affected": aff})
}
