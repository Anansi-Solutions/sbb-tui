package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/necrom4/sbb-tui/config"
)

func TestNormalizeHex(t *testing.T) {
	tests := []struct{ in, want string }{
		{"f00", "#f00"},
		{"8e224d", "#8e224d"},
		{"#fff", "#fff"},
		{" 039 ", "#039"},
		{"ABC", "#ABC"},
		{"", ""},
		{"ggg", ""},
		{"12345", ""},
		{"f0", ""},
	}
	for _, tt := range tests {
		if got := normalizeHex(tt.in); got != tt.want {
			t.Errorf("normalizeHex(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLineStyle(t *testing.T) {
	s := newStyles(config.DefaultTheme())

	colored := s.lineStyle("fff", "039")
	if got := colored.GetBackground(); got != lipgloss.Color("#039") {
		t.Errorf("background = %v, want #039", got)
	}
	if got := colored.GetForeground(); got != lipgloss.Color("#fff") {
		t.Errorf("foreground = %v, want #fff", got)
	}

	// Without API colors the theme badge style applies unchanged.
	fallback := s.lineStyle("", "")
	if got := fallback.GetBackground(); got != s.vehicleModel.GetBackground() {
		t.Errorf("fallback background = %v, want theme badge bg", got)
	}

	// Invalid colors are ignored rather than producing a broken style.
	invalid := s.lineStyle("zzz", "12345")
	if got := invalid.GetBackground(); got != s.vehicleModel.GetBackground() {
		t.Errorf("invalid background = %v, want theme badge bg", got)
	}
}

func TestThemeColorSentinels(t *testing.T) {
	if _, ok := themeColor("white").(lipgloss.AdaptiveColor); !ok {
		t.Error("themeColor(white) should be adaptive")
	}
	if _, ok := themeColor("black").(lipgloss.AdaptiveColor); !ok {
		t.Error("themeColor(black) should be adaptive")
	}
	if got := themeColor("#123456"); got != lipgloss.Color("#123456") {
		t.Errorf("themeColor(hex) = %v", got)
	}
}
