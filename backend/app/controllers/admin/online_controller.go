package admin

import (
	"fst/backend/app/models"
	"fst/backend/pkg/presence"
	"fst/backend/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// OnlineController 提供 Presence 在线统计与会话强制下线能力。
type OnlineController struct{}

func NewOnlineController() *OnlineController { return &OnlineController{} }

func (ctrl *OnlineController) Stats(c *gin.Context) {
	users, err := models.CountOnlineUsers()
	if err != nil {
		utils.Fail(c, 500, "Failed to count online users")
		return
	}
	sessions, err := models.CountOnlineSessions()
	if err != nil {
		utils.Fail(c, 500, "Failed to count online sessions")
		return
	}
	utils.Success(c, gin.H{"online_users": users, "online_sessions": sessions})
}

func (ctrl *OnlineController) ListSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	sessions, total, err := models.ListOnlineSessions(
		strings.TrimSpace(c.Query("keyword")),
		strings.TrimSpace(c.Query("client_type")),
		strings.TrimSpace(c.Query("auth_guard")),
		page, pageSize,
	)
	if err != nil {
		utils.Fail(c, 500, "Failed to list online sessions")
		return
	}
	utils.Success(c, gin.H{"list": sessions, "total": total, "page": page, "page_size": pageSize})
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
	if err := models.AdminRevokeSessionByID(sessionID); err != nil {
		utils.Fail(c, 500, "Failed to revoke session")
		return
	}
	presence.DefaultHub().Kick(sessionID, "revoked_by_admin")
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
	ids, err := models.ListActiveSessionIDsExceptCurrent(userID, authGuard, "")
	if err != nil {
		utils.Fail(c, 500, "Failed to load sessions")
		return
	}
	if err := models.AdminRevokeAllUserSessions(userID, authGuard); err != nil {
		utils.Fail(c, 500, "Failed to revoke sessions")
		return
	}
	presence.DefaultHub().KickMany(ids, "revoked_by_admin")
	utils.Success(c, gin.H{"message": "All sessions revoked", "count": len(ids)})
}
