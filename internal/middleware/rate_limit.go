package middleware

import (
	"linked-clone/pkg/logger"
	"linked-clone/pkg/response"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	visitors map[string]*visitor
	mutex    sync.RWMutex
	rate     time.Duration
	burst    int
}

type visitor struct {
	tokens   int
	lastSeen time.Time
	mutex    sync.Mutex
}

func NewRateLimiter(rate time.Duration, burst int) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
	}

	go rl.cleanupVisitors()

	return rl
}

func (rl *rateLimiter) getVisitor(ip string) *visitor {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			tokens:   rl.burst,
			lastSeen: time.Now(),
		}
		rl.visitors[ip] = v
	}

	return v
}

func (rl *rateLimiter) allow(ip string) bool {
	v := rl.getVisitor(ip)
	v.mutex.Lock()
	defer v.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(v.lastSeen)

	tokensToAdd := int(elapsed / rl.rate)
	if tokensToAdd > 0 {
		v.tokens += tokensToAdd
		if v.tokens > rl.burst {
			v.tokens = rl.burst
		}
		v.lastSeen = now
	}

	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

func (rl *rateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > time.Hour {
				delete(rl.visitors, ip)
			}
		}
		rl.mutex.Unlock()
	}
}

func RateLimitMiddleware(rate time.Duration, burst int, logger logger.Logger) gin.HandlerFunc {

	if os.Getenv("APP_ENV") == "test" {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	limiter := NewRateLimiter(rate, burst)

	return gin.HandlerFunc(func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.allow(ip) {
			logger.Warn("Rate limit exceeded", map[string]interface{}{
				"ip":         ip,
				"path":       c.Request.URL.Path,
				"user_agent": c.Request.UserAgent(),
				"method":     c.Request.Method,
			})
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded", "Too many requests")
			c.Abort()
			return
		}

		c.Next()
	})
}
