//go:build ignore

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"fst/backend/app/models"
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

func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	query := u.Query()
	if query.Get("key") != "" {
		query.Set("key", "***")
	}
	u.RawQuery = query.Encode()
	return u.String()
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
		log.Fatal("请提供订单ID，例如: go run ./backend/cmd/diagnose_epay_query.go 7")
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
	apiURL := strings.TrimRight(gateway.ApiURL, "/") + "/api.php"

	fmt.Printf("order_id=%d order_no=%s trade_no=%s status=%d gateway_id=%d gateway=%s api=%s pid=%s\n", order.ID, orderNo, maskMiddle(tradeNo), order.Status, order.GatewayID, gateway.Name, strings.TrimRight(gateway.ApiURL, "/"), maskMiddle(gateway.PID))

	attempts := make([]url.Values, 0, 2)
	if orderNo != "" {
		query := url.Values{}
		query.Set("act", "order")
		query.Set("pid", gateway.PID)
		query.Set("key", gateway.Key)
		query.Set("out_trade_no", orderNo)
		attempts = append(attempts, query)
	}
	if tradeNo != "" {
		query := url.Values{}
		query.Set("act", "order")
		query.Set("pid", gateway.PID)
		query.Set("key", gateway.Key)
		query.Set("trade_no", tradeNo)
		attempts = append(attempts, query)
	}
	if len(attempts) == 0 {
		log.Fatal("当前订单没有可用于查单的订单号或交易号")
	}

	client := &http.Client{Timeout: 20 * time.Second}
	for index, query := range attempts {
		fullURL := apiURL + "?" + query.Encode()
		fmt.Printf("attempt=%d url=%s\n", index+1, redactURL(fullURL))
		resp, err := client.Get(fullURL)
		if err != nil {
			fmt.Printf("attempt=%d error=%v\n", index+1, err)
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("attempt=%d read_error=%v\n", index+1, err)
			continue
		}
		fmt.Printf("attempt=%d http_status=%s body=%s\n", index+1, resp.Status, strings.TrimSpace(string(body)))
	}
}
