package ratelimiter

import (
	"fmt"
	"sync"
	"time"
)

type RateLimitAlgo interface {
	Allow() bool
}

type TokenBucket struct {
	capacity       int
	tokens         int
	refillTokens   int
	refillInterval time.Duration
	lastRefill     time.Time
	mu             sync.Mutex
}

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

func (tb *TokenBucket) GetTokens() int {
	return tb.tokens
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func NewTokenBucket(capacity, refillTokens int, refillInterval time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity:       capacity,
		tokens:         capacity,
		refillTokens:   refillTokens,
		refillInterval: refillInterval,
		lastRefill:     time.Now(),
	}
}

type FixedWindowCounter struct {
	limit        int
	requestCount int
	mu           sync.Mutex
}

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

type Record struct {
	algo       RateLimitAlgo
	lastActive time.Time
}

type RateLimiter struct {
	records         map[string]*Record
	ttl             time.Duration
	cleanupInterval time.Duration
	mu              sync.Mutex
}

func (rl *RateLimiter) Exists(key string) bool {
	_, exists := rl.records[key]
	return exists
}

func (rl *RateLimiter) Add(key string, algo RateLimitAlgo) {
	rl.records[key] = &Record{
		algo:       algo,
		lastActive: time.Now(),
	}
}

func (rl *RateLimiter) GetRecords() map[string]*Record {
	return rl.records
}

func (rl *RateLimiter) Allow(key string) bool {
	record := rl.records[key]
	record.lastActive = time.Now()
	return record.algo.Allow()
}

// really a quick and dirt solution. ideally I would like a cache struct
// injected in rate limiter. But for now is included so runtime memory
// is not cluttered

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

func NewRateLimiter(ttl, cleanupInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		records:         make(map[string]*Record),
		ttl:             ttl,
		cleanupInterval: cleanupInterval,
	}

	go rl.cleanup()

	return rl
}
