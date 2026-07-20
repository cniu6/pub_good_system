package models

import (
	"log"
	"strings"
	"time"

	"fst/backend/pkg/db"
)

// 公告状态
const (
	AnnouncementStatusDraft       uint8 = 0 // 草稿
	AnnouncementStatusPublished   uint8 = 1 // 已发布
	AnnouncementStatusUnpublished uint8 = 2 // 下架
)

// Announcement 站内公告（管理员发布，按用户记已读）
type Announcement struct {
	ID          uint64 `db:"id" json:"id"`
	Title       string `db:"title" json:"title"`
	Summary     string `db:"summary" json:"summary"` // 列表预览纯文本（铃铛/工作台），建议不超过 2 行
	Content     string `db:"content" json:"content"` // Markdown/HTML 原文
	Type        string `db:"type" json:"type"`       // info/success/warning/error
	Status      uint8  `db:"status" json:"status"`   // 0草稿 1已发布 2下架
	Priority    int    `db:"priority" json:"priority"`
	Popup       uint8  `db:"popup" json:"popup"`               // 1=登录后可弹窗提示
	TargetType  string `db:"target_type" json:"target_type"`   // all / role
	TargetValue string `db:"target_value" json:"target_value"` // user / admin；all 时空
	StartAt     int64  `db:"start_at" json:"start_at"`         // 0=不限
	EndAt       int64  `db:"end_at" json:"end_at"`             // 0=不限
	PublishedAt int64  `db:"published_at" json:"published_at"`
	CreatedBy   uint64 `db:"created_by" json:"created_by"`
	UpdatedBy   uint64 `db:"updated_by" json:"updated_by"`
	CreatedAt   int64  `db:"created_at" json:"created_at"`
	UpdatedAt   int64  `db:"updated_at" json:"updated_at"`
	DeletedAt   int64  `db:"deleted_at" json:"deleted_at"` // 0=未删
}

// AnnouncementSummaryMaxRunes 列表预览最大字符数（约 2 行）
const AnnouncementSummaryMaxRunes = 80

// announcementSelectColumns 显式列出 announcements 表全部列（带 a. 前缀）
// 不用 a.* 是为了避免和后面拼接的 is_read 计算列产生扫描歧义/顺序依赖，
// 同时表结构变更时能第一时间在这里发现列不匹配，而不是线上 500 才发现。
const announcementSelectColumns = `a.id, a.title, a.summary, a.content, a.type, a.status, a.priority, a.popup,
		a.target_type, a.target_value, a.start_at, a.end_at, a.published_at,
		a.created_by, a.updated_by, a.created_at, a.updated_at, a.deleted_at`

// AnnouncementWithRead 用户侧列表项（带已读）
type AnnouncementWithRead struct {
	Announcement
	IsReadInt int  `db:"is_read" json:"-"`
	IsRead    bool `db:"-" json:"is_read"`
}

// normalizeReadFlag 把 SQL 的 0/1 转成 bool（Scan 后调用）
func (a *AnnouncementWithRead) normalizeReadFlag() {
	a.IsRead = a.IsReadInt != 0
}

// UserAnnouncementRead 用户公告已读记录
type UserAnnouncementRead struct {
	ID             uint64 `db:"id" json:"id"`
	UserID         uint64 `db:"user_id" json:"user_id"`
	AnnouncementID uint64 `db:"announcement_id" json:"announcement_id"`
	ReadAt         int64  `db:"read_at" json:"read_at"`
}

