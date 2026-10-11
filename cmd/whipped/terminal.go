//go:build unix

package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// inheritSize keeps the size of the child's terminal in sync with ours
// until the returned function is called.
func inheritSize(terminal *os.File) (stop func()) {
	resizes := make(chan os.Signal, 1)
	signal.Notify(resizes, syscall.SIGWINCH)
	go func() {
		for range resizes {
			// A failed resize only distorts the layout until the next one.
			_ = pty.InheritSize(os.Stdin, terminal)
		}
	}()
	resizes <- syscall.SIGWINCH

	return func() {
		signal.Stop(resizes)
	}
}

// makeRaw switches the terminal to raw mode, which passes every key,
// including Ctrl-C, through to the child. It does nothing when the input
// is not a terminal.
func makeRaw(file *os.File) (restore func(), err error) {
	descriptor := int(file.Fd())
	if !term.IsTerminal(descriptor) {
		return func() {}, nil
	}

	state, err := term.MakeRaw(descriptor)
	if err != nil {
		return nil, fmt.Errorf("switch terminal to raw mode: %w", err)
	}

	return func() {
		_ = term.Restore(descriptor, state)
	}, nil
}

// copyOutput shows the child's output and closes the returned channel
// once the terminal is closed.
func copyOutput(terminal io.Reader) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.Copy(os.Stdout, terminal)
	}()

	return done
}
