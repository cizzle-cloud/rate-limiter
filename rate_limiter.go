package ratelimiter

import (
	"fmt"
	"sync"
	"time"
)

// RateLimitAlgo defines the interface for rate limiting algorithms.
type RateLimitAlgo interface {
	Allow() bool
}

// TokenBucket implements a token bucket rate limiting algorithm.
// It allows bursts of requests up to the bucket capacity and refills
// tokens at a specified rate.
type TokenBucket struct {
	capacity       int
	tokens         int
	refillTokens   int
	refillInterval time.Duration
	lastRefill     time.Time
	mu             sync.Mutex
}

// Allow checks if a request can be processed by consuming a token.
// It returns true if a token is available, false otherwise.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	if intervals := int(elapsed / tb.refillInterval); intervals > 0 {
		newTokens := intervals * tb.refillTokens
		tb.tokens = min(tb.tokens+newTokens, tb.capacity)
		tb.lastRefill = tb.lastRefill.Add(time.Duration(intervals) * tb.refillInterval)
	}
}

// NewTokenBucket creates a new TokenBucket with the specified capacity,
// refill rate, and refill interval.
func NewTokenBucket(capacity, refillTokens int, refillInterval time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity:       capacity,
		tokens:         capacity,
		refillTokens:   refillTokens,
		refillInterval: refillInterval,
		lastRefill:     time.Now(),
	}
}

// FixedWindowCounter implements a fixed window rate limiting algorithm.
// It allows a fixed number of requests within a specified time window.
type FixedWindowCounter struct {
	limit        int
	requestCount int
	mu           sync.Mutex
}

// Allow checks if a request can be processed within the current window.
// It returns true if the request count is below the limit, false otherwise.
func (fw *FixedWindowCounter) Allow() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.requestCount < fw.limit {
		fw.requestCount++
		fmt.Println("Request count", fw.requestCount)
		fmt.Println("Request allowed")
		return true
	}
	fmt.Println("Request denied")
	return false
}

// NewFixedWindowCounter creates a new FixedWindowCounter with the specified
// limit and window size. It automatically resets the counter after each window.
func NewFixedWindowCounter(limit int, windowSize time.Duration) *FixedWindowCounter {
	fw := &FixedWindowCounter{
		limit: limit,
	}

	go func() {
		ticker := time.NewTicker(windowSize)
		defer ticker.Stop()
		for range ticker.C {
			fw.mu.Lock()
			fw.requestCount = 0
			fw.mu.Unlock()
		}
	}()

	return fw

}

// Record holds a rate limiting algorithm instance and its last activity time.
type Record struct {
	algo       RateLimitAlgo
	lastActive time.Time
}

// RateLimiter manages multiple rate limiting records identified by keys.
// It automatically cleans up inactive records based on TTL.
type RateLimiter struct {
	records         map[string]*Record
	ttl             time.Duration
	cleanupInterval time.Duration
	mu              sync.Mutex
}

// Exists checks if a rate limiting record exists for the given key.
func (rl *RateLimiter) Exists(key string) bool {
	_, exists := rl.records[key]
	return exists
}

// Add creates a new rate limiting record for the given key with the specified algorithm.
func (rl *RateLimiter) Add(key string, algo RateLimitAlgo) {
	rl.records[key] = &Record{
		algo:       algo,
		lastActive: time.Now(),
	}
}

// GetRecords returns all current rate limiting records.
func (rl *RateLimiter) GetRecords() map[string]*Record {
	return rl.records
}

// Allow checks if a request for the given key is allowed by the associated
// rate limiting algorithm and updates the last activity time.
func (rl *RateLimiter) Allow(key string) bool {
	record := rl.records[key]
	record.lastActive = time.Now()
	return record.algo.Allow()
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, record := range rl.records {
			if now.Sub(record.lastActive) > rl.ttl {
				delete(rl.records, key)
			}
		}
		rl.mu.Unlock()
	}
}

// NewRateLimiter creates a new RateLimiter with the specified TTL for records
// and cleanup interval. It automatically starts a cleanup goroutine.
func NewRateLimiter(ttl, cleanupInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		records:         make(map[string]*Record),
		ttl:             ttl,
		cleanupInterval: cleanupInterval,
	}

	go rl.cleanup()

	return rl
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
