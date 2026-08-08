package dashboard

import (
	"fmt"
	"fst/backend/app/models"
	"log"
	"time"

	"fst/backend/pkg/db"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DashboardController 仪表盘控制器
type DashboardController struct{}

// DashboardStatistics 仪表盘统计数据
type DashboardStatistics struct {
	// 用户统计
	TotalUsers       int64 `json:"total_users"`
	TodayNewUsers    int64 `json:"today_new_users"`
	TodayActiveUsers int64 `json:"today_active_users"`
	ActiveUsers7d    int64 `json:"active_users_7d"`

	// 余额/积分日志统计
	TotalMoneyLogs int64 `json:"total_money_logs"`
	TotalScoreLogs int64 `json:"total_score_logs"`

	// 操作日志统计
	TotalOperationLogs int64 `json:"total_operation_logs"`
	TodayOperationLogs int64 `json:"today_operation_logs"`

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
	ID               uint64  `json:"id" gorm:"column:id"`
	Username         string  `json:"username" gorm:"column:username"`
	Nickname         string  `json:"nickname" gorm:"column:nickname"`
	Email            string  `json:"email" gorm:"column:email"`
	Role             string  `json:"role" gorm:"column:role"`
	Status           int     `json:"status" gorm:"column:status"`
	Money            float64 `json:"money" gorm:"column:money"`
	TotalPaidAmount  float64 `json:"total_paid_amount" gorm:"column:total_paid_amount"`
	BalancePaidRatio float64 `json:"balance_paid_ratio" gorm:"column:balance_paid_ratio"`
	CreateTime       int64   `json:"create_time" gorm:"column:create_time"`
	LastLoginTime    *int64  `json:"last_login_time" gorm:"column:last_login_time"`
}

type DashboardTrendPoint struct {
	Date          string  `json:"date"`
	Label         string  `json:"label"`
	NewUsers      int64   `json:"new_users"`
	ActiveUsers   int64   `json:"active_users"`
	PaidOrders    int64   `json:"paid_orders"`
	PaidAmount    float64 `json:"paid_amount"`
	OperationLogs int64   `json:"operation_logs"`
	APILogs       int64   `json:"api_logs"`
	EmailLogs     int64   `json:"email_logs"`
	SMSLogs       int64   `json:"sms_logs"`
}

type dashboardRealnameStats struct {
	TotalCount    int64 `gorm:"column:total_count"`
	PendingCount  int64 `gorm:"column:pending_count"`
	ApprovedCount int64 `gorm:"column:approved_count"`
	RejectedCount int64 `gorm:"column:rejected_count"`
}

// unixDayKey 将 Unix 秒时间戳格式化为 YYYY-MM-DD（应用层聚合，避免 FROM_UNIXTIME 方言差异）
func unixDayKey(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Format("2006-01-02")
}

// aggregateUnixTimestamps 按天计数
func aggregateUnixTimestamps(timestamps []int64) map[string]int64 {
	result := make(map[string]int64)
	for _, ts := range timestamps {
		day := unixDayKey(ts)
		if day == "" {
			continue
		}
		result[day]++
	}
	return result
}

// loadTrendCountByUnixField 查原始时间戳后在 Go 层按天聚合
func loadTrendCountByUnixField(database *gorm.DB, table, field string, startUnix int64) (map[string]int64, error) {
	type tsRow struct {
		Ts int64 `gorm:"column:ts"`
	}
	var rows []tsRow
	q := fmt.Sprintf("SELECT %s AS ts FROM %s WHERE %s >= ?", field, table, field)
	if err := database.Raw(q, startUnix).Scan(&rows).Error; err != nil {
		log.Printf("[Dashboard] 趋势计数查询失败(%s.%s): %v", table, field, err)
		return map[string]int64{}, err
	}
	timestamps := make([]int64, 0, len(rows))
	for _, row := range rows {
		timestamps = append(timestamps, row.Ts)
	}
	return aggregateUnixTimestamps(timestamps), nil
}

// loadTrendCountByTimeField 按天聚合 time.Time 类型时间列（email_logs/sms_logs 的 created_at 非 unix 秒）
func loadTrendCountByTimeField(database *gorm.DB, table, field string, start time.Time) (map[string]int64, error) {
	type tRow struct {
		Ts time.Time `gorm:"column:ts"`
	}
	var rows []tRow
	q := fmt.Sprintf("SELECT %s AS ts FROM %s WHERE %s >= ?", field, table, field)
	if err := database.Raw(q, start).Scan(&rows).Error; err != nil {
		log.Printf("[Dashboard] 趋势计数查询失败(%s.%s): %v", table, field, err)
		return map[string]int64{}, err
	}
	result := make(map[string]int64)
	for _, row := range rows {
		if row.Ts.IsZero() {
			continue
		}
		result[row.Ts.Format("2006-01-02")]++
	}
	return result, nil
}

type dashboardPaidRow struct {
	PaidAt int64   `gorm:"column:paid_at"`
	Amount float64 `gorm:"column:pay_amount"`
}

// loadTrendPaidOrders 已支付订单按天计数与金额汇总（应用层聚合）
func loadTrendPaidOrders(database *gorm.DB, startUnix int64) (map[string]int64, map[string]float64, error) {
	var rows []dashboardPaidRow
	q := fmt.Sprintf(`
		SELECT paid_at, pay_amount
		FROM payment_orders
		WHERE status = ? AND paid_at IS NOT NULL AND %s AND paid_at >= ?`, models.RealPaidOrderFilterSQL)
	if err := database.Raw(q, models.PaymentStatusPaid, startUnix).Scan(&rows).Error; err != nil {
		log.Printf("[Dashboard] 趋势支付订单查询失败: %v", err)
		return map[string]int64{}, map[string]float64{}, err
	}
	countMap := make(map[string]int64)
	amountMap := make(map[string]float64)
	for _, row := range rows {
		day := unixDayKey(row.PaidAt)
		if day == "" {
			continue
		}
		countMap[day]++
		amountMap[day] += row.Amount
	}
	return countMap, amountMap, nil
}

// buildDashboardTrends 任一子查询失败时仍返回已拼好的趋势点（失败项为 0），并带回 error 供 tracker 上报。
func buildDashboardTrends(database *gorm.DB, start time.Time, days int) ([]DashboardTrendPoint, error) {
	startUnix := start.Unix()
	var firstErr error

	newUsers, err := loadTrendCountByUnixField(database, "users", "create_time", startUnix)
	if err != nil && firstErr == nil {
		firstErr = err
	}
	activeUsers, err := loadTrendCountByUnixField(database, "users", "last_login_time", startUnix)
	if err != nil && firstErr == nil {
		firstErr = err
	}
	paidOrders, paidAmount, err := loadTrendPaidOrders(database, startUnix)
	if err != nil && firstErr == nil {
		firstErr = err
	}
	operationLogs, err := loadTrendCountByUnixField(database, "operation_logs", "create_time", startUnix)
	if err != nil && firstErr == nil {
		firstErr = err
	}
	apiLogs, err := loadTrendCountByUnixField(database, "api_access_logs", "create_time", startUnix)
	if err != nil && firstErr == nil {
		firstErr = err
	}
	emailLogs, err := loadTrendCountByTimeField(database, "email_logs", "created_at", start)
	if err != nil && firstErr == nil {
		firstErr = err
	}
	smsLogs, err := loadTrendCountByTimeField(database, "sms_logs", "created_at", start)
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
			APILogs:       apiLogs[dayKey],
			EmailLogs:     emailLogs[dayKey],
			SMSLogs:       smsLogs[dayKey],
		})
	}
	return trends, firstErr
}

