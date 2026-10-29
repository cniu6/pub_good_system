package models_test

import (
	"fmt"
	"testing"
	"time"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

// TestOnlineSessionReturnsUserIdentity 在线列表（含管理端按用户归并的接口）必须带出用户名/昵称。
// 回归背景：UserSession.Username/Nickname 曾标 `gorm:"-"`，JOIN users 选出的列被 Scan 忽略，
// 管理端在线列表的用户名一列为空。这里同时覆盖原始会话行和归并行，防止只在其中一条路径修好。
// TestOnlineUsersGroupedPagination 归并列表按用户分页，total 由数据库统计，不随页大小变化。
func TestOnlineUsersGroupedPagination(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	now := time.Now().Unix()
	for i := 0; i < 3; i++ {
		user := testutil.CreateTestUser(t, fmt.Sprintf("group-page-%d", i))
		if err := models.CreateUserSession(
			user.ID, "user", fmt.Sprintf("group-token-%d", i), fmt.Sprintf("group-refresh-%d", i),
			"127.0.0.1", "test-agent", "pc", "web", "",
			now+3600, now+7200,
		); err != nil {
			t.Fatalf("CreateUserSession(%d): %v", i, err)
		}
		// 让 last_seen_at 有明确先后，验证排序与分页边界
		if err := db.DB.Model(&models.UserSession{}).Where("user_id = ?", user.ID).
			Update("last_seen_at", now-int64(i)).Error; err != nil {
			t.Fatalf("更新 last_seen_at: %v", err)
		}
	}

	page1, total, err := models.ListOnlineUsersGrouped("", "", "user", 1, 2)
	if err != nil {
		t.Fatalf("ListOnlineUsersGrouped page1: %v", err)
	}
	if total != 3 || len(page1) != 2 {
		t.Fatalf("第一页应为 2 条、总数 3，实际 total=%d len=%d", total, len(page1))
	}
	if page1[0].Username != "group-page-0" || page1[1].Username != "group-page-1" {
		t.Fatalf("应按最近在线倒序: %q, %q", page1[0].Username, page1[1].Username)
	}
	for _, row := range page1 {
		if len(row.Devices) != 1 || row.Devices[0].Username != row.Username {
			t.Fatalf("设备明细未正确挂载: %+v", row)
		}
	}

	page2, total2, err := models.ListOnlineUsersGrouped("", "", "user", 2, 2)
	if err != nil {
		t.Fatalf("ListOnlineUsersGrouped page2: %v", err)
	}
	if total2 != 3 || len(page2) != 1 || page2[0].Username != "group-page-2" {
		t.Fatalf("第二页应只剩最旧一条: total=%d rows=%+v", total2, page2)
	}
}

func TestOnlineSessionReturnsUserIdentity(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	user := testutil.CreateTestUser(t, "xdb-online")
	now := time.Now().Unix()
	if err := models.CreateUserSession(
		user.ID, "user", "online-token-hash", "online-refresh-hash",
		"127.0.0.1", "test-agent", "pc", "web", "",
		now+3600, now+7200,
	); err != nil {
		t.Fatalf("CreateUserSession: %v", err)
	}

	sessions, total, err := models.ListOnlineSessions("", "", "user", 1, 20)
	if err != nil {
		t.Fatalf("ListOnlineSessions: %v", err)
	}
	if total != 1 || len(sessions) != 1 {
		t.Fatalf("在线会话数量不符: total=%d len=%d", total, len(sessions))
	}
	if sessions[0].Username != user.Username || sessions[0].Nickname != user.Nickname {
		t.Fatalf("在线会话用户名未回填: username=%q nickname=%q", sessions[0].Username, sessions[0].Nickname)
	}

	rows, groupedTotal, err := models.ListOnlineUsersGrouped("", "", "user", 1, 20)
	if err != nil {
		t.Fatalf("ListOnlineUsersGrouped: %v", err)
	}
	if groupedTotal != 1 || len(rows) != 1 {
		t.Fatalf("归并在线用户数量不符: total=%d len=%d", groupedTotal, len(rows))
	}
	if rows[0].Username != user.Username || rows[0].Nickname != user.Nickname {
		t.Fatalf("归并行用户名未回填: username=%q nickname=%q", rows[0].Username, rows[0].Nickname)
	}
	if len(rows[0].Devices) != 1 || rows[0].Devices[0].Username != user.Username {
		t.Fatalf("归并行内设备用户名未回填: %+v", rows[0].Devices)
	}
}
