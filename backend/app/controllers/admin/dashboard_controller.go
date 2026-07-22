package admin

import (
	"fmt"
	"fst/backend/app/models"
	"log"
	"time"

	"fst/backend/pkg/db"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// DashboardController 仪表盘控制器
type DashboardController struct{}

// DashboardStatistics 仪表盘统计数据
type DashboardStatistics struct {
	// 用户统计
	TotalUsers    int64 `json:"total_users"`
	TodayNewUsers int64 `json:"today_new_users"`
	TodayActiveUsers int64 `json:"today_active_users"`
	ActiveUsers7d int64 `json:"active_users_7d"`

	// 余额/积分日志统计
	TotalMoneyLogs int64 `json:"total_money_logs"`
	TotalScoreLogs int64 `json:"total_score_logs"`

	// 操作日志统计
	TotalOperationLogs  int64 `json:"total_operation_logs"`
	TodayOperationLogs  int64 `json:"today_operation_logs"`

	// 在线会话
	ActiveSessions int64 `json:"active_sessions"`

	TotalPaymentOrders   int64   `json:"total_payment_orders"`
	PaidPaymentOrders    int64   `json:"paid_payment_orders"`
	PendingPaymentOrders int64   `json:"pending_payment_orders"`
	TotalPaymentAmount   float64 `json:"total_payment_amount"`
	TodayPaymentOrders   int64   `json:"today_payment_orders"`
	TodayPaymentAmount   float64 `json:"today_payment_amount"`
	MonthPaymentAmount   float64 `json:"month_payment_amount"`
	YearPaymentAmount    float64 `json:"year_payment_amount"`
	TotalUserBalance     float64 `json:"total_user_balance"`

	PendingWithdrawCount  int64   `json:"pending_withdraw_count"`
	ApprovedWithdrawCount int64   `json:"approved_withdraw_count"`
	PaidWithdrawCount     int64   `json:"paid_withdraw_count"`
	PaidWithdrawAmount    float64 `json:"paid_withdraw_amount"`

	TotalRealnameRequests int64 `json:"total_realname_requests"`
	PendingRealnameCount  int64 `json:"pending_realname_count"`
	ApprovedRealnameCount int64 `json:"approved_realname_count"`
	RejectedRealnameCount int64 `json:"rejected_realname_count"`
}

// RecentUser 最近注册用户
type RecentUser struct {
	ID            uint64  `json:"id" db:"id"`
	Username      string  `json:"username" db:"username"`
	Nickname      string  `json:"nickname" db:"nickname"`
	Email         string  `json:"email" db:"email"`
	Role          string  `json:"role" db:"role"`
	Status        int     `json:"status" db:"status"`
	Money         float64 `json:"money" db:"money"`
	TotalPaidAmount float64 `json:"total_paid_amount" db:"total_paid_amount"`
	BalancePaidRatio float64 `json:"balance_paid_ratio" db:"balance_paid_ratio"`
	CreateTime    int64   `json:"create_time" db:"create_time"`
	LastLoginTime *int64  `json:"last_login_time" db:"last_login_time"`
}

type DashboardTrendPoint struct {
	Date          string  `json:"date"`
	Label         string  `json:"label"`
	NewUsers      int64   `json:"new_users"`
	ActiveUsers   int64   `json:"active_users"`
	PaidOrders    int64   `json:"paid_orders"`
	PaidAmount    float64 `json:"paid_amount"`
	OperationLogs int64   `json:"operation_logs"`
}

type dashboardCountRow struct {
	Day   string `db:"day"`
	Value int64  `db:"value"`
}

type dashboardAmountRow struct {
	Day   string  `db:"day"`
	Value float64 `db:"value"`
}

type dashboardRealnameStats struct {
	TotalCount    int64 `db:"total_count"`
	PendingCount  int64 `db:"pending_count"`
	ApprovedCount int64 `db:"approved_count"`
	RejectedCount int64 `db:"rejected_count"`
}

func loadDashboardCountMap(database *sqlx.DB, query string, args ...any) (map[string]int64, error) {
	rows := make([]dashboardCountRow, 0)
	if err := database.Select(&rows, query, args...); err != nil {
		log.Printf("[Dashboard] 趋势计数查询失败: %v", err)
		return map[string]int64{}, err
	}
	result := make(map[string]int64, len(rows))
	for _, row := range rows {
		result[row.Day] = row.Value
	}
	return result, nil
}

func loadDashboardAmountMap(database *sqlx.DB, query string, args ...any) (map[string]float64, error) {
	rows := make([]dashboardAmountRow, 0)
	if err := database.Select(&rows, query, args...); err != nil {
		log.Printf("[Dashboard] 趋势金额查询失败: %v", err)
		return map[string]float64{}, err
	}
	result := make(map[string]float64, len(rows))
	for _, row := range rows {
		result[row.Day] = row.Value
	}
	return result, nil
}

// buildDashboardTrends 任一子查询失败时仍返回已拼好的趋势点（失败项为 0），并带回 error 供 tracker 上报。
func buildDashboardTrends(database *sqlx.DB, start time.Time, days int) ([]DashboardTrendPoint, error) {
	startUnix := start.Unix()
	var firstErr error
	// db.Q：SQLite 下 DATE(FROM_UNIXTIME(...)) → date(..., 'unixepoch')
	newUsers, err := loadDashboardCountMap(database, db.Q("SELECT DATE(FROM_UNIXTIME(create_time)) AS day, COUNT(*) AS value FROM users WHERE create_time >= ? GROUP BY DATE(FROM_UNIXTIME(create_time))"), startUnix)
	if err != nil && firstErr == nil {
		firstErr = err
	}
	activeUsers, err := loadDashboardCountMap(database, db.Q("SELECT DATE(FROM_UNIXTIME(last_login_time)) AS day, COUNT(*) AS value FROM users WHERE last_login_time >= ? GROUP BY DATE(FROM_UNIXTIME(last_login_time))"), startUnix)
	if err != nil && firstErr == nil {
		firstErr = err
	}
	paidOrders, err := loadDashboardCountMap(database, db.Q(fmt.Sprintf("SELECT DATE(FROM_UNIXTIME(paid_at)) AS day, COUNT(*) AS value FROM payment_orders WHERE status = ? AND paid_at IS NOT NULL AND %s AND paid_at >= ? GROUP BY DATE(FROM_UNIXTIME(paid_at))", models.RealPaidOrderFilterSQL)), models.PaymentStatusPaid, startUnix)
	if err != nil && firstErr == nil {
		firstErr = err
	}
	paidAmount, err := loadDashboardAmountMap(database, db.Q(fmt.Sprintf("SELECT DATE(FROM_UNIXTIME(paid_at)) AS day, COALESCE(SUM(pay_amount), 0) AS value FROM payment_orders WHERE status = ? AND paid_at IS NOT NULL AND %s AND paid_at >= ? GROUP BY DATE(FROM_UNIXTIME(paid_at))", models.RealPaidOrderFilterSQL)), models.PaymentStatusPaid, startUnix)
	if err != nil && firstErr == nil {
		firstErr = err
	}
	operationLogs, err := loadDashboardCountMap(database, db.Q("SELECT DATE(FROM_UNIXTIME(create_time)) AS day, COUNT(*) AS value FROM operation_logs WHERE create_time >= ? GROUP BY DATE(FROM_UNIXTIME(create_time))"), startUnix)
	if err != nil && firstErr == nil {
		firstErr = err
	}

	trends := make([]DashboardTrendPoint, 0, days)
	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i)
		dayKey := day.Format("2006-01-02")
		trends = append(trends, DashboardTrendPoint{
			Date:          dayKey,
			Label:         day.Format("01-02"),
			NewUsers:      newUsers[dayKey],
			ActiveUsers:   activeUsers[dayKey],
			PaidOrders:    paidOrders[dayKey],
			PaidAmount:    paidAmount[dayKey],
			OperationLogs: operationLogs[dayKey],
		})
	}
	return trends, firstErr
}

