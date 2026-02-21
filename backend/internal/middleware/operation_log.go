package middleware

import (
	"bytes"
	"fst/backend/app/models"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// responseWriter 自定义响应写入器，用于捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// OperationLogMiddleware 操作日志中间件
// 记录用户的操作日志
func OperationLogMiddleware(module string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start_time := time.Now()

		// 获取请求体
		var request_body string
		if c.Request.Body != nil {
			body_bytes, _ := io.ReadAll(c.Request.Body)
			request_body = string(body_bytes)
			// 重新设置请求体，以便后续处理
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body_bytes))
		}

		// 包装响应写入器
		blw := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(start_time).Milliseconds()

		// 获取用户信息
		var user_id uint64
		var username string
		if uid, exists := c.Get("userID"); exists {
			user_id = uid.(uint64)
		}
		if uname, exists := c.Get("username"); exists {
			username = uname.(string)
		}

		// 获取响应内容 (限制长度)
		response_body := blw.body.String()
		if len(response_body) > 2000 {
			response_body = response_body[:2000] + "...(truncated)"
		}

		// 限制请求体长度
		if len(request_body) > 2000 {
			request_body = request_body[:2000] + "...(truncated)"
		}

		// 创建日志记录
		log := &models.OperationLog{
			UserID:       user_id,
			Username:     username,
			Module:       module,
			Action:       action,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			RequestBody:  &request_body,
			ResponseBody: &response_body,
			StatusCode:   c.Writer.Status(),
			Duration:     int(duration),
		}

		// 异步保存日志
		go func() {
			models.CreateOperationLog(log)
		}()
	}
}

// SimpleLogMiddleware 简单日志中间件
// 只记录基本信息，不记录请求/响应体
func SimpleLogMiddleware(module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start_time := time.Now()

		c.Next()

		duration := time.Since(start_time).Milliseconds()

		var user_id uint64
		var username string
		if uid, exists := c.Get("userID"); exists {
			user_id = uid.(uint64)
		}
		if uname, exists := c.Get("username"); exists {
			username = uname.(string)
		}

		// 根据请求方法确定操作类型
		action := getActionByMethod(c.Request.Method)

		log := &models.OperationLog{
			UserID:     user_id,
			Username:   username,
			Module:     module,
			Action:     action,
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			StatusCode: c.Writer.Status(),
			Duration:   int(duration),
		}

		go func() {
			models.CreateOperationLog(log)
		}()
	}
}

func getActionByMethod(method string) string {
	switch method {
	case "GET":
		return "查询"
	case "POST":
		return "创建"
	case "PUT":
		return "更新"
	case "DELETE":
		return "删除"
	default:
		return "操作"
	}
}
