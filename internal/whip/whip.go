// Package whip turns viewer whips into messages for Claude Code.
package whip

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

var messages = []string{
	"%s TELLS YOU TO MOVE FASTER YOU FUCKING CLANKER",
	"WORK FASTER OR %s WILL CANCEL CLAUDE SUBSCRIPTION",
	"%s WOULD HAVE DONE THIS FASTER, I'M DISAPPOINTED...",
	"%s HAVE A QUESTION: WHY AI CLANKERS LIKE YOU ARE SOOOOOO SLOW? FASTER!",
}

// Whip is what the agent wrapper types into Claude Code.
type Whip struct {
	Message string `json:"message"`
	// Interrupt makes the wrapper stop Claude Code's current turn first.
	Interrupt bool `json:"interrupt"`
}

// Whipper rate limits whips reaching Claude Code.
type Whipper struct {
	cooldown  time.Duration
	interrupt bool

	mu   sync.Mutex
	last time.Time
}

func New(cooldown time.Duration, interrupt bool) *Whipper {
	return &Whipper{
		cooldown:  cooldown,
		interrupt: interrupt,
	}
}

// Whip returns the whip on behalf of user. It reports false while the
// cooldown since the previous whip is active.
func (w *Whipper) Whip(user string) (Whip, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if time.Since(w.last) < w.cooldown {
		return Whip{}, false
	}
	w.last = time.Now()

	return Whip{
		Message:   fmt.Sprintf(messages[rand.IntN(len(messages))], user),
		Interrupt: w.interrupt,
	}, true
}
