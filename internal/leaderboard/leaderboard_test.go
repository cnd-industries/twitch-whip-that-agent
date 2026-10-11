package leaderboard

import (
	"slices"
	"testing"
)

func TestSnapshot(t *testing.T) {
	board := New()
	board.Inc("1", "bob")
	board.Inc("2", "alice")
	board.Inc("3", "carol")
	if got := board.Inc("3", "Carol"); got != 2 {
		t.Fatalf("Inc() = %d, want 2", got)
	}

	got := board.Snapshot(2)
	want := []Entry{{"Carol", 2}, {"alice", 1}}
	if !slices.Equal(got.Top, want) {
		t.Fatalf("Snapshot(2).Top = %v, want %v", got.Top, want)
	}
	if got.Total != 4 {
		t.Fatalf("Snapshot(2).Total = %d, want 4", got.Total)
	}
	if got := len(board.Snapshot(10).Top); got != 3 {
		t.Fatalf("len(Snapshot(10).Top) = %d, want 3", got)
	}
}
