package main

import (
	"fmt"
	"sync"
	"time"
)

type RateLimitAlgo interface {
	Allow() bool
}

type TokenBucket struct {
	tokens     int
	lastRefill time.Time
	refillRate int
	bucketSize int
	mu         sync.Mutex
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	// Add tokens based on time passed
	tb.tokens += int(elapsed * float64(tb.refillRate))
	if tb.tokens > tb.bucketSize {
		tb.tokens = tb.bucketSize
	}

	// Check if request is allowed
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}
	return false
}

func NewTokenBucket(refillRate, bucketSize int) *TokenBucket {
	return &TokenBucket{
		tokens:     bucketSize,
		lastRefill: time.Now(),
		refillRate: refillRate,
		bucketSize: bucketSize,
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

type RateLimiter struct {
	records map[string]RateLimitAlgo
}

func (rl *RateLimiter) Allow(key string) bool {
	rateLimitAlgo := rl.records[key]
	return rateLimitAlgo.Allow()
}

func main() {
	limiter1 := NewFixedWindowCounter(5, 10*time.Second) // 5 requests per 10 seconds
	limiter2 := NewFixedWindowCounter(5, 10*time.Second) // 5 requests per 10 seconds

	for i := 1; i <= 10; i++ {
		fmt.Println(" =========Request==========", i)
		limiter1.Allow()
		limiter2.Allow()
		time.Sleep(1 * time.Second) // Simulate request interval
	}
}
