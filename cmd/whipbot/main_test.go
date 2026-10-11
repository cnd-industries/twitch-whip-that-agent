package main

import "testing"

func TestIsCommand(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"!whip", true},
		{"  !WHIP faster please", true},
		{"!whipped", false},
		{"please !whip", false},
		{"", false},
	}
	for _, test := range tests {
		if got := isCommand(test.text, "!whip"); got != test.want {
			t.Errorf("isCommand(%q) = %v, want %v", test.text, got, test.want)
		}
	}
}
