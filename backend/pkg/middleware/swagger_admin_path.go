package middleware

import (
	"encoding/json"
	"fst/backend/pkg/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag"
)

// SwaggerAdminPathRewriteMiddleware 在返回 swagger doc.json 时，
// 将注解里写死的 /api/v1/admin 前缀按运行时 ADMIN_API_PATH 改写。
// 页面静态资源不改动；仅自定义 admin 前缀时生效。
func SwaggerAdminPathRewriteMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// gin-swagger 的 any 参数形如 /index.html、/doc.json
		anyPath := c.Param("any")
		if !isSwaggerDocJSONPath(anyPath) {
			c.Next()
			return
		}

		adminAPIPath := "/admin"
		if cfg := config.CloneGlobalConfig(); cfg != nil {
			adminAPIPath = config.NormalizeAdminAPIPath(cfg.AdminAPIPath)
		}
		// 默认 /admin 时注解路径已正确，直接交给 gin-swagger
		if adminAPIPath == "/admin" {
			c.Next()
			return
		}

		doc, err := buildRewrittenSwaggerDoc(adminAPIPath)
		if err != nil || doc == nil {
			c.Next()
			return
		}

		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Data(http.StatusOK, "application/json; charset=utf-8", doc)
		c.Abort()
	}
}

func isSwaggerDocJSONPath(anyPath string) bool {
	p := strings.TrimSpace(anyPath)
	if p == "" {
		return false
	}
	// 兼容 /doc.json、doc.json、/swagger/doc.json 等形态
	return strings.HasSuffix(strings.ToLower(p), "doc.json")
}

// buildRewrittenSwaggerDoc 读取已注册的 swag 文档，重写 paths 中的管理端前缀。
func buildRewrittenSwaggerDoc(adminAPIPath string) ([]byte, error) {
	// 取默认实例（与 docs 包 Register 的 InstanceName 一致，一般为 "swagger"）
	spec := swag.GetSwagger(swag.Name)
	if spec == nil {
		// 部分版本用固定名 "swagger"
		spec = swag.GetSwagger("swagger")
	}
	if spec == nil {
		return nil, nil
	}

	raw := strings.TrimSpace(spec.ReadDoc())
	if raw == "" {
		return nil, nil
	}

	var root map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &root); err != nil {
		return nil, err
	}

	paths, ok := root["paths"].(map[string]interface{})
	if !ok || paths == nil {
		return []byte(raw), nil
	}

	const oldPrefix = "/api/v1/admin"
	newPrefix := "/api/v1" + adminAPIPath
	rewritten := make(map[string]interface{}, len(paths))
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
	if info, ok := root["info"].(map[string]interface{}); ok && info != nil {
		desc, _ := info["description"].(string)
		note := " [Admin API Path: " + newPrefix + "]"
		if !strings.Contains(desc, "[Admin API Path:") {
			info["description"] = strings.TrimSpace(desc + note)
		}
	}

	return json.Marshal(root)
}
