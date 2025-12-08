package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // 每分钟允许的请求数
	cleanup  time.Duration // 清理间隔
}

type visitor struct {
	lastSeen time.Time
	count    int
	resetAt  time.Time
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     requestsPerMinute,
		cleanup:  5 * time.Minute,
	}

	// 启动清理goroutine
	go rl.cleanupVisitors()

	return rl
}

// RateLimitMiddleware 速率限制中间件
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端标识（IP地址）
		ip := c.ClientIP()

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		now := time.Now()

		if !exists {
			// 新访客
			rl.visitors[ip] = &visitor{
				lastSeen: now,
				count:    1,
				resetAt:  now.Add(time.Minute),
			}
			rl.mu.Unlock()
			c.Next()
			return
		}

		// 检查是否需要重置计数
		if now.After(v.resetAt) {
			v.count = 1
			v.resetAt = now.Add(time.Minute)
			v.lastSeen = now
			rl.mu.Unlock()
			c.Next()
			return
		}

		// 检查是否超过限制
		if v.count >= rl.rate {
			rl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "请求过于频繁，请稍后再试",
				"retry_after": int(time.Until(v.resetAt).Seconds()),
			})
			c.Abort()
			return
		}

		// 增加计数
		v.count++
		v.lastSeen = now
		rl.mu.Unlock()

		c.Next()
	}
}

// cleanupVisitors 定期清理过期的访客记录
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			// 删除5分钟未访问的记录
			if now.Sub(v.lastSeen) > rl.cleanup {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// LoginRateLimiter 登录专用速率限制（更严格）
type LoginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
}

type loginAttempt struct {
	count     int
	firstAttempt time.Time
	lockedUntil  time.Time
}

// NewLoginRateLimiter 创建登录速率限制器
func NewLoginRateLimiter() *LoginRateLimiter {
	lrl := &LoginRateLimiter{
		attempts: make(map[string]*loginAttempt),
	}

	// 启动清理goroutine
	go lrl.cleanupAttempts()

	return lrl
}

// LoginRateLimitMiddleware 登录速率限制中间件
func (lrl *LoginRateLimiter) LoginRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		lrl.mu.Lock()
		attempt, exists := lrl.attempts[ip]
		now := time.Now()

		if !exists {
			// 首次尝试
			lrl.attempts[ip] = &loginAttempt{
				count:        1,
				firstAttempt: now,
				lockedUntil:  time.Time{},
			}
			lrl.mu.Unlock()
			c.Next()
			return
		}

		// 检查是否被锁定
		if !attempt.lockedUntil.IsZero() && now.Before(attempt.lockedUntil) {
			lrl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "登录尝试次数过多，账户已被临时锁定",
				"retry_after": int(time.Until(attempt.lockedUntil).Seconds()),
			})
			c.Abort()
			return
		}

		// 重置时间窗口（5分钟）
		if now.Sub(attempt.firstAttempt) > 5*time.Minute {
			attempt.count = 1
			attempt.firstAttempt = now
			attempt.lockedUntil = time.Time{}
			lrl.mu.Unlock()
			c.Next()
			return
		}

		// 检查尝试次数
		if attempt.count >= 5 {
			// 5次失败后锁定15分钟
			attempt.lockedUntil = now.Add(15 * time.Minute)
			lrl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "登录失败次数过多，账户已被锁定15分钟",
				"retry_after": 900,
			})
			c.Abort()
			return
		}

		// 增加尝试次数
		attempt.count++
		lrl.mu.Unlock()

		c.Next()
	}
}

// ResetAttempts 重置登录尝试（登录成功后调用）
func (lrl *LoginRateLimiter) ResetAttempts(ip string) {
	lrl.mu.Lock()
	delete(lrl.attempts, ip)
	lrl.mu.Unlock()
}

// cleanupAttempts 定期清理过期的尝试记录
func (lrl *LoginRateLimiter) cleanupAttempts() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		lrl.mu.Lock()
		now := time.Now()
		for ip, attempt := range lrl.attempts {
			// 删除30分钟前的记录
			if now.Sub(attempt.firstAttempt) > 30*time.Minute {
				delete(lrl.attempts, ip)
			}
		}
		lrl.mu.Unlock()
	}
}
