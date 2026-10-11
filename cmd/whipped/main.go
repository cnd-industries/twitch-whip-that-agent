//go:build unix

// Command whipped runs Claude Code in a pseudo-terminal and types whips
// from whipbot into it the moment viewers send them.
//
// Usage: whipped [command [args...]]; the command defaults to Claude.
package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

func main() {
	code, err := run()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "whipped:", err)
		code = 1
	}
	os.Exit(code)
}

func run() (int, error) {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"claude"}
	}
	serverURL := cmp.Or(os.Getenv("WHIPBOT_URL"), "http://127.0.0.1:8080")

	command := exec.Command(args[0], args[1:]...)
	terminal, err := pty.Start(command)
	if err != nil {
		return 0, fmt.Errorf("start %s: %w", args[0], err)
	}
	defer func() {
		_ = terminal.Close()
	}()

	stopResizing := inheritSize(terminal)
	defer stopResizing()

	restore, err := makeRaw(os.Stdin)
	if err != nil {
		return 0, err
	}
	defer restore()

	keyboard := &input{
		writer: terminal,
	}
	go func() {
		_, _ = io.Copy(keyboard, os.Stdin)
	}()
	output := copyOutput(terminal)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go listen(ctx, serverURL+"/agent/events", keyboard.whip)

	return wait(command, output)
}

// wait waits for the command to exit and returns its exit code.
func wait(command *exec.Cmd, output <-chan struct{}) (int, error) {
	err := command.Wait()

	// Let the last output reach the terminal; a leftover grandchild may
	// keep the pty open, so do not wait for it forever.
	select {
	case <-output:
	case <-time.After(time.Second):
	}

	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		return exitErr.ExitCode(), nil
	}
	if err != nil {
		return 0, fmt.Errorf("wait for %s: %w", command.Path, err)
	}

	return 0, nil
}
