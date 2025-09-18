package coordinator

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter represents a rate limiter for a specific IP address
type RateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitMiddleware provides rate limiting functionality
type RateLimitMiddleware struct {
	limiters map[string]*RateLimiter
	mutex    sync.RWMutex
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(requestsPerSecond float64, burst int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiters: make(map[string]*RateLimiter),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
		cleanup:  5 * time.Minute, // Clean up old limiters every 5 minutes
	}
}

// getLimiter returns the rate limiter for the given IP address
func (rl *RateLimitMiddleware) getLimiter(ip string) *rate.Limiter {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = &RateLimiter{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.limiters[ip] = limiter
	} else {
		limiter.lastSeen = time.Now()
	}

	return limiter.limiter
}

// cleanupOldLimiters removes old rate limiters to prevent memory leaks
func (rl *RateLimitMiddleware) cleanupOldLimiters() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	for ip, limiter := range rl.limiters {
		if now.Sub(limiter.lastSeen) > rl.cleanup {
			delete(rl.limiters, ip)
		}
	}
}

// StartCleanup starts the cleanup goroutine
func (rl *RateLimitMiddleware) StartCleanup() {
	go func() {
		ticker := time.NewTicker(rl.cleanup)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanupOldLimiters()
		}
	}()
}

// getClientIP extracts the client IP address from the request
func (rl *RateLimitMiddleware) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for load balancers/proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if ip := net.ParseIP(xff); ip != nil {
			return ip.String()
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return ip.String()
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// RateLimit is the middleware function that enforces rate limiting
func (rl *RateLimitMiddleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := rl.getClientIP(r)
		limiter := rl.getLimiter(clientIP)

		// Check if the request is allowed
		if !limiter.Allow() {
			WriteJSONError(w, "rate limit exceeded", 429,
				fmt.Sprintf("too many requests from IP %s (limit: %.1f req/sec)", clientIP, float64(rl.rate)))
			return
		}

		// Request is allowed, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware provides request logging functionality
type LoggingMiddleware struct{}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

// LogRequest is the middleware function that logs HTTP requests
func (lm *LoggingMiddleware) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom ResponseWriter to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Log the request
		duration := time.Since(start)
		fmt.Printf("[%s] %s %s %d %v %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
			r.RemoteAddr,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RecoveryMiddleware provides panic recovery functionality
type RecoveryMiddleware struct{}

// NewRecoveryMiddleware creates a new recovery middleware
func NewRecoveryMiddleware() *RecoveryMiddleware {
	return &RecoveryMiddleware{}
}

// RecoverPanic is the middleware function that recovers from panics
func (rm *RecoveryMiddleware) RecoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Panic recovered: %v\n", err)
				WriteJSONError(w, "internal server error", 500,
					"an unexpected error occurred while processing the request")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware provides CORS functionality
type CORSMiddleware struct{}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{}
}

// HandleCORS is the middleware function that handles CORS
func (cm *CORSMiddleware) HandleCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
