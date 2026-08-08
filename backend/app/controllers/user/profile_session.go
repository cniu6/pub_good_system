package user

import (
	"fst/backend/app/models"
	"fst/backend/pkg/presence"
	"fst/backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetApiKey 获取当前用户的 API Key 明文（用户中心随时查看/复制；前端用 password 眼睛控制显隐）。
func (ctrl *ProfileController) GetApiKey(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	user, err := ctrl.user_svc.GetByID(uid)
	if err != nil {
		utils.Fail(c, 404, "User not found")
		return
	}

	plain := user.PlainApikeyForOwner()
	var apikeyValue any
	if plain != "" {
		apikeyValue = plain
	}
	utils.Success(c, gin.H{"apikey": apikeyValue})
}

// ========================================
// 登录设备/会话管理
// ========================================

// GetSessions 获取用户登录会话列表
// @Summary 获取登录会话
// @Description 获取当前用户的所有活跃登录会话
// @Tags User-会话
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/sessions [get]
func (ctrl *ProfileController) GetSessions(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	// 根据当前 token 的 authGuard 查询对应会话
	guard, _ := c.Get("authGuard")
	guardStr, _ := guard.(string)
	if guardStr == "" {
		guardStr = "user"
	}
	currentTokenHash := utils.HashToken(utils.ExtractBearerToken(c.GetHeader("Authorization")))
	sessions, err := models.GetUserSessionsWithGuard(uid, guardStr, currentTokenHash)
	if err != nil {
		utils.Fail(c, 500, "Failed to load sessions")
		return
	}

	utils.Success(c, sessions)
}

// RevokeSession 踢出指定会话
// @Summary 踢出会话
// @Description 撤销指定的登录会话
// @Tags User-会话
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "会话ID"
// @Success 200 {object} utils.Response
// @Router /api/v1/user/sessions/:id [delete]
func (ctrl *ProfileController) RevokeSession(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	session_id := c.Param("id")
	if session_id == "" {
		utils.Fail(c, 400, "Session ID is required")
		return
	}

	// 根据当前 token 的 authGuard 撤销对应会话
	guard, _ := c.Get("authGuard")
	guardStr, _ := guard.(string)
	if guardStr == "" {
		guardStr = "user"
	}
	if err := models.RevokeUserSessionWithGuard(uid, guardStr, session_id); err != nil {
		utils.Fail(c, 500, "Failed to revoke session")
		return
	}
	if sessionID, err := models.ParseSessionID(session_id); err == nil {
		presence.DefaultHub().Kick(sessionID, "session_revoked")
	}

	utils.Success(c, gin.H{"message": "Session revoked successfully"})
}

// RevokeAllSessions 踢出所有其他会话
// @Summary 踢出所有其他会话
// @Description 撤销当前会话以外的所有登录会话
// @Tags User-会话
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/sessions/revoke-all [post]
func (ctrl *ProfileController) RevokeAllSessions(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	currentTokenHash := utils.HashToken(utils.ExtractBearerToken(c.GetHeader("Authorization")))
	// 根据当前 token 的 authGuard 撤销对应会话
	guard, _ := c.Get("authGuard")
	guardStr, _ := guard.(string)
	if guardStr == "" {
		guardStr = "user"
	}
	sessionIDs, err := models.ListActiveSessionIDsExceptCurrent(uid, guardStr, currentTokenHash)
	if err != nil {
		utils.Fail(c, 500, "Failed to load sessions")
		return
	}
	if err := models.RevokeAllUserSessionsWithGuard(uid, guardStr, currentTokenHash); err != nil {
		utils.Fail(c, 500, "Failed to revoke sessions")
		return
	}
	presence.DefaultHub().KickMany(sessionIDs, "revoked_by_user")

	utils.Success(c, gin.H{"message": "All other sessions revoked"})
}

// Logout 撤销当前登录会话，并主动关闭其 Presence 连接。
func (ctrl *ProfileController) Logout(c *gin.Context) {
	userID, ok := c.Get("userID")
	uid, ok := userID.(uint64)
	if !ok {
		utils.Fail(c, 401, "User not logged in")
		return
	}
	guard, _ := c.Get("authGuard")
	guardStr, _ := guard.(string)
	if guardStr == "" {
		guardStr = "user"
	}
	tokenHash := utils.HashToken(utils.ExtractBearerToken(c.GetHeader("Authorization")))
	session, err := models.GetActiveSessionByTokenHash(tokenHash)
	if err != nil || session == nil || session.UserID != uid || session.AuthGuard != guardStr {
		utils.Fail(c, 401, "Session expired or revoked")
		return
	}
	if err := models.RevokeUserSessionWithGuard(uid, guardStr, strconv.FormatUint(session.ID, 10)); err != nil {
		utils.Fail(c, 500, "Failed to logout")
		return
	}
	presence.DefaultHub().Kick(session.ID, "logout")
	utils.Success(c, gin.H{"message": "Logged out"})
}

// ========================================
// API Key 管理
// ========================================

// ResetApiKey 重置API密钥
// @Summary 重置API密钥
// @Description 重置当前用户的API密钥
// @Tags User-会话
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Router /api/v1/user/resetapikey [post]
func (ctrl *ProfileController) ResetApiKey(c *gin.Context) {
	user_id, exists := c.Get("userID")
	if !exists {
		utils.Fail(c, 401, "User not logged in")
		return
	}
	uid, ok := user_id.(uint64)
	if !ok {
		utils.Fail(c, 401, "Invalid user session")
		return
	}

	new_key, err := models.ResetUserApiKey(uid)
	if err != nil {
		utils.Fail(c, 500, "Failed to reset API key")
		return
	}

	utils.Success(c, gin.H{"apikey": new_key})
}
