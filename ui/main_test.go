package ui

import (
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/necrom4/sbb-tui/config"
	"github.com/necrom4/sbb-tui/model"
)

// TestMain forces plain-text rendering so assertions don't have to
// account for ANSI escape sequences.
func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.Ascii)
	os.Exit(m.Run())
}

// newTestModel returns an appModel with a fixed size and the Unicode
// (non Nerd Font) icon set, which is easier to assert against.
func newTestModel() appModel {
	m := NewModel(config.Config{Theme: config.DefaultTheme()})
	m.width = 120
	m.height = 40
	return m
}

// ts builds a model.Timestamp on the fixture day.
func ts(hour, minute int) model.Timestamp {
	return model.Timestamp{Time: time.Date(2026, 6, 12, hour, minute, 0, 0, model.SwissLocation)}
}

// fixtureConnection returns a connection mirroring a real API response:
// S13 with delays and a platform change, a walk, a tram and the bare
// arrival node.
func fixtureConnection() model.Connection {
	return model.Connection{
		From:      "Einsiedeln",
		To:        "Zürich, Förrlibuckstr. 60",
		Departure: ts(12, 24),
		Arrival:   ts(13, 40),
		Duration:  4560,
		Occupancy: "12",
		Legs: []model.Leg{
			{
				Name:        "Einsiedeln",
				Departure:   ts(12, 24),
				DepDelay:    model.Delay{Minutes: 1},
				Type:        "strain",
				Line:        "S13",
				Category:    "S",
				LineNumber:  "13",
				Terminal:    "Wädenswil",
				Operator:    "SOB-sob",
				FgColor:     "fff",
				BgColor:     "039",
				Track:       "1!",
				Occupancy:   "11",
				RunningTime: 1560,
				Lat:         47.128578,
				Lon:         8.744481,
				Exit: &model.Exit{
					Name:     "Wädenswil",
					Arrival:  ts(12, 50),
					ArrDelay: model.Delay{Minutes: 2},
					Track:    "1",
					Lat:      47.229307,
					Lon:      8.675205,
				},
			},
			{
				Name:        "Wädenswil",
				Departure:   ts(12, 55),
				Type:        "walk",
				RunningTime: 240,
				Lat:         47.229307,
				Lon:         8.675205,
				Exit: &model.Exit{
					Name:    "Wädenswil, Bahnhof",
					Arrival: ts(12, 59),
					Lat:     47.23,
					Lon:     8.676,
				},
			},
			{
				Name:       "Wädenswil, Bahnhof",
				Departure:  ts(13, 2),
				Type:       "bus",
				Line:       "33",
				Category:   "B",
				LineNumber: "33",
				Terminal:   "Zürich, Werdhölzli",
				Operator:   "VBZ    F",
				Exit: &model.Exit{
					Name:    "Zürich, Förrlibuckstr. 60",
					Arrival: ts(13, 40),
				},
			},
			{
				Name:    "Zürich, Förrlibuckstr. 60",
				Arrival: ts(13, 40),
			},
		},
	}
}
