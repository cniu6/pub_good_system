package migrate

import (
	"log"

	"fst/backend/pkg/db"
)

// dropObsoleteFinanceApprovalArtifacts 清理已废弃的「财务审批 / 双人复核」残留（幂等）。
// 业务代码已移除该功能；此处仅删除旧库中的表与设置行。无残留时直接跳过。
func dropObsoleteFinanceApprovalArtifacts() {
	if db.DB == nil {
		return
	}

	if db.CheckTableExists("approval_requests") {
		if err := db.DB.Migrator().DropTable("approval_requests"); err != nil {
			log.Printf("[Migrate] 删除废弃表 approval_requests 失败: %v", err)
		} else {
			log.Println("[Migrate] 已删除废弃表 approval_requests")
		}
	}

	if !db.CheckTableExists("system_settings") {
		return
	}
	res := db.DB.Exec("DELETE FROM system_settings WHERE setting_key = ?", "finance_dual_approval")
	if res.Error != nil {
		log.Printf("[Migrate] 删除废弃设置 finance_dual_approval 失败: %v", res.Error)
		return
	}
	if res.RowsAffected > 0 {
		log.Printf("[Migrate] 已删除废弃设置 finance_dual_approval（%d 行）", res.RowsAffected)
	}
}
