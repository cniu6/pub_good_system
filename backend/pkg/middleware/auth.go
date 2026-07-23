package middleware

import (
	"fst/backend/app/models"
	"fst/backend/app/services"
	"fst/backend/pkg/config"
	"fst/backend/utils"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// AuthMethodJWT JWT Bearer 鉴权
	AuthMethodJWT = "jwt"
	// AuthMethodAPIKey X-Api-Key 鉴权
	AuthMethodAPIKey = "apikey"
	// AuthMethodNone 未鉴权（公开接口）
	AuthMethodNone = "none"
)

func AuthMiddleware() gin.HandlerFunc {
	return AuthMiddlewareForGuard(utils.UserAuthGuard)
}

// RejectApiKeyOnAuthPaths 认证公开接口禁止携带 X-Api-Key（必须走账密/验证码）。
func RejectApiKeyOnAuthPaths() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.TrimSpace(c.GetHeader("X-Api-Key")) != "" && IsAuthSensitiveAPIPath(c.Request.URL.Path) {
			utils.Fail(c, 403, "API Key cannot be used on authentication endpoints")
			c.Abort()
			return
		}
		c.Next()
	}
}

// AuthMiddlewareForGuard 支持传入一个或多个可接受的 guard。
// 例如用户路由传 ("user","admin") 表示管理员token也能访问用户接口。
// 同时支持 Authorization: Bearer <jwt> 与 X-Api-Key: <key>。
func AuthMiddlewareForGuard(acceptGuards ...string) gin.HandlerFunc {
	if len(acceptGuards) == 0 {
		acceptGuards = []string{utils.UserAuthGuard}
	}
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		apiKey := strings.TrimSpace(c.GetHeader("X-Api-Key"))

		if authHeader != "" {
			authenticateWithJWT(c, authHeader, acceptGuards)
			return
		}
		if apiKey != "" {
			authenticateWithAPIKey(c, apiKey, acceptGuards)
			return
		}

		utils.Fail(c, 401, "Authorization header or X-Api-Key is required")
		c.Abort()
	}
}

func authenticateWithJWT(c *gin.Context, authHeader string, acceptGuards []string) {
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		utils.Fail(c, 401, "Authorization header format must be Bearer {token}")
		c.Abort()
		return
	}

	var claims *utils.Claims
	var parseErr error
	for _, guard := range acceptGuards {
		claims, parseErr = utils.ParseTokenForGuard(parts[1], guard)
		if parseErr == nil {
			break
		}
	}
	if parseErr != nil {
		utils.Fail(c, 401, "Invalid or expired token")
		c.Abort()
		return
	}

	actualGuard := claims.AuthGuard
	if actualGuard == "" {
		actualGuard = utils.UserAuthGuard
	}
	active, err := models.IsUserSessionActive(claims.UserID, actualGuard, utils.HashToken(parts[1]))
	if err != nil || !active {
		utils.Fail(c, 401, "Session expired or revoked")
		c.Abort()
		return
	}

	user, err := models.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		utils.Fail(c, 401, "User not found")
		c.Abort()
		return
	}
	if !setAuthUserContext(c, user, actualGuard, AuthMethodJWT) {
		return
	}
	c.Next()
}

func authenticateWithAPIKey(c *gin.Context, apiKey string, acceptGuards []string) {
	// 全局总开关：默认关闭，管理员未在后台主动开启前，一律拒绝 X-Api-Key 鉴权方式（必须走 Bearer JWT）。
	// 该开关读的是内存缓存（services.GlobalSettingsService），不会每次请求都查库；
	// 管理员在后台保存后会立即刷新缓存，单机部署下无需重启即可生效。
	if !services.GetGlobalAPIKeyAuthEnabled() {
		utils.Fail(c, 403, "API Key authentication is disabled")
		c.Abort()
		return
	}

	// 防御：即便误挂到认证公开路由，也拒绝
	if IsAuthSensitiveAPIPath(c.Request.URL.Path) {
		utils.Fail(c, 403, "API Key cannot be used on authentication endpoints")
		c.Abort()
		return
	}

	// 默认禁止 API Key 调用管理端写接口（资金/配置变更等高危操作只允许 JWT 会话）
	if isAdminMutatingRequest(c) {
		log.Printf("[SECURITY] API Key blocked on admin write | ip=%s method=%s path=%s", c.ClientIP(), c.Request.Method, c.Request.URL.Path)
		utils.Fail(c, 403, "API Key cannot be used on admin write endpoints")
		c.Abort()
		return
	}

	user, err := models.GetUserByApiKey(apiKey)
	if err != nil || user == nil {
		utils.Fail(c, 401, "Invalid API Key")
		c.Abort()
		return
	}

	// 用户等级能力：是否允许使用 API Key
	if ok, msg := models.CheckUserLevelAllows(user.ID, "api_key"); !ok {
		utils.Fail(c, 403, msg)
		c.Abort()
		return
	}

	// IP 白名单（用户级 apikey_allow_ips，逗号分隔；空=不限制）
	if !apiKeyIPAllowed(user, c.ClientIP()) {
		log.Printf("[SECURITY] API Key IP denied | user_id=%d ip=%s", user.ID, c.ClientIP())
		utils.Fail(c, 403, "API Key not allowed from this IP")
		c.Abort()
		return
	}

	// 过期检查
	if user.ApikeyExpiresAt != nil && *user.ApikeyExpiresAt > 0 && *user.ApikeyExpiresAt < time.Now().Unix() {
		utils.Fail(c, 401, "API Key expired")
		c.Abort()
		return
	}

	// 解析 guard：管理员且路由接受 admin → admin；否则若接受 user（含管理员访问用户接口）→ user
	actualGuard := ""
	if user.Role == utils.AdminAuthGuard && guardAccepted(acceptGuards, utils.AdminAuthGuard) {
		actualGuard = utils.AdminAuthGuard
	} else if guardAccepted(acceptGuards, utils.UserAuthGuard) {
		actualGuard = utils.UserAuthGuard
	}
	if actualGuard == "" {
		utils.Fail(c, 403, "Insufficient permissions")
		c.Abort()
		return
	}

	// scope：若配置了 apikey_scopes，管理端路径要求含 admin 或 *
	if !apiKeyScopeAllowed(user, c.Request.URL.Path, actualGuard) {
		utils.Fail(c, 403, "API Key scope insufficient")
		c.Abort()
		return
	}

	if !setAuthUserContext(c, user, actualGuard, AuthMethodAPIKey) {
		return
	}
	log.Printf("[AUDIT][APIKEY] user_id=%d guard=%s method=%s path=%s ip=%s", user.ID, actualGuard, c.Request.Method, c.Request.URL.Path, c.ClientIP())
	c.Next()
}

