package ui

import "testing"

func TestVehicleForNerdFont(t *testing.T) {
	ic := newIconSet(true)

	// Every known type resolves to its mapped glyph, never the generic one.
	for legType, want := range ic.vehicleByType {
		if got := ic.vehicleFor(legType); got != want {
			t.Errorf("vehicleFor(%q) = %q, want mapped %q", legType, got, want)
		}
	}

	// Trains and trams must be visually distinct from each other.
	if ic.vehicleFor("express_train") == ic.vehicleFor("tram") {
		t.Error("train and tram share the same glyph")
	}
	if ic.vehicleFor("bus") == ic.vehicleFor("ship") {
		t.Error("bus and ship share the same glyph")
	}

	// Unknown subtype falls back by substring.
	if got := ic.vehicleFor("trolley_bus"); got != ic.vehicleFor("bus") {
		t.Errorf("vehicleFor(trolley_bus) = %q, want bus glyph", got)
	}
	if got := ic.vehicleFor("special_train"); got != ic.vehicleFor("train") {
		t.Errorf("vehicleFor(special_train) = %q, want train glyph", got)
	}

	// Fully unknown types use the generic vehicle glyph.
	if got := ic.vehicleFor("hovercraft"); got != ic.vehicle {
		t.Errorf("vehicleFor(hovercraft) = %q, want generic", got)
	}
	if got := ic.vehicleFor(""); got != ic.vehicle {
		t.Errorf("vehicleFor(empty) = %q, want generic", got)
	}
}

func TestVehicleForUnicodeFallback(t *testing.T) {
	ic := newIconSet(false)
	for _, legType := range []string{"train", "tram", "bus", "ship", "unknown", ""} {
		if got := ic.vehicleFor(legType); got != ic.vehicle {
			t.Errorf("vehicleFor(%q) = %q, want generic %q", legType, got, ic.vehicle)
		}
	}
}

func TestPlatformLabel(t *testing.T) {
	ic := newIconSet(false)

	tests := []struct{ platform, want string }{
		{"13", ic.platform},
		{"13!", ic.platform},
		{"A", ic.stop},
		{"B2", ic.stop},
		{"", ic.platform},
	}
	for _, tt := range tests {
		if got := ic.platformLabel(tt.platform); got != tt.want {
			t.Errorf("platformLabel(%q) = %q, want %q", tt.platform, got, tt.want)
		}
	}
}

func TestIconSetCompleteness(t *testing.T) {
	for _, nerd := range []bool{true, false} {
		ic := newIconSet(nerd)
		glyphs := map[string]string{
			"arrival":   ic.arrival,
			"departure": ic.departure,
			"search":    ic.search,
			"swap":      ic.swap,
			"vehicle":   ic.vehicle,
			"walk":      ic.walk,
			"prompt":    ic.prompt,
			"person":    ic.person,
			"warning":   ic.warning,
			"towards":   ic.towards,
			"filledDot": ic.filledDot,
			"hollowDot": ic.hollowDot,
			"horizLine": ic.horizLine,
			"vertLine":  ic.vertLine,
		}
		for name, g := range glyphs {
			if g == "" {
				t.Errorf("nerdFont=%v: icon %s is empty", nerd, name)
			}
		}
	}
}
