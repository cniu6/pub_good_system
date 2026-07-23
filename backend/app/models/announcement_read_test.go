package models_test

import (
	"testing"

	"fst/backend/app/models"
	"fst/backend/internal/testutil"
	"fst/backend/pkg/db"
)

// TestMarkAnnouncementReadIdempotent 回归 SQLite 下 MarkAnnouncementRead 幂等：
// user_announcement_reads 是复合唯一键；实现为 UPDATE-then-INSERT，应可重复标记且不报错。
func TestMarkAnnouncementReadIdempotent(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "ann_reader")

	annID, err := models.CreateAnnouncement(&models.Announcement{
		Title:      "hello",
		Content:    "world",
		Status:     models.AnnouncementStatusPublished,
		TargetType: "all",
	})
	if err != nil {
		t.Fatalf("CreateAnnouncement 失败: %v", err)
	}
	// 置为已发布并回填 published_at，确保用户端可见。
	if err := models.SetAnnouncementStatus(annID, models.AnnouncementStatusPublished, u.ID); err != nil {
		t.Fatalf("SetAnnouncementStatus 失败: %v", err)
	}

	if n, err := models.CountUnreadAnnouncements(u.ID, u.Role); err != nil || n < 1 {
		t.Fatalf("初始未读数应>=1，实际 n=%d err=%v", n, err)
	}

	// 连续标记两次：均应成功、不报错（回归点）。
	if err := models.MarkAnnouncementRead(u.ID, annID); err != nil {
		t.Fatalf("第一次 MarkAnnouncementRead 失败: %v", err)
	}
	if err := models.MarkAnnouncementRead(u.ID, annID); err != nil {
		t.Fatalf("第二次 MarkAnnouncementRead（幂等）失败: %v", err)
	}

	var cnt int64
	if err := db.DB.Raw("SELECT COUNT(*) FROM user_announcement_reads WHERE user_id=? AND announcement_id=?", u.ID, annID).Scan(&cnt).Error; err != nil {
		t.Fatalf("查询已读记录失败: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("已读记录应恰为 1 条，实际 %d", cnt)
	}

	if n, err := models.CountUnreadAnnouncements(u.ID, u.Role); err != nil || n != 0 {
		t.Fatalf("标记后未读数应为 0，实际 n=%d err=%v", n, err)
	}
}

// TestMarkAllAnnouncementsReadIdempotent 覆盖批量标记路径同样可在 SQLite 正常工作。
func TestMarkAllAnnouncementsReadIdempotent(t *testing.T) {
	cleanup := testutil.SetupSQLite(t)
	defer cleanup()

	u := testutil.CreateTestUser(t, "ann_reader_all")

	for i := 0; i < 3; i++ {
		annID, err := models.CreateAnnouncement(&models.Announcement{
			Title:      "t",
			Content:    "c",
			Status:     models.AnnouncementStatusPublished,
			TargetType: "all",
		})
		if err != nil {
			t.Fatalf("CreateAnnouncement 失败: %v", err)
		}
		if err := models.SetAnnouncementStatus(annID, models.AnnouncementStatusPublished, u.ID); err != nil {
			t.Fatalf("SetAnnouncementStatus 失败: %v", err)
		}
	}

	if err := models.MarkAllAnnouncementsRead(u.ID, u.Role); err != nil {
		t.Fatalf("MarkAllAnnouncementsRead 失败: %v", err)
	}
	// 再调一次应幂等。
	if err := models.MarkAllAnnouncementsRead(u.ID, u.Role); err != nil {
		t.Fatalf("MarkAllAnnouncementsRead（幂等）失败: %v", err)
	}

	if n, err := models.CountUnreadAnnouncements(u.ID, u.Role); err != nil || n != 0 {
		t.Fatalf("全部标记后未读数应为 0，实际 n=%d err=%v", n, err)
	}
}
