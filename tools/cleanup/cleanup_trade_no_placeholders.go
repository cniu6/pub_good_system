//go:build ignore

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"fst/backend/app/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type paymentOrderTradeNoRow struct {
	ID      uint64
	OrderNo string
	TradeNo string
}

type cleanupItem struct {
	ID         uint64
	OrderNo    string
	OldTradeNo string
	NewTradeNo string
}

func main() {
	apply := flag.Bool("apply", false, "执行数据库更新，默认仅预览")
	dsn := flag.String("dsn", "", "MySQL DSN，优先级高于环境变量")
	flag.Parse()

	_ = godotenv.Load(".env")

	resolvedDSN := resolveDSN(*dsn)
	db, err := sql.Open("mysql", resolvedDSN)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库 Ping 失败: %v", err)
	}

	rows, err := db.Query("SELECT id, order_no, trade_no FROM payment_orders WHERE trade_no <> '' ORDER BY id ASC")
	if err != nil {
		log.Fatalf("查询 payment_orders 失败: %v", err)
	}
	defer rows.Close()

	var scanned []paymentOrderTradeNoRow
	for rows.Next() {
		var row paymentOrderTradeNoRow
		if err := rows.Scan(&row.ID, &row.OrderNo, &row.TradeNo); err != nil {
			log.Fatalf("扫描 payment_orders 行失败: %v", err)
		}
		scanned = append(scanned, row)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("遍历 payment_orders 结果失败: %v", err)
	}

	items := make([]cleanupItem, 0)
	for _, row := range scanned {
		normalized := models.NormalizeTradeNo(row.TradeNo)
		if normalized == row.TradeNo {
			continue
		}
		items = append(items, cleanupItem{
			ID:         row.ID,
			OrderNo:    row.OrderNo,
			OldTradeNo: row.TradeNo,
			NewTradeNo: normalized,
		})
	}

	fmt.Printf("扫描完成：共检查 %d 条 trade_no 非空订单，发现 %d 条需要清理。\n", len(scanned), len(items))
	if len(items) == 0 {
		fmt.Println("没有发现需要处理的历史 trade_no 脏数据。")
		return
	}

	for _, item := range items {
		fmt.Printf("order_no=%s old=%q new=%q\n", item.OrderNo, item.OldTradeNo, item.NewTradeNo)
	}

	if !*apply {
		fmt.Println("当前为预览模式，未写入数据库。添加 -apply 参数后会执行更新。")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("开启事务失败: %v", err)
	}
	defer tx.Rollback()

	now := time.Now().Unix()
	for _, item := range items {
		_, err = tx.Exec("UPDATE payment_orders SET trade_no = ?, update_time = ? WHERE id = ?", item.NewTradeNo, now, item.ID)
		if err != nil {
			log.Fatalf("更新订单 %s 失败: %v", item.OrderNo, err)
		}
	}

	if err = tx.Commit(); err != nil {
		log.Fatalf("提交事务失败: %v", err)
	}

	fmt.Printf("清理完成：已更新 %d 条 payment_orders.trade_no。\n", len(items))
}

func resolveDSN(flagDSN string) string {
	if flagDSN != "" {
		return flagDSN
	}
	if envDSN := os.Getenv("DB_DSN"); envDSN != "" {
		return envDSN
	}

	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "root")
	dbName := getEnv("DB_NAME", "fst_platform")

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

