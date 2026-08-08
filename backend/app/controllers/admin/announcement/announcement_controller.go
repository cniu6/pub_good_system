package announcement

import (
	"log"
	"strconv"
	"strings"

	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/presence"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// AnnouncementController 管理端公告 CRUD
type AnnouncementController struct{}

func NewAnnouncementController() *AnnouncementController {
	return &AnnouncementController{}
}

func adminUserID(c *gin.Context) uint64 {
	id, _ := utils.GetUserID(c)
	return id
}

func parseAnnouncementID(c *gin.Context) (uint64, error) {
	return strconv.ParseUint(c.Param("id"), 10, 64)
}

func normalizeAnnouncementType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "success", "warning", "error":
		return strings.ToLower(t)
	default:
		return "info"
	}
}

// List 公告列表
// @Summary 管理端公告列表
// @Tags Admin-公告
// @Security BearerAuth
// @Router /v1/admin/announcements [get]
func (ctrl *AnnouncementController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = utils.NormalizePagination(page, pageSize)
	keyword := strings.TrimSpace(c.Query("keyword"))
	typ := strings.TrimSpace(c.Query("type"))
	var statusPtr *uint8
	if s := c.Query("status"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 8); err == nil {
			u := uint8(v)
			statusPtr = &u
		}
	}
	list, total, err := models.AdminListAnnouncements(page, pageSize, statusPtr, typ, keyword)
	if err != nil {
		utils.Fail(c, 500, "Failed to list announcements")
		return
	}
	utils.Success(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

// Detail 公告详情
// @Summary 管理端公告详情
// @Tags Admin-公告
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/announcements/{id} [get]
func (ctrl *AnnouncementController) Detail(c *gin.Context) {
	id, err := parseAnnouncementID(c)
	if err != nil {
		utils.Fail(c, 400, "Invalid id")
		return
	}
	a, err := models.GetAnnouncementByID(id)
	if err != nil {
		utils.Fail(c, 404, "Announcement not found")
		return
	}
	utils.Success(c, a)
}

// announcementTargetType/announcementTargetValue：产品定调公告面向全体登录用户，不做管理员/用户分层。
// DB 的 target_type/target_value 列与读取路径（ListVisibleAnnouncementsForUser 等）仍支持按角色定向，
// 只是当前创建/编辑入口固定写 "all"/""，先不在 API/前端暴露这两个字段，避免出现「传了但被服务端忽略」的死字段。
const (
	announcementTargetType  = "all"
	announcementTargetValue = ""
)

type announcementUpsertReq struct {
	Title    string `json:"title" binding:"required"`
	Summary  string `json:"summary"` // 列表预览，可选；空则从正文截取
	Content  string `json:"content" binding:"required"`
	Type     string `json:"type"`
	Priority int    `json:"priority"`
	Popup    *uint8 `json:"popup"`
	StartAt  int64  `json:"start_at"`
	EndAt    int64  `json:"end_at"`
}

// Create 创建草稿
// @Summary 管理端公告创建
// @Tags Admin-公告
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/announcements [post]
func (ctrl *AnnouncementController) Create(c *gin.Context) {
	var req announcementUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	popup := uint8(0)
	if req.Popup != nil {
		popup = *req.Popup
	}
	adminID := adminUserID(c)
	title := strings.TrimSpace(req.Title)
	summary := strings.TrimSpace(req.Summary)
	if err := utils.ValidateRuneLen(title, "标题", utils.MaxAnnouncementTitle); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	if err := utils.ValidateRuneLen(summary, "摘要", utils.MaxAnnouncementSummary); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	a := &models.Announcement{
		Title:       title,
		Summary:     summary,
		Content:     req.Content, // 管理员可信内容，保留 Markdown/HTML
		Type:        normalizeAnnouncementType(req.Type),
		Status:      models.AnnouncementStatusDraft,
		Priority:    req.Priority,
		Popup:       popup,
		TargetType:  announcementTargetType,
		TargetValue: announcementTargetValue,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
		CreatedBy:   adminID,
		UpdatedBy:   adminID,
	}
	if a.Title == "" {
		utils.Fail(c, 400, "Title required")
		return
	}
	id, err := models.CreateAnnouncement(a)
	if err != nil {
		utils.Fail(c, 500, "Failed to create announcement")
		return
	}
	a.ID = id
	utils.Success(c, a)
}

// Update 更新
// @Summary 管理端公告更新
// @Tags Admin-公告
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/announcements/{id} [put]
func (ctrl *AnnouncementController) Update(c *gin.Context) {
	id, err := parseAnnouncementID(c)
	if err != nil {
		utils.Fail(c, 400, "Invalid id")
		return
	}
	existing, err := models.GetAnnouncementByID(id)
	if err != nil {
		utils.Fail(c, 404, "Announcement not found")
		return
	}
	var req announcementUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	existing.Title = strings.TrimSpace(req.Title)
	existing.Summary = strings.TrimSpace(req.Summary)
	if err := utils.ValidateRuneLen(existing.Title, "标题", utils.MaxAnnouncementTitle); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	if err := utils.ValidateRuneLen(existing.Summary, "摘要", utils.MaxAnnouncementSummary); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}
	existing.Content = req.Content
	existing.Type = normalizeAnnouncementType(req.Type)
	existing.Priority = req.Priority
	if req.Popup != nil {
		existing.Popup = *req.Popup
	}
	existing.TargetType = announcementTargetType
	existing.TargetValue = announcementTargetValue
	existing.StartAt = req.StartAt
	existing.EndAt = req.EndAt
	existing.UpdatedBy = adminUserID(c)
	if existing.Title == "" {
		utils.Fail(c, 400, "Title required")
		return
	}
	if err := models.UpdateAnnouncement(existing); err != nil {
		utils.Fail(c, 500, "Failed to update announcement")
		return
	}
	utils.Success(c, existing)
}

