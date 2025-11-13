package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter represents a simple in-memory rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// RateLimit creates a rate limiting middleware
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use client IP as the key for rate limiting
		key := c.ClientIP()

		rl.mutex.Lock()
		defer rl.mutex.Unlock()

		now := time.Now()

		// Get existing requests for this key
		requests, exists := rl.requests[key]
		if !exists {
			requests = []time.Time{}
		}

		// Remove requests outside the time window
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}

		// Check if limit is exceeded
		if len(validRequests) >= rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		// Add current request
		validRequests = append(validRequests, now)
		rl.requests[key] = validRequests

		// Clean up old entries periodically (simple cleanup)
		if len(rl.requests) > 1000 {
			rl.cleanup(now)
		}

		c.Next()
	}
}

// cleanup removes old entries from the rate limiter
func (rl *RateLimiter) cleanup(now time.Time) {
	for key, requests := range rl.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if now.Sub(reqTime) < rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = validRequests
		}
	}
}

// PATRateLimit creates a rate limiter specifically for PAT endpoints
// Allows 10 requests per minute for PAT creation and management
func PATRateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(10, time.Minute)
	return limiter.RateLimit()
}

// PATAuthRateLimit creates a rate limiter for PAT authentication attempts
// Allows 100 authentication attempts per minute per IP
func PATAuthRateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(100, time.Minute)
	return limiter.RateLimit()
}
