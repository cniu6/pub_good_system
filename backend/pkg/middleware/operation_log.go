package middleware

import (
	"fst/backend/app/models"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// SimpleLogMiddleware 简单操作日志中间件（不记录请求/响应体）
func SimpleLogMiddleware(module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start_time := time.Now()

		c.Next()

		duration := time.Since(start_time).Milliseconds()

		var user_id uint64
		var username string
		if uid, exists := c.Get("userID"); exists {
			if parsedUID, ok := uid.(uint64); ok {
				user_id = parsedUID
			}
		}
		if uname, exists := c.Get("username"); exists {
			if parsedUsername, ok := uname.(string); ok {
				username = parsedUsername
			}
		}

		record := &models.OperationLog{
			UserID:     user_id,
			Username:   username,
			Module:     module,
			Action:     getActionByMethod(c.Request.Method),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			StatusCode: c.Writer.Status(),
			Duration:   int(duration),
		}

		go func(entry *models.OperationLog) {
			if err := models.CreateOperationLog(entry); err != nil {
				log.Printf("[OperationLog] 保存失败: %v", err)
			}
		}(record)
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
