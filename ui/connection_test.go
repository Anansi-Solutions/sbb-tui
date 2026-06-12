package ui

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/necrom4/sbb-tui/model"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{900, "15 min"},
		{59, "0 min"},
		{3600, "1 h 00 min"},
		{4560, "1 h 16 min"},
		{7500, "2 h 05 min"},
		{0, "0 min"},
	}
	for _, tt := range tests {
		if got := formatDuration(tt.seconds); got != tt.want {
			t.Errorf("formatDuration(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		in     string
		maxLen int
		want   string
	}{
		{"Zurich HB", 0, ""},
		{"Zurich HB", 3, "Zur"},
		{"Bern", 10, "Bern"},
		{"Wadenswil Gare Centrale", 10, "Wadensw..."},
	}
	for _, tt := range tests {
		if got := truncateString(tt.in, tt.maxLen); got != tt.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.in, tt.maxLen, got, tt.want)
		}
	}
}

func TestWalkMinutes(t *testing.T) {
	withRunningTime := model.Leg{RunningTime: 420}
	if got := walkMinutes(withRunningTime); got != 7 {
		t.Errorf("walkMinutes(runningtime) = %d, want 7", got)
	}

	viaTimes := model.Leg{
		Departure: ts(12, 0),
		Exit:      &model.Exit{Arrival: ts(12, 4)},
	}
	if got := walkMinutes(viaTimes); got != 4 {
		t.Errorf("walkMinutes(times) = %d, want 4", got)
	}

	if got := walkMinutes(model.Leg{}); got != 0 {
		t.Errorf("walkMinutes(empty) = %d, want 0", got)
	}
}

func TestGoogleMapsURL(t *testing.T) {
	valid := model.Leg{
		Lat: 47.37, Lon: 8.54,
		Exit: &model.Exit{Lat: 47.38, Lon: 8.53},
	}
	url := googleMapsURL(valid)
	if !strings.Contains(url, "origin=47.37") || !strings.Contains(url, "destination=47.38") {
		t.Errorf("url = %q, want origin/destination coords", url)
	}

	if got := googleMapsURL(model.Leg{Lat: 47.37, Lon: 8.54}); got != "" {
		t.Errorf("url without exit = %q, want empty", got)
	}

	noCoords := model.Leg{Lat: 47.37, Lon: 8.54, Exit: &model.Exit{}}
	if got := googleMapsURL(noCoords); got != "" {
		t.Errorf("url with zero exit coords = %q, want empty", got)
	}
}

func TestFormatDelay(t *testing.T) {
	m := newTestModel()
	if got := m.formatDelay(0); got != "" {
		t.Errorf("formatDelay(0) = %q, want empty", got)
	}
	if got := m.formatDelay(3); got != " +3'" {
		t.Errorf("formatDelay(3) = %q, want \" +3'\"", got)
	}
}

func TestRenderOccupancy(t *testing.T) {
	m := newTestModel()
	person := m.icons.person

	if got := m.renderOccupancy(""); got != "" {
		t.Errorf("renderOccupancy(unknown) = %q, want empty", got)
	}

	both := m.renderOccupancy("12")
	if !strings.Contains(both, "1.") || !strings.Contains(both, "2.") {
		t.Errorf("renderOccupancy(12) = %q, want both class labels", both)
	}
	if got := strings.Count(both, person); got != 6 {
		t.Errorf("renderOccupancy(12) has %d person icons, want 6", got)
	}

	secondOnly := m.renderOccupancy("03")
	if strings.Contains(secondOnly, "1.") || !strings.Contains(secondOnly, "2.") {
		t.Errorf("renderOccupancy(03) = %q, want second class only", secondOnly)
	}
	if got := strings.Count(secondOnly, person); got != 3 {
		t.Errorf("renderOccupancy(03) has %d person icons, want 3", got)
	}
}

func TestRenderStopsLine(t *testing.T) {
	m := newTestModel()
	c := fixtureConnection()

	line := m.renderStopsLine(c, 40)
	if got := strings.Count(line, m.icons.filledDot); got != 2 {
		t.Errorf("stops line has %d filled dots, want 2: %q", got, line)
	}
	if got := strings.Count(line, m.icons.hollowDot); got != 1 {
		t.Errorf("stops line has %d hollow dots, want 1 (one transfer): %q", got, line)
	}

	empty := m.renderStopsLine(model.Connection{}, 40)
	if !strings.Contains(empty, m.icons.filledDot) {
		t.Errorf("empty connection stops line = %q", empty)
	}
}

