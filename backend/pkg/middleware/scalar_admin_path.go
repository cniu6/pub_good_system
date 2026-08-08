package middleware

import (
	"encoding/json"
	"fst/backend/pkg/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag"
)

// ScalarAdminPathRewriteMiddleware 为 Scalar 返回运行时改写后的 openapi.json。
// 控制器注解仍写 /api/v1/admin/*，当 ADMIN_API_PATH 不是默认 /admin 时，
// 将 openapi.json 里的 /api/v1/admin 前缀改写为实际前缀。
// 同时把 host 替换为当前请求的真实 Host（含端口），让 Scalar Try It 直接命中本服务。
func ScalarAdminPathRewriteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminAPIPath := "/admin"
		if cfg := config.CloneGlobalConfig(); cfg != nil {
			adminAPIPath = config.NormalizeAdminAPIPath(cfg.AdminAPIPath)
		}

		realHost := c.Request.Host
		if forwardedHost := c.GetHeader("X-Forwarded-Host"); forwardedHost != "" {
			realHost = forwardedHost
		}

		doc, err := buildRewrittenSwaggerDoc(adminAPIPath, realHost)
		if err != nil || doc == nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusOK, "application/json; charset=utf-8", doc)
		c.Abort()
	}
}

// readSwaggerDocJSON 读取已注册 swag 文档并序列化为 JSON 字节。
func readSwaggerDocJSON() ([]byte, error) {
	spec := swag.GetSwagger(swag.Name)
	if spec == nil {
		spec = swag.GetSwagger("swagger")
	}
	if spec == nil {
		return nil, nil
	}

	raw := strings.TrimSpace(spec.ReadDoc())
	if raw == "" {
		return nil, nil
	}
	// 即使原样输出也做一次 JSON 校验/紧凑化，确保格式正确
	var root map[string]any
	if err := json.Unmarshal([]byte(raw), &root); err != nil {
		return nil, err
	}
	return json.Marshal(root)
}

// buildRewrittenSwaggerDoc 读取已注册的 swag 文档，重写 paths 中的管理端前缀，
// 并把 host 替换为当前请求的真实 Host（含端口）。
func buildRewrittenSwaggerDoc(adminAPIPath, realHost string) ([]byte, error) {
	doc, err := readSwaggerDocJSON()
	if err != nil || doc == nil {
		return doc, err
	}

	var root map[string]any
	if err := json.Unmarshal(doc, &root); err != nil {
		return nil, err
	}

	// 替换 host 为当前请求的真实 Host，确保 Scalar Try It 请求到正确的地址
	if realHost != "" {
		hostValue, _ := root["host"].(string)
		if hostValue != realHost {
			root["host"] = realHost
		}
	}

	paths, ok := root["paths"].(map[string]any)
	if !ok || paths == nil {
		return doc, nil
	}

	const oldPrefix = "/api/v1/admin"
	newPrefix := "/api/v1" + adminAPIPath
	rewritten := make(map[string]any, len(paths))
	for pathKey, pathVal := range paths {
		// 仅精确前缀段匹配，避免误伤 /api/v1/administrator 一类路径
		if pathKey == oldPrefix || strings.HasPrefix(pathKey, oldPrefix+"/") {
			// /api/v1/admin/users -> /api/v1{自定义}/users
			rewritten[newPrefix+strings.TrimPrefix(pathKey, oldPrefix)] = pathVal
			continue
		}
		rewritten[pathKey] = pathVal
	}
	root["paths"] = rewritten

	// 文档里补充说明当前实际管理端前缀，方便排查
	if info, ok := root["info"].(map[string]any); ok && info != nil {
		desc, _ := info["description"].(string)
		note := " [Admin API Path: " + newPrefix + "]"
		if !strings.Contains(desc, "[Admin API Path:") {
			info["description"] = strings.TrimSpace(desc + note)
		}
	}

	// 注入 x-tagGroups，使 Scalar 侧边栏按 Public/User/Admin/System 四大组规整
	injectTagGroups(root)

	return json.Marshal(root)
}

// injectTagGroups 注入 x-tagGroups 扩展，将各 tag 分组到四大导航组下。
// Scalar / Redoc 支持 x-tagGroups 实现分组侧边栏。
func injectTagGroups(root map[string]any) {
	root["x-tagGroups"] = []map[string]any{
		{"name": "📢 Public 公共接口", "tags": []string{"Public-认证", "Public-配置", "Public-区号", "Public-回调", "Public-会话"}},
		{"name": "👤 User 用户接口", "tags": []string{"User-资料", "User-支付", "User-实名", "User-提现", "User-公告", "User-会话"}},
		{"name": "🔧 Admin 管理员接口", "tags": []string{
			"Admin-公告", "Admin-邮件日志", "Admin-邮件模板", "Admin-短信日志", "Admin-短信模板",
			"Admin-用户", "Admin-用户等级", "Admin-用户积分", "Admin-支付", "Admin-实名",
			"Admin-设置", "Admin-操作日志", "Admin-API日志", "Admin-仪表盘", "Admin-自动任务",
			"Admin-在线用户", "Admin-调试", "Admin-终端", "Admin-待办", "Admin-数据库",
		}},
		{"name": "⚙ System 系统管理", "tags": []string{"System-管理"}},
	}
}
