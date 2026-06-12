package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimestampUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		wantZero bool
		wantErr  bool
	}{
		{"valid", `"2026-06-12 12:35:00"`, false, false},
		{"null", `null`, true, false},
		{"empty", `""`, true, false},
		{"wrong format", `"12.06.2026"`, true, true},
		{"old API format", `"2026-06-12T12:35:00+0200"`, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts Timestamp
			err := json.Unmarshal([]byte(tt.in), &ts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if ts.IsZero() != tt.wantZero {
				t.Fatalf("IsZero() = %v, want %v", ts.IsZero(), tt.wantZero)
			}
		})
	}
}

func TestTimestampSwissLocal(t *testing.T) {
	var ts Timestamp
	if err := json.Unmarshal([]byte(`"2026-06-12 12:35:00"`), &ts); err != nil {
		t.Fatal(err)
	}

	want := time.Date(2026, 6, 12, 12, 35, 0, 0, SwissLocation)
	if !ts.Equal(want) {
		t.Fatalf("parsed %v, want %v", ts.Time, want)
	}
	if got := ts.Local().Format("15:04"); got != "12:35" {
		t.Fatalf("Local() = %s, want 12:35", got)
	}
}

func TestTimestampSub(t *testing.T) {
	a := Timestamp{Time: time.Date(2026, 6, 12, 12, 0, 0, 0, SwissLocation)}
	b := Timestamp{Time: time.Date(2026, 6, 12, 12, 45, 0, 0, SwissLocation)}
	if got := b.Sub(a); got != 45*time.Minute {
		t.Fatalf("Sub() = %v, want 45m", got)
	}
}

func TestDelayUnmarshalJSON(t *testing.T) {
	tests := []struct {
		in            string
		wantMinutes   int
		wantCancelled bool
	}{
		{`"+0"`, 0, false},
		{`"+7"`, 7, false},
		{`"-2"`, -2, false},
		{`"X"`, 0, true},
		{`"x"`, 0, true},
		{`""`, 0, false},
		{`null`, 0, false},
		{`"garbage"`, 0, false}, // tolerated, not an error
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			var d Delay
			if err := json.Unmarshal([]byte(tt.in), &d); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Minutes != tt.wantMinutes || d.Cancelled != tt.wantCancelled {
				t.Fatalf("got {%d %v}, want {%d %v}", d.Minutes, d.Cancelled, tt.wantMinutes, tt.wantCancelled)
			}
		})
	}
}

func TestOccupancy(t *testing.T) {
	tests := []struct {
		in            Occupancy
		first, second int
	}{
		{"", 0, 0},
		{"11", 1, 1},
		{"23", 2, 3},
		{"0", 0, 0},
		{"3", 3, 0},
		{"ab", 0, 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.in), func(t *testing.T) {
			if got := tt.in.FirstClass(); got != tt.first {
				t.Errorf("FirstClass() = %d, want %d", got, tt.first)
			}
			if got := tt.in.SecondClass(); got != tt.second {
				t.Errorf("SecondClass() = %d, want %d", got, tt.second)
			}
		})
	}
}

func TestLegKindHelpers(t *testing.T) {
	walk := Leg{Type: "walk"}
	if !walk.IsWalk() || walk.IsVehicle() {
		t.Error("walk leg misclassified")
	}

	vehicle := Leg{Type: "strain", Line: "S13"}
	if vehicle.IsWalk() || !vehicle.IsVehicle() {
		t.Error("vehicle leg misclassified")
	}

	arrivalNode := Leg{Name: "St. Gallen"}
	if arrivalNode.IsWalk() || arrivalNode.IsVehicle() {
		t.Error("bare arrival node misclassified")
	}
}

