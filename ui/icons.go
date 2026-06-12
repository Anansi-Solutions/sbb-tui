package ui

import "strings"

// iconSet groups every glyph the UI uses, so callers don't need to
// branch between Nerd Font and Unicode-fallback variants at use sites.
type iconSet struct {
	// vehicleByType maps the API's leg `type` to a glyph; empty in
	// Unicode-fallback mode so every type falls back to `vehicle`.
	vehicleByType map[string]string

	// Mode-dependent (Nerd Font vs Unicode fallback)
	arrival   string
	departure string
	platform  string
	stop      string
	search    string
	swap      string
	vehicle   string
	walk      string
	prompt    string
	person    string
	warning   string

	// Mode-invariant
	towards   string
	filledDot string
	hollowDot string
	horizLine string
	vertLine  string
	keyTab    string
	keyEnter  string
	keySpace  string
	keyUpDw   string
	keyUPDW   string
	keyRight  string
	keyEsc    string
}

// newIconSet returns the glyphs to use, swapping the mode-dependent ones based on nerdFont.
func newIconSet(nerdFont bool) iconSet {
	icons := iconSet{
		platform: "Pl.",
		stop:     "Stop",
		towards:  "→",

		filledDot: "●",
		hollowDot: "○",
		horizLine: "─",
		vertLine:  "│",

		keyTab:   "⇥",
		keyEnter: "↵",
		keySpace: "␣",
		keyUpDw:  "↕",
		keyUPDW:  "⇧↕",
		keyRight: "→",
		keyEsc:   "⎋",
	}

	if nerdFont {
		icons.vehicleByType = map[string]string{
			"train":         "\uf238",     //
			"express_train": "\uf238",     //
			"strain":        "\uf239",     //
			"metro":         "\uf239",     //
			"tram":          "\U000F052D", // 󰔭
			"bus":           "\uf207",     //
			"night_bus":     "\uf207",     //
			"post":          "\uf207",     //
			"ship":          "\uf21a",     //
			"cableway":      "\U000F0686", // 󰚆
			"gondola":       "\U000F0686", // 󰚆
			"funicular":     "\U000F0686", // 󰚆
		}
		icons.warning = "\uf071" //
		icons.person = "\uf007"  //
		icons.arrival = "󰗔"
		icons.departure = ""
		icons.search = ""
		icons.swap = ""
		icons.vehicle = ""
		icons.walk = ""
		icons.prompt = " "
	} else {
		icons.arrival = "⤙"
		icons.departure = "⤚"
		icons.search = "⌕"
		icons.swap = "↔"
		icons.vehicle = "◇"
		icons.walk = "walk:"
		icons.prompt = "⏵ "
		icons.person = "●"
		icons.warning = "⚠"
	}

	return icons
}

// vehicleFor returns the glyph for the API's leg `type`, falling back
// to the generic vehicle glyph for unknown types and in Unicode mode.
func (ic iconSet) vehicleFor(legType string) string {
	if icon, ok := ic.vehicleByType[legType]; ok {
		return icon
	}
	if len(ic.vehicleByType) > 0 {
		switch {
		case strings.Contains(legType, "bus"):
			return ic.vehicleByType["bus"]
		case strings.Contains(legType, "train"):
			return ic.vehicleByType["train"]
		}
	}
	return ic.vehicle
}

// platformLabel picks "Stop" for letter-prefixed platform strings, otherwise "Pl.".
func (ic iconSet) platformLabel(platform string) string {
	if len(platform) > 0 && platform[0] >= 'A' && platform[0] <= 'Z' {
		return ic.stop
	}
	return ic.platform
}
