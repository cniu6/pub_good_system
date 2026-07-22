package models_test

import (
	"fmt"
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
)

// TestCleanExcessEmailLogs_KeepsNewest 回归：email/sms/operation/api_access 四类日志的
// CleanExcessX 抽取成通用 cleanExcessRowsGeneric 后，行为必须和抽取前完全一致——
// 只保留最新 maxCount 条，多余的按时间+id 顺序删掉。
func TestCleanExcessEmailLogs_KeepsNewest(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		if err := models.CreateEmailLog("to@example.com", "subject", "content", "tpl", 1, ""); err != nil {
			t.Fatalf("CreateEmailLog 失败: %v", err)
		}
	}

	affected, err := models.CleanExcessEmailLogs(3)
	if err != nil {
		t.Fatalf("CleanExcessEmailLogs 失败: %v", err)
	}
	if affected != 2 {
		t.Fatalf("应删除 2 条，实际删除 %d 条", affected)
	}

	// 注意：EmailLogQuery.Status 零值是 0（不是 -1），查询时必须显式传 -1 表示「不筛选状态」，
	// 否则会被当成「只查失败邮件」。
	q := &models.EmailLogQuery{Page: 1, PageSize: 20, Status: -1}
	logs, total, err := models.GetEmailLogList(q)
	if err != nil {
		t.Fatalf("GetEmailLogList 失败: %v", err)
	}
	if total != 3 || len(logs) != 3 {
		t.Fatalf("应剩 3 条，实际 total=%d len=%d", total, len(logs))
	}
}

// TestCleanExcessEmailLogsPerRecipient_PerGroupLimit 验证按分组（收件邮箱）保留最新 N 条，
// 不同分组互不影响；并验证 extraWhere（排除空邮箱）生效。
func TestCleanExcessEmailLogsPerRecipient_PerGroupLimit(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	for i := 0; i < 4; i++ {
		if err := models.CreateEmailLog("a@example.com", "s", "c", "tpl", 1, ""); err != nil {
			t.Fatalf("CreateEmailLog(a) 失败: %v", err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := models.CreateEmailLog("b@example.com", "s", "c", "tpl", 1, ""); err != nil {
			t.Fatalf("CreateEmailLog(b) 失败: %v", err)
		}
	}

	affected, err := models.CleanExcessEmailLogsPerRecipient(2)
	if err != nil {
		t.Fatalf("CleanExcessEmailLogsPerRecipient 失败: %v", err)
	}
	if affected != 2 {
		t.Fatalf("a@example.com 超出 2 条应删 2 条，实际删除 %d 条", affected)
	}

	q := &models.EmailLogQuery{Page: 1, PageSize: 20, Status: -1, ToEmail: "a@example.com"}
	_, totalA, err := models.GetEmailLogList(q)
	if err != nil {
		t.Fatalf("GetEmailLogList(a) 失败: %v", err)
	}
	if totalA != 2 {
		t.Fatalf("a@example.com 应剩 2 条，实际 %d 条", totalA)
	}

	q2 := &models.EmailLogQuery{Page: 1, PageSize: 20, Status: -1, ToEmail: "b@example.com"}
	_, totalB, err := models.GetEmailLogList(q2)
	if err != nil {
		t.Fatalf("GetEmailLogList(b) 失败: %v", err)
	}
	if totalB != 2 {
		t.Fatalf("b@example.com 未超限，应仍是 2 条，实际 %d 条", totalB)
	}
}

// TestCleanExcessOperationLogsPerUser_IntColumn 验证 int64 时间列（create_time）的分组清理路径，
// 覆盖和 email/sms（time.Time 列）不同的泛型类型参数分支。
func TestCleanExcessOperationLogsPerUser_IntColumn(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		if err := models.CreateOperationLog(&models.OperationLog{
			UserID: 42, Username: "u42", Module: "m", Action: "a", Method: "GET", Path: "/x",
		}); err != nil {
			t.Fatalf("CreateOperationLog 失败: %v", err)
		}
	}

	affected, err := models.CleanExcessOperationLogsPerUser(3)
	if err != nil {
		t.Fatalf("CleanExcessOperationLogsPerUser 失败: %v", err)
	}
	if affected != 2 {
		t.Fatalf("应删除 2 条，实际删除 %d 条", affected)
	}

	_, total, err := models.GetOperationLogList(&models.OperationLogQuery{Page: 1, PageSize: 20, UserID: 42})
	if err != nil {
		t.Fatalf("GetOperationLogList 失败: %v", err)
	}
	if total != 3 {
		t.Fatalf("user_id=42 应剩 3 条，实际 %d 条", total)
	}
}

// TestCleanExcessAPIAccessLogsPerUser_ExcludesAnonymous 验证 extraWhere="user_id > 0" 生效，
// 未登录（user_id=0）的记录不参与「按用户保留 N 条」清理。
func TestCleanExcessAPIAccessLogsPerUser_ExcludesAnonymous(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		if err := models.CreateAPIAccessLog(&models.APIAccessLog{
			RequestID: fmt.Sprintf("test-anon-%d", i),
			UserID:    0, Method: "GET", Path: "/anon", RoutePath: "/anon", StatusCode: 200,
		}); err != nil {
			t.Fatalf("CreateAPIAccessLog(匿名) 失败: %v", err)
		}
	}

	affected, err := models.CleanExcessAPIAccessLogsPerUser(2)
	if err != nil {
		t.Fatalf("CleanExcessAPIAccessLogsPerUser 失败: %v", err)
	}
	if affected != 0 {
		t.Fatalf("user_id=0 不应参与按用户清理，实际删除了 %d 条", affected)
	}
}
