package rate

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Limiter provides rate limiting functionality
type Limiter struct {
	globalLimiter *rate.Limiter
	ipLimiters    map[string]*rate.Limiter
	mu            sync.RWMutex
	cleanupTicker *time.Ticker
	config        Config
}

// Config holds rate limiting configuration
type Config struct {
	GlobalRPS  int           `json:"global_rps"`
	PerIPRPS   int           `json:"per_ip_rps"`
	BurstSize  int           `json:"burst_size"`
	WindowSize time.Duration `json:"window_size"`
}

// NewLimiter creates a new rate limiter
func NewLimiter(config Config) *Limiter {
	limiter := &Limiter{
		globalLimiter: rate.NewLimiter(rate.Limit(config.GlobalRPS), config.BurstSize),
		ipLimiters:    make(map[string]*rate.Limiter),
		config:        config,
	}

	// Start cleanup goroutine
	limiter.cleanupTicker = time.NewTicker(time.Minute)
	go limiter.cleanup()

	return limiter
}

// cleanup removes old IP limiters to prevent memory leaks
func (l *Limiter) cleanup() {
	for range l.cleanupTicker.C {
		l.mu.Lock()
			// Remove limiters that haven't been used in the last 10 minutes
	cutoff := time.Now().Add(-10 * time.Minute)
	for ip := range l.ipLimiters {
		// This is a simplified cleanup - in production you might want to track last access
		if time.Since(cutoff) > 0 {
			delete(l.ipLimiters, ip)
		}
	}
		l.mu.Unlock()
	}
}

// getIPLimiter gets or creates a rate limiter for a specific IP
func (l *Limiter) getIPLimiter(ip string) *rate.Limiter {
	l.mu.RLock()
	limiter, exists := l.ipLimiters[ip]
	l.mu.RUnlock()

	if exists {
		return limiter
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists = l.ipLimiters[ip]; exists {
		return limiter
	}

	// Create new limiter for this IP
	limiter = rate.NewLimiter(rate.Limit(l.config.PerIPRPS), l.config.BurstSize)
	l.ipLimiters[ip] = limiter

	return limiter
}

// Allow checks if a request is allowed
func (l *Limiter) Allow(ip string) bool {
	// Check global rate limit first
	if !l.globalLimiter.Allow() {
		return false
	}

	// Check per-IP rate limit
	ipLimiter := l.getIPLimiter(ip)
	return ipLimiter.Allow()
}

// Wait waits for a request to be allowed
func (l *Limiter) Wait(ctx context.Context, ip string) error {
	// Wait for global rate limit
	if err := l.globalLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("global rate limit wait failed: %w", err)
	}

	// Wait for per-IP rate limit
	ipLimiter := l.getIPLimiter(ip)
	if err := ipLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("per-IP rate limit wait failed: %w", err)
	}

	return nil
}

// Close stops the cleanup goroutine
func (l *Limiter) Close() {
	if l.cleanupTicker != nil {
		l.cleanupTicker.Stop()
	}
}

// RateLimitMiddleware creates a Gin middleware for rate limiting
func RateLimitMiddleware(limiter *Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)

		// Check if request is allowed
		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitWaitMiddleware creates a Gin middleware that waits for rate limits
func RateLimitWaitMiddleware(limiter *Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)

		// Wait for rate limit
		if err := limiter.Wait(c.Request.Context(), ip); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "rate_limit_error",
				"message": "Rate limiting error occurred",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (for proxy scenarios)
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		// Take the first IP in the chain
		if idx := len(forwardedFor); idx > 0 {
			if commaIdx := len(forwardedFor); commaIdx > 0 {
				forwardedFor = forwardedFor[:commaIdx]
			}
		}
		return forwardedFor
	}

	// Check X-Real-IP header
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fallback to remote address
	return c.ClientIP()
}

// GetStats returns rate limiting statistics
func (l *Limiter) GetStats() map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["global_rps"] = l.config.GlobalRPS
	stats["per_ip_rps"] = l.config.PerIPRPS
	stats["burst_size"] = l.config.BurstSize
	stats["window_size"] = l.config.WindowSize
	stats["active_ip_limiters"] = len(l.ipLimiters)

	return stats
}
