package utilities

import (
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

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.LastRefillTime).Seconds()
	rl.Tokens += elapsed * rl.RefillRate
	if rl.Tokens > rl.MaxTokens {
		rl.Tokens = rl.MaxTokens
	}
	rl.LastRefillTime = now

	if rl.Tokens >= 1 {
		rl.Tokens--
		return true
	}

	return false
}

var (
	limiters = make(map[string]*RateLimiter)
	mu       sync.RWMutex
)

// GetRateLimiter returns a rate limiter for a given key (e.g., department:transaction)
func GetRateLimiter(key string, limit int, windowSeconds int) *RateLimiter {
	mu.RLock()
	rl, exists := limiters[key]
	mu.RUnlock()

	if exists {
		// Update refill rate if limit or window changed (simplified)
		// For now, we assume once created it stays the same or we could re-create it.
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
