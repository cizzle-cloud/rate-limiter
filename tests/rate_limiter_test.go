package tests

import (
	"testing"
	"time"

	ratelimiter "github.com/cizzle-cloud/rate-limiter/rate_limiter"
)

func TestCleanup(t *testing.T) {
	limiter := ratelimiter.NewRateLimiter(time.Second, time.Second)

	capacity := 5
	refillTokens := 2
	refillInterval := time.Second

	// Create a record
	limiter.Add(
		"client1",
		ratelimiter.NewTokenBucket(capacity, refillTokens, refillInterval),
	)

	// sleep for one cleanup interval
	time.Sleep(time.Second)

	if len(limiter.GetRecords()) != 0 {
		t.Error("should be zero")
	}

}
