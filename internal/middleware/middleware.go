package middleware

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

// Claims represents JWT token claims
type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuth returns a gin middleware for JWT authentication
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "NO_TOKEN",
					"message": "No authorization token provided",
				},
			})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix
		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		// Validate token
		if !validateJWT(tokenString) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Token is invalid or expired",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// validateJWT validates a JWT token
func validateJWT(tokenString string) bool {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key"
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return false
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check if token is expired
		return claims.ExpiresAt.After(time.Now())
	}

	return false
}
