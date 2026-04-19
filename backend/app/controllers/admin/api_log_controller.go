package admin

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

func invalidateAPILogStatsCache() {
	apiLogStatsCache.mu.Lock()
	apiLogStatsCache.data = nil
	apiLogStatsCache.expiresAt = time.Time{}
	apiLogStatsCache.mu.Unlock()
}

// List API接口日志列表
func (ctrl *APILogController) List(c *gin.Context) {
	utils.SanitizeQueryParams(c)

	var query models.APIAccessLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

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

// Detail API接口日志详情
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
		utils.Fail(c, 404, "记录不存在")
		return
	}

	utils.Success(c, item)
}

// Stats API接口日志统计
func (ctrl *APILogController) Stats(c *gin.Context) {
	stats, err := getCachedAPILogStats()
	if err != nil {
		utils.Fail(c, 500, "统计失败")
		return
	}
	utils.Success(c, stats)
}

// Clean 清理API接口日志
func (ctrl *APILogController) Clean(c *gin.Context) {
	var req struct {
		BeforeTime int64 `json:"before_time" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误")
		return
	}

	affected, err := models.DeleteAPIAccessLogsBefore(req.BeforeTime)
	if err != nil {
		utils.Fail(c, 500, "清理失败")
		return
	}

	invalidateAPILogStatsCache()

	utils.Success(c, gin.H{
		"affected": affected,
	})
}
