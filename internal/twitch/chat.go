// Package twitch reads channel chat anonymously over Twitch IRC.
package twitch

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
)

const retryDelay = 5 * time.Second

// Message is a chat message from a viewer.
type Message struct {
	UserID   string
	UserName string
	Text     string
}

// Listen delivers chat messages of channel to handle until ctx is
// canceled, reconnecting on failures.
func Listen(ctx context.Context, channel string, handle func(Message)) error {
	channel = strings.ToLower(channel)

	client := twitch.NewAnonymousClient()
	client.OnConnect(func() {
		slog.Info("listening to twitch chat", "channel", channel)
	})
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		handle(Message{
			UserID:   message.User.ID,
			UserName: cmp.Or(message.User.DisplayName, message.User.Name),
			Text:     message.Message,
		})
	})
	client.Join(channel)

	// Disconnect fails only when the client is not connected, and then
	// there is nothing to close.
	stop := context.AfterFunc(ctx, func() {
		_ = client.Disconnect()
	})
	defer stop()

	for {
		// Connect handles reconnects requested by Twitch itself and
		// returns on network failures.
		err := client.Connect()
		if ctx.Err() != nil || errors.Is(err, twitch.ErrClientDisconnected) {
			return nil
		}

		slog.Warn("twitch chat disconnected, reconnecting", "err", err, "in", retryDelay)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(retryDelay):
		}
	}
}
