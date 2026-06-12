package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/necrom4/sbb-tui/config"
	"github.com/necrom4/sbb-tui/model"
)

// ----- Pure input helpers -----

func TestCompleteDateAndTime(t *testing.T) {
	if got := completeTime("1"); got != "10:00" {
		t.Errorf("completeTime(1) = %q, want 10:00", got)
	}
	if got := completeTime("14:3"); got != "14:30" {
		t.Errorf("completeTime(14:3) = %q, want 14:30", got)
	}
	if got := completeTime("14:30"); got != "14:30" {
		t.Errorf("completeTime(full) = %q, want unchanged", got)
	}

	// completeDate pads with today's date, so only shape is asserted.
	if got := completeDate(""); len(got) != 10 {
		t.Errorf("completeDate(empty) = %q, want full DD.MM.YYYY", got)
	}
	if got := completeDate("01.02.2030"); got != "01.02.2030" {
		t.Errorf("completeDate(full) = %q, want unchanged", got)
	}
}

func TestDigitHelpers(t *testing.T) {
	if got := stripDelimiters("12.06.2026", '.'); got != "12062026" {
		t.Errorf("stripDelimiters = %q", got)
	}
	if got := countDigitsBefore("12.06", 4); got != 3 {
		t.Errorf("countDigitsBefore = %d, want 3", got)
	}
	if got := posOfDigit("12.06", 2); got != 3 {
		t.Errorf("posOfDigit = %d, want 3", got)
	}
	if got := posOfDigit("12", 5); got != 2 {
		t.Errorf("posOfDigit(out of range) = %d, want len", got)
	}
	if got := formatDate("12062026"); got != "12.06.2026" {
		t.Errorf("formatDate = %q", got)
	}
	if got := formatTime("1430"); got != "14:30" {
		t.Errorf("formatTime = %q", got)
	}
}

func TestValidateDateDigits(t *testing.T) {
	valid := []string{"", "1", "12", "120", "1206", "12062026", "3"}
	for _, d := range valid {
		if !validateDateDigits(d) {
			t.Errorf("validateDateDigits(%q) = false, want true", d)
		}
	}
	invalid := []string{"4", "32", "1213", "1200"}
	for _, d := range invalid {
		if validateDateDigits(d) {
			t.Errorf("validateDateDigits(%q) = true, want false", d)
		}
	}
}

func TestValidateTimeDigits(t *testing.T) {
	valid := []string{"", "1", "23", "235", "2359", "0000"}
	for _, d := range valid {
		if !validateTimeDigits(d) {
			t.Errorf("validateTimeDigits(%q) = false, want true", d)
		}
	}
	invalid := []string{"3", "24", "236"}
	for _, d := range invalid {
		if validateTimeDigits(d) {
			t.Errorf("validateTimeDigits(%q) = true, want false", d)
		}
	}
}

func TestToAPIDate(t *testing.T) {
	if got := toAPIDate("12.06.2026"); got != "2026-06-12" {
		t.Errorf("toAPIDate = %q, want 2026-06-12", got)
	}
	if got := toAPIDate("garbage"); got != "garbage" {
		t.Errorf("toAPIDate(garbage) = %q, want passthrough", got)
	}
}

func TestAdaptSuggestions(t *testing.T) {
	got := adaptSuggestions("zur", []string{"Zürich HB", "Bern"})
	if len(got) != 1 {
		t.Fatalf("adaptSuggestions = %v, want 1 match", got)
	}
	// The user's literal input is grafted onto the suggestion tail.
	if got[0] != "zur"+"ich HB" {
		t.Errorf("adaptSuggestions = %q, want zurich HB", got[0])
	}

	if got := adaptSuggestions("", []string{"Bern"}); len(got) != 1 || got[0] != "Bern" {
		t.Errorf("empty input should pass suggestions through, got %v", got)
	}
}

func TestPrefixMatchLen(t *testing.T) {
	tests := []struct {
		suggestion, input string
		want              int
	}{
		{"zürich hb", "zur", len("zür")},    // diacritic folding
		{"st. gallen", "stg", len("st. g")}, // punctuation skipping
		{"bern", "luz", 0},
		{"be", "bern", 0}, // suggestion shorter than input
	}
	for _, tt := range tests {
		if got := prefixMatchLen(tt.suggestion, tt.input); got != tt.want {
			t.Errorf("prefixMatchLen(%q, %q) = %d, want %d", tt.suggestion, tt.input, got, tt.want)
		}
	}
}

