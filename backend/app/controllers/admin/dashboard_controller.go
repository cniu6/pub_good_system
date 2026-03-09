package admin

import (
	"time"

	"fst/backend/internal/db"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// DashboardStatistics 仪表盘统计数据
type DashboardStatistics struct {
	// 用户统计
	TotalUsers    int64 `json:"total_users"`
	TodayNewUsers int64 `json:"today_new_users"`
	ActiveUsers7d int64 `json:"active_users_7d"`

	// 余额/积分日志统计
	TotalMoneyLogs int64 `json:"total_money_logs"`
	TotalScoreLogs int64 `json:"total_score_logs"`

	// 操作日志统计
	TotalOperationLogs  int64 `json:"total_operation_logs"`
	TodayOperationLogs  int64 `json:"today_operation_logs"`

	// 在线会话
	ActiveSessions int64 `json:"active_sessions"`
}

// RecentUser 最近注册用户
type RecentUser struct {
	ID            uint64  `json:"id" db:"id"`
	Username      string  `json:"username" db:"username"`
	Nickname      string  `json:"nickname" db:"nickname"`
	Email         string  `json:"email" db:"email"`
	Role          string  `json:"role" db:"role"`
	Status        int     `json:"status" db:"status"`
	CreateTime    int64   `json:"create_time" db:"create_time"`
	LastLoginTime *int64  `json:"last_login_time" db:"last_login_time"`
}

// GetDashboard 获取仪表盘统计数据
func GetDashboard(ctx *gin.Context) {
	database := db.GetDB()
	stats := DashboardStatistics{}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	sevenDaysAgo := now.AddDate(0, 0, -7)

	todayStartUnix := todayStart.Unix()
	sevenDaysAgoUnix := sevenDaysAgo.Unix()

	// 用户统计
	_ = database.Get(&stats.TotalUsers, "SELECT COUNT(*) FROM users")
	_ = database.Get(&stats.TodayNewUsers, "SELECT COUNT(*) FROM users WHERE create_time >= ?", todayStartUnix)
	_ = database.Get(&stats.ActiveUsers7d, "SELECT COUNT(*) FROM users WHERE last_login_time >= ?", sevenDaysAgoUnix)

	// 日志统计
	_ = database.Get(&stats.TotalMoneyLogs, "SELECT COUNT(*) FROM user_money_logs")
	_ = database.Get(&stats.TotalScoreLogs, "SELECT COUNT(*) FROM user_score_logs")

	// 操作日志
	_ = database.Get(&stats.TotalOperationLogs, "SELECT COUNT(*) FROM operation_logs")
	_ = database.Get(&stats.TodayOperationLogs, "SELECT COUNT(*) FROM operation_logs WHERE create_time >= ?", todayStartUnix)

	// 活跃会话
	_ = database.Get(&stats.ActiveSessions, "SELECT COUNT(*) FROM user_sessions WHERE expires_at > ?", now.Unix())

	// 最近注册用户
	var recentUsers []RecentUser
	_ = database.Select(&recentUsers, "SELECT id, username, COALESCE(nickname,'') as nickname, COALESCE(email,'') as email, role, status, create_time, last_login_time FROM users ORDER BY create_time DESC LIMIT 5")

	utils.Success(ctx, gin.H{
		"statistics":   stats,
		"recent_users": recentUsers,
	})
}
