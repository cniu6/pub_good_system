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

// TestMySQLIntegration 真机 MySQL 集成测试。
//
// 需要环境变量 FST_MYSQL_DSN 或 TEST_MYSQL_DSN 指向可写库；未设置则 Skip。
// 覆盖：InitDB + RunAutoMigrate + 用户/订单 CRUD + CastToText 关键词搜索。
//
//	set FST_MYSQL_DSN=user:pass@tcp(127.0.0.1:3306)/fst_test?charset=utf8mb4&parseTime=True&loc=Local
//	go test -tags integration ./backend/pkg/db/ -run MySQL -count=1
func TestMySQLIntegration(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("FST_MYSQL_DSN"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("TEST_MYSQL_DSN"))
	}
	if dsn == "" {
		t.Skip("MySQL 未提供认证：请设置 FST_MYSQL_DSN 或 TEST_MYSQL_DSN 后再跑本集成测试")
	}

	t.Setenv("DB_DRIVER", "mysql")
	t.Setenv("JWT_SECRET", "mysql-integration-test-secret-please-change")
	t.Setenv("CORS_ORIGINS", "*")
	t.Setenv("APP_ENV", "development")

	config.InitConfig()
	config.UpdateGlobalConfig(func(cfg *config.Config) {
		cfg.DBDriver = "mysql"
		cfg.DBDSN = dsn
	})

	db.InitDB()
	t.Cleanup(func() {
		if db.DB != nil {
			_ = db.Close()
		}
	})

	if !db.IsMySQL() {
		t.Fatalf("期望 mysql 驱动，实际 %q", db.DriverName())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	sqlDB, err := db.SQLDB()
	if err != nil {
		t.Fatalf("获取底层连接失败: %v", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		t.Fatalf("MySQL Ping 失败: %v", err)
	}

	migrate.RunAutoMigrate()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano()%1_000_000_000)
	user := &models.User{
		Username: "mysql-it-" + suffix,
		Nickname: "MySQL 集成测试",
		Email:    "mysql-it-" + suffix + "@example.test",
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

	now := time.Now().Unix()
	order := &models.PaymentOrder{
		OrderNo:        "MySQLIT" + suffix,
		UserID:         user.ID,
		GatewayID:      1,
		TradeNo:        "T" + suffix,
		PaymentChannel: "epay",
		PaymentType:    "alipay",
		Amount:         12.34,
		PayAmount:      12.34,
		Subject:        "mysql-integration",
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

	list, total, err := models.GetPaymentOrderList(user.ID, 1, 20, -1, order.OrderNo)
	if err != nil {
		t.Fatalf("GetPaymentOrderList 失败: %v", err)
	}
	if total < 1 || len(list) < 1 {
		t.Fatalf("订单列表应命中，total=%d list=%d", total, len(list))
	}

	if _, err := models.CreateUserMoneyLog(user.ID, 12.34, 10, 22.34, "mysql-cast-test"); err != nil {
		t.Fatalf("CreateUserMoneyLog 失败: %v", err)
	}
	moneyLogs, moneyTotal, err := models.GetUserMoneyLogList(user.ID, 1, 20, "12.34")
	if err != nil {
		t.Fatalf("GetUserMoneyLogList(CastToText) 失败: %v", err)
	}
	if moneyTotal < 1 || len(moneyLogs) < 1 {
		t.Fatalf("金额关键词搜索应命中，total=%d", moneyTotal)
	}

	var castOut string
	expr := db.CastToText("SUM(amount)")
	if err := db.DB.Raw(`SELECT COALESCE(`+expr+`, '0') FROM payment_orders WHERE user_id = ?`, user.ID).Scan(&castOut).Error; err != nil {
		t.Fatalf("CastToText 聚合查询失败 expr=%s: %v", expr, err)
	}

	t.Logf("MySQL 集成测试通过（driver=%s user_id=%d order=%s cast=%s）。",
		db.DriverName(), user.ID, order.OrderNo, castOut)
}
