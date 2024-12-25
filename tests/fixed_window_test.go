package tests

import (
	"testing"
	"time"

	"rate_limiter/algorithms"
	ratelimiter "rate_limiter/rate_limiter"
)

// Fixed Window test cases
// 1. Requests passing
// 2. Some Requests passing, Some requests are denied
// 3. Some Requests passing, Some requests are denied, then reset and then requests apssing

var tests = []struct {
	name            string
	limit           int
	windowSize      time.Duration
	allowedRequests int
	requests        int
	resetTimes      int
	sleepInterval   time.Duration
}{
	{
		name:            "all requests accepted",
		limit:           5,
		windowSize:      time.Second,
		allowedRequests: 5,
		requests:        5,
		resetTimes:      0,
		sleepInterval:   time.Duration(0),
	},
	{
		name:            "all requests accepted",
		limit:           5,
		windowSize:      time.Second,
		allowedRequests: 5,
		requests:        5,
		resetTimes:      5,
		sleepInterval:   time.Second,
	},
	{
		name:            "some requests accepted and some denied",
		limit:           5,
		windowSize:      time.Second,
		allowedRequests: 5,
		requests:        6,
		resetTimes:      0,
		sleepInterval:   time.Duration(0),
	},
	{
		name:            "some requests accepted, some denied and then a reset of time window occured",
		limit:           3,
		windowSize:      time.Second,
		allowedRequests: 5,
		requests:        7,
		resetTimes:      1,
		sleepInterval:   time.Second / 5,
	},
}

func TestFixedWindow(t *testing.T) {

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fw := algorithms.NewFixedWindow(tt.limit, tt.windowSize)
			rl := ratelimiter.NewRateLimiter(fw)

			allowedCount := 0
			deniedCount := 0
			resetCount := 0

			ticker := time.NewTicker(tt.windowSize)
			defer ticker.Stop()

			go func() {
				for range ticker.C {
					resetCount++
				}
			}()

			for i := 0; i < tt.requests; i++ {
				if rl.Allow("client") {
					allowedCount++
				} else {
					deniedCount++
				}

				time.Sleep(tt.sleepInterval)
			}

			if allowedCount != tt.allowedRequests {
				t.Errorf("expected %d allowed requests, got %d", tt.allowedRequests, allowedCount)
			}
			deniedRequests := tt.requests - tt.allowedRequests
			if deniedCount != deniedRequests {
				t.Errorf("expected %d denied requests, got %d", deniedRequests, deniedCount)
			}
			if resetCount != tt.resetTimes {
				t.Errorf("expected %d reset times, got %d", tt.resetTimes, resetCount)
			}
		})
	}
}
