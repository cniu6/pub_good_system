//go:build integration

package db_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"fst/backend/app/controllers/admin/dashboard"
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/internal/migrate"
	"fst/backend/internal/task"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"

	"github.com/gin-gonic/gin"
)

// TestCrossDBGormQueries 在真实 PostgreSQL / MySQL 上跑一遍由原生 SQL 改写为 GORM 的查询，
// 确认生成的 SQL 在各库都合法且结果一致。SQLite 由常规单元测试覆盖。
//
//	set FST_PG_DSN=postgres://postgres@127.0.0.1:55432/fst_test?sslmode=disable
//	set FST_MYSQL_DSN=root@tcp(127.0.0.1:53306)/fst_test?charset=utf8mb4&parseTime=True&loc=Local
//	go test -tags integration ./backend/pkg/db/ -run CrossDB -count=1 -v
//
// 注意：会删除目标库中的全部表，只能指向测试库。
func TestCrossDBGormQueries(t *testing.T) {
	drivers := []struct{ name, env string }{
		{"postgres", "FST_PG_DSN"},
		{"mysql", "FST_MYSQL_DSN"},
	}
	for _, d := range drivers {
		t.Run(d.name, func(t *testing.T) {
			dsn := strings.TrimSpace(os.Getenv(d.env))
			if dsn == "" {
				t.Skipf("未设置 %s，跳过", d.env)
			}
			setupCrossDB(t, d.name, dsn)
			runCrossDBChecks(t)
		})
	}
}

func setupCrossDB(t *testing.T, driver, dsn string) {
	t.Helper()
	t.Setenv("DB_DRIVER", driver)
	t.Setenv("JWT_SECRET", "test-jwt-secret-at-least-16-chars")
	t.Setenv("ADMIN_JWT_SECRET", "test-admin-jwt-secret-16")
	t.Setenv("CORS_ORIGINS", "*")
	t.Setenv("APP_ENV", "development")
	config.InitConfig()
	config.UpdateGlobalConfig(func(cfg *config.Config) {
		cfg.DBDriver = driver
		cfg.DBDSN = dsn
	})
	db.InitDB()
	t.Cleanup(func() { _ = db.Close() })

	tables, err := db.DB.Migrator().GetTables()
	if err != nil {
		t.Fatalf("列出表失败: %v", err)
	}
	for _, name := range tables {
		if err := db.DB.Migrator().DropTable(name); err != nil {
			t.Fatalf("清空测试库失败(%s): %v", name, err)
		}
	}
	migrate.RunAutoMigrate()
}

func must(t *testing.T, what string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", what, err)
	}
}

