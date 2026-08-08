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

	// 翻译模型名为中文（从 ModelNameCN 映射读取）
	translateModelNames(root)

	return json.Marshal(root)
}

// translateModelNames 翻译 definitions 的 key 为中文，并更新所有 $ref 引用。
// 映射表来自 modelname_cn.go 的 ModelNameCN。
func translateModelNames(root map[string]any) {
	defs, ok := root["definitions"].(map[string]any)
	if !ok || len(defs) == 0 {
		return
	}
	if len(ModelNameCN) == 0 {
		return
	}

	// 构建旧名 -> 新名映射（完整 key）
	oldToNew := make(map[string]string, len(defs))
	for key := range defs {
		for en, cn := range ModelNameCN {
			if key == en || strings.HasSuffix(key, "."+en) {
				oldToNew[key] = cn
				break
			}
		}
	}
	if len(oldToNew) == 0 {
		return
	}

	// 替换 definitions 的 key
	newDefs := make(map[string]any, len(defs))
	for oldKey, val := range defs {
		if newKey, ok := oldToNew[oldKey]; ok {
			newDefs[newKey] = val
		} else {
			newDefs[oldKey] = val
		}
	}
	root["definitions"] = newDefs

	// 递归替换整个文档中所有 $ref 字符串
	replaceRefs(root)
}

// replaceRefs 递归遍历 JSON 树，替换所有形如 `#/definitions/旧名` 的 $ref 值。
func replaceRefs(v any) {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			if k == "$ref" {
				if ref, ok := child.(string); ok && strings.HasPrefix(ref, "#/definitions/") {
					suffix := ref[len("#/definitions/"):]
					if cn, found := modelNameFromOld(suffix); found {
						val[k] = "#/definitions/" + cn
					}
				}
			} else {
				replaceRefs(child)
			}
		}
	case []any:
		for i := range val {
			replaceRefs(val[i])
		}
	}
}

// modelNameFromOld 从 ModelNameCN 中查找旧名对应的中文名。
func modelNameFromOld(old string) (string, bool) {
	for en, cn := range ModelNameCN {
		if old == en || strings.HasSuffix(old, "."+en) {
			return cn, true
		}
	}
	return "", false
}