func loadDashboardUsers(database *sqlx.DB, orderBy string) ([]RecentUser, error) {
	users := make([]RecentUser, 0)
	query := fmt.Sprintf(`SELECT
		u.id,
		u.username,
		COALESCE(u.nickname, '') AS nickname,
		COALESCE(u.email, '') AS email,
		u.role,
		u.status,
		COALESCE(u.money, 0) AS money,
		COALESCE(p.total_paid_amount, 0) AS total_paid_amount,
		CASE WHEN COALESCE(p.total_paid_amount, 0) > 0 THEN COALESCE(u.money, 0) / p.total_paid_amount ELSE 0 END AS balance_paid_ratio,
		u.create_time,
		u.last_login_time
	FROM users u
	LEFT JOIN (
		SELECT user_id, COALESCE(SUM(pay_amount), 0) AS total_paid_amount
		FROM payment_orders
		WHERE status = ? AND %s
		GROUP BY user_id
	) p ON p.user_id = u.id
	WHERE u.delete_time IS NULL
	ORDER BY %s
	LIMIT 5`, models.RealPaidOrderFilterSQL, orderBy)
	if err := database.Select(&users, query, models.PaymentStatusPaid); err != nil {
		log.Printf("[Dashboard] 最近用户列表查询失败: %v", err)
		return users, err
	}
	return users, nil
}

