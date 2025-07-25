package ratelimiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAllowFw(t *testing.T) {
	limit := 5
	windowSize := 500 * time.Millisecond
	fw := NewFixedWindowCounter(limit, windowSize)
	for i := 0; i < limit; i++ {
		if !fw.Allow() {
			t.Errorf("Expected Allow() to return true on iteration %d", i)
		}
	}

	if fw.Allow() {
		t.Error("Expected Allow() to return false when limit is exceeded on the given window size.")
	}

}

func TestRefillFw(t *testing.T) {
	limit := 5
	windowSize := 500 * time.Millisecond
	fw := NewFixedWindowCounter(limit, windowSize)
	for i := 0; i < limit; i++ {
		if !fw.Allow() {
			t.Errorf("Expected Allow() to return true on iteration %d", i)
		}
	}

	time.Sleep(windowSize + 10*time.Millisecond)

	if !fw.Allow() {
		t.Error("Expected Allow() to return true after moving to the next window.")
	}

}

func TestConcurrencyFw(t *testing.T) {
	limit := 5
	windowSize := 500 * time.Millisecond
	fw := NewFixedWindowCounter(limit, windowSize)

	var wg sync.WaitGroup
	var count int32
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if fw.Allow() {
				atomic.AddInt32(&count, 1)
			}
		}()
	}

	wg.Wait()

	if int(count) > limit {
		t.Errorf("Allowed count (%d) exceeded bucket capacity (%d)", count, limit)
	}
}

func TestAllowTb(t *testing.T) {
	capacity := 5
	refillTokens := 1
	refillInterval := 500 * time.Millisecond // Make sure is big enough to let the capacity be exhausted

	tb := NewTokenBucket(capacity, refillTokens, refillInterval)

	// Capacity should be exhausted
	for i := 0; i < capacity; i++ {
		if !tb.Allow() {
			t.Errorf("Expected Allow() to return true on iteration %d", i)
		}
	}

	// Now the bucket should be empty
	if tb.Allow() {
		t.Error("Expected Allow() to return false when tokens are deleted")
	}
}

func TestRefillTb(t *testing.T) {
	capacity := 5
	refillTokens := 1
	refillInterval := 500 * time.Millisecond

	tb := NewTokenBucket(capacity, refillTokens, refillInterval)

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

func TestConcurrencyTb(t *testing.T) {
	capacity := 10
	refillTokens := 2
	refillInterval := 500 * time.Millisecond

	tb := NewTokenBucket(capacity, refillTokens, refillInterval)

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

func TestCleanup(t *testing.T) {
	limiter := NewRateLimiter(time.Second, time.Second)

	capacity := 5
	refillTokens := 2
	refillInterval := time.Second

	// Create a record
	limiter.Add(
		"client1",
		NewTokenBucket(capacity, refillTokens, refillInterval),
	)

	// Sleep for one cleanup interval
	time.Sleep(time.Second)

	if len(limiter.GetRecords()) != 0 {
		t.Error("should be zero")
	}

}
