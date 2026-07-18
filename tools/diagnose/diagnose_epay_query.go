//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"fst/backend/app/models"
	"fst/backend/app/plugins/pay_balance/epay"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func maskMiddle(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "***" + value[len(value)-4:]
}

func main() {
	log.SetFlags(0)

	orderID := uint64(0)
	if len(os.Args) > 1 {
		parsed, err := strconv.ParseUint(strings.TrimSpace(os.Args[1]), 10, 64)
		if err != nil {
			log.Fatalf("订单ID参数非法: %v", err)
		}
		orderID = parsed
	}
	if orderID == 0 {
		log.Fatal("请提供订单ID，例如: go run ./tools/diagnose/diagnose_epay_query.go 7")
	}

	config.InitConfig()
	conn, err := sqlx.Connect(config.GlobalConfig.DBDriver, config.GlobalConfig.DBDSN)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer conn.Close()
	db.DB = conn

	order, err := models.GetPaymentOrderByID(orderID)
	if err != nil {
		log.Fatalf("读取订单失败: %v", err)
	}
	gateway, err := models.GetPayGatewayByID(order.GatewayID)
	if err != nil {
		log.Fatalf("读取支付通道失败: %v", err)
	}

	orderNo := strings.TrimSpace(order.OrderNo)
	tradeNo := models.NormalizeTradeNo(order.TradeNo)
	cfg := epay.ConfigFromGateway(gateway)

	fmt.Printf("order_id=%d order_no=%s trade_no=%s status=%d gateway_id=%d gateway=%s api=%s pid=%s\n",
		order.ID, orderNo, maskMiddle(tradeNo), order.Status, order.GatewayID, gateway.Name,
		cfg.ApiURL, maskMiddle(gateway.PID))

	result, err := epay.QueryOrder(cfg, orderNo, tradeNo)
	if err != nil {
		log.Fatalf("查单失败: %v", err)
	}
	if result == nil {
		log.Fatal("查单结果为空")
	}

	fmt.Printf("code=%d msg=%s trade_no=%s out_trade_no=%s type=%s money=%s trade_status=%s\n",
		result.Code, result.Msg, result.TradeNo, result.OutTradeNo, result.Type, result.Money, result.TradeStatus)
}
