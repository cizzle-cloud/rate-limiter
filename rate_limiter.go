package ratelimiter

type RateLimitAlgo interface {
	Allow(clientID string) bool
}

type RateLimiter struct {
	rateLimitAlgo RateLimitAlgo
}

func (rl *RateLimiter) Allow(clientID string) bool {
	return rl.rateLimitAlgo.Allow(clientID)
}

func NewRateLimiter(rateLimitAlgo RateLimitAlgo) *RateLimiter {
	return &RateLimiter{rateLimitAlgo: rateLimitAlgo}
}
