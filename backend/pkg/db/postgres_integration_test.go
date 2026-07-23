//go:build integration

package db_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/migrate"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
)

// TestPostgresIntegration 真机 PostgreSQL 集成测试。
//
// 需要环境变量 FST_PG_DSN 或 TEST_POSTGRES_DSN 指向可写库；未设置则 Skip。
// 覆盖：InitDB + RunAutoMigrate + 用户 CRUD + 支付/提现冒烟 + CastToText 关键词搜索。
//
//	set FST_PG_DSN=postgres://user:pass@127.0.0.1:5432/fst_test?sslmode=disable
//	go test -tags integration ./backend/pkg/db/ -run Postgres -count=1
func TestPostgresIntegration(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("FST_PG_DSN"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("TEST_POSTGRES_DSN"))
	}
	if dsn == "" {
		t.Skip("Postgres 未生产认证：请设置 FST_PG_DSN 或 TEST_POSTGRES_DSN 后再跑本集成测试")
	}

	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("JWT_SECRET", "postgres-integration-test-secret-please-change")
	t.Setenv("CORS_ORIGINS", "*")
	t.Setenv("APP_ENV", "development")
	t.Setenv("DB_SSLMODE", "disable")

	config.InitConfig()
	config.UpdateGlobalConfig(func(cfg *config.Config) {
		cfg.DBDriver = "postgres"
		cfg.DBDSN = dsn
	})

	db.InitDB()
	t.Cleanup(func() {
		if db.DB != nil {
			_ = db.Close()
		}
	})

	if !db.IsPostgres() {
		t.Fatalf("期望 postgres 驱动，实际 %q", db.DriverName())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sqlDB, err := db.SQLDB()
	if err != nil {
		t.Fatalf("获取底层连接失败: %v", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		t.Fatalf("Postgres Ping 失败: %v", err)
	}

	// 全量自迁移：真机验证 AutoMigrate + 种子/补丁
	migrate.RunAutoMigrate()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000_000)
	user := &models.User{
		Username: "pg-it-" + suffix,
		Nickname: "PG 集成测试",
		Email:    "pg-it-" + suffix + "@example.test",
		Password: "not-used-in-test",
		Money:    10,
		Score:    1,
		Role:     "user",
		Status:   1,
	}
	if err := models.CreateUser(user); err != nil {
		t.Fatalf("models.CreateUser 失败: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("CreateUser 未回填 ID")
	}

	got, err := models.GetUserByID(user.ID)
	if err != nil || got == nil || got.Username != user.Username {
		t.Fatalf("GetUserByID 失败: got=%+v err=%v", got, err)
	}

	// 支付订单：建单 + RealPaid 过滤 SQL + 关键词搜索（触发 CastToText）
	now := time.Now().Unix()
	orderNo := "PGIT" + suffix
	order := &models.PaymentOrder{
		OrderNo:        orderNo,
		UserID:         user.ID,
		GatewayID:      1,
		TradeNo:        "T" + suffix,
		PaymentChannel: "epay",
		PaymentType:    "alipay",
		Amount:         12.34,
		PayAmount:      12.34,
		Subject:        "pg-integration",
		Status:         models.PaymentStatusPaid,
		ExpireAt:       now + 3600,
		ClientIP:       "127.0.0.1",
		CreateTime:     now,
		UpdateTime:     now,
	}
	paidAt := now
	order.PaidAt = &paidAt
	if err := models.CreatePaymentOrder(order); err != nil {
		t.Fatalf("CreatePaymentOrder 失败: %v", err)
	}

	list, total, err := models.GetPaymentOrderList(user.ID, 1, 20, -1, orderNo)
	if err != nil {
		t.Fatalf("GetPaymentOrderList 失败: %v", err)
	}
	if total < 1 {
		t.Fatalf("订单列表应至少命中 1 笔，total=%d list=%d", total, len(list))
	}

	// 余额日志关键词搜索（CastToText 数字列）
	if _, err := models.CreateUserMoneyLog(user.ID, 12.34, 10, 22.34, "pg-cast-test"); err != nil {
		t.Fatalf("CreateUserMoneyLog 失败: %v", err)
	}
	moneyLogs, moneyTotal, err := models.GetUserMoneyLogList(user.ID, 1, 20, "12.34")
	if err != nil {
		t.Fatalf("GetUserMoneyLogList(CastToText) 失败: %v", err)
	}
	if moneyTotal < 1 || len(moneyLogs) < 1 {
		t.Fatalf("金额关键词搜索应命中，total=%d", moneyTotal)
	}

	var paidCount int64
	if err := db.DB.Raw(
		`SELECT COUNT(*) FROM payment_orders WHERE status = ? AND `+models.RealPaidOrderFilterSQL,
		models.PaymentStatusPaid,
	).Scan(&paidCount).Error; err != nil {
		t.Fatalf("RealPaidOrderFilterSQL 在 Postgres 上执行失败: %v", err)
	}
	if paidCount < 1 {
		t.Fatalf("真实已付过滤应命中至少 1 笔，got=%d", paidCount)
	}

	// 提现：建申请 + 列表
	wr := &models.WithdrawRequest{
		UserID:      user.ID,
		Amount:      1.23,
		AccountType: "alipay",
		AccountName: "pg-test",
		AccountNo:   "pg-account-" + suffix,
		Status:      0,
		CreateTime:  now,
		UpdateTime:  now,
	}
	if err := models.CreateWithdrawRequest(wr); err != nil {
		t.Fatalf("CreateWithdrawRequest 失败: %v", err)
	}
	if wr.ID == 0 {
		t.Fatal("CreateWithdrawRequest 未回填 ID")
	}
	gotWR, err := models.GetWithdrawRequestByID(wr.ID)
	if err != nil || gotWR == nil || gotWR.UserID != user.ID {
		t.Fatalf("GetWithdrawRequestByID 失败: got=%+v err=%v", gotWR, err)
	}

	// CastToText 冒烟：聚合字符串
	var castOut string
	expr := db.CastToText("SUM(amount)")
	if err := db.DB.Raw(`SELECT COALESCE(`+expr+`, '0') FROM payment_orders WHERE user_id = ?`, user.ID).Scan(&castOut).Error; err != nil {
		t.Fatalf("CastToText 聚合查询失败 expr=%s: %v", expr, err)
	}

	t.Logf("Postgres 集成测试通过（driver=%s user_id=%d order=%s withdraw_id=%d cast=%s）。",
		db.DriverName(), user.ID, orderNo, wr.ID, castOut)
}
