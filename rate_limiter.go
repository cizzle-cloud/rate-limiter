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
	capacity float32
	tokens   float32
	mu       sync.Mutex
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.tokens >= 1 {
		tb.tokens--
		fmt.Println("Request allowed")
		fmt.Println("Number of tokens", tb.tokens)
		return true
	}
	fmt.Println("Request denied")
	fmt.Println("Number of tokens", tb.tokens)
	return false
}

func NewTokenBucket(refillTokens float32, refillInterval time.Duration, capacity float32) *TokenBucket {
	tb := &TokenBucket{
		capacity: capacity,
		tokens:   capacity,
	}

	go func() {
		ticker := time.NewTicker(refillInterval)
		defer ticker.Stop()

		for range ticker.C {
			tb.mu.Lock()
			tb.tokens += refillTokens
			if tb.tokens > tb.capacity {
				tb.tokens = tb.capacity
			}
			fmt.Println("tokens after adding", tb.tokens)
			tb.mu.Unlock()

		}
	}()

	return tb
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
	// limiter1 := NewFixedWindowCounter(5, 10*time.Second) // 5 requests per 10 seconds
	// limiter2 := NewFixedWindowCounter(5, 10*time.Second) // 5 requests per 10 seconds
	limiter3 := NewTokenBucket(1, 2*time.Second, 5)

	for i := 1; i <= 10; i++ {
		fmt.Println(" =========Request==========", i)
		// limiter1.Allow()
		// limiter2.Allow()
		limiter3.Allow()
		time.Sleep(500 * time.Millisecond) // Simulate request interval
	}
}
