package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LogEntry 日志条目
type LogEntry struct {
	Timestamp   string `json:"timestamp"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Status      int    `json:"status"`
	Latency     string `json:"latency"`
	LatencyMs   int64  `json:"latency_ms"`
	IP          string `json:"ip"`
	UserAgent   string `json:"user_agent"`
	BodySize    int    `json:"body_size"`
	Error       string `json:"error,omitempty"`
}

// ResponseWriter 包装 gin.ResponseWriter 以获取状态码
type ResponseWriter struct {
	gin.ResponseWriter
	statusCode int
}

func (w *ResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *ResponseWriter) Status() int {
	return w.statusCode
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 包装 ResponseWriter
		wrapper := &ResponseWriter{
			ResponseWriter: c.Writer,
			statusCode:     200,
		}
		c.Writer = wrapper

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		// 获取错误信息
		var errors []string
		for _, e := range c.Errors {
			errors = append(errors, e.Error())
		}

		// 构建日志条目
		entry := LogEntry{
			Timestamp: start.Format(time.RFC3339),
			Method:    c.Request.Method,
			Path:      path,
			Status:    wrapper.Status(),
			Latency:   latency.String(),
			LatencyMs: latency.Milliseconds(),
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			BodySize:  wrapper.Size(),
		}

		if len(errors) > 0 {
			entry.Error = errors[0]
		}

		// 日志输出
		logEntry := map[string]interface{}{
			"type":     "access",
			"timestamp": entry.Timestamp,
			"method":   entry.Method,
			"path":     entry.Path,
			"query":    query,
			"status":   entry.Status,
			"latency":  entry.Latency,
			"ip":       entry.IP,
			"user_agent": entry.UserAgent,
			"body_size": entry.BodySize,
		}

		if entry.Error != "" {
			logEntry["error"] = entry.Error
		}

		// JSON 格式日志
		logJSON, _ := json.Marshal(logEntry)
		log.Println(string(logJSON))

		// 根据状态码输出颜色
		statusStr := statusCodeColor(wrapper.Status())
		log.Printf("[%s] %s %s %s %s",
			statusStr,
			c.Request.Method,
			path,
			latency,
			c.ClientIP(),
		)
	}
}

// statusCodeColor 返回状态码的颜色代码
func statusCodeColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "\033[32m" + string(rune(code)) + "\033[0m"
	case code >= 300 && code < 400:
		return "\033[33m" + string(rune(code)) + "\033[0m"
	case code >= 400 && code < 500:
		return "\033[31m" + string(rune(code)) + "\033[0m"
	default:
		return "\033[35m" + string(rune(code)) + "\033[0m"
	}
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)

				c.AbortWithStatusJSON(500, gin.H{
					"error": "Internal server error",
				})
			}
		}()
		c.Next()
	}
}

// RequestIDMiddleware 请求 ID 中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID 生成请求 ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// SecurityHeadersMiddleware 安全头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware 速率限制中间件 (简化版)
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	// 这里应该使用 Redis 或内存存储来跟踪请求
	// 简化实现，仅供参考
	return func(c *gin.Context) {
		// TODO: 实现真正的速率限制
		c.Next()
	}
}

// Helper function for random string
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
