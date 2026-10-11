// Command whipbot lets Twitch viewers whip the streamer's Claude Code
// session from chat and shows the whips on an OBS overlay.
package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cnd-industries/twitch-whip-that-agent/internal/config"
	"github.com/cnd-industries/twitch-whip-that-agent/internal/leaderboard"
	"github.com/cnd-industries/twitch-whip-that-agent/internal/ratelimit"
	"github.com/cnd-industries/twitch-whip-that-agent/internal/sse"
	"github.com/cnd-industries/twitch-whip-that-agent/internal/twitch"
	"github.com/cnd-industries/twitch-whip-that-agent/internal/whip"
	"github.com/cnd-industries/twitch-whip-that-agent/web"
)

const leaderboardSize = 10

func main() {
	if err := run(); err != nil {
		slog.Error("failed to run", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	settings, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	application := &app{
		board:   leaderboard.New(),
		limiter: ratelimit.New(settings.RateLimit, time.Minute),
		whipper: whip.New(settings.Cooldown, settings.Interrupt),
		overlay: sse.NewBroker(),
		agents:  sse.NewBroker(),
		command: settings.Command,
	}

	failures := make(chan error, 2)

	if settings.TwitchChannel == "" {
		slog.Warn("TWITCH_CHANNEL is not set, Twitch chat is disabled, use POST /api/whip to test")
	} else {
		go func() {
			err := twitch.Listen(ctx, settings.TwitchChannel, func(message twitch.Message) {
				if isCommand(message.Text, settings.Command) {
					application.whip(message.UserID, message.UserName)
				}
			})
			if err != nil {
				err = fmt.Errorf("listen to twitch chat: %w", err)
			}
			failures <- err
		}()
	}

	routes, err := application.routes()
	if err != nil {
		return fmt.Errorf("set up routes: %w", err)
	}

	server := &http.Server{
		Addr:              settings.HTTPAddr,
		Handler:           routes,
		ReadHeaderTimeout: 5 * time.Second,
		// Cancel long-lived SSE requests on shutdown.
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		slog.Info("server started", "addr", "http://"+settings.HTTPAddr)
		failures <- fmt.Errorf("serve http: %w", server.ListenAndServe())
	}()

	select {
	case <-ctx.Done():
	case err = <-failures:
	}
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if shutdownErr := server.Shutdown(shutdownCtx); shutdownErr != nil {
		err = errors.Join(err, fmt.Errorf("shut down server: %w", shutdownErr))
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

type app struct {
	board   *leaderboard.Leaderboard
	limiter *ratelimit.Limiter
	whipper *whip.Whipper
	// overlay receives every whip, agents only those passing the cooldown.
	overlay *sse.Broker
	agents  *sse.Broker
	// command is the chat command, shown on the overlay as a hint.
	command string
}

// overlayState is sent to the overlay on connect.
type overlayState struct {
	Command     string               `json:"command"`
	Leaderboard leaderboard.Snapshot `json:"leaderboard"`
}

type whipEvent struct {
	Name        string               `json:"name"`
	Count       int                  `json:"count"`
	Leaderboard leaderboard.Snapshot `json:"leaderboard"`
}

// whip handles a whip from a viewer and reports whether it was accepted.
func (a *app) whip(userID, name string) bool {
	if !a.limiter.Allow(userID) {
		slog.Info("whip rate limited", "user", name)
		return false
	}

	count := a.board.Inc(userID, name)
	a.overlay.Publish(sse.Event{
		Name: "whip",
		Data: whipEvent{
			Name:        name,
			Count:       count,
			Leaderboard: a.board.Snapshot(leaderboardSize),
		},
	})

	agentWhip, ok := a.whipper.Whip(name)
	agents := 0
	if ok {
		agents = a.agents.Publish(sse.Event{
			Name: "whip",
			Data: agentWhip,
		})
	}
	slog.Info("whip", "user", name, "count", count, "cooldown", !ok, "agents", agents)

	return true
}

func (a *app) routes() (http.Handler, error) {
	router := http.NewServeMux()

	router.Handle("GET /events", a.overlay.HandlerWithInitial(func() sse.Event {
		return sse.Event{
			Name: "state",
			Data: overlayState{
				Command:     a.command,
				Leaderboard: a.board.Snapshot(leaderboardSize),
			},
		}
	}))
	// Whips for Claude Code sessions run through cmd/whipped.
	router.Handle("GET /agent/events", a.agents.Handler())

	// Lets the streamer test the setup without chat:
	// curl -X POST 'localhost:8080/api/whip?user=tester'
	router.HandleFunc("POST /api/whip", func(w http.ResponseWriter, r *http.Request) {
		name := cmp.Or(r.FormValue("user"), "streamer")
		if !a.whip("local:"+name, name) {
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})

	overlay, err := web.Overlay()
	if err != nil {
		return nil, fmt.Errorf("open overlay: %w", err)
	}
	router.Handle("GET /", http.FileServerFS(overlay))

	return router, nil
}

// isCommand reports whether text starts with the chat command.
func isCommand(text, command string) bool {
	first, _, _ := strings.Cut(strings.TrimSpace(text), " ")
	return strings.EqualFold(first, command)
}
