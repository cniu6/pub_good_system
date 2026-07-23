package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"fst/backend/pkg/db"

	"gorm.io/gorm"
)

// 公告状态
const (
	AnnouncementStatusDraft       uint8 = 0 // 草稿
	AnnouncementStatusPublished   uint8 = 1 // 已发布
	AnnouncementStatusUnpublished uint8 = 2 // 下架
)

// Announcement 站内公告（管理员发布，按用户记已读）
type Announcement struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title       string `gorm:"column:title;size:200;not null" json:"title"`
	Summary     string `gorm:"column:summary;size:255;not null;default:''" json:"summary"`
	Content     string `gorm:"column:content;type:mediumtext;not null" json:"content"`
	Type        string `gorm:"column:type;size:20;not null;default:'info'" json:"type"`
	Status      uint8  `gorm:"column:status;not null;default:0;index:idx_ann_status_time,priority:1;index:idx_ann_deleted_status,priority:2" json:"status"`
	Priority    int    `gorm:"column:priority;not null;default:0;index:idx_ann_priority_pub,priority:1" json:"priority"`
	Popup       uint8  `gorm:"column:popup;not null;default:0" json:"popup"`
	TargetType  string `gorm:"column:target_type;size:20;not null;default:'all'" json:"target_type"`
	TargetValue string `gorm:"column:target_value;size:50;not null;default:''" json:"target_value"`
	StartAt     int64  `gorm:"column:start_at;not null;default:0;index:idx_ann_status_time,priority:2" json:"start_at"`
	EndAt       int64  `gorm:"column:end_at;not null;default:0;index:idx_ann_status_time,priority:3" json:"end_at"`
	PublishedAt int64  `gorm:"column:published_at;not null;default:0;index:idx_ann_priority_pub,priority:2" json:"published_at"`
	CreatedBy   uint64 `gorm:"column:created_by;not null;default:0" json:"created_by"`
	UpdatedBy   uint64 `gorm:"column:updated_by;not null;default:0" json:"updated_by"`
	CreatedAt   int64  `gorm:"column:created_at;not null;default:0" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at;not null;default:0" json:"updated_at"`
	DeletedAt   int64  `gorm:"column:deleted_at;not null;default:0;index:idx_ann_deleted_status,priority:1" json:"deleted_at"`
}

// TableName 表名
func (Announcement) TableName() string { return "announcements" }

// AnnouncementSummaryMaxRunes 列表预览最大字符数（约 2 行）
const AnnouncementSummaryMaxRunes = 80

const announcementSelectColumns = `a.id, a.title, a.summary, a.content, a.type, a.status, a.priority, a.popup,
		a.target_type, a.target_value, a.start_at, a.end_at, a.published_at,
		a.created_by, a.updated_by, a.created_at, a.updated_at, a.deleted_at`

// AnnouncementWithRead 用户侧列表项（带已读）
type AnnouncementWithRead struct {
	Announcement
	IsReadInt int  `gorm:"column:is_read" json:"-"`
	IsRead    bool `gorm:"-" json:"is_read"`
}

func (a *AnnouncementWithRead) normalizeReadFlag() {
	a.IsRead = a.IsReadInt != 0
}

// UserAnnouncementRead 用户公告已读记录
type UserAnnouncementRead struct {
	ID             uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID         uint64 `gorm:"column:user_id;not null;uniqueIndex:uk_user_announcement,priority:1;index:idx_uar_user" json:"user_id"`
	AnnouncementID uint64 `gorm:"column:announcement_id;not null;uniqueIndex:uk_user_announcement,priority:2;index:idx_uar_announcement" json:"announcement_id"`
	ReadAt         int64  `gorm:"column:read_at;not null;default:0" json:"read_at"`
}

// TableName 表名
func (UserAnnouncementRead) TableName() string { return "user_announcement_reads" }

// IsAnnouncementEnabled 读取系统开关 announcement_enabled
func IsAnnouncementEnabled() bool {
	s, err := GetSettingByKey("announcement_enabled")
	if err != nil || s == nil {
		return false
	}
	return s.Value == "true" || s.Value == "1"
}