// InitAnnouncementTables 创建公告表与已读表
func InitAnnouncementTables() {
	if !db.CheckTableExists("announcements") {
		schema := `CREATE TABLE IF NOT EXISTS announcements (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(200) NOT NULL COMMENT '标题',
			summary VARCHAR(255) NOT NULL DEFAULT '' COMMENT '列表预览纯文本',
			content MEDIUMTEXT NOT NULL COMMENT '正文 Markdown/HTML',
			type VARCHAR(20) NOT NULL DEFAULT 'info' COMMENT '类型 info/success/warning/error',
			status TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0草稿 1已发布 2下架',
			priority INT NOT NULL DEFAULT 0 COMMENT '越大越靠前',
			popup TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否登录弹窗',
			target_type VARCHAR(20) NOT NULL DEFAULT 'all' COMMENT 'all/role',
			target_value VARCHAR(50) NOT NULL DEFAULT '' COMMENT 'role 目标值',
			start_at BIGINT NOT NULL DEFAULT 0 COMMENT '生效开始 Unix，0不限',
			end_at BIGINT NOT NULL DEFAULT 0 COMMENT '生效结束 Unix，0不限',
			published_at BIGINT NOT NULL DEFAULT 0 COMMENT '发布时间',
			created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
			updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
			created_at BIGINT NOT NULL DEFAULT 0,
			updated_at BIGINT NOT NULL DEFAULT 0,
			deleted_at BIGINT NOT NULL DEFAULT 0 COMMENT '软删，0未删',
			INDEX idx_ann_status_time (status, start_at, end_at),
			INDEX idx_ann_deleted_status (deleted_at, status),
			INDEX idx_ann_priority_pub (priority, published_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
		if _, err := db.Exec(schema); err != nil {
			log.Printf("[Init] Failed to create announcements: %v", err)
		} else {
			log.Println("[Init] Created announcements table")
		}
	}

	// 旧表补 summary 预览列
	if db.CheckTableExists("announcements") && !db.CheckColumnExists("announcements", "summary") {
		if _, err := db.Exec(`ALTER TABLE announcements ADD COLUMN summary VARCHAR(255) NOT NULL DEFAULT '' COMMENT '列表预览纯文本' AFTER title`); err != nil {
			log.Printf("[Init] Failed to add announcements.summary: %v", err)
		} else {
			log.Println("[Init] Added announcements.summary")
		}
	}
	if !db.CheckTableExists("user_announcement_reads") {
		schema := `CREATE TABLE IF NOT EXISTS user_announcement_reads (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT UNSIGNED NOT NULL,
			announcement_id BIGINT UNSIGNED NOT NULL,
			read_at BIGINT NOT NULL DEFAULT 0,
			UNIQUE KEY uk_user_announcement (user_id, announcement_id),
			INDEX idx_uar_user (user_id),
			INDEX idx_uar_announcement (announcement_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
		if _, err := db.Exec(schema); err != nil {
			log.Printf("[Init] Failed to create user_announcement_reads: %v", err)
		} else {
			log.Println("[Init] Created user_announcement_reads table")
		}
	}
}

// IsAnnouncementEnabled 读取系统开关 announcement_enabled
func IsAnnouncementEnabled() bool {
	s, err := GetSettingByKey("announcement_enabled")
	if err != nil || s == nil {
		return false
	}
	return s.Value == "true" || s.Value == "1"
}

// CreateAnnouncement 创建公告（默认草稿）
func CreateAnnouncement(a *Announcement) (uint64, error) {
	now := time.Now().Unix()
	a.CreatedAt = now
	a.UpdatedAt = now
	if a.Type == "" {
		a.Type = "info"
	}
	if a.TargetType == "" {
		a.TargetType = "all"
	}
	a.Summary = NormalizeAnnouncementSummary(a.Summary, a.Content)
	res, err := db.DB.NamedExec(`INSERT INTO announcements (
		title, summary, content, type, status, priority, popup, target_type, target_value,
		start_at, end_at, published_at, created_by, updated_by, created_at, updated_at, deleted_at
	) VALUES (
		:title, :summary, :content, :type, :status, :priority, :popup, :target_type, :target_value,
		:start_at, :end_at, :published_at, :created_by, :updated_by, :created_at, :updated_at, :deleted_at
	)`, a)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return uint64(id), err
}

// UpdateAnnouncement 更新公告（不含发布状态流转）
func UpdateAnnouncement(a *Announcement) error {
	a.UpdatedAt = time.Now().Unix()
	a.Summary = NormalizeAnnouncementSummary(a.Summary, a.Content)
	_, err := db.DB.NamedExec(`UPDATE announcements SET
		title=:title, summary=:summary, content=:content, type=:type, priority=:priority, popup=:popup,
		target_type=:target_type, target_value=:target_value, start_at=:start_at, end_at=:end_at,
		updated_by=:updated_by, updated_at=:updated_at
	WHERE id=:id AND deleted_at=0`, a)
	return err
}

// GetAnnouncementByID 管理端按 ID 取（含草稿，不含已硬删）
func GetAnnouncementByID(id uint64) (*Announcement, error) {
	var a Announcement
	err := db.DB.Get(&a, `SELECT * FROM announcements WHERE id=? AND deleted_at=0`, id)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// SoftDeleteAnnouncement 软删
func SoftDeleteAnnouncement(id, adminID uint64) error {
	now := time.Now().Unix()
	_, err := db.DB.Exec(`UPDATE announcements SET deleted_at=?, updated_by=?, updated_at=? WHERE id=? AND deleted_at=0`,
		now, adminID, now, id)
	return err
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
	_, err = db.DB.Exec(`UPDATE announcements SET status=?, published_at=?, updated_by=?, updated_at=? WHERE id=? AND deleted_at=0`,
		status, publishedAt, adminID, now, id)
	return err
}

// AdminListAnnouncements 管理端分页列表
func AdminListAnnouncements(page, pageSize int, status *uint8, typ, keyword string) ([]Announcement, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	where := []string{"deleted_at=0"}
	args := []interface{}{}
	if status != nil {
		where = append(where, "status=?")
		args = append(args, *status)
	}
	if typ != "" {
		where = append(where, "type=?")
		args = append(args, typ)
	}
	if keyword != "" {
		where = append(where, "(title LIKE ? OR content LIKE ?)")
		like := "%" + keyword + "%"
		args = append(args, like, like)
	}
	w := strings.Join(where, " AND ")
	var total int
	if err := db.DB.Get(&total, "SELECT COUNT(*) FROM announcements WHERE "+w, args...); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	var list []Announcement
	err := db.DB.Select(&list, `SELECT * FROM announcements WHERE `+w+
		` ORDER BY priority DESC, published_at DESC, id DESC LIMIT ? OFFSET ?`, args...)
	return list, total, err
}

// ListVisibleAnnouncementsForUser 用户可见公告列表（含已读）
func ListVisibleAnnouncementsForUser(userID uint64, userRole string, unreadOnly, popupOnly bool, limit int) ([]AnnouncementWithRead, error) {
	now := time.Now().Unix()
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	visAliased := `a.deleted_at=0 AND a.status=? AND (a.start_at=0 OR a.start_at<=?) AND (a.end_at=0 OR a.end_at>=?)
		AND (a.target_type='all' OR (a.target_type='role' AND a.target_value=?))`
	if popupOnly {
		visAliased += " AND a.popup=1"
	}
	if unreadOnly {
		visAliased += " AND r.id IS NULL"
	}
	q := `SELECT ` + announcementSelectColumns + `, CASE WHEN r.id IS NULL THEN 0 ELSE 1 END AS is_read
		FROM announcements a
		LEFT JOIN user_announcement_reads r ON r.announcement_id=a.id AND r.user_id=?
		WHERE ` + visAliased + `
		ORDER BY a.priority DESC, a.published_at DESC, a.id DESC
		LIMIT ?`
	args := []interface{}{userID, AnnouncementStatusPublished, now, now, userRole, limit}
	var list []AnnouncementWithRead
	err := db.DB.Select(&list, q, args...)
	for i := range list {
		list[i].normalizeReadFlag()
	}
	return list, err
}

// GetVisibleAnnouncementForUser 用户取单条可见公告
func GetVisibleAnnouncementForUser(userID, id uint64, userRole string) (*AnnouncementWithRead, error) {
	now := time.Now().Unix()
	var item AnnouncementWithRead
	err := db.DB.Get(&item, `SELECT `+announcementSelectColumns+`, CASE WHEN r.id IS NULL THEN 0 ELSE 1 END AS is_read
		FROM announcements a
		LEFT JOIN user_announcement_reads r ON r.announcement_id=a.id AND r.user_id=?
		WHERE a.id=? AND a.deleted_at=0 AND a.status=?
		  AND (a.start_at=0 OR a.start_at<=?) AND (a.end_at=0 OR a.end_at>=?)
		  AND (a.target_type='all' OR (a.target_type='role' AND a.target_value=?))`,
		userID, id, AnnouncementStatusPublished, now, now, userRole)
	if err != nil {
		return nil, err
	}
	item.normalizeReadFlag()
	return &item, nil
}

// CountUnreadAnnouncements 未读数
func CountUnreadAnnouncements(userID uint64, userRole string) (int, error) {
	now := time.Now().Unix()
	var n int
	err := db.DB.Get(&n, `SELECT COUNT(*) FROM announcements a
		LEFT JOIN user_announcement_reads r ON r.announcement_id=a.id AND r.user_id=?
		WHERE a.deleted_at=0 AND a.status=?
		  AND (a.start_at=0 OR a.start_at<=?) AND (a.end_at=0 OR a.end_at>=?)
		  AND (a.target_type='all' OR (a.target_type='role' AND a.target_value=?))
		  AND r.id IS NULL`,
		userID, AnnouncementStatusPublished, now, now, userRole)
	return n, err
}

// MarkAnnouncementRead 标记单条已读（幂等）
func MarkAnnouncementRead(userID, announcementID uint64) error {
	now := time.Now().Unix()
	_, err := db.DB.Exec(db.Q(`INSERT INTO user_announcement_reads (user_id, announcement_id, read_at)
		VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE read_at=VALUES(read_at)`),
		userID, announcementID, now)
	return err
}

// MarkAllAnnouncementsRead 将当前可见未读全部标已读
func MarkAllAnnouncementsRead(userID uint64, userRole string) error {
	now := time.Now().Unix()
	var ids []uint64
	err := db.DB.Select(&ids, `SELECT a.id FROM announcements a
		LEFT JOIN user_announcement_reads r ON r.announcement_id=a.id AND r.user_id=?
		WHERE a.deleted_at=0 AND a.status=?
		  AND (a.start_at=0 OR a.start_at<=?) AND (a.end_at=0 OR a.end_at>=?)
		  AND (a.target_type='all' OR (a.target_type='role' AND a.target_value=?))
		  AND r.id IS NULL`,
		userID, AnnouncementStatusPublished, now, now, userRole)
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

// NormalizeAnnouncementSummary 规范化列表预览：空则从正文生成；超长截断
func NormalizeAnnouncementSummary(summary, content string) string {
	s := strings.TrimSpace(summary)
	if s == "" {
		return PlainTextSummary(content, AnnouncementSummaryMaxRunes)
	}
	// 去掉换行压成列表可用的短文案（前端再 line-clamp 2 行）
	s = strings.Join(strings.Fields(strings.ReplaceAll(s, "\r", "\n")), " ")
	r := []rune(s)
	if len(r) > AnnouncementSummaryMaxRunes {
		return string(r[:AnnouncementSummaryMaxRunes]) + "…"
	}
	return s
}

// DisplaySummary 用户列表展示用：优先 summary，否则从正文截
func DisplaySummary(a Announcement) string {
	if strings.TrimSpace(a.Summary) != "" {
		return NormalizeAnnouncementSummary(a.Summary, "")
	}
	return PlainTextSummary(a.Content, AnnouncementSummaryMaxRunes)
}
