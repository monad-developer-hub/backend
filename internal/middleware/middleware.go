package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a gin middleware for logging
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return ""
	})
}

// ErrorHandler returns a gin middleware for handling errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":      "INTERNAL_SERVER_ERROR",
					"message":   err.Error(),
					"timestamp": time.Now(),
				},
			})
		}
	}
}

// Simple in-memory rate limiter
type rateLimiter struct {
	mu      sync.RWMutex
	clients map[string][]time.Time
	limit   int
	window  time.Duration
}

func newRateLimiter(requestsPerMinute int) *rateLimiter {
	return &rateLimiter{
		clients: make(map[string][]time.Time),
		limit:   requestsPerMinute,
		window:  time.Minute,
	}
}

func (rl *rateLimiter) allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Clean old entries
	if requests, exists := rl.clients[clientID]; exists {
		validRequests := make([]time.Time, 0)
		for _, requestTime := range requests {
			if now.Sub(requestTime) < rl.window {
				validRequests = append(validRequests, requestTime)
			}
		}
		rl.clients[clientID] = validRequests
	}

	// Check if under limit
	if len(rl.clients[clientID]) >= rl.limit {
		return false
	}

	// Add current request
	rl.clients[clientID] = append(rl.clients[clientID], now)
	return true
}

// RateLimit returns a gin middleware for rate limiting
func RateLimit(requestsPerMinute int) gin.HandlerFunc {
	limiter := newRateLimiter(requestsPerMinute)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": "Too many requests. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
