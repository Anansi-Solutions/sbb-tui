package ui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/necrom4/sbb-tui/model"
)

func TestCapitalise(t *testing.T) {
	tests := []struct{ in, want string }{
		{"hello", "Hello"},
		{"Hello", "Hello"},
		{"", ""},
		{"123", "123"},
	}
	for _, tt := range tests {
		if got := capitalise(tt.in); got != tt.want {
			t.Errorf("capitalise(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestUserError(t *testing.T) {
	// Known sentinel errors are shown as-is (capitalised).
	if got := userError(errNoConnections); !strings.HasPrefix(got, "No connections") {
		t.Errorf("userError(known) = %q", got)
	}

	// Unknown errors get the generic prefix.
	got := userError(errors.New("boom"))
	if !strings.Contains(got, "Something went wrong") || !strings.Contains(got, "boom") {
		t.Errorf("userError(unknown) = %q", got)
	}

	// Wrapped sentinels are still recognised (no generic prefix).
	wrapped := userError(fmt.Errorf("context: %w", errMissingDeparture))
	if strings.Contains(wrapped, "Something went wrong") {
		t.Errorf("userError(wrapped) = %q, want sentinel handling", wrapped)
	}
}

func TestViewTerminalTooSmall(t *testing.T) {
	m := newTestModel()
	m.width, m.height = 10, 5

	out := m.View()
	if !strings.Contains(out, "Terminal too small") {
		t.Errorf("View() = %q, want size warning", out)
	}
}

func TestViewRendersResults(t *testing.T) {
	m := newTestModel()
	m.connections = []model.Connection{fixtureConnection()}
	m.searched = true

	out := m.View()
	for _, want := range []string{"S13", "Einsiedeln", "12:24"} {
		if !strings.Contains(out, want) {
			t.Errorf("View() missing %q", want)
		}
	}
}

func TestViewStartScreen(t *testing.T) {
	m := newTestModel()
	out := m.View()
	if !strings.Contains(out, "Enter stations above to see timetables") {
		t.Error("View() missing start screen tagline")
	}
}
