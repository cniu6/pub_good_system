// Package migrate 数据库自迁移统一入口（系统内部骨架）。
//
// 职责边界（别跟 pkg/db 搞混）：
//   - pkg/db：只负责连库（InitDB）+ GORM 小工具（ForUpdate / WithTx / Migrator 检查）
//   - 本包：唯一「建表/补丁/种子」编排入口 —— 调 db.DB.AutoMigrate，再跑存量补丁与种子
//
// 调用链：cmd/server → appinit.Bootstrap → db.InitDB → migrate.RunAutoMigrate
package migrate

import (
	"log"

	"fst/backend/app/models"
	"fst/backend/internal/task"
	"fst/backend/pkg/db"
)

// RunAutoMigrate 统一执行 GORM AutoMigrate、存量补丁与业务种子。
func RunAutoMigrate() {
	log.Println("[Migrate] 开始数据库自迁移（GORM）...")

	if db.DB == nil {
		log.Fatalf("[Migrate] 数据库未初始化，无法 AutoMigrate")
	}

	// 业务模型清单在 models.AllGormModels；自动任务表在 task 包，这里一并注册
	modelsList := models.AllGormModels()
	modelsList = append(modelsList, &task.JobDefinition{}, &task.JobRun{})

	if err := db.DB.AutoMigrate(modelsList...); err != nil {
		log.Fatalf("[Migrate] GORM AutoMigrate 失败: %v", err)
	}
	log.Printf("[Migrate] GORM AutoMigrate 完成，共 %d 个模型", len(modelsList))

	// AutoMigrate 覆盖不到的存量补丁
	models.EnsureRealnameCertUniqueConstraint()
	models.RepairVerificationCodeTable()
	dropObsoleteFinanceApprovalArtifacts()

	// 业务种子与聚合表历史回填
	models.SeedEmailTemplates()
	models.SeedSystemSettings()
	models.SeedRBACDefaults()
	models.SeedSMSTemplates()
	models.BackfillAPIAccessLogAggregateIfNeeded()
	models.BackfillEmailLogAggregateIfNeeded()
	models.BackfillSMSLogAggregateIfNeeded()
	models.BackfillOperationLogAggregateIfNeeded()

	log.Println("[Migrate] 数据库自迁移全部完成")
}
