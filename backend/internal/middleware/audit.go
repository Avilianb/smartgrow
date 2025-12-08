package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLogger 安全审计日志中间件
type AuditLogger struct {
	logSensitiveOps bool
}

// NewAuditLogger 创建审计日志记录器
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logSensitiveOps: true,
	}
}

// AuditMiddleware 审计日志中间件
func (al *AuditLogger) AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		ip := c.ClientIP()

		// 处理请求
		c.Next()

		// 记录敏感操作
		if al.shouldLog(path, method) {
			latency := time.Since(start)
			statusCode := c.Writer.Status()

			// 获取用户信息
			username := "anonymous"
			role := "unknown"
			if val, exists := c.Get("username"); exists {
				username = val.(string)
			}
			if val, exists := c.Get("role"); exists {
				role = val.(string)
			}

			// 记录审计日志
			log.Printf("[AUDIT] %s %s | Status: %d | User: %s | Role: %s | IP: %s | Latency: %v",
				method, path, statusCode, username, role, ip, latency)

			// 记录失败的认证尝试
			if statusCode == 401 || statusCode == 403 {
				log.Printf("[SECURITY] Authentication/Authorization failure | Path: %s | IP: %s | User: %s",
					path, ip, username)
			}
		}
	}
}

// shouldLog 判断是否需要记录审计日志
func (al *AuditLogger) shouldLog(path, method string) bool {
	// 敏感操作列表
	sensitivePaths := []string{
		"/api/auth/login",
		"/api/auth/admin/login",
		"/api/admin/users",
	}

	// 所有 POST, PUT, DELETE 操作都记录
	if method == "POST" || method == "PUT" || method == "DELETE" {
		return true
	}

	// 敏感路径的 GET 操作也记录
	for _, sp := range sensitivePaths {
		if path == sp {
			return true
		}
	}

	return false
}