func TestFoldRune(t *testing.T) {
	tests := []struct{ in, want rune }{
		{'ü', 'u'},
		{'é', 'e'},
		{'a', 'a'},
	}
	for _, tt := range tests {
		if got := foldRune(tt.in); got != tt.want {
			t.Errorf("foldRune(%c) = %c, want %c", tt.in, got, tt.want)
		}
	}
}

// ----- Update behavior -----

// keyMsg builds a tea.KeyMsg for a printable string or special key.
func keyMsg(s string) tea.KeyMsg {
	switch s {
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

// update is a helper unwrapping the returned tea.Model.
func update(t *testing.T, m appModel, msg tea.Msg) (appModel, tea.Cmd) {
	t.Helper()
	updated, cmd := m.Update(msg)
	am, ok := updated.(appModel)
	if !ok {
		t.Fatalf("Update returned %T, want appModel", updated)
	}
	return am, cmd
}

func TestTabNavigationCycles(t *testing.T) {
	m := newTestModel()
	if m.tabIndex != 0 {
		t.Fatalf("initial tabIndex = %d, want 0", m.tabIndex)
	}

	for i := 1; i < len(m.headerOrder); i++ {
		m, _ = update(t, m, keyMsg("tab"))
		if m.tabIndex != i {
			t.Fatalf("after %d tabs, tabIndex = %d", i, m.tabIndex)
		}
	}

	// One more tab wraps to the start; shift+tab wraps back to the end.
	m, _ = update(t, m, keyMsg("tab"))
	if m.tabIndex != 0 {
		t.Fatalf("tab wrap: tabIndex = %d, want 0", m.tabIndex)
	}
	m, _ = update(t, m, keyMsg("shift+tab"))
	if m.tabIndex != len(m.headerOrder)-1 {
		t.Fatalf("shift+tab wrap: tabIndex = %d, want last", m.tabIndex)
	}
}

// tabTo focuses the header item with the given id.
func tabTo(t *testing.T, m appModel, id string) appModel {
	t.Helper()
	for i := range m.headerOrder {
		if m.headerOrder[i].id == id {
			m.tabIndex = i
			return m
		}
	}
	t.Fatalf("no header item %q", id)
	return m
}

func TestSwapButton(t *testing.T) {
	m := newTestModel()
	m.inputs[0].SetValue("Bern")
	m.inputs[1].SetValue("Luzern")

	m = tabTo(t, m, "swap")
	m, _ = update(t, m, keyMsg(" "))

	if m.inputs[0].Value() != "Luzern" || m.inputs[1].Value() != "Bern" {
		t.Fatalf("swap failed: %q / %q", m.inputs[0].Value(), m.inputs[1].Value())
	}
}

func TestArrivalToggle(t *testing.T) {
	m := newTestModel()
	m = tabTo(t, m, "isArrivalTime")

	m, _ = update(t, m, keyMsg(" "))
	if !m.isArrivalTime {
		t.Fatal("toggle to arrival failed")
	}
	m, _ = update(t, m, keyMsg(" "))
	if m.isArrivalTime {
		t.Fatal("toggle back to departure failed")
	}
}

func TestEnterValidatesInputs(t *testing.T) {
	m := newTestModel()
	m, cmd := update(t, m, keyMsg("enter"))

	if !errors.Is(m.errorMsg, errMissingDeparture) {
		t.Fatalf("errorMsg = %v, want errMissingDeparture", m.errorMsg)
	}
	if cmd != nil {
		t.Fatal("no search should start on invalid input")
	}

	m.inputs[0].SetValue("Bern")
	m, cmd = update(t, m, keyMsg("enter"))
	if !errors.Is(m.errorMsg, errMissingArrival) {
		t.Fatalf("errorMsg = %v, want errMissingArrival", m.errorMsg)
	}
	if cmd != nil {
		t.Fatal("no search should start on invalid input")
	}
}

func TestEnterStartsSearch(t *testing.T) {
	m := newTestModel()
	m.inputs[0].SetValue("Bern")
	m.inputs[1].SetValue("Luzern")

	m, cmd := update(t, m, keyMsg("enter"))
	if !m.loading || !m.searched {
		t.Fatalf("loading = %v, searched = %v; want both true", m.loading, m.searched)
	}
	if cmd == nil {
		t.Fatal("expected a search command")
	}
}

func TestPendingSearchFiresOnWindowSize(t *testing.T) {
	m := NewModel(config.Config{
		Theme: config.DefaultTheme(),
		From:  "Bern",
		To:    "Luzern",
	})
	if !m.pendingSearch {
		t.Fatal("pendingSearch should be set when from and to are pre-filled")
	}

	m, cmd := update(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	if !m.loading || cmd == nil {
		t.Fatal("window size should trigger the pending search")
	}
	if m.pendingSearch {
		t.Fatal("pendingSearch should be consumed")
	}

	// A resize must not retrigger it.
	m.loading = false
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 100, Height: 30})
	if m.loading {
		t.Fatal("resize retriggered the search")
	}
}

func TestNoPendingSearchWithoutFlags(t *testing.T) {
	m := newTestModel()
	if m.pendingSearch {
		t.Fatal("pendingSearch should be false without from/to flags")
	}
}

func TestDataMsgSuccess(t *testing.T) {
	m := newTestModel()
	m.loading = true
	m.resultIndex = 3

	m, _ = update(t, m, dataMsg{connections: []model.Connection{fixtureConnection()}})

	if m.loading {
		t.Fatal("loading should stop")
	}
	if len(m.connections) != 1 || m.resultIndex != 0 {
		t.Fatalf("connections = %d, resultIndex = %d", len(m.connections), m.resultIndex)
	}
	if m.errorMsg != nil {
		t.Fatalf("errorMsg = %v, want nil", m.errorMsg)
	}
}

func TestDataMsgError(t *testing.T) {
	m := newTestModel()
	m.loading = true

	m, _ = update(t, m, dataMsg{err: errors.New("Stop xyz not found.")})
	if m.errorMsg == nil {
		t.Fatal("errorMsg should be set")
	}
}

func TestDataMsgEmpty(t *testing.T) {
	m := newTestModel()
	m, _ = update(t, m, dataMsg{})
	if !errors.Is(m.errorMsg, errNoConnections) {
		t.Fatalf("errorMsg = %v, want errNoConnections", m.errorMsg)
	}
}

func TestResultNavigationBounds(t *testing.T) {
	m := newTestModel()
	m.connections = []model.Connection{fixtureConnection(), fixtureConnection()}

	m, _ = update(t, m, keyMsg("up"))
	if m.resultIndex != 0 {
		t.Fatalf("up at top moved index to %d", m.resultIndex)
	}

	m, _ = update(t, m, keyMsg("down"))
	if m.resultIndex != 1 {
		t.Fatalf("down: index = %d, want 1", m.resultIndex)
	}

	m, _ = update(t, m, keyMsg("down"))
	if m.resultIndex != 1 {
		t.Fatalf("down at bottom moved index to %d", m.resultIndex)
	}
}

func TestQuitKeys(t *testing.T) {
	m := newTestModel()

	_, cmd := update(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc should quit")
	}
	if msg := cmd(); msg != (tea.QuitMsg{}) {
		t.Fatalf("esc cmd = %v, want QuitMsg", msg)
	}

	// "q" quits only when a button is focused, not while typing.
	mBtn := tabTo(t, m, "search")
	_, cmd = update(t, mBtn, keyMsg("q"))
	if cmd == nil {
		t.Fatal("q on button should quit")
	}
	if msg := cmd(); msg != (tea.QuitMsg{}) {
		t.Fatalf("q cmd = %v, want QuitMsg", msg)
	}
}

func TestSuggestionsMsg(t *testing.T) {
	m := newTestModel()
	m.inputs[0].SetValue("zur")

	m, _ = update(t, m, suggestionsMsg{inputIndex: 0, names: []string{"Zürich HB"}})

	got := m.inputs[0].AvailableSuggestions()
	if len(got) != 1 || got[0] != "zurich HB" {
		t.Fatalf("suggestions = %v, want [zurich HB]", got)
	}
}

func TestVersionCheckMsg(t *testing.T) {
	m := newTestModel()
	m, _ = update(t, m, versionCheckMsg{newerVersion: "v9.9.9"})
	if m.newerVersion != "v9.9.9" {
		t.Fatalf("newerVersion = %q", m.newerVersion)
	}
}
