package migrate

import (
	"fmt"
	"fst/backend/app/models"
	"fst/backend/pkg/db"
	"fst/backend/pkg/payment"
	"log"
	"strings"
)

// legacyPayGateway 仅用于迁移已废弃列的临时结构体
type legacyPayGateway struct {
	ID                   uint64
	SignType             string
	Key                  string
	TargetCurrency       string
	ExchangeRateMode     string
	ExchangeRate         float64
	ExchangeFixedAmount  float64
	ExchangeRateSource   string
	TargetFeeRate        int
	TargetFeeFixed       float64
	TargetFeeMode        string
	ActiveQueryEnabled   int
	QueryIntervalSeconds int
	QueryBatchSize       int
	ExtConfig            string
}

// migratePayGatewayExtConfig 把 pay_gateways 表中已迁移到 ext_config 的列数据复制到 ext_config JSON，
// 然后删除旧列。幂等：若 sign_type 列已不存在则直接跳过。
func migratePayGatewayExtConfig() {
	if db.DB == nil || !db.CheckTableExists("pay_gateways") {
		return
	}

	if !db.CheckColumnExists("pay_gateways", "sign_type") {
		log.Println("[Migrate] pay_gateways 旧列已迁移，跳过")
		return
	}

	var rows []legacyPayGateway

	quote := "`"
	if db.IsPostgres() || db.IsSQLite() {
		quote = "\""
	}

	columns := []string{
		"id", "sign_type", "key", "target_currency", "exchange_rate_mode",
		"exchange_rate", "exchange_fixed_amount", "exchange_rate_source",
		"target_fee_rate", "target_fee_fixed", "target_fee_mode",
		"active_query_enabled", "query_interval_seconds", "query_batch_size",
		"ext_config",
	}
	for i, c := range columns {
		columns[i] = quote + c + quote
	}
	table := quote + "pay_gateways" + quote

	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ","), table)
	err := db.DB.Raw(query).Scan(&rows).Error
	if err != nil {
		log.Printf("[Migrate] 读取 pay_gateways 旧列失败: %v", err)
		return
	}

	for _, row := range rows {
		extMap := payment.ParseExtConfigMap(row.ExtConfig)

		// 旧列值写入 ext_config，仅当 ext_config 中不存在该键时才写入，避免覆盖已有数据
		setLegacyExtValue(extMap, "sign_type", row.SignType)
		setLegacyExtValue(extMap, "key", row.Key)
		setLegacyExtValue(extMap, "target_currency", row.TargetCurrency)
		setLegacyExtValue(extMap, "exchange_rate_mode", row.ExchangeRateMode)
		setLegacyExtValue(extMap, "exchange_rate", row.ExchangeRate)
		setLegacyExtValue(extMap, "exchange_fixed_amount", row.ExchangeFixedAmount)
		setLegacyExtValue(extMap, "exchange_rate_source", row.ExchangeRateSource)
		setLegacyExtValue(extMap, "target_fee_rate", row.TargetFeeRate)
		setLegacyExtValue(extMap, "target_fee_fixed", row.TargetFeeFixed)
		setLegacyExtValue(extMap, "target_fee_mode", row.TargetFeeMode)
		setLegacyExtValue(extMap, "active_query_enabled", row.ActiveQueryEnabled)
		setLegacyExtValue(extMap, "query_interval_seconds", row.QueryIntervalSeconds)
		setLegacyExtValue(extMap, "query_batch_size", row.QueryBatchSize)

		newExt := payment.MarshalExtConfigMap(extMap)
		if err := db.DB.Model(&models.PayGateway{}).Where("id = ?", row.ID).Update("ext_config", newExt).Error; err != nil {
			log.Printf("[Migrate] 更新 gateway id=%d ext_config 失败: %v", row.ID, err)
		}
	}

	// 删除旧列
	dropColumns := []string{
		"sign_type",
		"key",
		"target_currency",
		"exchange_rate_mode",
		"exchange_rate",
		"exchange_fixed_amount",
		"exchange_rate_source",
		"target_fee_rate",
		"target_fee_fixed",
		"target_fee_mode",
		"active_query_enabled",
		"query_interval_seconds",
		"query_batch_size",
	}
	for _, col := range dropColumns {
		if !db.CheckColumnExists("pay_gateways", col) {
			continue
		}
		if err := db.DB.Migrator().DropColumn(&models.PayGateway{}, col); err != nil {
			log.Printf("[Migrate] 删除列 %s 失败: %v", col, err)
		} else {
			log.Printf("[Migrate] 已删除列 %s", col)
		}
	}

	log.Println("[Migrate] pay_gateways 列迁移完成")
}

// setLegacyExtValue 将旧列值写入 ext_config map，仅当 key 不存在时写入。
// 即使旧值为零值也写入，以保留显式配置（如 active_query_enabled=0）。
func setLegacyExtValue(extMap map[string]interface{}, key string, value interface{}) {
	if _, ok := extMap[key]; !ok {
		extMap[key] = value
	}
}
