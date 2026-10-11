package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	limiter := New(2, 50*time.Millisecond)

	if !limiter.Allow("a") || !limiter.Allow("a") {
		t.Fatal("first two events must be allowed")
	}
	if limiter.Allow("a") {
		t.Fatal("third event within the window must be denied")
	}
	if !limiter.Allow("b") {
		t.Fatal("other keys must not be affected")
	}

	time.Sleep(60 * time.Millisecond)

	if !limiter.Allow("a") {
		t.Fatal("event after the window must be allowed")
	}
}
