package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter IP 限流器管理器
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  sync.RWMutex
	r   rate.Limit // 每秒产生多少令牌 (速率)
	b   int        // 桶的大小 (突发容量)
}

// NewIPRateLimiter 创建限流器
// r: 每秒允许多少个请求
// b: 允许瞬间突发多少个请求
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		r:   r,
		b:   b,
	}

	// 启动一个协程定期清理过期的 IP，防止内存泄漏
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			i.mu.Lock()
			// 简单粗暴：每10分钟清空一次记录，生产环境可以用 LRU
			i.ips = make(map[string]*rate.Limiter)
			i.mu.Unlock()
		}
	}()

	return i
}

// GetLimiter 获取指定 IP 的限流器
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// RateLimitMiddleware Gin 中间件
func RateLimitMiddleware(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.GetLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "操作太频繁，请喝口水歇会儿 (IP限制)",
			})
			return
		}
		c.Next()
	}
}