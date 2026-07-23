package middleware

import (
	"strconv"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

// 简易 HTTP 计数（无 prometheus 依赖时供 /metrics 使用）
var (
	httpRequestsTotal   atomic.Uint64
	httpRequests2xx     atomic.Uint64
	httpRequests4xx     atomic.Uint64
	httpRequests5xx     atomic.Uint64
)

// MetricsMiddleware 统计请求次数（按最终状态码分桶）。
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		httpRequestsTotal.Add(1)
		code := c.Writer.Status()
		switch {
		case code >= 200 && code < 300:
			httpRequests2xx.Add(1)
		case code >= 400 && code < 500:
			httpRequests4xx.Add(1)
		case code >= 500:
			httpRequests5xx.Add(1)
		}
	}
}

// HTTPMetricsSnapshot 返回当前计数快照。
func HTTPMetricsSnapshot() (total, c2xx, c4xx, c5xx uint64) {
	return httpRequestsTotal.Load(), httpRequests2xx.Load(), httpRequests4xx.Load(), httpRequests5xx.Load()
}

// FormatPrometheusHTTPCounters 输出 prometheus 文本格式的计数行。
func FormatPrometheusHTTPCounters() string {
	total, c2xx, c4xx, c5xx := HTTPMetricsSnapshot()
	return "# HELP fst_http_requests_total Total HTTP requests\n" +
		"# TYPE fst_http_requests_total counter\n" +
		"fst_http_requests_total " + strconv.FormatUint(total, 10) + "\n" +
		"# HELP fst_http_requests_by_class HTTP requests by status class\n" +
		"# TYPE fst_http_requests_by_class counter\n" +
		`fst_http_requests_by_class{class="2xx"} ` + strconv.FormatUint(c2xx, 10) + "\n" +
		`fst_http_requests_by_class{class="4xx"} ` + strconv.FormatUint(c4xx, 10) + "\n" +
		`fst_http_requests_by_class{class="5xx"} ` + strconv.FormatUint(c5xx, 10) + "\n"
}
