package apilog

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// APILogController API接口日志管理控制器
type APILogController struct{}

func NewAPILogController() *APILogController {
	return &APILogController{}
}

const apiLogStatsCacheTTL = 15 * time.Second

var apiLogStatsCache struct {
	mu        sync.RWMutex
	expiresAt time.Time
	data      *models.APIAccessLogStats
}

func cloneAPILogStats(stats *models.APIAccessLogStats) *models.APIAccessLogStats {
	if stats == nil {
		return nil
	}

	cloned := *stats
	cloned.TopPaths = append([]models.APIAccessPathStat(nil), stats.TopPaths...)
	cloned.MethodStats = append([]models.APIAccessMethodStat(nil), stats.MethodStats...)
	cloned.SceneStats = append([]models.APIAccessSceneStat(nil), stats.SceneStats...)
	return &cloned
}

func getCachedAPILogStats() (*models.APIAccessLogStats, error) {
	now := time.Now()

	apiLogStatsCache.mu.RLock()
	if apiLogStatsCache.data != nil && now.Before(apiLogStatsCache.expiresAt) {
		cached := cloneAPILogStats(apiLogStatsCache.data)
		apiLogStatsCache.mu.RUnlock()
		return cached, nil
	}
	apiLogStatsCache.mu.RUnlock()

	stats, err := models.GetAPIAccessLogStats()
	if err != nil {
		return nil, err
	}

	apiLogStatsCache.mu.Lock()
	apiLogStatsCache.data = cloneAPILogStats(stats)
	apiLogStatsCache.expiresAt = now.Add(apiLogStatsCacheTTL)
	apiLogStatsCache.mu.Unlock()

	return stats, nil
}

// InvalidateAPILogStatsCache 清空 API 日志统计缓存，供设置/清理接口跨包调用。
func InvalidateAPILogStatsCache() {
	apiLogStatsCache.mu.Lock()
	apiLogStatsCache.data = nil
	apiLogStatsCache.expiresAt = time.Time{}
	apiLogStatsCache.mu.Unlock()
}

// List API接口日志列表
// @Summary API接口日志列表
// @Tags Admin-API日志
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/api-logs [get]
func (ctrl *APILogController) List(c *gin.Context) {
	utils.SanitizeQueryParams(c)

	var query models.APIAccessLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.Fail(c, 400, "Invalid parameters")
		return
	}

	defaultQueryDays := services.GetGlobalAPILogRuntimeConfig().QueryDays
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

// Detail API接口日志详情
// @Summary API接口日志详情
// @Tags Admin-API日志
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/api-logs/{id} [get]
func (ctrl *APILogController) Detail(c *gin.Context) {
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
	if err != nil {
		utils.Fail(c, 404, "Record does not exist")
		return
	}

	utils.Success(c, item)
}

// Stats API接口日志统计
// @Summary API接口日志统计
// @Tags Admin-API日志
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/api-logs/stats [get]
func (ctrl *APILogController) Stats(c *gin.Context) {
	stats, err := getCachedAPILogStats()
	if err != nil {
		utils.Fail(c, 500, "Statistics failed")
		return
	}
	utils.Success(c, stats)
}

// Clean 清理API接口日志
// @Summary 清理API接口日志
// @Tags Admin-API日志
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/api-logs/clean [post]
func (ctrl *APILogController) Clean(c *gin.Context) {
	var req struct {
		BeforeTime int64 `json:"before_time" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "Invalid parameters")
		return
	}

	affected, err := models.DeleteAPIAccessLogsBefore(req.BeforeTime)
	if err != nil {
		utils.Fail(c, 500, "Cleanup failed")
		return
	}

	InvalidateAPILogStatsCache()

	utils.Success(c, gin.H{
		"affected": affected,
	})
}
