package algorithms

import (
	"sync"
	"time"
)

type FixedWindow struct {
	limit         int
	windowSize    time.Duration
	requestCounts map[string]int
	mu            sync.RWMutex // allows concurrent reads if write contention is low. TODO @NG: investigate comparison
	resetTicker   *time.Ticker
}

func (fw *FixedWindow) Allow(clientID string) bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.requestCounts[clientID] >= fw.limit {
		return false
	}

	fw.requestCounts[clientID]++

	return true
}

func NewFixedWindow(limit int, windowSize time.Duration) *FixedWindow {
	fw := &FixedWindow{
		limit:         limit,
		windowSize:    windowSize,
		requestCounts: make(map[string]int),
		resetTicker:   time.NewTicker(windowSize),
	}

	go func() {
		for range fw.resetTicker.C {
			fw.mu.Lock()
			fw.requestCounts = make(map[string]int)
			fw.mu.Unlock()
		}
	}()

	return fw
}
