package models

import (
	"fmt"
	"fst/backend/pkg/db"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// log_retention.go：email_logs / sms_logs / operation_logs / api_access_logs 四类日志的
// 「保留清理」逻辑高度重复（各自一份几乎一样的 CleanExcessX / CleanExcessXPerY / 节流调度 /
// 配置读取），这里抽成通用实现，四个 model 文件里只保留极薄的具名包装函数（保持导出函数名、
// 签名不变，不影响任何调用方）。
//
// 踩坑记录：最早的实现是「查出第 maxCount 名那一行的时间值当 cutoff，再把这个 Go 时间值当参数
// 传回 DELETE ... WHERE time_col < ? OR (time_col = ? AND id < ?)」（改造前 4 个文件各自的旧实现
// 也是这个写法）。这个写法在 SQLite（modernc.org/sqlite 驱动）下会整表删空：驱动把 time.Time 参数
// 用 Go 的 time.Time.String() 格式（"2006-01-02 15:04:05 +0000 UTC"）序列化绑定，跟表里实际存的
// ISO8601 字符串（"2006-01-02T15:04:05Z"）格式不一致，字符串比较直接错乱。改成 id NOT IN (子查询)
// 后完全不需要把时间值序列化后再传回去，天然避开这个坑，三种驱动下语义一致。

// cleanExcessRowsGeneric 全局保留最新 maxCount 条（按 timeCol 排序），删除更旧的记录。
// table/timeCol 均为代码里硬编码的表名/列名常量，不接受外部输入，拼 SQL 是安全的。
func cleanExcessRowsGeneric(table, timeCol string, maxCount int) (int64, error) {
	if maxCount <= 0 {
		return 0, nil
	}
	var total int64
	if err := db.DB.Get(&total, "SELECT COUNT(*) FROM "+table); err != nil {
		return 0, err
	}
	if total <= int64(maxCount) {
		return 0, nil
	}

	// 内层再套一层派生表（AS keep_rows）是为了绕开 MySQL「不能在 UPDATE/DELETE 里把目标表本身
	// 当子查询表」的限制（Error 1093），SQLite/PostgreSQL 对此没有限制但套一层也无害。
	deleteSQL := fmt.Sprintf(
		"DELETE FROM %s WHERE id NOT IN (SELECT id FROM (SELECT id FROM %s ORDER BY %s DESC, id DESC LIMIT ?) AS keep_rows)",
		table, table, timeCol,
	)
	result, err := db.Exec(deleteSQL, maxCount)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// cleanExcessRowsPerGroupGeneric 按 groupCol 分组，每组最多保留 maxPerGroup 条，多的按时间清理。
// G 是分组列的 Go 类型（如 string 收件人/手机号，uint64 用户 ID）。
// extraWhere 可选（形如 "user_id > 0"），用于像 api_access_logs 那样排除未登录（user_id=0）分组；
// 传空字符串表示不加额外过滤。
func cleanExcessRowsPerGroupGeneric[G any](table, groupCol, timeCol string, maxPerGroup int, extraWhere string) (int64, error) {
	if maxPerGroup <= 0 {
		return 0, nil
	}

	groupWhere := ""
	if strings.TrimSpace(extraWhere) != "" {
		groupWhere = " WHERE " + extraWhere
	}

	var groups []struct {
		Grp G     `db:"grp"`
		Cnt int64 `db:"cnt"`
	}
	groupSQL := fmt.Sprintf("SELECT %s AS grp, COUNT(*) AS cnt FROM %s%s GROUP BY %s HAVING COUNT(*) > ?", groupCol, table, groupWhere, groupCol)
	if err := db.DB.Select(&groups, groupSQL, maxPerGroup); err != nil {
		return 0, err
	}

	deleteSQL := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = ? AND id NOT IN (SELECT id FROM (SELECT id FROM %s WHERE %s = ? ORDER BY %s DESC, id DESC LIMIT ?) AS keep_rows)",
		table, groupCol, table, groupCol, timeCol,
	)

	var totalAffected int64
	for _, g := range groups {
		result, err := db.Exec(deleteSQL, g.Grp, g.Grp, maxPerGroup)
		if err != nil {
			// 单个分组清理失败不应中断其余分组的清理，但也不能完全静默丢弃，
			// 否则某个分组持续清理失败会长期悄悄超限占用存储。
			log.Printf("[LogRetention] 按分组清理超限记录失败 table=%s group_col=%s grp=%v: %v", table, groupCol, g.Grp, err)
			continue
		}
		n, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			log.Printf("[LogRetention] 读取清理影响行数失败 table=%s group_col=%s grp=%v: %v", table, groupCol, g.Grp, rowsErr)
			continue
		}
		if n > 0 {
			totalAffected += n
		}
	}
	return totalAffected, nil
}

