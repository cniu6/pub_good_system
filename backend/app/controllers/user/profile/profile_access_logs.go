package profile

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// currentUserID 从上下文取当前登录用户 ID；缺失/无效时直接写 401 响应。
func currentUserID(c *gin.Context) (uint64, bool) {
	uid, ok := utils.GetUserID(c)
	if !ok || uid == 0 {
		utils.Fail(c, 401, "User not logged in")
		return 0, false
	}
	return uid, true
}

// ListMyOperationLogs 当前用户自己的操作日志列表（强制按 user_id 过滤）
// @Summary 我的操作日志列表
// @Tags User-资料
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/user/logs [get]
func (ctrl *ProfileController) ListMyOperationLogs(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !services.GetGlobalOperationLogRuntimeConfig().UserVisible {
		utils.Fail(c, 403, "User operation logs not open")
		return
	}
	utils.SanitizeQueryParams(c)

	var query models.OperationLogQuery
	_ = c.ShouldBindQuery(&query)
	// 严禁看别人：忽略客户端传入的 user_id，强制覆盖为当前用户
	query.UserID = uid
	query.Username = ""

	defaultQueryDays := 30
	opCfg := services.GetGlobalOperationLogRuntimeConfig()
	if opCfg.QueryDays > 0 {
		defaultQueryDays = opCfg.QueryDays
	}
	query.Page, query.PageSize = utils.NormalizePagination(query.Page, query.PageSize)

	var rangeErr error
	query.StartTime, query.EndTime, rangeErr = utils.NormalizeTimeRange(query.StartTime, query.EndTime, defaultQueryDays, 365)
	if rangeErr != nil {
		utils.Fail(c, 400, "Invalid parameters")
		return
	}

	logs, total, err := models.GetOperationLogList(&query)
	if err != nil {
		utils.Fail(c, 500, "Query failed")
		return
	}
	// 列表不给用户看请求/响应体
	for i := range logs {
		logs[i].RequestBody = nil
		logs[i].ResponseBody = nil
	}

	utils.Success(c, gin.H{
		"list":      logs,
		"total":     total,
		"page":      query.Page,
		"page_size": query.PageSize,
	})
}

// GetMyOperationLogDetail 当前用户自己的操作日志详情
// @Summary 我的操作日志详情
// @Tags User-资料
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/user/logs/{id} [get]
func (ctrl *ProfileController) GetMyOperationLogDetail(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	opCfg := services.GetGlobalOperationLogRuntimeConfig()
	if !opCfg.UserVisible {
		utils.Fail(c, 403, "User operation logs not open")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "Invalid ID")
		return
	}
	item, err := models.GetOperationLogByID(id)
	if err != nil || item == nil {
		utils.Fail(c, 404, "Record does not exist")
		return
	}
	// 归属校验：只能看自己的
	if item.UserID != uid {
		utils.Fail(c, 403, "No permission to view this record")
		return
	}
	// 默认不展示请求/响应体，避免把失败堆栈等内部细节暴露给用户
	if !opCfg.UserShowBody {
		item.RequestBody = nil
		item.ResponseBody = nil
	}
	utils.Success(c, item)
}

// ListMyAPILogs 当前用户自己的 API 访问日志（仅 API Key 鉴权；不含 JWT 网页请求）
// @Summary 我的 API 访问日志列表
// @Tags User-资料
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/user/api-logs [get]
func (ctrl *ProfileController) ListMyAPILogs(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	apiCfg := services.GetGlobalAPILogRuntimeConfig()
	if !apiCfg.UserVisible {
		utils.Fail(c, 403, "User API logs not open")
		return
	}
	utils.SanitizeQueryParams(c)

	var query models.APIAccessLogQuery
	_ = c.ShouldBindQuery(&query)
	// 严禁看别人：强制覆盖 user_id
	query.UserID = uid
	query.Username = ""
	// 用户中心只展示 API Key 调用，不展示 JWT（网页登录态）
	query.AuthMethod = "apikey"

	defaultQueryDays := apiCfg.QueryDays
	if defaultQueryDays <= 0 {
		defaultQueryDays = 7
	}
	query.Page, query.PageSize = utils.NormalizePagination(query.Page, query.PageSize)

	var rangeErr error
	query.StartTime, query.EndTime, rangeErr = utils.NormalizeTimeRange(query.StartTime, query.EndTime, defaultQueryDays, 365)
	if rangeErr != nil {
		utils.Fail(c, 400, "Invalid parameters")
		return
	}

	list, total, err := models.GetAPIAccessLogList(&query)
	if err != nil {
		utils.Fail(c, 500, "Query failed")
		return
	}

	utils.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      query.Page,
		"page_size": query.PageSize,
	})
}

// GetMyAPILogStats 当前用户自己的 API Key 调用统计
// @Summary 我的 API 调用统计
// @Tags User-资料
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/user/api-logs/stats [get]
func (ctrl *ProfileController) GetMyAPILogStats(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !services.GetGlobalAPILogRuntimeConfig().UserVisible {
		utils.Fail(c, 403, "User API logs not open")
		return
	}
	stats, err := models.GetAPIAccessLogStatsByUserID(uid)
	if err != nil {
		utils.Fail(c, 500, "Statistics failed")
		return
	}
	utils.Success(c, stats)
}

// GetMyAPILogDetail 当前用户自己的 API 访问日志详情（仅 API Key）
// @Summary 我的 API 访问日志详情
// @Tags User-资料
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/user/api-logs/{id} [get]
func (ctrl *ProfileController) GetMyAPILogDetail(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	if !services.GetGlobalAPILogRuntimeConfig().UserVisible {
		utils.Fail(c, 403, "User API logs not open")
		return
	}
	param := c.Param("id")
	var (
		item *models.APIAccessLog
		err  error
	)
	if id, parseErr := strconv.ParseUint(param, 10, 64); parseErr == nil {
		item, err = models.GetAPIAccessLogByID(id)
	} else {
		item, err = models.GetAPIAccessLogByRequestID(param)
	}
	if err != nil || item == nil {
		utils.Fail(c, 404, "Record does not exist")
		return
	}
	if item.UserID != uid {
		utils.Fail(c, 403, "No permission to view this record")
		return
	}
	// 仅允许查看本人 API Key 调用；JWT/未鉴权一律拒绝
	if item.AuthMethod != "apikey" {
		utils.Fail(c, 403, "No permission to view this record")
		return
	}
	utils.Success(c, item)
}
