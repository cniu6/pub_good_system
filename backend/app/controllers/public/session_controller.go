package public

import (
	"fst/backend/app/models"
	"fst/backend/utils"

	"github.com/gin-gonic/gin"
)

// SessionController 公开会话清理接口（无需鉴权中间件，自行校验签名）
type SessionController struct{}

// NewSessionController 创建会话控制器
func NewSessionController() *SessionController {
	return &SessionController{}
}

// ForceLogout 尽力撤销「当前这枚 access token」对应的会话，即使 token 已过期也能清理。
// 场景：前端因 access token 过期被强制退出登录时，正常 /user/logout 需要鉴权中间件放行，
// 过期 token 根本进不到业务代码；这里改用"只验证签名、不校验 exp"的解析方式，
// 按 token_hash 定位并撤销，避免过期会话一直残留到定时清理任务才被扫掉。
//
// 范围（刻意收窄）：只吊销 user_id + auth_guard + token_hash 完全匹配的那一条会话，
// 同账号其他设备/浏览器的会话不受影响（不会走 RevokeAllUserSessions）。
// 出于安全考虑：仅验证 JWT 签名合法性，不放宽签名/角色校验；始终返回成功，不暴露 token 状态。
// @Summary 强制退出登录（容忍 token 已过期）
// @Tags Public-会话
// @Accept json
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/public/session/force-logout [post]
func (ctrl *SessionController) ForceLogout(c *gin.Context) {
	token := utils.ExtractBearerToken(c.GetHeader("Authorization"))
	if token == "" {
		utils.Success(c, gin.H{"message": "ok"})
		return
	}

	// 用本枚 token 的哈希精确匹配会话行；其它会话的 token_hash 不同，不会被更新。
	tokenHash := utils.HashToken(token)
	for _, guard := range []string{utils.UserAuthGuard, utils.AdminAuthGuard} {
		claims, err := utils.ParseTokenForGuardIgnoreExpiry(token, guard)
		if err != nil {
			continue
		}
		_ = models.RevokeSessionByTokenHash(claims.UserID, guard, tokenHash)
		break
	}

	// 无论是否命中会话，都返回成功：避免向未认证方暴露 token 是否存在/有效等信息。
	utils.Success(c, gin.H{"message": "ok"})
}

// RegisterRoutes 注册会话相关公开路由
func (ctrl *SessionController) RegisterRoutes(group *gin.RouterGroup) {
	session := group.Group("/session")
	{
		session.POST("/force-logout", ctrl.ForceLogout)
	}
}