// logRetentionConfig 三个通用配置项：全局上限 / 是否启用按用户（收件人）上限 / 按用户上限。
type logRetentionConfig struct {
	MaxCount            int
	PerUserLimitEnabled bool
	PerUserMaxCount     int
}

// loadLogRetentionConfigGeneric 按 "<keyPrefix>_max_count" / "<keyPrefix>_per_user_limit_enabled" /
// "<keyPrefix>_per_user_max_count" 三个 key 读取配置，email/sms/operation 三类日志共用同一套 key 命名规则。
func loadLogRetentionConfigGeneric(keyPrefix string) logRetentionConfig {
	cfg := logRetentionConfig{MaxCount: 1000, PerUserMaxCount: 1000}
	settingsMap, err := GetSettingsMap([]string{
		keyPrefix + "_max_count",
		keyPrefix + "_per_user_limit_enabled",
		keyPrefix + "_per_user_max_count",
	})
	if err != nil {
		return cfg
	}
	if v, ok := settingsMap[keyPrefix+"_max_count"]; ok {
		if n, parseErr := strconv.Atoi(strings.TrimSpace(v)); parseErr == nil && n > 0 {
			cfg.MaxCount = n
		}
	}
	if v, ok := settingsMap[keyPrefix+"_per_user_limit_enabled"]; ok {
		lower := strings.ToLower(strings.TrimSpace(v))
		cfg.PerUserLimitEnabled = lower == "true" || lower == "1"
	}
	if v, ok := settingsMap[keyPrefix+"_per_user_max_count"]; ok {
		if n, parseErr := strconv.Atoi(strings.TrimSpace(v)); parseErr == nil && n > 0 {
			cfg.PerUserMaxCount = n
		}
	}
	return cfg
}

// scheduleLogRetentionCleanupGeneric 通用节流调度：30 秒内只真正触发一次清理，避免每写一条日志
// 都去读配置 + 扫表。email/sms/operation 三类日志的 CreateXLog 写入后都异步调用这个。
func scheduleLogRetentionCleanupGeneric(
	nextAt *atomic.Int64,
	keyPrefix, logTag string,
	cleanGlobal func(maxCount int) (int64, error),
	cleanPerUser func(maxPerUser int) (int64, error),
) {
	now := time.Now().UnixNano()
	prev := nextAt.Load()
	if prev > now {
		return
	}
	if !nextAt.CompareAndSwap(prev, now+int64(30*time.Second)) {
		return
	}

	cfg := loadLogRetentionConfigGeneric(keyPrefix)
	if cfg.MaxCount > 0 {
		if _, err := cleanGlobal(cfg.MaxCount); err != nil {
			log.Printf("[%s] 自动清理超限日志失败: %v", logTag, err)
		}
	}
	if cfg.PerUserLimitEnabled && cfg.PerUserMaxCount > 0 {
		if _, err := cleanPerUser(cfg.PerUserMaxCount); err != nil {
			log.Printf("[%s] 按用户/收件人清理超限日志失败: %v", logTag, err)
		}
	}
}