func isAdminMutatingRequest(c *gin.Context) bool {
	method := strings.ToUpper(c.Request.Method)
	if method == "GET" || method == "HEAD" || method == "OPTIONS" {
		return false
	}
	path := strings.ToLower(c.Request.URL.Path)
	// 匹配 /api/v1/{admin_api_path}/...
	return strings.Contains(path, "/api/v1/") && (strings.Contains(path, "/admin/") || isConfiguredAdminAPIPath(path))
}

func isConfiguredAdminAPIPath(path string) bool {
	cfg := config.CloneGlobalConfig()
	if cfg == nil {
		return false
	}
	adminPath := config.NormalizeAdminAPIPath(cfg.AdminAPIPath)
	if adminPath == "" || adminPath == "/admin" {
		return false
	}
	needle := "/api/v1" + strings.ToLower(adminPath) + "/"
	return strings.Contains(path, needle)
}

func apiKeyIPAllowed(user *models.User, clientIP string) bool {
	if user == nil || user.ApikeyAllowIPs == nil {
		return true
	}
	raw := strings.TrimSpace(*user.ApikeyAllowIPs)
	if raw == "" {
		return true
	}
	clientIP = strings.TrimSpace(clientIP)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if part == "*" || part == clientIP {
			return true
		}
	}
	return false
}

func apiKeyScopeAllowed(user *models.User, path, guard string) bool {
	if user == nil || user.ApikeyScopes == nil {
		return true
	}
	raw := strings.TrimSpace(*user.ApikeyScopes)
	if raw == "" {
		return true
	}
	scopes := map[string]struct{}{}
	for _, s := range strings.Split(strings.ToLower(raw), ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			scopes[s] = struct{}{}
		}
	}
	if _, ok := scopes["*"]; ok {
		return true
	}
	path = strings.ToLower(path)
	if guard == utils.AdminAuthGuard || strings.Contains(path, "/admin/") {
		_, ok := scopes["admin"]
		return ok
	}
	_, ok := scopes["user"]
	return ok
}

func guardAccepted(acceptGuards []string, guard string) bool {
	for _, g := range acceptGuards {
		if g == guard {
			return true
		}
	}
	return false
}

func setAuthUserContext(c *gin.Context, user *models.User, actualGuard, authMethod string) bool {
	if user.Status == 0 {
		utils.Fail(c, 403, "Account is inactive")
		c.Abort()
		return false
	}
	if user.LockUntil != nil && *user.LockUntil > time.Now().Unix() {
		utils.Fail(c, 403, "Account is locked")
		c.Abort()
		return false
	}
	if actualGuard == utils.AdminAuthGuard && user.Role != utils.AdminAuthGuard {
		utils.Fail(c, 403, "Admin access only")
		c.Abort()
		return false
	}

	c.Set("username", user.Username)
	c.Set("userID", user.ID)
	c.Set("role", user.Role)
	c.Set("authGuard", actualGuard)
	c.Set("authMethod", authMethod)
	return true
}

// AdminOnly 验证用户是否为管理员
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		authGuard, _ := c.Get("authGuard")
		if authGuard != utils.AdminAuthGuard {
			utils.Fail(c, 403, "Admin access only")
			c.Abort()
			return
		}

		role, exists := c.Get("role")
		if !exists || role != "admin" {
			userID, _ := c.Get("userID")
			path := c.Request.URL.Path
			method := c.Request.Method
			clientIP := c.ClientIP()
			log.Printf("[SECURITY WARNING] non-admin access attempt | user_id=%v | ip=%s | method=%s | path=%s", userID, clientIP, method, path)
			utils.Fail(c, 403, "Admin access only")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireRole 通用角色验证中间件
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			utils.Fail(c, 403, "Role not found")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			utils.Fail(c, 403, "Invalid role type")
			c.Abort()
			return
		}

		allowed := false
		for _, r := range allowedRoles {
			if roleStr == r {
				allowed = true
				break
			}
		}

		if !allowed {
			utils.Fail(c, 403, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}
