package middleware

import (
	"encoding/json"
	"fst/backend/pkg/config"
	"net/http"
	"sort"
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

	const oldPrefix = "/v1/admin"
	newPrefix := "/v1" + adminAPIPath
	rewritten := make(map[string]any, len(paths))
	for pathKey, pathVal := range paths {
		// 仅精确前缀段匹配，避免误伤 /api/v1/administrator 一类路径
		if pathKey == oldPrefix || strings.HasPrefix(pathKey, oldPrefix+"/") {
			// /v1/admin/users -> /v1{自定义}/users
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

	// 注入嵌套 tag 层级（v1/v2 母目录 + scope/function 子目录）
	// 优先于 x-tagGroups；若 Scalar 不支持 x-parent，再回退到 x-tagGroups
	injectNestedTagHierarchy(root)

	// 将 Swagger 2.0 转换为 OpenAPI 3.1，使 Scalar 能识别 tags.parent 嵌套
	convertToOpenAPI31(root)

	// 翻译模型名为中文（运行时处理）
	// swaggo 不支持 @name 自定义模型名，只能文档生成后替换
	translateModelNames(root)

	return json.Marshal(root)
}

// translateModelNames 翻译 definitions 的 key 为中文，并更新所有 $ref 引用。
func translateModelNames(root map[string]any) {
	defs, ok := root["definitions"].(map[string]any)
	if !ok || len(defs) == 0 {
		return
	}

	oldToNew := make(map[string]string, len(defs))
	for key := range defs {
		for en, cn := range modelNameCN {
			if key == en || strings.HasSuffix(key, "."+en) {
				oldToNew[key] = cn
				break
			}
		}
	}
	if len(oldToNew) == 0 {
		return
	}

	newDefs := make(map[string]any, len(defs))
	for oldKey, val := range defs {
		if newKey, ok := oldToNew[oldKey]; ok {
			newDefs[newKey] = val
		} else {
			newDefs[oldKey] = val
		}
	}
	root["definitions"] = newDefs

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

// pathScopeMap 根据 URL 第三段推断 scope。
var pathScopeMap = map[string]string{
	"public": "Public",
	"user":   "User",
	"admin":  "Admin",
	"system": "System",
}

// pathFunctionMap 根据路径关键词推断 function 中文名。
// 新增业务目录时在这里加一行即可自动归类到 Scalar。
var pathFunctionMap = map[string]string{
	"auth":            "认证",
	"login":           "认证",
	"register":        "认证",
	"send-register":   "认证",
	"forgot-password": "认证",
	"reset-password":  "认证",
	"refresh-token":   "认证",
	"session":         "会话",
	"force-logout":    "会话",
	"phone-country":   "区号",
	"dial-countries":  "区号",
	"geo":             "区号",
	"app-config":      "配置",
	"payment":         "支付",
	"payment-gateway": "支付",
	"gateways":        "支付",
	"orders":          "支付",
	"notify":          "回调",
	"return":          "回调",
	"callback":        "回调",
	"profile":         "资料",
	"apikey":          "资料",
	"password":        "资料",
	"settings":        "设置",
	"avatar":          "设置",
	"stats":           "统计",
	"realname":        "实名",
	"withdraw":        "提现",
	"user":            "用户",
	"users":           "用户",
	"user-levels":     "用户等级",
	"money":           "用户积分",
	"score":           "用户积分",
	"money-logs":      "用户积分",
	"score-logs":      "用户积分",
	"online":          "在线用户",
	"presence":        "在线用户",
	"announcements":   "公告",
	"announcement":    "公告",
	"email-logs":      "邮件日志",
	"email-templates": "邮件模板",
	"sms-logs":        "短信日志",
	"sms-templates":   "短信模板",
	"api-logs":        "API日志",
	"logs":            "日志",
	"dashboard":       "仪表盘",
	"terminal":        "调试",
	"debug":           "调试",
	"db":              "数据库",
	"auto-jobs":       "自动任务",
	"todos":           "待办",
	"system":          "管理",
	"device":          "本机数据",
}

// inferTagFromPath 根据接口路径自动推断 tag。
// 当 controller 没写 @Tags 或 tag 格式不对时，按 /api/{version}/{scope}/... 自动归类。
func inferTagFromPath(path string) string {
	// 去掉可能存在的 /api 前缀
	path = strings.TrimPrefix(path, "/api")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return "未分类"
	}

	scope := "未分类"
	if s, ok := pathScopeMap[parts[1]]; ok {
		scope = s
	}

	// 从后续路径段中推断 function
	function := "其他"
	for i := 2; i < len(parts); i++ {
		seg := parts[i]
		if seg == "" || strings.HasPrefix(seg, ":") {
			continue
		}
		if fn, ok := pathFunctionMap[seg]; ok {
			function = fn
			break
		}
	}

	return scope + "-" + function
}

// tagNode 描述一个 tag 节点及其父节点、显示名。
type tagNode struct {
	parent      string
	displayName string
}

// appendUnique 向字符串切片追加元素，已存在则跳过。
func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

// versionFromPath 从接口路径提取版本前缀（v1/v2），默认为 v1。
func versionFromPath(path string) string {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) > 0 && (parts[0] == "v1" || parts[0] == "v2") {
		return parts[0]
	}
	// 兼容 /api/v1/... /api/v2/... 的写法
	if strings.HasPrefix(path, "api/") && len(parts) >= 2 {
		if parts[1] == "v1" || parts[1] == "v2" {
			return parts[1]
		}
	}
	return "v1"
}

// splitTag 把 tag 拆分为 scope 和 function 两部分（如 "User-资料" -> User, 资料）。
// 兼容旧格式中可能带 V2- 前缀的 tag，遇到时先剥掉版本前缀。
func splitTag(tag string) (scope, function string) {
	tag = strings.TrimPrefix(tag, "V1-")
	tag = strings.TrimPrefix(tag, "V2-")
	idx := strings.Index(tag, "-")
	if idx < 0 {
		return tag, tag
	}
	return tag[:idx], tag[idx+1:]
}

// injectNestedTagHierarchy 根据接口路径给 tag 加版本母目录，
// 并在 root.tags 中声明 x-parent/x-displayName，使 Scalar 侧边栏以 v1/v2 为顶层嵌套展开。
func injectNestedTagHierarchy(root map[string]any) {
	paths, ok := root["paths"].(map[string]any)
	if !ok || paths == nil {
		return
	}

	nodes := make(map[string]tagNode)     // tag 名 -> 节点信息
	groups := make(map[string][]string)   // versionScope -> 该版本作用域下的 function tag 列表
	children := make(map[string][]string) // parent tag -> 子 tag 名列表

	for pathKey, pathVal := range paths {
		version := versionFromPath(pathKey)

		ops, ok := pathVal.(map[string]any)
		if !ok {
			continue
		}
		for _, opVal := range ops {
			op, ok := opVal.(map[string]any)
			if !ok {
				continue
			}
			rawTags, _ := op["tags"].([]any)
			// 如果没有 tag，按路径自动推断，实现"自动挂"
			if len(rawTags) == 0 {
				rawTags = []any{inferTagFromPath(pathKey)}
			}
			newTags := make([]any, 0, len(rawTags))

			for _, t := range rawTags {
				tag, ok := t.(string)
				if !ok || tag == "" {
					continue
				}

				scope, function := splitTag(tag)
				versionScope := version + "-" + scope
				newTag := versionScope + "-" + function
				newTags = append(newTags, newTag)

				// 创建/更新节点：v1 -> v1-User -> v1-User-资料
				if _, exists := nodes[version]; !exists {
					nodes[version] = tagNode{parent: "", displayName: version}
				}
				if _, exists := nodes[versionScope]; !exists {
					nodes[versionScope] = tagNode{parent: version, displayName: scope}
					children[version] = appendUnique(children[version], versionScope)
				}
				nodes[newTag] = tagNode{parent: versionScope, displayName: function}
				children[versionScope] = appendUnique(children[versionScope], newTag)
				groups[versionScope] = appendUnique(groups[versionScope], newTag)
			}

			op["tags"] = newTags
		}
	}

	// 注入 tags 数组
	if len(nodes) > 0 {
		tags := make([]map[string]any, 0, len(nodes))
		for name, node := range nodes {
			tag := map[string]any{
				"name":          name,
				"x-displayName": node.displayName,
			}
			if node.parent != "" {
				// OpenAPI 3.2 原生 parent 字段（Scalar 1.64 支持）
				// 同时保留 x-parent 扩展以兼容旧版解析器
				tag["parent"] = node.parent
				tag["x-parent"] = node.parent
			}
			// Scalar 内部也会读取 children 属性构建树，显式声明更保险
			if childList := children[name]; len(childList) > 0 {
				sort.Strings(childList)
				tag["children"] = childList
			}
			tags = append(tags, tag)
		}
		// 稳定排序，避免顺序影响 diff
		sort.Slice(tags, func(i, j int) bool {
			return tags[i]["name"].(string) < tags[j]["name"].(string)
		})
		root["tags"] = tags
	}

	// 生成 Scalar x-tagGroups 兜底分组：用 "v1------Admin" 作为 group 名，
	// 在 Scalar 不支持原生 parent 嵌套时也能直观区分版本和作用域。
	if len(groups) > 0 {
		xTagGroups := make([]map[string]any, 0, len(groups))
		for versionScope, funcTags := range groups {
			sort.Strings(funcTags)
			// group 显示名：v1-Admin -> v1------Admin，视觉上分隔版本与 scope
			displayGroup := strings.ReplaceAll(versionScope, "-", "------")
			xTagGroups = append(xTagGroups, map[string]any{
				"name": displayGroup,
				"tags": funcTags,
			})
		}
		sort.Slice(xTagGroups, func(i, j int) bool {
			return xTagGroups[i]["name"].(string) < xTagGroups[j]["name"].(string)
		})
		root["x-tagGroups"] = xTagGroups
	}
}

// convertToOpenAPI31 将 Swagger 2.0 文档最小转换为 OpenAPI 3.1，
// 主要目的是让 Scalar 1.64+ 能识别 tags.parent 嵌套层级。
// 注意：这是运行时给文档展示用的最小转换，不完美支持所有 OpenAPI 3.1 特性，
// 但足以让 Scalar 正确渲染嵌套侧边栏和请求测试。
func convertToOpenAPI31(root map[string]any) {
	// 改版本号。tags.parent 是 OpenAPI 3.2 特性，Scalar 1.64+ 支持。
	delete(root, "swagger")
	root["openapi"] = "3.2.0"

	// servers
	host, _ := root["host"].(string)
	basePath, _ := root["basePath"].(string)
	schemes := []string{"http"}
	if raw, ok := root["schemes"].([]any); ok && len(raw) > 0 {
		schemes = make([]string, 0, len(raw))
		for _, s := range raw {
			if ss, ok := s.(string); ok {
				schemes = append(schemes, ss)
			}
		}
	}
	serverURL := "/" + basePath
	if host != "" {
		serverURL = schemes[0] + "://" + host + basePath
	}
	root["servers"] = []map[string]any{{"url": serverURL}}
	delete(root, "host")
	delete(root, "basePath")
	delete(root, "schemes")

	// components
	components := make(map[string]any)

	// definitions -> components/schemas
	if defs, ok := root["definitions"].(map[string]any); ok && len(defs) > 0 {
		components["schemas"] = defs
		delete(root, "definitions")
	}

	// securityDefinitions -> components/securitySchemes
	if secDefs, ok := root["securityDefinitions"].(map[string]any); ok && len(secDefs) > 0 {
		components["securitySchemes"] = secDefs
		delete(root, "securityDefinitions")
	}

	// responses -> components/responses
	if respDefs, ok := root["responses"].(map[string]any); ok && len(respDefs) > 0 {
		components["responses"] = respDefs
		delete(root, "responses")
	}

	// parameters -> components/parameters
	if paramDefs, ok := root["parameters"].(map[string]any); ok && len(paramDefs) > 0 {
		components["parameters"] = paramDefs
		delete(root, "parameters")
	}

	if len(components) > 0 {
		root["components"] = components
	}

	// 把 operation 里的 schema/produces/consumes 转换为 OpenAPI 3.1 的 content 结构
	paths, _ := root["paths"].(map[string]any)
	if paths == nil {
		return
	}
	for _, pathVal := range paths {
		ops, ok := pathVal.(map[string]any)
		if !ok {
			continue
		}
		for _, opVal := range ops {
			op, ok := opVal.(map[string]any)
			if !ok {
				continue
			}
			convertOperationToOpenAPI31(op)
		}
	}

	// definitions 已迁移到 components/schemas，需要把 $ref 前缀同步改写
	rewriteRefsInNode(root, "#/definitions/", "#/components/schemas/")
}

// rewriteRefsInNode 递归改写节点中所有 $ref 的前缀。
func rewriteRefsInNode(node any, oldPrefix, newPrefix string) {
	switch v := node.(type) {
	case map[string]any:
		for key, val := range v {
			if key == "$ref" {
				if s, ok := val.(string); ok && strings.HasPrefix(s, oldPrefix) {
					v[key] = newPrefix + strings.TrimPrefix(s, oldPrefix)
				}
				continue
			}
			rewriteRefsInNode(val, oldPrefix, newPrefix)
		}
	case []any:
		for _, item := range v {
			rewriteRefsInNode(item, oldPrefix, newPrefix)
		}
	}
}

// convertOperationToOpenAPI31 将单个 Swagger 2.0 operation 转换为 3.1 最小结构。
func convertOperationToOpenAPI31(op map[string]any) {
	// consumes / produces 处理请求/响应的 content type
	consumes := []string{"application/json"}
	if raw, ok := op["consumes"].([]any); ok && len(raw) > 0 {
		consumes = make([]string, 0, len(raw))
		for _, c := range raw {
			if s, ok := c.(string); ok {
				consumes = append(consumes, s)
			}
		}
	}
	produces := []string{"application/json"}
	if raw, ok := op["produces"].([]any); ok && len(raw) > 0 {
		produces = make([]string, 0, len(raw))
		for _, p := range raw {
			if s, ok := p.(string); ok {
				produces = append(produces, s)
			}
		}
	}
	delete(op, "consumes")
	delete(op, "produces")

	// 转换 parameters：body -> requestBody；其他保留
	if rawParams, ok := op["parameters"].([]any); ok && len(rawParams) > 0 {
		var queryParams []any
		var bodySchema map[string]any
		var bodyDesc string
		for _, p := range rawParams {
			param, ok := p.(map[string]any)
			if !ok {
				continue
			}
			in, _ := param["in"].(string)
			if in == "body" {
				bodySchema, _ = param["schema"].(map[string]any)
				bodyDesc, _ = param["description"].(string)
			} else {
				queryParams = append(queryParams, p)
			}
		}
		if bodySchema != nil {
			content := make(map[string]any)
			for _, ct := range consumes {
				content[ct] = map[string]any{"schema": bodySchema, "description": bodyDesc}
			}
			op["requestBody"] = map[string]any{
				"content":     content,
				"description": bodyDesc,
			}
		}
		if len(queryParams) > 0 {
			op["parameters"] = queryParams
		} else {
			delete(op, "parameters")
		}
	}

	// 转换 responses
	if rawResponses, ok := op["responses"].(map[string]any); ok && len(rawResponses) > 0 {
		responses := make(map[string]any)
		for code, rawResp := range rawResponses {
			resp, ok := rawResp.(map[string]any)
			if !ok {
				continue
			}
			desc, _ := resp["description"].(string)
			newResp := map[string]any{"description": desc}
			if schema, ok := resp["schema"].(map[string]any); ok {
				content := make(map[string]any)
				for _, ct := range produces {
					content[ct] = map[string]any{"schema": schema}
				}
				newResp["content"] = content
			}
			// 保留 headers / examples 等
			for k, v := range resp {
				if k != "schema" && k != "description" {
					newResp[k] = v
				}
			}
			responses[code] = newResp
		}
		op["responses"] = responses
	}
}

// modelNameFromOld 从 modelNameCN 查找旧名对应的中文名。
func modelNameFromOld(old string) (string, bool) {
	for en, cn := range modelNameCN {
		if old == en || strings.HasSuffix(old, "."+en) {
			return cn, true
		}
	}
	return "", false
}

// modelNameCN 模型名英文 -> 中文翻译映射。
// 新增模型时直接在这里加一行；不要硬编码到业务代码中。
var modelNameCN = map[string]string{
	"AdminCompleteOrderRequest":    "管理端补单请求",
	"AdminResolveExceptionRequest": "管理端订单异常处理请求",
	"AdminUserDetailResponse":      "管理端用户详情响应",
	"AdminUserRealnameSummary":     "管理端用户实名认证摘要",
	"BatchUpdateSettingsRequest":   "批量更新系统设置请求",
	"CreateSettingRequest":         "创建系统设置请求",
	"EmailPreviewRequest":          "邮件内容预览请求",
	"EmailSendTestRequest":         "邮件发送测试请求",
	"EmailTemplateUpdateRequest":   "邮件模板更新请求",
	"ReviewRealnameRequest":        "实名认证审核请求",
	"SMSPreviewRequest":            "短信内容预览请求",
	"SMSSendTestRequest":           "短信发送测试请求",
	"SMSTemplateUpdateRequest":     "短信模板更新请求",
	"SettingsListResponse":         "系统设置列表响应",
	"UpdateSettingMetaRequest":     "更新系统设置元信息请求",
	"UpdateSettingRequest":         "更新系统设置请求",
	"AppConfigResponse":            "应用配置响应",
	"DialCountriesResponse":        "国际区号列表响应",
	"LoginRequest":                 "登录请求",
	"PhoneCountryResponse":         "手机号国家区号响应",
	"RefreshTokenRequest":          "刷新令牌请求",
	"RegisterRequest":              "注册请求",
	"ResetEmailRequest":            "重置邮箱请求",
	"ResetPasswordConfirmRequest":  "重置密码确认请求",
	"SendCodeRequest":              "发送验证码请求",
	"ChangePasswordRequest":        "修改密码请求",
	"CreateOrderRequest":           "创建订单请求",
	"DeactivateAccountRequest":     "注销账号请求",
	"ProfileRealnameSummary":       "用户实名认证摘要",
	"ProfileResponse":              "用户资料响应",
	"SendEmailCodeRequest":         "发送邮箱验证码请求",
	"SendPhoneCodeRequest":         "发送手机验证码请求",
	"SubmitRealnameRequest":        "提交实名认证请求",
	"UpdateProfileRequest":         "更新个人资料请求",
	"UpdateSettingsRequest":        "更新用户设置请求",
	"VerifyEmailChangeRequest":     "验证邮箱变更请求",
	"VerifyPhoneChangeRequest":     "验证手机号变更请求",
	"AdminUserListItem":            "管理端用户列表项",
	"UserCreateRequest":            "创建用户请求",
	"UserUpdateRequest":            "更新用户请求",
	"SettingDTO":                   "系统设置 DTO",
	"SettingsGroup":                "设置分组",
	"DialCountry":                  "国际区号",
	"Response":                     "通用响应",
}

// Scalar / Redoc 支持 x-tagGroups 实现分组侧边栏。
func injectTagGroups(root map[string]any) {
	root["x-tagGroups"] = []map[string]any{
		{"name": "📢 Public 公共接口", "tags": []string{"Public-认证", "Public-配置", "Public-区号", "Public-回调", "Public-会话"}},
		{"name": "👤 User 用户接口", "tags": []string{"User-资料", "User-支付", "User-实名", "User-提现", "User-公告", "User-会话", "User-设置"}},
		{"name": "🔧 Admin 管理员接口", "tags": []string{
			"Admin-公告", "Admin-邮件日志", "Admin-邮件模板", "Admin-短信日志", "Admin-短信模板",
			"Admin-用户", "Admin-用户等级", "Admin-用户积分", "Admin-支付", "Admin-实名",
			"Admin-设置", "Admin-操作日志", "Admin-API日志", "Admin-仪表盘", "Admin-自动任务",
			"Admin-在线用户", "Admin-调试", "Admin-终端", "Admin-待办", "Admin-数据库",
		}},
		{"name": "⚙ System 系统管理", "tags": []string{"System-管理"}},
		{"name": "🧪 V2 测试", "tags": []string{"V2-测试"}},
	}
}
