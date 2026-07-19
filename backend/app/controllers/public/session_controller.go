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

// ForceLogout 尽力撤销当前会话记录，即使 access token 已过期也能清理。
// 场景：前端因 access token 过期被强制退出登录时，正常 /user/logout 需要鉴权中间件放行，
// 过期 token 根本进不到业务代码；这里改用"只验证签名、不校验 exp"的解析方式，
// 定位到具体会话记录并标记撤销，避免过期会话一直残留到定时清理任务才被扫掉。
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
