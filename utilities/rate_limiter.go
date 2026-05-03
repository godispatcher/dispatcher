package utilities

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter struct to hold tokens and metadata
type RateLimiter struct {
	Tokens         float64
	MaxTokens      float64
	RefillRate     float64 // tokens per second
	LastRefillTime time.Time
	mu             sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit float64, window time.Duration) *RateLimiter {
	refillRate := limit / window.Seconds()
	return &RateLimiter{
		Tokens:         limit,
		MaxTokens:      limit,
		RefillRate:     refillRate,
		LastRefillTime: time.Now(),
	}
}

// Result contains rate limit status and headers
type RateLimitResult struct {
	Allowed    bool
	Limit      int
	Remaining  int
	Reset      int64
	RetryAfter int
}

// Allow checks if a request is allowed and returns detailed result
func (rl *RateLimiter) Allow() *RateLimitResult {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.LastRefillTime).Seconds()
	rl.Tokens += elapsed * rl.RefillRate
	if rl.Tokens > rl.MaxTokens {
		rl.Tokens = rl.MaxTokens
	}
	rl.LastRefillTime = now

	res := &RateLimitResult{
		Limit: int(rl.MaxTokens),
		Reset: now.Add(time.Duration(1/rl.RefillRate) * time.Second).Unix(), // Approximate
	}

	if rl.Tokens >= 1 {
		rl.Tokens--
		res.Allowed = true
		res.Remaining = int(rl.Tokens)
		return res
	}

	res.Allowed = false
	res.Remaining = 0
	res.RetryAfter = int(1 / rl.RefillRate)
	return res
}

var (
	limiters = make(map[string]*RateLimiter)
	mu       sync.RWMutex
)

// GetRateLimiter returns a rate limiter for a given key
func GetRateLimiter(key string, limit int, windowSeconds int) *RateLimiter {
	mu.RLock()
	rl, exists := limiters[key]
	mu.RUnlock()

	if exists {
		// In a real system, we might want to update the limiter if config changed
		return rl
	}

	mu.Lock()
	defer mu.Unlock()
	// Double check after lock
	if rl, exists = limiters[key]; exists {
		return rl
	}

	rl = NewRateLimiter(float64(limit), time.Duration(windowSeconds)*time.Second)
	limiters[key] = rl
	return rl
}

// GenerateKey generates a unique key based on the scope and request data
func GenerateKey(department, transaction, scope, remoteAddr, licence, verifyCode string) string {
	base := fmt.Sprintf("%s:%s", department, transaction)

	switch scope {
	case "ip":
		return fmt.Sprintf("ip:%s:%s", remoteAddr, base)
	case "user":
		if verifyCode != "" {
			return fmt.Sprintf("user:%s:%s", verifyCode, base)
		}
		return fmt.Sprintf("ip:%s:%s", remoteAddr, base) // Fallback to IP
	case "api_key":
		if licence != "" {
			return fmt.Sprintf("apikey:%s:%s", licence, base)
		}
		return fmt.Sprintf("ip:%s:%s", remoteAddr, base) // Fallback to IP
	case "route":
		return fmt.Sprintf("route:%s", base)
	case "global":
		return "global:system"
	default:
		return fmt.Sprintf("default:%s", base)
	}
}