func TestDisplayLine(t *testing.T) {
	tests := []struct {
		name string
		leg  Leg
		want string
	}{
		{"bus gets B prefix", Leg{Line: "33", Category: "B", LineNumber: "33"}, "B 33"},
		{"tram gets T prefix", Leg{Line: "51", Category: "T", LineNumber: "51"}, "T 51"},
		{"ship gets BAT prefix", Leg{Line: "3600", Category: "BAT", LineNumber: "3600"}, "BAT 3600"},
		{"sbahn stays combined", Leg{Line: "S13", Category: "S", LineNumber: "13"}, "S13"},
		{"ic stays combined", Leg{Line: "IC 1", Category: "IC", LineNumber: "1"}, "IC 1"},
		{"no category", Leg{Line: "17", LineNumber: "17"}, "17"},
		{"empty", Leg{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.leg.DisplayLine(); got != tt.want {
				t.Fatalf("DisplayLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOperatorName(t *testing.T) {
	tests := []struct{ in, want string }{
		{"VBZ    F", "VBZ F"},
		{"SBB", "SBB"},
		{"  SOB-sob ", "SOB-sob"},
		{"", ""},
	}
	for _, tt := range tests {
		leg := Leg{Operator: tt.in}
		if got := leg.OperatorName(); got != tt.want {
			t.Errorf("OperatorName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestVehicleLegIndexes(t *testing.T) {
	c := Connection{Legs: []Leg{
		{Type: "walk"},
		{Type: "strain", Line: "S13"},
		{Type: "walk"},
		{Type: "tram", Line: "17"},
		{Name: "arrival node"},
	}}

	if got := c.FirstVehicleLeg(); got != 1 {
		t.Errorf("FirstVehicleLeg() = %d, want 1", got)
	}
	if got := c.LastVehicleLeg(); got != 3 {
		t.Errorf("LastVehicleLeg() = %d, want 3", got)
	}

	empty := Connection{Legs: []Leg{{Type: "walk"}}}
	if got := empty.FirstVehicleLeg(); got != -1 {
		t.Errorf("FirstVehicleLeg() = %d, want -1", got)
	}
	if got := empty.LastVehicleLeg(); got != -1 {
		t.Errorf("LastVehicleLeg() = %d, want -1", got)
	}
}

func TestDisruptionList(t *testing.T) {
	t.Run("object keyed by id", func(t *testing.T) {
		c := Connection{Disruptions: json.RawMessage(`{
			"id-b": {"id": "id-b", "texts": {"M": {"summary": "M summary"}}},
			"id-a": {"id": "id-a", "texts": {"S": {"summary": "Info", "consequence": "Expect delays"}}}
		}`)}

		list := c.DisruptionList()
		if len(list) != 2 {
			t.Fatalf("len = %d, want 2", len(list))
		}
		// Sorted by ID for stable output.
		if list[0].ID != "id-a" || list[1].ID != "id-b" {
			t.Fatalf("order = %s, %s; want id-a, id-b", list[0].ID, list[1].ID)
		}
		if got := list[0].Text().Consequence; got != "Expect delays" {
			t.Fatalf("Text().Consequence = %q", got)
		}
		// Falls back to M when S is missing.
		if got := list[1].Text().Summary; got != "M summary" {
			t.Fatalf("Text().Summary = %q", got)
		}
	})

	t.Run("legacy empty array", func(t *testing.T) {
		c := Connection{Disruptions: json.RawMessage(`[]`)}
		if got := c.DisruptionList(); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("absent", func(t *testing.T) {
		var c Connection
		if got := c.DisruptionList(); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("null", func(t *testing.T) {
		c := Connection{Disruptions: json.RawMessage(`null`)}
		if got := c.DisruptionList(); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})
}

func TestDisruptionTextFallback(t *testing.T) {
	d := Disruption{Texts: map[string]DisruptionTexts{
		"L": {Summary: "long"},
	}}
	if got := d.Text().Summary; got != "long" {
		t.Fatalf("Text() = %q, want long", got)
	}

	if got := (Disruption{}).Text(); got != (DisruptionTexts{}) {
		t.Fatalf("empty disruption Text() = %+v, want zero", got)
	}
}

// routeFixture is a trimmed-down real response from route.json.
const routeFixture = `{
	"duration": 4560,
	"from": "Einsiedeln",
	"departure": "2026-06-12 12:24:00",
	"dep_delay": "+1",
	"to": "Z\u00fcrich, F\u00f6rrlibuckstr. 60",
	"arrival": "2026-06-12 13:40:00",
	"occupancy": "12",
	"disruptions": {"d1": {"id": "d1", "texts": {"S": {"summary": "Information", "consequence": "Expect delays"}}}},
	"legs": [
		{
			"departure": "2026-06-12 12:24:00",
			"tripid": "T1",
			"stopid": "8503283",
			"name": "Einsiedeln",
			"sbb_name": "Einsiedeln",
			"type": "strain",
			"line": "S13",
			"terminal": "W\u00e4denswil",
			"fgcolor": "fff",
			"bgcolor": "039",
			"*G": "S",
			"*L": "13",
			"operator": "SOB-sob",
			"stops": [
				{"arrival": "2026-06-12 12:31:00", "departure": "2026-06-12 12:32:00", "dep_delay": "+1", "occupancy": "11", "name": "Biberbrugg", "stopid": "8503284", "lon": 8.72, "lat": 47.15}
			],
			"runningtime": 1560,
			"exit": {
				"arrival": "2026-06-12 12:50:00",
				"stopid": "8503206",
				"name": "W\u00e4denswil",
				"sbb_name": "W\u00e4denswil",
				"waittime": 600,
				"track": "1",
				"arr_delay": "+2",
				"lon": 8.675205,
				"lat": 47.229307
			},
			"occupancy": "11",
			"dep_delay": "+1",
			"track": "1!",
			"type_name": "S-Bahn",
			"lon": 8.744481,
			"lat": 47.128578
		},
		{
			"arrival": "2026-06-12 13:36:00",
			"departure": "2026-06-12 13:36:00",
			"type": "walk",
			"name": "Z\u00fcrich, F\u00f6rrlibuckstrasse",
			"runningtime": 240,
			"exit": {"arrival": "2026-06-12 13:40:00", "isaddress": true, "name": "Z\u00fcrich, F\u00f6rrlibuckstr. 60"}
		},
		{"arrival": "2026-06-12 13:40:00", "isaddress": true, "name": "Z\u00fcrich, F\u00f6rrlibuckstr. 60"}
	]
}`

func TestConnectionDecode(t *testing.T) {
	var c Connection
	if err := json.Unmarshal([]byte(routeFixture), &c); err != nil {
		t.Fatal(err)
	}

	if c.From != "Einsiedeln" || c.Duration != 4560 {
		t.Fatalf("connection header decoded wrong: %+v", c)
	}
	if c.DepDelay.Minutes != 1 {
		t.Errorf("DepDelay = %d, want 1", c.DepDelay.Minutes)
	}
	if c.Occupancy.SecondClass() != 2 {
		t.Errorf("Occupancy.SecondClass() = %d, want 2", c.Occupancy.SecondClass())
	}
	if len(c.Legs) != 3 {
		t.Fatalf("legs = %d, want 3", len(c.Legs))
	}

	leg := c.Legs[0]
	if !leg.IsVehicle() || leg.Line != "S13" || leg.Category != "S" || leg.LineNumber != "13" {
		t.Errorf("vehicle leg decoded wrong: %+v", leg)
	}
	if leg.Track != "1!" {
		t.Errorf("Track = %q, want 1! (platform change marker)", leg.Track)
	}
	if leg.BgColor != "039" || leg.FgColor != "fff" {
		t.Errorf("colors = %s/%s, want 039/fff", leg.BgColor, leg.FgColor)
	}
	if leg.Exit == nil || leg.Exit.Name != "Wädenswil" || leg.Exit.ArrDelay.Minutes != 2 {
		t.Errorf("exit decoded wrong: %+v", leg.Exit)
	}
	if len(leg.Stops) != 1 || leg.Stops[0].Name != "Biberbrugg" || leg.Stops[0].Occupancy != "11" {
		t.Errorf("stops decoded wrong: %+v", leg.Stops)
	}

	if !c.Legs[1].IsWalk() || c.Legs[1].RunningTime != 240 {
		t.Errorf("walk leg decoded wrong: %+v", c.Legs[1])
	}
	if c.Legs[1].Exit == nil || !c.Legs[1].Exit.IsAddress {
		t.Errorf("walk exit decoded wrong: %+v", c.Legs[1].Exit)
	}

	if c.Legs[2].IsVehicle() || c.Legs[2].IsWalk() {
		t.Errorf("arrival node misclassified: %+v", c.Legs[2])
	}

	if len(c.DisruptionList()) != 1 {
		t.Errorf("disruptions = %d, want 1", len(c.DisruptionList()))
	}
}
