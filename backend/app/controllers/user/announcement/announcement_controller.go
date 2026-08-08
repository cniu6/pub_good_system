package announcement

import (
	"log"
	"strconv"

	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// AnnouncementController 用户端公告
type AnnouncementController struct {
	userSvc *services.UserService
}

func NewAnnouncementController() *AnnouncementController {
	return &AnnouncementController{userSvc: services.NewUserService()}
}

func userIDFromCtx(c *gin.Context) (uint64, bool) {
	return utils.GetUserID(c)
}

func (ctrl *AnnouncementController) userRole(uid uint64) string {
	u, err := ctrl.userSvc.GetByID(uid)
	if err != nil || u == nil {
		return "user"
	}
	if u.Role == "admin" {
		return "admin"
	}
	return "user"
}

// List 我的公告列表
// @Summary 我的公告列表
// @Tags User-公告
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/user/announcements [get]
func (ctrl *AnnouncementController) List(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		utils.Fail(c, 401, "Unauthorized")
		return
	}
	if !models.IsAnnouncementEnabled() {
		utils.Success(c, gin.H{"list": []interface{}{}, "enabled": false})
		return
	}
	unreadOnly := c.Query("unread_only") == "1" || c.Query("unread_only") == "true"
	popupOnly := c.Query("popup") == "1" || c.Query("popup") == "true"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	role := ctrl.userRole(uid)
	list, err := models.ListVisibleAnnouncementsForUser(uid, role, unreadOnly, popupOnly, limit)
	if err != nil {
		// 打印真实错误方便排查（用户消息仍保持友好，不暴露内部细节）
		log.Printf("[Announcement] ListVisibleAnnouncementsForUser failed: uid=%d role=%s unreadOnly=%v popupOnly=%v limit=%d err=%v",
			uid, role, unreadOnly, popupOnly, limit, err)
		utils.Fail(c, 500, "Failed to list announcements")
		return
	}
	// 列表附带摘要字段，方便铃铛展示
	out := make([]gin.H, 0, len(list))
	for _, item := range list {
		out = append(out, gin.H{
			"id":           item.ID,
			"title":        item.Title,
			"content":      item.Content,
			"summary":      models.DisplaySummary(item.Announcement),
			"type":         item.Type,
			"popup":        item.Popup,
			"published_at": item.PublishedAt,
			"is_read":      item.IsRead,
		})
	}
	utils.Success(c, gin.H{"list": out, "enabled": true})
}

// Detail 公告详情
// @Summary 公告详情
// @Tags User-公告
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/user/announcements/{id} [get]
func (ctrl *AnnouncementController) Detail(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		utils.Fail(c, 401, "Unauthorized")
		return
	}
	if !models.IsAnnouncementEnabled() {
		utils.Fail(c, 404, "Announcement not found")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "Invalid id")
		return
	}
	item, err := models.GetVisibleAnnouncementForUser(uid, id, ctrl.userRole(uid))
	if err != nil {
		utils.Fail(c, 404, "Announcement not found")
		return
	}
	utils.Success(c, item)
}

// MarkRead 标记已读
// @Summary 标记公告已读
// @Tags User-公告
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/user/announcements/{id}/read [post]
func (ctrl *AnnouncementController) MarkRead(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		utils.Fail(c, 401, "Unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Fail(c, 400, "Invalid id")
		return
	}
	// 必须可见才允许记已读
	if _, err := models.GetVisibleAnnouncementForUser(uid, id, ctrl.userRole(uid)); err != nil {
		utils.Fail(c, 404, "Announcement not found")
		return
	}
	if err := models.MarkAnnouncementRead(uid, id); err != nil {
		log.Printf("[Announcement] MarkAnnouncementRead failed: uid=%d id=%d err=%v", uid, id, err)
		utils.Fail(c, 500, "Failed to mark read")
		return
	}
	utils.Success(c, gin.H{"message": "ok"})
}

// MarkAllRead 全部已读
// @Summary 标记全部公告已读
// @Tags User-公告
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/user/announcements/read-all [post]
func (ctrl *AnnouncementController) MarkAllRead(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		utils.Fail(c, 401, "Unauthorized")
		return
	}
	if err := models.MarkAllAnnouncementsRead(uid, ctrl.userRole(uid)); err != nil {
		log.Printf("[Announcement] MarkAllAnnouncementsRead failed: uid=%d err=%v", uid, err)
		utils.Fail(c, 500, "Failed to mark all read")
		return
	}
	utils.Success(c, gin.H{"message": "ok"})
}

// RegisterRoutes 注册用户端公告路由（统一在控制器内管理，避免 routes 层散落）。
func (ctrl *AnnouncementController) RegisterRoutes(group *gin.RouterGroup) {
	anns := group.Group("/announcements")
	{
		anns.GET("", ctrl.List)
		anns.GET("/unread-count", ctrl.UnreadCount)
		anns.POST("/read-all", ctrl.MarkAllRead)
		anns.GET("/:id", ctrl.Detail)
		anns.POST("/:id/read", ctrl.MarkRead)
	}
}

// UnreadCount 未读角标
// @Summary 未读公告数量
// @Tags User-公告
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /v1/user/announcements/unread-count [get]
func (ctrl *AnnouncementController) UnreadCount(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		utils.Fail(c, 401, "Unauthorized")
		return
	}
	if !models.IsAnnouncementEnabled() {
		utils.Success(c, gin.H{"count": 0, "enabled": false})
		return
	}
	n, err := models.CountUnreadAnnouncements(uid, ctrl.userRole(uid))
	if err != nil {
		log.Printf("[Announcement] CountUnreadAnnouncements failed: uid=%d err=%v", uid, err)
		utils.Fail(c, 500, "Failed to count")
		return
	}
	utils.Success(c, gin.H{"count": n, "enabled": true})
}
