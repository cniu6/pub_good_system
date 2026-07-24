package admin

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/presence"
	"fst/backend/utils"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// OnlineController 提供 Presence 在线统计与会话强制下线能力。
type OnlineController struct{}

func NewOnlineController() *OnlineController { return &OnlineController{} }

// requirePresenceEnabled 在线列表/统计依赖心跳；总开关关闭时直接拒绝。
// 用户详情里的会话列表/踢下线不走这里，仍可管理 JWT 会话。
func requirePresenceEnabled(c *gin.Context) bool {
	if services.GetGlobalPresenceEnabled() {
		return true
	}
	utils.Fail(c, 403, "在线状态功能未启用")
	return false
}

func (ctrl *OnlineController) Stats(c *gin.Context) {
	if !requirePresenceEnabled(c) {
		return
	}
	users, err := models.CountOnlineUsers()
	if err != nil {
		log.Printf("[ADMIN][ONLINE] count online users failed: %v", err)
		utils.Fail(c, 500, "Failed to count online users")
		return
	}
	sessions, err := models.CountOnlineSessions()
	if err != nil {
		log.Printf("[ADMIN][ONLINE] count online sessions failed: %v", err)
		utils.Fail(c, 500, "Failed to count online sessions")
		return
	}
	utils.Success(c, gin.H{"online_users": users, "online_sessions": sessions})
}

func (ctrl *OnlineController) ListSessions(c *gin.Context) {
	if !requirePresenceEnabled(c) {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = utils.NormalizePagination(page, pageSize)
	// 按用户归并：一行多设备，便于管理端展示
	rows, total, err := models.ListOnlineUsersGrouped(
		strings.TrimSpace(c.Query("keyword")),
		strings.TrimSpace(c.Query("client_type")),
		strings.TrimSpace(c.Query("auth_guard")),
		page, pageSize,
	)
	if err != nil {
		log.Printf("[ADMIN][ONLINE] list sessions failed: %v", err)
		utils.Fail(c, 500, "Failed to list online sessions")
		return
	}
	utils.Success(c, gin.H{"list": rows, "total": total, "page": page, "page_size": pageSize})
}

// UserSessions 返回指定用户的有效会话，包含当前离线但未撤销的设备。
func (ctrl *OnlineController) UserSessions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID == 0 {
		utils.Fail(c, 400, "Invalid user ID")
		return
	}
	authGuard := strings.TrimSpace(c.DefaultQuery("auth_guard", "user"))
	sessions, err := models.GetUserSessionsWithGuard(userID, authGuard, "")
	if err != nil {
		log.Printf("[ADMIN][ONLINE] load user sessions failed user_id=%d guard=%s: %v", userID, authGuard, err)
		utils.Fail(c, 500, "Failed to load user sessions")
		return
	}
	utils.Success(c, sessions)
}

func (ctrl *OnlineController) RevokeSession(c *gin.Context) {
	sessionID, err := models.ParseSessionID(c.Param("id"))
	if err != nil {
		utils.Fail(c, 400, "Invalid session ID")
		return
	}
	adminID, _ := c.Get("userID")
	if err := models.AdminRevokeSessionByID(sessionID); err != nil {
		log.Printf("[ADMIN][ONLINE] revoke session failed admin_id=%v session_id=%d: %v", adminID, sessionID, err)
		utils.Fail(c, 500, "Failed to revoke session")
		return
	}
	presence.DefaultHub().Kick(sessionID, "revoked_by_admin")
	log.Printf("[SECURITY AUDIT] admin revoke session | admin_id=%v session_id=%d ip=%s", adminID, sessionID, c.ClientIP())
	utils.Success(c, gin.H{"message": "Session revoked"})
}

func (ctrl *OnlineController) RevokeAllUserSessions(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || userID == 0 {
		utils.Fail(c, 400, "Invalid user ID")
		return
	}
	var request struct {
		AuthGuard string `json:"auth_guard"`
	}
	_ = c.ShouldBindJSON(&request)
	authGuard := strings.TrimSpace(request.AuthGuard)
	if authGuard == "" {
		authGuard = "user"
	}
	if authGuard != "user" && authGuard != "admin" {
		utils.Fail(c, 400, "Invalid auth_guard")
		return
	}
	adminID, _ := c.Get("userID")
	ids, err := models.ListActiveSessionIDsExceptCurrent(userID, authGuard, "")
	if err != nil {
		log.Printf("[ADMIN][ONLINE] list active sessions failed admin_id=%v target_user_id=%d: %v", adminID, userID, err)
		utils.Fail(c, 500, "Failed to load sessions")
		return
	}
	if err := models.AdminRevokeAllUserSessions(userID, authGuard); err != nil {
		log.Printf("[ADMIN][ONLINE] revoke all sessions failed admin_id=%v target_user_id=%d guard=%s: %v",
			adminID, userID, authGuard, err)
		utils.Fail(c, 500, "Failed to revoke sessions")
		return
	}
	presence.DefaultHub().KickMany(ids, "revoked_by_admin")
	log.Printf("[SECURITY AUDIT] admin revoke all sessions | admin_id=%v target_user_id=%d guard=%s count=%d ip=%s",
		adminID, userID, authGuard, len(ids), c.ClientIP())
	utils.Success(c, gin.H{"message": "All sessions revoked", "count": len(ids)})
}
