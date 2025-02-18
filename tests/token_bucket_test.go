package tests

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ratelimiter "github.com/cizzle-cloud/rate-limiter/rate_limiter"
)

func TestAllow(t *testing.T) {
	capacity := 5
	refillTokens := 1
	refillInterval := 500 * time.Millisecond // Make sure is big enough to let the capacity be exhausted

	tb := ratelimiter.NewTokenBucket(capacity, refillTokens, refillInterval)

	// Capacity should be exhausted
	for i := 0; i < capacity; i++ {
		if !tb.Allow() {
			t.Errorf("Expected Allow() to return true on iteration %d", i)
		}
	}

	// Now the bucket should be empty
	if tb.Allow() {
		t.Error("Expected Allow() to return false when tokens are depleted")
	}
}

func TestRefill(t *testing.T) {
	capacity := 5
	refillTokens := 1
	refillInterval := 500 * time.Millisecond

	tb := ratelimiter.NewTokenBucket(capacity, refillTokens, refillInterval)

	// Empty the bucket
	for i := 0; i < capacity; i++ {
		if !tb.Allow() {
			t.Fatalf("Expected Allow() to return true on iteration %d", i)
		}
	}

	// Ensure bucket is empty
	if tb.Allow() {
		t.Error("Expected Allow() to return false after emptying the bucket")
	}

	// Wait for one refill interval + a small tolerance
	time.Sleep(refillInterval + 10*time.Millisecond)

	// One token should have been refilled
	if !tb.Allow() {
		t.Error("Expected Allow() to return true after one refill interval")
	}
}

func TestConcurrency(t *testing.T) {
	capacity := 10
	refillTokens := 2
	refillInterval := 500 * time.Millisecond

	tb := ratelimiter.NewTokenBucket(capacity, refillTokens, refillInterval)

	var wg sync.WaitGroup
	var count int32

	// Launch 20 goroutines and call Allow concurrently
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if tb.Allow() {
				atomic.AddInt32(&count, 1)
			}
		}()
	}

	wg.Wait()

	// Count should be smaller than bucket's capacity
	if int(count) > capacity {
		t.Errorf("Allowed count (%d) exceeded bucket capacity (%d)", count, capacity)
	}
}
