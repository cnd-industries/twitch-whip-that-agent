// Package leaderboard keeps per-session whip counts in memory.
package leaderboard

import (
	"cmp"
	"slices"
	"sync"
)

type Entry struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Snapshot is the leaderboard state shown on the overlay.
type Snapshot struct {
	// Total is the number of whips by all users, not only the top.
	Total int     `json:"total"`
	Top   []Entry `json:"top"`
}

type Leaderboard struct {
	mu      sync.Mutex
	entries map[string]*Entry
	total   int
}

func New() *Leaderboard {
	return &Leaderboard{
		entries: make(map[string]*Entry),
	}
}

// Inc adds a whip for the user and returns their new total.
func (l *Leaderboard) Inc(userID, name string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.entries[userID]
	if !ok {
		entry = &Entry{}
		l.entries[userID] = entry
	}
	// Keep the latest display name in case the user renamed.
	entry.Name = name
	entry.Count++
	l.total++

	return entry.Count
}

// Snapshot returns the total and up to limit entries ordered by count
// descending.
func (l *Leaderboard) Snapshot(limit int) Snapshot {
	l.mu.Lock()
	top := make([]Entry, 0, len(l.entries))
	for _, entry := range l.entries {
		top = append(top, *entry)
	}
	total := l.total
	l.mu.Unlock()

	slices.SortFunc(top, func(first, second Entry) int {
		return cmp.Or(cmp.Compare(second.Count, first.Count), cmp.Compare(first.Name, second.Name))
	})

	return Snapshot{
		Total: total,
		Top:   top[:min(limit, len(top))],
	}
}
