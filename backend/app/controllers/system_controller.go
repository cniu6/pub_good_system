package controllers

import (
	"fst/backend/app/models"
	"fst/backend/internal/task"
	"fst/backend/pkg/config"
	"fst/backend/pkg/presence"
	"fst/backend/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SystemController struct{}

// getCleanupIntervalMinutes 读取清理间隔（分钟）
func getCleanupIntervalMinutes() int {
	cfg := config.GlobalConfig
	if cfg == nil {
		return 10
	}
	interval := cfg.CleanupIntervalMinutes
	if interval <= 0 {
		return 10
	}
	return interval
}

// GetCleanupStatus 查询验证码清理任务的运行状态
// @Summary 获取清理任务状态
// @Description 返回验证码清理任务的运行状态、间隔、上次/下次执行时间
// @Tags 系统管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/system/cleanup-status [get]
func (ctrl *SystemController) GetCleanupStatus(c *gin.Context) {
	intervalMinutes := getCleanupIntervalMinutes()
	result := map[string]interface{}{
		"running":          false,
		"interval_minutes": intervalMinutes,
	}

	// 正式调度与执行记录由 internal/task（cleanup_sessions_codes）负责
	def, err := task.GetDefinition("cleanup_sessions_codes")
	if err == nil && def != nil {
		result["running"] = def.LastStatus == task.StatusRunning
		result["last_status"] = def.LastStatus
		result["last_message"] = def.LastError
		if def.LastFinishedAt > 0 {
			t := time.Unix(def.LastFinishedAt, 0)
			result["last_cleanup_time"] = t.Format("2006-01-02 15:04:05")
			next := t.Add(time.Duration(def.IntervalSeconds) * time.Second)
			if def.IntervalSeconds <= 0 {
				next = t.Add(time.Duration(intervalMinutes) * time.Minute)
			}
			result["next_cleanup_time"] = next.Format("2006-01-02 15:04:05")
		}
		if def.IntervalSeconds > 0 {
			result["interval_minutes"] = def.IntervalSeconds / 60
		}
	}

	utils.Success(c, result)
}

// CreatePresenceTicket 签发 Presence WebSocket 一次性短时票据（JWT 不进 URL）
// @Summary 获取 Presence WS 票据
// @Tags 系统管理
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/system/ws-ticket [post]
func (ctrl *SystemController) CreatePresenceTicket(c *gin.Context) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		utils.Fail(c, 401, "未登录")
		return
	}
	userID, _ := userIDVal.(uint64)
	guardVal, _ := c.Get("authGuard")
	guard, _ := guardVal.(string)
	if guard == "" {
		guard = utils.UserAuthGuard
	}

	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		utils.Fail(c, 401, "需要 Bearer Token")
		return
	}
	sess, err := models.GetActiveSessionByTokenHash(utils.HashToken(strings.TrimSpace(parts[1])))
	if err != nil || sess == nil || sess.UserID != userID || sess.AuthGuard != guard {
		utils.Fail(c, 401, "会话无效")
		return
	}
	if !presence.AllowWSTicketIssue(userID, guard) {
		utils.Fail(c, 429, "WebSocket 连接请求过于频繁，请稍后重试")
		return
	}

	ticket, exp, err := presence.IssueWSTicket(userID, guard, sess.ID)
	if err != nil {
		utils.Fail(c, 500, "签发票据失败")
		return
	}
	utils.Success(c, gin.H{
		"ticket":     ticket,
		"expires_at": exp.Unix(),
	})
}
