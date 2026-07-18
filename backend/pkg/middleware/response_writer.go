package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

// responseWriter 自定义响应写入器，用于捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	maxCapture int
	written    int64
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.written += int64(len(b))
	if w.body != nil && w.maxCapture > 0 {
		remaining := w.maxCapture - w.body.Len()
		if remaining > 0 {
			if len(b) > remaining {
				_, _ = w.body.Write(b[:remaining])
			} else {
				_, _ = w.body.Write(b)
			}
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func newResponseWriter(writer gin.ResponseWriter, maxCapture int) *responseWriter {
	return &responseWriter{
		ResponseWriter: writer,
		body:           bytes.NewBufferString(""),
		maxCapture:     maxCapture,
	}
}