func runCrossDBChecks(t *testing.T) {
	now := time.Now().Unix()
	user := &models.User{Username: "xdb-user", Nickname: "跨库", Email: "xdb@example.test", Password: "x", Money: 50, Role: "user", Status: 1}
	must(t, "CreateUser", models.CreateUser(user))
	paidAt := now
	must(t, "CreatePaymentOrder", models.CreatePaymentOrder(&models.PaymentOrder{
		OrderNo: "XDB1", UserID: user.ID, GatewayID: 1, TradeNo: "T-XDB-1", PaymentChannel: "epay",
		PaymentType: "alipay", Amount: 20, PayAmount: 20, Subject: "xdb", Status: models.PaymentStatusPaid,
		PaidAt: &paidAt, ExpireAt: now + 3600,
	}))

	t.Run("用户列表与详情", func(t *testing.T) {
		svc := services.NewUserService()
		status := uint8(1)
		res, err := svc.GetList(&services.UserListQuery{Keyword: "xdb", Status: &status})
		must(t, "GetList", err)
		if res.Total != 1 || len(res.List) != 1 || res.List[0].TotalPaidAmount != 20 {
			t.Fatalf("GetList 结果不符: total=%d list=%+v", res.Total, res.List)
		}
		detail, err := svc.GetAdminDetail(user.ID)
		must(t, "GetAdminDetail", err)
		if detail.ID != user.ID || detail.BalancePaidRatio != 2.5 {
			t.Fatalf("GetAdminDetail 结果不符: %+v", detail)
		}
		simple, err := svc.BatchGetUserSimpleInfo([]uint64{user.ID})
		must(t, "BatchGetUserSimpleInfo", err)
		if simple[user.ID].Username != user.Username {
			t.Fatalf("BatchGetUserSimpleInfo 结果不符: %+v", simple)
		}
	})

	t.Run("在线会话", func(t *testing.T) {
		must(t, "CreateUserSession", models.CreateUserSession(user.ID, "user", "th-xdb", "rh-xdb", "127.0.0.1", "ua", "pc", "web", "", now+3600, now+7200))
		list, total, err := models.ListOnlineSessions(user.Username, "web", "user", 1, 20)
		must(t, "ListOnlineSessions", err)
		if total != 1 || len(list) != 1 || list[0].Username != user.Username {
			t.Fatalf("ListOnlineSessions 结果不符: total=%d list=%+v", total, list)
		}
		grouped, gTotal, err := models.ListOnlineUsersGrouped("", "", "", 1, 20)
		must(t, "ListOnlineUsersGrouped", err)
		if gTotal != 1 || len(grouped) != 1 {
			t.Fatalf("ListOnlineUsersGrouped 结果不符: total=%d", gTotal)
		}
	})

	t.Run("公告可见与已读", func(t *testing.T) {
		id, err := models.CreateAnnouncement(&models.Announcement{Title: "xdb", Content: "hello"})
		must(t, "CreateAnnouncement", err)
		must(t, "SetAnnouncementStatus", models.SetAnnouncementStatus(id, models.AnnouncementStatusPublished, 1))
		list, err := models.ListVisibleAnnouncementsForUser(user.ID, "user", true, false, 10)
		must(t, "ListVisibleAnnouncementsForUser", err)
		if len(list) != 1 || list[0].IsRead {
			t.Fatalf("可见未读公告应为 1 条: %+v", list)
		}
		must(t, "MarkAllAnnouncementsRead", models.MarkAllAnnouncementsRead(user.ID, "user"))
		n, err := models.CountUnreadAnnouncements(user.ID, "user")
		must(t, "CountUnreadAnnouncements", err)
		item, err := models.GetVisibleAnnouncementForUser(user.ID, id, "user")
		must(t, "GetVisibleAnnouncementForUser", err)
		if n != 0 || !item.IsRead {
			t.Fatalf("全部已读后未读数应为 0: unread=%d isRead=%v", n, item.IsRead)
		}
	})

	t.Run("支付统计与仪表盘", func(t *testing.T) {
		stats, err := models.GetPaymentStats()
		must(t, "GetPaymentStats", err)
		if stats.PaidOrders != 1 || stats.TodayAmount != 20 {
			t.Fatalf("GetPaymentStats 结果不符: %+v", stats)
		}
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		dashboard.NewDashboardController().GetDashboard(c)
		var resp struct {
			Data struct {
				PartialOK     bool     `json:"partial_ok"`
				FailedMetrics []string `json:"failed_metrics"`
				RecentUsers   []any    `json:"recent_users"`
			} `json:"data"`
		}
		must(t, "解析仪表盘响应", json.Unmarshal(w.Body.Bytes(), &resp))
		if !resp.Data.PartialOK || len(resp.Data.RecentUsers) == 0 {
			t.Fatalf("仪表盘指标查询失败: %s", w.Body.String())
		}
	})

	t.Run("日志聚合回填与统计", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			ts := now - int64(i)
			must(t, "插入 API 日志", db.DB.Create(&models.APIAccessLog{
				RequestID: fmt.Sprintf("xdb-req-%d", i), UserID: user.ID, AuthMethod: "apikey",
				Method: "GET", Path: "/x", StatusCode: 200 + i*300, Duration: 10, IP: "127.0.0.1", CreateTime: &ts,
			}).Error)
			must(t, "插入操作日志", db.DB.Create(&models.OperationLog{
				UserID: user.ID, Module: "m", Action: "a", Method: "POST", StatusCode: 200, CreateTime: &ts,
			}).Error)
			must(t, "插入邮件日志", db.DB.Create(&models.EmailLog{ToEmail: "a@x.test", Subject: "s", Content: "c", TemplateName: "t", Status: 1}).Error)
			must(t, "插入短信日志", db.DB.Create(&models.SMSLog{Phone: "13800000000", Provider: "p", TemplateName: "t", Status: uint8(i)}).Error)
		}

		// 先走明细兜底统计，再清空聚合表走回填 + 聚合表读取
		for _, m := range []any{&models.APIAccessLogStat{}, &models.OperationLogStat{}, &models.EmailLogStat{}, &models.SMSLogStat{}} {
			must(t, "清空聚合表", db.DB.Where("1 = 1").Delete(m).Error)
		}
		for pass := 0; pass < 2; pass++ {
			api, err := models.GetAPIAccessLogStats()
			must(t, "GetAPIAccessLogStats", err)
			op, err := models.GetOperationLogStatsDetail()
			must(t, "GetOperationLogStatsDetail", err)
			email, err := models.GetEmailLogStatsDetail()
			must(t, "GetEmailLogStatsDetail", err)
			sms, err := models.GetSMSLogStatsDetail()
			must(t, "GetSMSLogStatsDetail", err)
			if api.TotalCount != 2 || api.ServerErrorCount != 1 || len(api.TopPaths) != 1 ||
				op.TotalCount != 2 || len(op.TopModules) != 1 ||
				email.SuccessCount != 2 || len(email.TopTemplates) != 1 ||
				sms.FailCount != 1 || len(sms.ProviderStats) != 1 {
				t.Fatalf("第 %d 轮统计不符: api=%+v op=%+v email=%+v sms=%+v", pass, api, op, email, sms)
			}
			models.BackfillAPIAccessLogAggregateIfNeeded()
			models.BackfillOperationLogAggregateIfNeeded()
			models.BackfillEmailLogAggregateIfNeeded()
			models.BackfillSMSLogAggregateIfNeeded()
		}

		userStats, err := models.GetAPIAccessLogStatsByUserID(user.ID)
		must(t, "GetAPIAccessLogStatsByUserID", err)
		if userStats.TotalCount != 2 || userStats.TodayCount != 2 ||
			userStats.SuccessCount != 1 || userStats.ServerErrorCount != 1 ||
			userStats.ClientErrorCount != 0 || userStats.AvgDuration != 10 {
			t.Fatalf("用户 API 统计不符: %+v", userStats)
		}
	})

	t.Run("管理员余额补单并发幂等", func(t *testing.T) {
		u := &models.User{Username: "xdb-money", Nickname: "补单", Email: "xdb-money@example.test", Password: "x", Role: "user", Status: 1}
		must(t, "CreateUser(money)", models.CreateUser(u))
		req := services.MoneyOperationRequest{
			Operation: "balance_order", Amount: 10, OrderNo: "XDB-MONEY-1", OrderStatus: models.PaymentStatusPaid,
		}

		start := make(chan struct{})
		errs := make([]error, 2)
		var wg sync.WaitGroup
		for i := range errs {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				<-start
				_, errs[i] = services.OperateUserMoney(u.ID, req)
			}(i)
		}
		close(start)
		wg.Wait()

		success := 0
		for _, err := range errs {
			if err == nil {
				success++
			}
		}
		after, err := models.GetUserByID(u.ID)
		must(t, "GetUserByID(money)", err)
		if success != 1 || after.Money != 10 {
			t.Fatalf("同一订单并发补单应只成功一次且只加款一次: success=%d money=%v errs=%v", success, after.Money, errs)
		}
	})

	t.Run("日志保留清理", func(t *testing.T) {
		n, err := models.CleanExcessAPIAccessLogsPerUser(1)
		must(t, "CleanExcessAPIAccessLogsPerUser", err)
		m, err := models.CleanExcessSMSLogsPerRecipient(1)
		must(t, "CleanExcessSMSLogsPerRecipient", err)
		k, err := models.CleanExcessOperationLogs(1)
		must(t, "CleanExcessOperationLogs", err)
		if n != 1 || m != 1 || k != 1 {
			t.Fatalf("清理条数不符: api=%d sms=%d op=%d", n, m, k)
		}
	})

	t.Run("自动任务执行记录", func(t *testing.T) {
		var ids []uint64
		for i := 0; i < 3; i++ {
			id, err := task.InsertRun(&task.JobRun{
				RunUID: fmt.Sprintf("%036d", i+1), JobCode: "xdb-job", Status: task.StatusSuccess, StartedAt: now - int64(10-i),
			})
			must(t, "InsertRun", err)
			ids = append(ids, id)
		}
		_, err := task.MarkKeepForever(ids[:1], true)
		must(t, "MarkKeepForever", err)
		pruned, err := task.PruneSuccessRuns(1)
		must(t, "PruneSuccessRuns", err)
		list, total, err := task.ListRuns(1, 20, "xdb", task.StatusSuccess, "", "xdb-job", 0, 0)
		must(t, "ListRuns", err)
		if pruned != 2 || total != 1 || len(list) != 1 || list[0].ID != ids[0] {
			t.Fatalf("修剪后应只剩永久保留的一条: pruned=%d total=%d", pruned, total)
		}
		kept, keptTotal, err := task.ListKeptRuns(1, 20, "xdb", "", "", "", 0, 0)
		must(t, "ListKeptRuns", err)
		deleted, err := task.DeleteKeptRuns("all", "")
		must(t, "DeleteKeptRuns", err)
		if keptTotal != 1 || len(kept) != 1 || deleted != 1 {
			t.Fatalf("保留记录不符: kept=%d deleted=%d", keptTotal, deleted)
		}
		if _, err := task.InsertRun(&task.JobRun{RunUID: "bad", JobCode: "xdb-job", Status: task.StatusFailed, StartedAt: now}); err != nil {
			t.Fatalf("InsertRun(bad): %v", err)
		}
		fixed, err := task.RepairBadRunUIDs()
		must(t, "RepairBadRunUIDs", err)
		cleaned, err := task.CleanRuns(task.CleanRunsRequest{Scope: "failed", JobCode: "xdb-job"})
		must(t, "CleanRuns", err)
		if fixed != 1 || cleaned != 1 {
			t.Fatalf("修复/清理不符: fixed=%d cleaned=%d", fixed, cleaned)
		}
		_, _, err = task.CountDefinitions()
		must(t, "CountDefinitions", err)
		_, err = task.ListDefinitions("", "", nil)
		must(t, "ListDefinitions", err)
		_, err = task.SumLifetimeRuns()
		must(t, "SumLifetimeRuns", err)
	})
}
