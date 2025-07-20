# Rate Limiter

[![Go Report Card](https://goreportcard.com/badge/github.com/cizzle-cloud/rate-limiter)](https://goreportcard.com/report/github.com/cizzle-cloud/rate-limiter)
[![codecov](https://codecov.io/gh/cizzle-cloud/rate-limiter/branch/main/graph/badge.svg)](https://codecov.io/gh/cizzle-cloud/rate-limiter)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

A flexible rate limiting library for Go applications with support for multiple algorithms and automatic cleanup of inactive records.

## Features

- **Multiple Rate Limiting Algorithms**
  - Token Bucket: Allows bursts up to bucket capacity with configurable refill rates
  - Fixed Window Counter: Limits requests within fixed time windows
  - Extensible interface for custom algorithms

- **Key-Based Rate Limiting**: Associate rate limiters with specific keys (users, IPs, API keys, etc.)
- **Automatic Cleanup**: Automatically removes inactive rate limiting records based on TTL
- **Thread-Safe**: All operations are protected with mutexes for concurrent use
- **Zero Dependencies**: Pure Go implementation with no external dependencies

## Installation

```bash
go get github.com/cizzle-cloud/rate-limiter
```

## Quick Start

### Token Bucket Example

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/cizzle-cloud/rate-limiter"
)

func main() {
    // Create a token bucket with capacity of 10, refilling 5 tokens every second
    tb := ratelimiter.NewTokenBucket(10, 5, time.Second)
    
    // Check if request is allowed
    if tb.Allow() {
        fmt.Println("Request allowed")
    } else {
        fmt.Println("Request denied - rate limit exceeded")
    }
}
```

### Fixed Window Counter Example

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/cizzle-cloud/rate-limiter"
)

func main() {
    // Allow 100 requests per minute
    fwc := ratelimiter.NewFixedWindowCounter(100, time.Minute)
    
    if fwc.Allow() {
        fmt.Println("Request allowed")
    } else {
        fmt.Println("Request denied - rate limit exceeded")
    }
}
```

### Rate Limiter with Keys

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/cizzle-cloud/rate-limiter"
)

func main() {
    // Create rate limiter with 1 hour TTL and 10 minute cleanup interval
    rl := ratelimiter.NewRateLimiter(time.Hour, 10*time.Minute)
    
    userID := "user123"
    
    // Check if user record exists
    if !rl.Exists(userID) {
        // Create a token bucket for this user (10 requests, refill 1 per second)
        tb := ratelimiter.NewTokenBucket(10, 1, time.Second)
        rl.Add(userID, tb)
    }
    
    // Check if user's request is allowed
    if rl.Allow(userID) {
        fmt.Println("User request allowed")
    } else {
        fmt.Println("User request denied")
    }
}
```

## API Reference

### Interfaces

#### RateLimitAlgo
```go
type RateLimitAlgo interface {
    Allow() bool
}
```

### Types

#### TokenBucket
Implements a token bucket rate limiting algorithm.

**Constructor:**
```go
func NewTokenBucket(capacity, refillTokens int, refillInterval time.Duration) *TokenBucket
```

**Methods:**
- `Allow() bool` - Consumes a token if available

#### FixedWindowCounter
Implements a fixed window rate limiting algorithm.

**Constructor:**
```go
func NewFixedWindowCounter(limit int, windowSize time.Duration) *FixedWindowCounter
```

**Methods:**
- `Allow() bool` - Checks if request count is within limit

#### RateLimiter
Manages multiple rate limiting records with automatic cleanup.

**Constructor:**
```go
func NewRateLimiter(ttl, cleanupInterval time.Duration) *RateLimiter
```

**Methods:**
- `Exists(key string) bool` - Checks if record exists
- `Add(key string, algo RateLimitAlgo)` - Adds new rate limiting record
- `Allow(key string) bool` - Checks if request is allowed for key
- `GetRecords() map[string]*Record` - Returns all records

## Use Cases

- **API Rate Limiting**: Limit requests per user, IP address, or API key
- **Resource Protection**: Prevent abuse of expensive operations
- **Traffic Shaping**: Control request flow to downstream services
- **DDoS Mitigation**: Basic protection against request floods
- **Fair Usage**: Ensure equal resource access among users

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---