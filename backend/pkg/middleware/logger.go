package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerConfig 日志中间件配置
type LoggerConfig struct {
	// SkipPaths 跳过日志记录的路径
	SkipPaths []string
	// SkipPathPrefixes 跳过日志记录的路径前缀
	SkipPathPrefixes []string
	// LogRequestBody 是否记录请求体
	LogRequestBody bool
	// LogResponseBody 是否记录响应体
	LogResponseBody bool
	// MaxBodyLength 最大记录体长度
	MaxBodyLength int
}

// DefaultLoggerConfig 默认配置
var DefaultLoggerConfig = LoggerConfig{
	SkipPaths: []string{
		"/health",
		"/favicon.ico",
	},
	SkipPathPrefixes: []string{
		"/scalar",
		"/static",
	},
	LogRequestBody:  false,
	LogResponseBody: false,
	MaxBodyLength:   500,
}

// LoggerMiddleware 请求日志中间件
// 使用默认配置
func LoggerMiddleware() gin.HandlerFunc {
	return LoggerMiddlewareWithConfig(DefaultLoggerConfig)
}

// LoggerMiddlewareWithConfig 使用自定义配置的请求日志中间件
func LoggerMiddlewareWithConfig(config LoggerConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过该路径
		path := c.Request.URL.Path
		if shouldSkip(path, config) {
			c.Next()
			return
		}

		// 开始时间
		start_time := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		end_time := time.Now()
		latency := end_time.Sub(start_time)

		// 获取状态码
		status_code := c.Writer.Status()

		// 获取客户端IP
		client_ip := c.ClientIP()

		// 获取请求方法
		method := c.Request.Method

		// 构建日志（附带 request_id / user_id，便于与访问日志、鉴权审计关联）
		requestID := GetRequestID(c)
		log_line := buildLogLine(start_time, status_code, latency, client_ip, method, path, requestID)
		if uid, exists := c.Get("userID"); exists {
			if parsed, ok := uid.(uint64); ok && parsed > 0 {
				log_line += fmt.Sprintf(" | user_id=%d", parsed)
			}
		}
		if len(c.Errors) > 0 {
			log_line += " | gin_errors=" + c.Errors.String()
		}

		// 根据状态码选择日志级别
		if status_code >= 500 {
			gin.DefaultErrorWriter.Write([]byte("[ERROR] " + log_line + "\n"))
		} else if status_code >= 400 {
			gin.DefaultErrorWriter.Write([]byte("[WARN] " + log_line + "\n"))
		} else {
			gin.DefaultWriter.Write([]byte("[INFO] " + log_line + "\n"))
		}
	}
}

// shouldSkip 检查是否应该跳过该路径
func shouldSkip(path string, config LoggerConfig) bool {
	// 检查完整路径
	for _, skip_path := range config.SkipPaths {
		if path == skip_path {
			return true
		}
	}

	// 检查路径前缀
	for _, prefix := range config.SkipPathPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// buildLogLine 构建日志行（含可选 request_id）
func buildLogLine(start_time time.Time, status_code int, latency time.Duration, client_ip, method, path, requestID string) string {
	// 格式化时间
	time_str := start_time.Format("2006/01/02 - 15:04:05")

	// 格式化延迟
	latency_str := formatLatency(latency)

	if requestID == "" {
		return fmt.Sprintf("%s | %3d | %13s | %15s | %-7s %s",
			time_str, status_code, latency_str, client_ip, method, path)
	}
	return fmt.Sprintf("%s | %3d | %13s | %15s | %-7s %s | request_id=%s",
		time_str, status_code, latency_str, client_ip, method, path, requestID)
}

// formatLatency 格式化延迟时间
func formatLatency(latency time.Duration) string {
	if latency < time.Millisecond {
		return fmt.Sprintf("%dns", latency.Nanoseconds())
	} else if latency < time.Second {
		return fmt.Sprintf("%.2fms", float64(latency.Nanoseconds())/1000000)
	} else {
		return fmt.Sprintf("%.2fs", latency.Seconds())
	}
}

// RequestLogger 带有更多信息的请求日志中间件
// 包含用户信息、请求参数等
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if shouldSkip(path, DefaultLoggerConfig) {
			c.Next()
			return
		}

		start_time := time.Now()
		c.Next()
		latency := time.Since(start_time)

		// 获取用户信息（类型断言失败时视为匿名，避免 panic）
		var user_id uint64
		var role string
		if uid, exists := c.Get("userID"); exists {
			if parsed, ok := uid.(uint64); ok {
				user_id = parsed
			}
		}
		if r, exists := c.Get("role"); exists {
			if parsed, ok := r.(string); ok {
				role = parsed
			}
		}

		// 构建详细日志
		time_str := start_time.Format("2006/01/02 - 15:04:05")
		status_code := c.Writer.Status()
		client_ip := c.ClientIP()
		method := c.Request.Method
		requestID := GetRequestID(c)

		var user_info string
		if user_id > 0 {
			user_info = fmt.Sprintf("[uid:%d|%s]", user_id, role)
		} else {
			user_info = "[anonymous]"
		}

		log_line := fmt.Sprintf("%s | %3d | %13s | %15s | %-7s %s %s",
			time_str,
			status_code,
			formatLatency(latency),
			client_ip,
			method,
			path,
			user_info,
		)
		if requestID != "" {
			log_line += " | request_id=" + requestID
		}

		if status_code >= 500 {
			gin.DefaultErrorWriter.Write([]byte("[ERROR] " + log_line + "\n"))
		} else if status_code >= 400 {
			gin.DefaultErrorWriter.Write([]byte("[WARN] " + log_line + "\n"))
		} else {
			gin.DefaultWriter.Write([]byte("[INFO] " + log_line + "\n"))
		}
	}
}

