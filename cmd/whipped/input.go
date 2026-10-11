//go:build unix

package main

import (
	"io"
	"sync"
	"time"

	"github.com/cnd-industries/twitch-whip-that-agent/internal/whip"
)

// input is the child's keyboard, shared by the user and whips.
type input struct {
	mu     sync.Mutex
	writer io.Writer
}

func (i *input) Write(keys []byte) (int, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.writer.Write(keys)
}

// whip types the message as if the user did. The user's keystrokes wait
// until it is done so they do not get mixed into it. Write errors are
// ignored: they mean Claude Code has exited, and run reports that.
func (i *input) whip(received whip.Whip) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if received.Interrupt {
		// Escape interrupts the current Claude Code turn.
		_, _ = i.writer.Write([]byte{0x1b})
		// A pause keeps the next bytes from being read as an Alt
		// sequence and lets Claude Code stop.
		time.Sleep(200 * time.Millisecond)
	}

	_, _ = io.WriteString(i.writer, received.Message)
	// Enter arriving separately is not taken as part of a paste.
	time.Sleep(50 * time.Millisecond)
	_, _ = i.writer.Write([]byte{'\r'})
}
