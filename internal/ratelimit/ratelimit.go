// Package ratelimit limits events per key with golang.org/x/time/rate.
package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter allows bursts of up to limit events per key and refills them
// evenly over window.
type Limiter struct {
	limit  int
	window time.Duration

	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func New(limit int, window time.Duration) *Limiter {
	return &Limiter{
		limit:    limit,
		window:   window,
		limiters: make(map[string]*rate.Limiter),
	}
}

// Allow records an event for key and reports whether it fits the limit.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	limiter, ok := l.limiters[key]
	if !ok {
		timing := rate.Every(l.window / time.Duration(l.limit))
		limiter = rate.NewLimiter(timing, l.limit)
		l.limiters[key] = limiter
	}
	l.mu.Unlock()

	return limiter.Allow()
}
