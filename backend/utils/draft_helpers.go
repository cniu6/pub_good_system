package utils

import (
	"html"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 以下为草稿业务（backend/controllers、services 等）共用的轻量辅助函数。
// 密码哈希请使用 password.go 的 HashPassword / CheckPasswordHash，勿再另起一套。

// ParseUint 将字符串解析为 uint（解析失败返回 0）
func ParseUint(s string) uint {
	val, err := strconv.ParseUint(strings.TrimSpace(s), 10, 32)
	if err != nil {
		return 0
	}
	return uint(val)
}

// ParseInt 将字符串解析为 int（解析失败返回 0）
func ParseInt(s string) int {
	val, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return val
}

// GenerateAPIKey 生成随机 API 密钥（UUID 字符串）
func GenerateAPIKey() string {
	return uuid.New().String()
}

// GenerateOrderNumber 生成订单编号：FST + UUID 前 8 位
func GenerateOrderNumber() string {
	return "FST" + uuid.New().String()[:8]
}

// CleanXSS 轻量清理输入（HTML 转义 + 去控制字符），避免草稿依赖 bluemonday
// 说明：现网复杂过滤见 validate.go 的 Clean_XSS；草稿表单字段用本函数即可
func CleanXSS(input string) string {
	if input == "" {
		return ""
	}
	input = strings.TrimSpace(input)
	input = html.EscapeString(input)
	// 去掉常见控制字符（保留换行/制表以外的不可见字符）
	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		if r == '\n' || r == '\r' || r == '\t' || r >= 32 {
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

// CleanXSSFields 批量清理字符串指针字段中的 XSS 风险内容
func CleanXSSFields(fields ...*string) {
	for _, field := range fields {
		if field != nil && *field != "" {
			*field = CleanXSS(*field)
		}
	}
}

// 草稿中间件写入上下文时使用的键名（与现网 userID/role 对齐，并兼容草稿侧 user_id）
const (
	// ContextKeyUser 完整用户对象（草稿 GORM User）
	ContextKeyUser = "user"
	// ContextKeyUserID 用户 ID（uint）
	ContextKeyUserID = "user_id"
	// ContextKeyUsername 用户名
	ContextKeyUsername = "username"
	// ContextKeyRole 角色
	ContextKeyRole = "role"
	// ContextKeyStatus 状态
	ContextKeyStatus = "status"
)

// GetUserID 从 Gin 上下文读取用户 ID
// 优先读现网中间件写入的 userID(uint64)，再读草稿键 user_id(uint)
func GetUserID(c *gin.Context) uint {
	if c == nil {
		return 0
	}
	// 现网 AuthMiddleware: c.Set("userID", claims.UserID) 类型为 uint64
	if v, ok := c.Get("userID"); ok {
		switch id := v.(type) {
		case uint64:
			return uint(id)
		case uint:
			return id
		case int:
			if id > 0 {
				return uint(id)
			}
		case int64:
			if id > 0 {
				return uint(id)
			}
		}
	}
	if v, ok := c.Get(ContextKeyUserID); ok {
		switch id := v.(type) {
		case uint:
			return id
		case uint64:
			return uint(id)
		case int:
			if id > 0 {
				return uint(id)
			}
		}
	}
	return 0
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get(ContextKeyUsername); ok {
		if name, ok := v.(string); ok {
			return name
		}
	}
	if v, ok := c.Get("username"); ok {
		if name, ok := v.(string); ok {
			return name
		}
	}
	return ""
}

// GetUserRole 从上下文获取角色字符串
func GetUserRole(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get(ContextKeyRole); ok {
		switch r := v.(type) {
		case string:
			return r
		}
	}
	if v, ok := c.Get("role"); ok {
		if r, ok := v.(string); ok {
			return r
		}
	}
	return ""
}

// IsAdmin 当前请求是否管理员角色
func IsAdmin(c *gin.Context) bool {
	return GetUserRole(c) == "admin"
}

// IsAuthenticated 是否已登录（userID > 0）
func IsAuthenticated(c *gin.Context) bool {
	return GetUserID(c) > 0
}

// SuccessWithMessage 成功响应并自定义 message（草稿控制器使用）
func SuccessWithMessage(c *gin.Context, message string, data any) {
	SuccessMsg(c, message, data)
}

// PageResponse 分页列表统一响应
// data 为当前页数据，total 为总条数，page/pageSize 为分页参数
func PageResponse(c *gin.Context, data any, total int64, page, pageSize int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	Success(c, gin.H{
		"list":      data,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
