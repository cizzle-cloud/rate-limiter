package tests

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ratelimiter "github.com/cizzle-cloud/rate-limiter"
)

func TestAllowFw(t *testing.T) {
	limit := 5
	windowSize := 500 * time.Millisecond
	fw := ratelimiter.NewFixedWindowCounter(limit, windowSize)
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
	fw := ratelimiter.NewFixedWindowCounter(limit, windowSize)
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
	fw := ratelimiter.NewFixedWindowCounter(limit, windowSize)

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
