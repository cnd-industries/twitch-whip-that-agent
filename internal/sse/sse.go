// Package sse streams events to clients with Server-Sent Events.
package sse

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Event is a named Server-Sent Event.
type Event struct {
	Name string
	Data any
}

// Broker fans out events to all connected clients.
type Broker struct {
	mu          sync.Mutex
	subscribers map[chan []byte]struct{}
}

func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[chan []byte]struct{}),
	}
}

// Publish sends the event to every subscriber, dropping it for those
// that are too slow to keep up, and returns the number of subscribers.
func (b *Broker) Publish(event Event) int {
	frame, err := encode(event)
	if err != nil {
		slog.Error("encode event", "err", err)
		return 0
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for subscriber := range b.subscribers {
		select {
		case subscriber <- frame:
		default:
		}
	}

	return len(b.subscribers)
}

func (b *Broker) subscribe() chan []byte {
	subscriber := make(chan []byte, 16)

	b.mu.Lock()
	b.subscribers[subscriber] = struct{}{}
	b.mu.Unlock()

	return subscriber
}

func (b *Broker) unsubscribe(subscriber chan []byte) {
	b.mu.Lock()
	delete(b.subscribers, subscriber)
	b.mu.Unlock()
}

// Handler streams events to the client.
func (b *Broker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b.stream(w, r, func() ([]byte, error) {
			// A comment flushes the headers so the client knows it is
			// subscribed.
			return []byte(": connected\n\n"), nil
		})
	}
}

// HandlerWithInitial streams events to the client like Handler, sending
// the initial event first so the client can render the current state.
func (b *Broker) HandlerWithInitial(initial func() Event) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b.stream(w, r, func() ([]byte, error) {
			return encode(initial())
		})
	}
}

// stream subscribes the client and sends it the greeting followed by
// published events until the client disconnects. The greeting is built
// after subscribing, so no event published in between is missed.
func (b *Broker) stream(w http.ResponseWriter, r *http.Request, greeting func() ([]byte, error)) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	subscriber := b.subscribe()
	defer b.unsubscribe(subscriber)

	frame, err := greeting()
	if err != nil {
		slog.Error("build greeting", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Write errors mean the client has disconnected, which ends the
	// stream as intended.
	client := http.NewResponseController(w)
	if err := send(w, client, frame); err != nil {
		return
	}

	// Comments keep proxies from closing an idle stream and detect
	// disconnected clients.
	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case frame := <-subscriber:
			if err := send(w, client, frame); err != nil {
				return
			}
		case <-ping.C:
			if err := send(w, client, []byte(": ping\n\n")); err != nil {
				return
			}
		}
	}
}

// send writes the frame and flushes it to the client right away.
func send(w http.ResponseWriter, client *http.ResponseController, frame []byte) error {
	if _, err := w.Write(frame); err != nil {
		return fmt.Errorf("write frame: %w", err)
	}
	if err := client.Flush(); err != nil {
		return fmt.Errorf("flush frame: %w", err)
	}

	return nil
}

func encode(event Event) ([]byte, error) {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return nil, fmt.Errorf("marshal %s event: %w", event.Name, err)
	}

	return fmt.Appendf(nil, "event: %s\ndata: %s\n\n", event.Name, data), nil
}