// Publish 发布
// @Summary 管理端公告发布
// @Tags Admin-公告
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/announcements/{id}/publish [post]
func (ctrl *AnnouncementController) Publish(c *gin.Context) {
	id, err := parseAnnouncementID(c)
	if err != nil {
		utils.Fail(c, 400, "Invalid id")
		return
	}
	if _, err := models.GetAnnouncementByID(id); err != nil {
		utils.Fail(c, 404, "Announcement not found")
		return
	}
	if err := models.SetAnnouncementStatus(id, models.AnnouncementStatusPublished, adminUserID(c)); err != nil {
		utils.Fail(c, 500, "Failed to publish")
		return
	}
	// 发布即自动打开总开关，避免「发了但用户端空白」
	if !models.IsAnnouncementEnabled() {
		var enableErr error
		if services.GlobalSettingsService != nil {
			enableErr = services.GlobalSettingsService.UpdateSingleSettingWithCache("announcement_enabled", "true")
		} else {
			enableErr = models.UpdateSetting("announcement_enabled", "true")
		}
		if enableErr != nil {
			// 公告已发布，开关落库失败不回滚发布，但必须可观测，否则会出现「已发布但用户端空白」难排查
			log.Printf("[Announcement] 发布成功但自动开启 announcement_enabled 失败: id=%d, err=%v", id, enableErr)
		}
	}
	a, _ := models.GetAnnouncementByID(id)
	// Presence 实时推送（仅在开启在线状态时；关闭时用户仍可通过 HTTP 拉公告）
	if a != nil && services.GetGlobalPresenceEnabled() {
		presence.DefaultHub().BroadcastJSON(map[string]interface{}{
			"type":  "announcement",
			"id":    a.ID,
			"title": a.Title,
			"popup": a.Popup == 1,
		})
	}
	utils.Success(c, a)
}

// Unpublish 下架
// @Summary 管理端公告下架
// @Tags Admin-公告
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/announcements/{id}/unpublish [post]
func (ctrl *AnnouncementController) Unpublish(c *gin.Context) {
	id, err := parseAnnouncementID(c)
	if err != nil {
		utils.Fail(c, 400, "Invalid id")
		return
	}
	if err := models.SetAnnouncementStatus(id, models.AnnouncementStatusUnpublished, adminUserID(c)); err != nil {
		utils.Fail(c, 500, "Failed to unpublish")
		return
	}
	a, _ := models.GetAnnouncementByID(id)
	utils.Success(c, a)
}

// Delete 软删
// @Summary 管理端公告删除
// @Tags Admin-公告
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/admin/announcements/{id} [delete]
func (ctrl *AnnouncementController) Delete(c *gin.Context) {
	id, err := parseAnnouncementID(c)
	if err != nil {
		utils.Fail(c, 400, "Invalid id")
		return
	}
	if err := models.SoftDeleteAnnouncement(id, adminUserID(c)); err != nil {
		utils.Fail(c, 500, "Failed to delete")
		return
	}
	utils.Success(c, gin.H{"message": "deleted"})
}
