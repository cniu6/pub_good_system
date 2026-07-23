package models

// AllGormModels 返回需要 GORM AutoMigrate 的全部业务模型（含聚合统计表）。
// migrate 包统一调用；不要在此处放种子逻辑。
func AllGormModels() []interface{} {
	return []interface{}{
		&User{},
		&UserSession{},
		&UserSettings{},
		&VerificationCode{},
		&SystemSetting{},
		&EmailLog{},
		&EmailTemplate{},
		&SMSLog{},
		&SMSTemplate{},
		&OperationLog{},
		&APIAccessLog{},
		&PaymentOrder{},
		&PaymentException{},
		&WithdrawRequest{},
		&UserMoneyLog{},
		&UserScoreLog{},
		&RealnameVerification{},
		&PayGateway{},
		&IdempotencyKey{},
		&Announcement{},
		&UserAnnouncementRead{},
		&UserLevelCap{},
		// 聚合统计
		&APIAccessLogStat{},
		&APIAccessLogDailyStat{},
		&APIAccessLogPathStatRow{},
		&APIAccessLogMethodStatRow{},
		&APIAccessLogSceneStatRow{},
		&APIAccessLogIPStatRow{},
		&EmailLogStat{},
		&EmailLogDailyStat{},
		&EmailLogTemplateStatRow{},
		&SMSLogStat{},
		&SMSLogDailyStat{},
		&SMSLogTemplateStatRow{},
		&SMSLogProviderStatRow{},
		&OperationLogStat{},
		&OperationLogDailyStat{},
		&OperationLogModuleStatRow{},
		&OperationLogActionStatRow{},
		&OperationLogMethodStatRow{},
	}
}