func loadDashboardUsers(database *gorm.DB, orderBy string) ([]RecentUser, error) {
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
	if err := database.Raw(query, models.PaymentStatusPaid).Scan(&users).Error; err != nil {
		log.Printf("[Dashboard] 最近用户列表查询失败: %v", err)
		return users, err
	}
	return users, nil
}

// NewDashboardController 创建仪表盘控制器实例
func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

// dashboardErrTracker 统一记录各指标查询失败：之前每条静默吞错误，
// DB 故障时接口仍返回 200 + 全零统计，运维/前端没法区分「真没数据」和「查库失败」。
type dashboardErrTracker struct {
	database *gorm.DB
	failed   []string
}

func (t *dashboardErrTracker) get(name string, dest any, query string, args ...any) {
	if err := t.database.Raw(query, args...).Scan(dest).Error; err != nil {
		log.Printf("[Dashboard] 查询指标 %s 失败: %v", name, err)
		t.failed = append(t.failed, name)
	}
}

// GetDashboard 获取仪表盘统计数据
// @Summary 管理端仪表盘
// @Tags Admin-仪表盘
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/dashboard [get]
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

	// 用户统计（排除软删，与 total_user_balance 一致）
	tracker.get("total_users", &stats.TotalUsers, "SELECT COUNT(*) FROM users WHERE delete_time IS NULL")
	tracker.get("today_new_users", &stats.TodayNewUsers, "SELECT COUNT(*) FROM users WHERE delete_time IS NULL AND create_time >= ?", todayStartUnix)
	tracker.get("today_active_users", &stats.TodayActiveUsers, "SELECT COUNT(*) FROM users WHERE delete_time IS NULL AND last_login_time >= ?", todayStartUnix)
	tracker.get("active_users_7d", &stats.ActiveUsers7d, "SELECT COUNT(*) FROM users WHERE delete_time IS NULL AND last_login_time >= ?", sevenDaysAgoUnix)
	tracker.get("total_user_balance", &stats.TotalUserBalance, "SELECT COALESCE(SUM(money), 0) FROM users WHERE delete_time IS NULL")

	// 日志统计（排除软删）
	tracker.get("total_money_logs", &stats.TotalMoneyLogs, "SELECT COUNT(*) FROM user_money_logs WHERE delete_time IS NULL")
	tracker.get("total_score_logs", &stats.TotalScoreLogs, "SELECT COUNT(*) FROM user_score_logs WHERE delete_time IS NULL")

	// 操作日志
	tracker.get("total_operation_logs", &stats.TotalOperationLogs, "SELECT COUNT(*) FROM operation_logs")
	tracker.get("today_operation_logs", &stats.TodayOperationLogs, "SELECT COUNT(*) FROM operation_logs WHERE create_time >= ?", todayStartUnix)

	// 活跃会话
	// 活跃会话（与在线模块口径一致：未撤销 + 心跳窗口内）
	tracker.get("active_sessions", &stats.ActiveSessions,
		"SELECT COUNT(*) FROM user_sessions WHERE is_active = ? AND last_seen_at >= ?",
		true, now.Unix()-models.GetOnlineHeartbeatGraceSeconds())

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
		"partial_ok":         len(tracker.failed) == 0,
		"failed_metrics":     tracker.failed,
	})
}
