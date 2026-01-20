package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// corsMiddleware adds CORS headers for local development
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// securityHeadersMiddleware adds security headers to all responses
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Writer.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Next()
	}
}

// rateLimitMiddleware implements simple per-IP rate limiting
func rateLimitMiddleware(rps int) gin.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.RWMutex
		clients = make(map[string]*client)
		once    sync.Once // Fix Issue #2: Ensure cleanup goroutine starts only once
	)

	// Fix Issue #2: Use sync.Once to prevent multiple cleanup goroutines
	// Start cleanup goroutine only once, even if middleware is called multiple times
	once.Do(func() {
		go func() {
			for {
				time.Sleep(5 * time.Minute)
				mu.Lock()
				for ip, c := range clients {
					if time.Since(c.lastSeen) > 10*time.Minute {
						delete(clients, ip)
					}
				}
				mu.Unlock()
			}
		}()
	})

	return func(c *gin.Context) {
		// Skip rate limiting for health checks
		if c.Request.URL.Path == "/api/health" {
			c.Next()
			return
		}

		ip := c.ClientIP()

		mu.Lock()
		if _, exists := clients[ip]; !exists {
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Limit(rps), rps*2), // Allow bursts
			}
		}
		clients[ip].lastSeen = time.Now()
		mu.Unlock()

		mu.RLock()
		limiter := clients[ip].limiter
		mu.RUnlock()

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