func prepareAnnouncement(a *Announcement) {
	now := time.Now().Unix()
	if a.CreatedAt == 0 {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	if a.Type == "" {
		a.Type = "info"
	}
	if a.TargetType == "" {
		a.TargetType = "all"
	}
	a.Summary = NormalizeAnnouncementSummary(a.Summary, a.Content)
}

// CreateAnnouncement 创建公告（默认草稿）
func CreateAnnouncement(a *Announcement) (uint64, error) {
	prepareAnnouncement(a)
	if err := db.DB.Create(a).Error; err != nil {
		return 0, err
	}
	return a.ID, nil
}

// UpdateAnnouncement 更新公告（不含发布状态流转）
func UpdateAnnouncement(a *Announcement) error {
	a.UpdatedAt = time.Now().Unix()
	a.Summary = NormalizeAnnouncementSummary(a.Summary, a.Content)
	return db.DB.Model(&Announcement{}).Where("id = ? AND deleted_at = 0", a.ID).Updates(map[string]any{
		"title": a.Title, "summary": a.Summary, "content": a.Content, "type": a.Type,
		"priority": a.Priority, "popup": a.Popup, "target_type": a.TargetType, "target_value": a.TargetValue,
		"start_at": a.StartAt, "end_at": a.EndAt, "updated_by": a.UpdatedBy, "updated_at": a.UpdatedAt,
	}).Error
}

// GetAnnouncementByID 管理端按 ID 取（含草稿，不含已硬删）
func GetAnnouncementByID(id uint64) (*Announcement, error) {
	var a Announcement
	err := db.DB.Where("id = ? AND deleted_at = 0", id).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// SoftDeleteAnnouncement 软删
func SoftDeleteAnnouncement(id, adminID uint64) error {
	now := time.Now().Unix()
	return db.DB.Model(&Announcement{}).Where("id = ? AND deleted_at = 0", id).Updates(map[string]any{
		"deleted_at": now, "updated_by": adminID, "updated_at": now,
	}).Error
}

// SetAnnouncementStatus 发布/下架
func SetAnnouncementStatus(id uint64, status uint8, adminID uint64) error {
	now := time.Now().Unix()
	existing, err := GetAnnouncementByID(id)
	if err != nil {
		return err
	}
	publishedAt := existing.PublishedAt
	if status == AnnouncementStatusPublished && publishedAt == 0 {
		publishedAt = now
	}
	return db.DB.Model(&Announcement{}).Where("id = ? AND deleted_at = 0", id).Updates(map[string]any{
		"status": status, "published_at": publishedAt, "updated_by": adminID, "updated_at": now,
	}).Error
}

// AdminListAnnouncements 管理端分页列表
func AdminListAnnouncements(page, pageSize int, status *uint8, typ, keyword string) ([]Announcement, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	q := db.DB.Model(&Announcement{}).Where("deleted_at = 0")
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if typ != "" {
		q = q.Where("type = ?", typ)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("title LIKE ? OR content LIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []Announcement
	err := q.Order("priority DESC, published_at DESC, id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, int(total), err
}

func visibleAnnouncementSQL(popupOnly, unreadOnly bool) string {
	vis := `a.deleted_at=0 AND a.status=? AND (a.start_at=0 OR a.start_at<=?) AND (a.end_at=0 OR a.end_at>=?)
		AND (a.target_type='all' OR (a.target_type='role' AND a.target_value=?))`
	if popupOnly {
		vis += " AND a.popup=1"
	}
	if unreadOnly {
		vis += " AND r.id IS NULL"
	}
	return vis
}

// ListVisibleAnnouncementsForUser 用户可见公告列表（含已读）
func ListVisibleAnnouncementsForUser(userID uint64, userRole string, unreadOnly, popupOnly bool, limit int) ([]AnnouncementWithRead, error) {
	now := time.Now().Unix()
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	q := `SELECT ` + announcementSelectColumns + `, CASE WHEN r.id IS NULL THEN 0 ELSE 1 END AS is_read
		FROM announcements a
		LEFT JOIN user_announcement_reads r ON r.announcement_id=a.id AND r.user_id=?
		WHERE ` + visibleAnnouncementSQL(popupOnly, unreadOnly) + `
		ORDER BY a.priority DESC, a.published_at DESC, a.id DESC
		LIMIT ?`
	var list []AnnouncementWithRead
	err := db.DB.Raw(q, userID, AnnouncementStatusPublished, now, now, userRole, limit).Scan(&list).Error
	for i := range list {
		list[i].normalizeReadFlag()
	}
	return list, err
}

// GetVisibleAnnouncementForUser 用户取单条可见公告
func GetVisibleAnnouncementForUser(userID, id uint64, userRole string) (*AnnouncementWithRead, error) {
	now := time.Now().Unix()
	var item AnnouncementWithRead
	err := db.DB.Raw(`SELECT `+announcementSelectColumns+`, CASE WHEN r.id IS NULL THEN 0 ELSE 1 END AS is_read
		FROM announcements a
		LEFT JOIN user_announcement_reads r ON r.announcement_id=a.id AND r.user_id=?
		WHERE a.id=? AND a.deleted_at=0 AND a.status=?
		  AND (a.start_at=0 OR a.start_at<=?) AND (a.end_at=0 OR a.end_at>=?)
		  AND (a.target_type='all' OR (a.target_type='role' AND a.target_value=?))`,
		userID, id, AnnouncementStatusPublished, now, now, userRole).Scan(&item).Error
	if err != nil {
		return nil, err
	}
	if item.ID == 0 {
		return nil, sql.ErrNoRows
	}
	item.normalizeReadFlag()
	return &item, nil
}

// CountUnreadAnnouncements 未读数
func CountUnreadAnnouncements(userID uint64, userRole string) (int, error) {
	now := time.Now().Unix()
	var n int64
	err := db.DB.Raw(`SELECT COUNT(*) FROM announcements a
		LEFT JOIN user_announcement_reads r ON r.announcement_id=a.id AND r.user_id=?
		WHERE a.deleted_at=0 AND a.status=?
		  AND (a.start_at=0 OR a.start_at<=?) AND (a.end_at=0 OR a.end_at>=?)
		  AND (a.target_type='all' OR (a.target_type='role' AND a.target_value=?))
		  AND r.id IS NULL`,
		userID, AnnouncementStatusPublished, now, now, userRole).Scan(&n).Error
	return int(n), err
}

// MarkAnnouncementRead 标记单条已读（幂等）
func MarkAnnouncementRead(userID, announcementID uint64) error {
	now := time.Now().Unix()
	r := db.DB.Model(&UserAnnouncementRead{}).
		Where("user_id = ? AND announcement_id = ?", userID, announcementID).
		Update("read_at", now)
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected > 0 {
		return nil
	}
	err := db.DB.Create(&UserAnnouncementRead{
		UserID: userID, AnnouncementID: announcementID, ReadAt: now,
	}).Error
	if db.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

// MarkAllAnnouncementsRead 将当前可见未读全部标已读
func MarkAllAnnouncementsRead(userID uint64, userRole string) error {
	now := time.Now().Unix()
	var ids []uint64
	err := db.DB.Raw(`SELECT a.id FROM announcements a
		LEFT JOIN user_announcement_reads r ON r.announcement_id=a.id AND r.user_id=?
		WHERE a.deleted_at=0 AND a.status=?
		  AND (a.start_at=0 OR a.start_at<=?) AND (a.end_at=0 OR a.end_at>=?)
		  AND (a.target_type='all' OR (a.target_type='role' AND a.target_value=?))
		  AND r.id IS NULL`,
		userID, AnnouncementStatusPublished, now, now, userRole).Scan(&ids).Error
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := MarkAnnouncementRead(userID, id); err != nil {
			return err
		}
	}
	return nil
}

// ListDashboardAnnouncements 工作台最近 N 条摘要
func ListDashboardAnnouncements(userID uint64, userRole string, limit int) ([]AnnouncementWithRead, error) {
	if limit <= 0 {
		limit = 5
	}
	return ListVisibleAnnouncementsForUser(userID, userRole, false, false, limit)
}

// PlainTextSummary 从 content 粗提纯文本摘要（列表用）
func PlainTextSummary(content string, maxLen int) string {
	s := content
	for {
		start := strings.Index(s, "<")
		if start < 0 {
			break
		}
		end := strings.Index(s[start:], ">")
		if end < 0 {
			break
		}
		s = s[:start] + " " + s[start+end+1:]
	}
	s = strings.ReplaceAll(s, "#", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.Join(strings.Fields(s), " ")
	r := []rune(s)
	if maxLen > 0 && len(r) > maxLen {
		return string(r[:maxLen]) + "…"
	}
	return s
}

// NormalizeAnnouncementSummary 规范化列表预览
func NormalizeAnnouncementSummary(summary, content string) string {
	s := strings.TrimSpace(summary)
	if s == "" {
		return PlainTextSummary(content, AnnouncementSummaryMaxRunes)
	}
	s = strings.Join(strings.Fields(strings.ReplaceAll(s, "\r", "\n")), " ")
	r := []rune(s)
	if len(r) > AnnouncementSummaryMaxRunes {
		return string(r[:AnnouncementSummaryMaxRunes]) + "…"
	}
	return s
}

// DisplaySummary 用户列表展示用
func DisplaySummary(a Announcement) string {
	if strings.TrimSpace(a.Summary) != "" {
		return NormalizeAnnouncementSummary(a.Summary, "")
	}
	return PlainTextSummary(a.Content, AnnouncementSummaryMaxRunes)
}
