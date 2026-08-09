// Package migrate 数据库自迁移统一入口（系统内部骨架）。
//
// 职责边界（别跟 pkg/db 搞混）：
//   - pkg/db：只负责连库（InitDB）+ GORM 小工具（ForUpdate / WithTx / Migrator 检查）
//   - 本包：唯一「建表/补丁/种子」编排入口 —— 调 db.DB.AutoMigrate，再跑存量补丁与种子
//
// 调用链：cmd/server → appinit.Bootstrap → db.InitDB → migrate.RunAutoMigrate
package migrate

import (
	"fmt"
	"log"

	"fst/backend/app/models"
	"fst/backend/internal/task"
	"fst/backend/pkg/config"
	"fst/backend/pkg/db"
)

// RunAutoMigrate 统一执行 GORM AutoMigrate、存量补丁与业务种子。
func RunAutoMigrate() {
	log.Println("[Migrate] 开始数据库自迁移（GORM）...")

	if db.DB == nil {
		log.Fatalf("[Migrate] 数据库未初始化，无法 AutoMigrate")
	}

	// MySQL 统一数据库默认排序规则，确保 AutoMigrate 新建表继承一致规则。
	ensureMySQLDatabaseCollation()

	// 业务模型清单在 models.AllGormModels；自动任务表在 task 包，这里一并注册
	modelsList := models.AllGormModels()
	modelsList = append(modelsList, &task.JobDefinition{}, &task.JobRun{}, &task.JobRunKeep{})

	// AutoMigrate 时临时设置 gorm:table_options，确保新建表使用统一字符集/排序规则。
	// 不能在全局 db.DB 上 Set —— Set 返回的实例 clone==0，会导致后续查询共享 Statement 污染。
	migrateDB := db.DB
	if db.IsMySQL() {
		collation := config.DefaultMySQLCollation
		if cfg := config.GetGlobalConfig(); cfg != nil && cfg.DBCollation != "" {
			collation = cfg.DBCollation
		}
		migrateDB = db.DB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE="+collation)
	}

	if err := migrateDB.AutoMigrate(modelsList...); err != nil {
		log.Fatalf("[Migrate] GORM AutoMigrate 失败: %v", err)
	}
	log.Printf("[Migrate] GORM AutoMigrate 完成，共 %d 个模型", len(modelsList))

	// AutoMigrate 覆盖不到的存量补丁
	models.EnsureRealnameCertUniqueConstraint()
	models.RepairVerificationCodeTable()
	models.RepairHashedApiKeys()
	dropObsoleteFinanceApprovalArtifacts()
	dropObsoleteRBACArtifacts()
	migratePayGatewayExtConfig()

	// 业务种子与聚合表历史回填
	models.SeedEmailTemplates()
	models.SeedSystemSettings()
	models.SeedUserLevelCaps()
	models.SeedSMSTemplates()
	models.SeedExchangeRates()
	models.BackfillAPIAccessLogAggregateIfNeeded()
	models.BackfillEmailLogAggregateIfNeeded()
	models.BackfillSMSLogAggregateIfNeeded()
	models.BackfillOperationLogAggregateIfNeeded()

	log.Println("[Migrate] 数据库自迁移全部完成")
}

// ensureMySQLDatabaseCollation 启动时将数据库默认字符集/排序规则统一为项目默认值，
// 避免新建表因数据库默认规则不一致而产生 utf8mb4_general_ci 等多 collation 混用。
func ensureMySQLDatabaseCollation() {
	if !db.IsMySQL() {
		return
	}

	cfg := config.GetGlobalConfig()
	collation := config.DefaultMySQLCollation
	if cfg != nil && cfg.DBCollation != "" {
		collation = cfg.DBCollation
	}

	var dbName string
	if err := db.DB.Raw("SELECT DATABASE()").Scan(&dbName).Error; err != nil {
		log.Printf("[Migrate] 查询当前数据库名失败: %v", err)
		return
	}
	if dbName == "" {
		log.Println("[Migrate] 当前数据库名为空，跳过数据库默认排序规则设置")
		return
	}

	sql := fmt.Sprintf("ALTER DATABASE `%s` CHARACTER SET utf8mb4 COLLATE %s", dbName, collation)
	if err := db.DB.Exec(sql).Error; err != nil {
		log.Printf("[Migrate] 设置数据库默认排序规则失败: %v", err)
		return
	}
	log.Printf("[Migrate] 数据库 `%s` 默认排序规则已统一为 utf8mb4 / %s", dbName, collation)
}
