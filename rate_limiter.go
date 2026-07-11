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
	mu         sync.Mutex
}

// Allow checks if a request is allowed by the rate limiting algorithm
// and updates the last activity time.
func (r *Record) Allow() bool {
	r.mu.Lock()
	r.lastActive = time.Now()
	r.mu.Unlock()
	return r.algo.Allow()
}

// RateLimiter manages multiple rate limiting records identified by keys.
// It automatically cleans up inactive records based on TTL.
type RateLimiter struct {
	records         map[string]*Record
	ttl             time.Duration
	cleanupInterval time.Duration
	mu              sync.Mutex
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

// GetOrCreate returns the existing record for the given key, or creates a new one
// with the provided algorithm.
func (rl *RateLimiter) GetOrCreate(key string, algo RateLimitAlgo) *Record {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if record, exists := rl.records[key]; exists {
		return record
	}
	record := &Record{
		algo:       algo,
		lastActive: time.Now(),
	}
	rl.records[key] = record
	return record
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, record := range rl.records {
			record.mu.Lock()
			expired := now.Sub(record.lastActive) > rl.ttl
			record.mu.Unlock()
			if expired {
				delete(rl.records, key)
			}
		}
		rl.mu.Unlock()
	}
}

// getRecords returns a copy of all current rate limiting records.
func (rl *RateLimiter) getRecords() map[string]*Record {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	records := make(map[string]*Record, len(rl.records))
	for k, v := range rl.records {
		records[k] = v
	}
	return records
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