func TestRenderSimpleConnection(t *testing.T) {
	m := newTestModel()
	c := fixtureConnection()

	out := m.renderSimpleConnection(c, 0, 60)

	for _, want := range []string{
		"S13",        // line badge of the first vehicle
		"SOB-sob",    // operator
		"Wädenswil",  // terminal
		"12:24",      // departure of first vehicle
		"13:40",      // connection arrival
		"+1'",        // departure delay
		"1 h 16 min", // duration
		"Pl. 1!",     // platform with track change marker
	} {
		if !strings.Contains(out, want) {
			t.Errorf("simple connection missing %q:\n%s", want, out)
		}
	}
}

func TestRenderGondolaConnection(t *testing.T) {
	m := newTestModel()
	c := fixtureGondolaConnection()

	out := m.renderSimpleConnection(c, 0, 60)
	if strings.Contains(out, "unavailable") {
		t.Fatalf("gondola connection rendered as malformed:\n%s", out)
	}
	for _, want := range []string{"GB 2042", "TCSA", "Vounetse", "17:13", "17:28", "15 min"} {
		if !strings.Contains(out, want) {
			t.Errorf("gondola card missing %q:\n%s", want, out)
		}
	}

	detail := strings.Join(m.buildDetailLines(c, 80), "\n")
	for _, want := range []string{"GB 2042", "Charmey", "→ Vounetse"} {
		if !strings.Contains(detail, want) {
			t.Errorf("gondola detail missing %q:\n%s", want, detail)
		}
	}
}

func TestRenderSimpleConnectionMalformed(t *testing.T) {
	m := newTestModel()
	c := model.Connection{Legs: []model.Leg{{Type: "walk"}}}

	out := m.renderSimpleConnection(c, 0, 60)
	if !strings.Contains(out, "Connection details unavailable") {
		t.Errorf("malformed connection should show user error, got:\n%s", out)
	}
}

func TestBuildDetailLines(t *testing.T) {
	m := newTestModel()
	c := fixtureConnection()

	lines := m.buildDetailLines(c, 80)
	all := strings.Join(lines, "\n")

	for _, want := range []string{
		"Einsiedeln",           // departure stop
		"Wädenswil",            // exit of first leg
		"S13",                  // line badge
		"B 33",                 // bus with category prefix
		"VBZ F",                // normalized operator
		"→ Zürich, Werdhölzli", // direction
		"12:24", "12:50",       // first leg times
		m.icons.walk,         // walk leg present
		"travelmode=walking", // walk Maps link
		"+1'", "+2'",         // both delays
	} {
		if !strings.Contains(all, want) {
			t.Errorf("detail view missing %q:\n%s", want, all)
		}
	}

	// The bare arrival node must not be rendered as its own leg: its
	// station name appears exactly once (as the bus leg's exit).
	if got := strings.Count(all, "Zürich, Förrlibuckstr. 60"); got != 1 {
		t.Errorf("arrival station appears %d times, want 1:\n%s", got, all)
	}
}

func TestBuildDetailLinesDisruptions(t *testing.T) {
	m := newTestModel()
	c := fixtureConnection()
	c.Disruptions = json.RawMessage(`{
		"d1": {"id": "d1", "texts": {"S": {"summary": "Information", "consequence": "Expect delays"}}},
		"d2": {"id": "d2", "texts": {"S": {"summary": "Information", "consequence": "Expect delays"}}}
	}`)

	lines := m.buildDetailLines(c, 80)
	all := strings.Join(lines, "\n")

	if !strings.Contains(all, m.icons.warning+" Information") {
		t.Errorf("detail view missing disruption banner:\n%s", all)
	}
	// Identical disruption texts are deduplicated.
	if got := strings.Count(all, "Expect delays"); got != 1 {
		t.Errorf("disruption consequence appears %d times, want 1 (deduplicated)", got)
	}
}

func TestRenderFullConnectionFits(t *testing.T) {
	m := newTestModel()
	c := fixtureConnection()

	out := m.renderFullConnection(c, 60)
	if !strings.Contains(out, "Einsiedeln") {
		t.Errorf("full connection render missing content:\n%s", out)
	}
}
