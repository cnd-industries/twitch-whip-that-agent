package whip

import (
	"strings"
	"testing"
	"time"
)

func TestWhipperCooldown(t *testing.T) {
	whipper := New(50*time.Millisecond, true)

	got, ok := whipper.Whip("alice")
	if !ok || !got.Interrupt || !strings.Contains(got.Message, "alice") {
		t.Fatalf("Whip() = %+v, %v; want interrupting message for alice", got, ok)
	}
	if _, ok := whipper.Whip("bob"); ok {
		t.Fatal("whip during cooldown must be rejected")
	}

	time.Sleep(60 * time.Millisecond)

	if _, ok := whipper.Whip("bob"); !ok {
		t.Fatal("whip after cooldown must be accepted")
	}
}