// NewDashboardController 创建仪表盘控制器实例
func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

// dashboardErrTracker 统一记录各指标查询失败：之前每条 `_ = database.Get(...)` 都静默吞错误，
// DB 故障时接口仍返回 200 + 全零统计，运维/前端没法区分「真没数据」和「查库失败」。
// 现在失败会打日志，并通过响应里的 partial_ok / failed_metrics 告知前端。
type dashboardErrTracker struct {
	database *sqlx.DB
	failed   []string
}

func (t *dashboardErrTracker) get(name string, dest any, query string, args ...any) {
	if err := t.database.Get(dest, query, args...); err != nil {
		log.Printf("[Dashboard] 查询指标 %s 失败: %v", name, err)
		t.failed = append(t.failed, name)
	}
}

// GetDashboard 获取仪表盘统计数据
func (ctrl *DashboardController) GetDashboard(ctx *gin.Context) {
	database := db.GetDB()
	stats := DashboardStatistics{}
	if database == nil {
		utils.Success(ctx, gin.H{
			"statistics":         stats,
			"recent_users":       []RecentUser{},
			"recent_login_users": []RecentUser{},
			"trends":             []DashboardTrendPoint{},
		})
		return
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	sevenDaysAgo := now.AddDate(0, 0, -7)
	trendStart := todayStart.AddDate(0, 0, -6)

	todayStartUnix := todayStart.Unix()
	sevenDaysAgoUnix := sevenDaysAgo.Unix()

	tracker := &dashboardErrTracker{database: database}

	// 用户统计
	tracker.get("total_users", &stats.TotalUsers, "SELECT COUNT(*) FROM users")
	tracker.get("today_new_users", &stats.TodayNewUsers, "SELECT COUNT(*) FROM users WHERE create_time >= ?", todayStartUnix)
	tracker.get("today_active_users", &stats.TodayActiveUsers, "SELECT COUNT(*) FROM users WHERE last_login_time >= ?", todayStartUnix)
	tracker.get("active_users_7d", &stats.ActiveUsers7d, "SELECT COUNT(*) FROM users WHERE last_login_time >= ?", sevenDaysAgoUnix)
	tracker.get("total_user_balance", &stats.TotalUserBalance, "SELECT COALESCE(SUM(money), 0) FROM users WHERE delete_time IS NULL")

	// 日志统计
	tracker.get("total_money_logs", &stats.TotalMoneyLogs, "SELECT COUNT(*) FROM user_money_logs")
	tracker.get("total_score_logs", &stats.TotalScoreLogs, "SELECT COUNT(*) FROM user_score_logs")

	// 操作日志
	tracker.get("total_operation_logs", &stats.TotalOperationLogs, "SELECT COUNT(*) FROM operation_logs")
	tracker.get("today_operation_logs", &stats.TodayOperationLogs, "SELECT COUNT(*) FROM operation_logs WHERE create_time >= ?", todayStartUnix)

	// 活跃会话
	tracker.get("active_sessions", &stats.ActiveSessions, "SELECT COUNT(*) FROM user_sessions WHERE expires_at > ?", now.Unix())

	if paymentStats, err := models.GetPaymentStats(); err == nil && paymentStats != nil {
		stats.TotalPaymentOrders = paymentStats.TotalOrders
		stats.PaidPaymentOrders = paymentStats.PaidOrders
		stats.PendingPaymentOrders = paymentStats.PendingOrders
		stats.TotalPaymentAmount = paymentStats.TotalAmount
		stats.TodayPaymentOrders = paymentStats.TodayOrders
		stats.TodayPaymentAmount = paymentStats.TodayAmount
		stats.MonthPaymentAmount = paymentStats.MonthAmount
		stats.YearPaymentAmount = paymentStats.YearAmount
	} else if err != nil {
		log.Printf("[Dashboard] 查询指标 payment_stats 失败: %v", err)
		tracker.failed = append(tracker.failed, "payment_stats")
	}

	if withdrawStats, err := models.GetWithdrawRequestStats(&models.WithdrawListQuery{}); err == nil && withdrawStats != nil {
		stats.PendingWithdrawCount = withdrawStats.PendingCount
		stats.ApprovedWithdrawCount = withdrawStats.ApprovedCount
		stats.PaidWithdrawCount = withdrawStats.PaidCount
		stats.PaidWithdrawAmount = withdrawStats.PaidAmount
	} else if err != nil {
		log.Printf("[Dashboard] 查询指标 withdraw_stats 失败: %v", err)
		tracker.failed = append(tracker.failed, "withdraw_stats")
	}

	realnameStats := dashboardRealnameStats{}
	tracker.get("realname_stats", &realnameStats, `SELECT
		COUNT(*) AS total_count,
		COALESCE(SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END), 0) AS pending_count,
		COALESCE(SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END), 0) AS approved_count,
		COALESCE(SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END), 0) AS rejected_count
		FROM user_realname_verifications
		WHERE delete_time IS NULL`)
	stats.TotalRealnameRequests = realnameStats.TotalCount
	stats.PendingRealnameCount = realnameStats.PendingCount
	stats.ApprovedRealnameCount = realnameStats.ApprovedCount
	stats.RejectedRealnameCount = realnameStats.RejectedCount

	recentUsers, err := loadDashboardUsers(database, "u.create_time DESC")
	if err != nil {
		tracker.failed = append(tracker.failed, "recent_users")
	}
	recentLoginUsers, err := loadDashboardUsers(database, "u.last_login_time DESC, u.create_time DESC")
	if err != nil {
		tracker.failed = append(tracker.failed, "recent_login_users")
	}
	trends, err := buildDashboardTrends(database, trendStart, 7)
	if err != nil {
		tracker.failed = append(tracker.failed, "trends")
	}

	utils.Success(ctx, gin.H{
		"statistics":         stats,
		"recent_users":       recentUsers,
		"recent_login_users": recentLoginUsers,
		"trends":             trends,
		// partial_ok=false 时表示部分指标查询失败（已用 0 兜底展示），failed_metrics 列出具体哪些，
		// 前端可据此提示「部分数据可能不准确」，而不是让运维误以为「就是没数据」。
		"partial_ok":     len(tracker.failed) == 0,
		"failed_metrics": tracker.failed,
	})
}
