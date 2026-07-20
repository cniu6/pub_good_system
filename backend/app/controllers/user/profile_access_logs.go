package user

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// currentUserID 从上下文取当前登录用户 ID
func currentUserID(c *gin.Context) (uint64, bool) {
	userIDAny, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "用户未登录")
		return 0, false
	}
	uid, ok := userIDAny.(uint64)
	if !ok || uid == 0 {
		utils.Fail(c, 401, "用户未登录")
		return 0, false
	}
	return uid, true
}

// ListMyOperationLogs 当前用户自己的操作日志列表（强制按 user_id 过滤）
// GET /api/v1/user/logs
func (ctrl *ProfileController) ListMyOperationLogs(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	utils.SanitizeQueryParams(c)

	var query models.OperationLogQuery
	_ = c.ShouldBindQuery(&query)
	// 严禁看别人：忽略客户端传入的 user_id，强制覆盖为当前用户
	query.UserID = uid
	query.Username = ""

	defaultQueryDays := 30
	if settingsMap, err := models.GetSettingsMap([]string{"operation_log_query_days"}); err == nil {
		if v, ok := settingsMap["operation_log_query_days"]; ok {
			if parsed, parseErr := strconv.Atoi(v); parseErr == nil && parsed > 0 {
				defaultQueryDays = parsed
			}
		}
	}
	if defaultQueryDays > 365 {
		defaultQueryDays = 365
	}

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
		utils.Fail(c, 400, "参数错误")
		return
	}

	logs, total, err := models.GetOperationLogList(&query)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":      logs,
		"total":     total,
		"page":      query.Page,
		"page_size": query.PageSize,
	})
}

// GetMyOperationLogDetail 当前用户自己的操作日志详情
// GET /api/v1/user/logs/:id
func (ctrl *ProfileController) GetMyOperationLogDetail(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的 ID")
		return
	}
	item, err := models.GetOperationLogByID(id)
	if err != nil || item == nil {
		utils.Fail(c, 404, "记录不存在")
		return
	}
	// 归属校验：只能看自己的
	if item.UserID != uid {
		utils.Fail(c, 403, "无权查看该记录")
		return
	}
	utils.Success(c, item)
}

// ListMyAPILogs 当前用户自己的 API 访问日志列表（强制按 user_id 过滤）
// GET /api/v1/user/api-logs
func (ctrl *ProfileController) ListMyAPILogs(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	utils.SanitizeQueryParams(c)

	var query models.APIAccessLogQuery
	_ = c.ShouldBindQuery(&query)
	// 严禁看别人：强制覆盖 user_id
	query.UserID = uid
	query.Username = ""

	defaultQueryDays := services.GetGlobalAPILogRuntimeConfig().QueryDays
	if defaultQueryDays <= 0 {
		defaultQueryDays = 7
	}
	if defaultQueryDays > 365 {
		defaultQueryDays = 365
	}

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
		utils.Fail(c, 400, "参数错误")
		return
	}

	list, total, err := models.GetAPIAccessLogList(&query)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}

	utils.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      query.Page,
		"page_size": query.PageSize,
	})
}

// GetMyAPILogStats 当前用户自己的 API 访问日志统计（仅本人数据，不含其他用户信息）
// GET /api/v1/user/api-logs/stats
func (ctrl *ProfileController) GetMyAPILogStats(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		return
	}
	stats, err := models.GetAPIAccessLogStatsByUserID(uid)
	if err != nil {
		utils.Fail(c, 500, "统计失败")
		return
	}
	utils.Success(c, stats)
}

// GetMyAPILogDetail 当前用户自己的 API 访问日志详情
// GET /api/v1/user/api-logs/:id
func (ctrl *ProfileController) GetMyAPILogDetail(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
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
		utils.Fail(c, 404, "记录不存在")
		return
	}
	if item.UserID != uid {
		utils.Fail(c, 403, "无权查看该记录")
		return
	}
	utils.Success(c, item)
}
