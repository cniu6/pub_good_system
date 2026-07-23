package migrate

import (
	"log"

	"fst/backend/pkg/db"
)

// dropObsoleteRBACArtifacts 清理已废弃的 RBAC 表（幂等）。
// 业务已改回仅用 users.role=admin 鉴权；无残留时直接跳过。
func dropObsoleteRBACArtifacts() {
	if db.DB == nil {
		return
	}
	// 先删关联表，再删主表
	for _, name := range []string{"user_roles", "role_permissions", "permissions", "roles"} {
		if !db.CheckTableExists(name) {
			continue
		}
		if err := db.DB.Migrator().DropTable(name); err != nil {
			log.Printf("[Migrate] 删除废弃表 %s 失败: %v", name, err)
		} else {
			log.Printf("[Migrate] 已删除废弃表 %s", name)
		}
	}
}
