//go:build unix

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cnd-industries/twitch-whip-that-agent/internal/whip"
)

const retryDelay = 2 * time.Second

// listen receives whips from whipbot until ctx is canceled. Errors are
// not reported: the terminal belongs to Claude Code, and whipbot may
// simply not be running yet, so it keeps retrying quietly.
func listen(ctx context.Context, url string, handle func(whip.Whip)) {
	for {
		_ = stream(ctx, url, handle)

		select {
		case <-ctx.Done():
			return
		case <-time.After(retryDelay):
		}
	}
}

// stream reads one Server-Sent Events connection. whipbot sends a single
// data line per whip.
func stream(ctx context.Context, url string, handle func(whip.Whip)) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("connect to whipbot: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s", response.Status)
	}

	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		data, ok := strings.CutPrefix(scanner.Text(), "data: ")
		if !ok {
			continue
		}

		var received whip.Whip
		if err := json.Unmarshal([]byte(data), &received); err != nil {
			// A malformed whip is skipped so the following ones still
			// arrive.
			continue
		}
		handle(received)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read events: %w", err)
	}

	return nil
}
