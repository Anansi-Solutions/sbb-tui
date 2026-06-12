package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultTheme(t *testing.T) {
	th := DefaultTheme()
	if th.BadgeModelBg != "#7F7F7F" {
		t.Errorf("BadgeModelBg = %q, want #7F7F7F", th.BadgeModelBg)
	}
	if th.Text == "" || th.Error == "" || th.Warning == "" {
		t.Error("default theme has empty colors")
	}
}

func TestMergeTheme(t *testing.T) {
	base := DefaultTheme()

	// Empty override keeps every base value.
	if got := mergeTheme(base, Theme{}); got != base {
		t.Errorf("mergeTheme(empty) = %+v, want base", got)
	}

	// Each set field overrides its base value.
	got := mergeTheme(base, Theme{Text: "#101010", BadgeModelBg: "#202020"})
	if got.Text != "#101010" || got.BadgeModelBg != "#202020" {
		t.Errorf("override not applied: %+v", got)
	}
	if got.Error != base.Error {
		t.Errorf("unset field changed: %q", got.Error)
	}
}

// withTempHome points $HOME at a temp dir for the duration of the test.
func withTempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Make sure the OS fallback path never resolves to the real config.
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	return home
}

func writeConfig(t *testing.T, home, content string) {
	t.Helper()
	dir := filepath.Join(home, ".config", "sbb-tui")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	withTempHome(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.NerdFont || !cfg.Animations {
		t.Errorf("defaults wrong: %+v", cfg)
	}
	if cfg.Theme != DefaultTheme() {
		t.Errorf("theme = %+v, want default", cfg.Theme)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	home := withTempHome(t)
	writeConfig(t, home, `
ui:
  animations: false
  nerdfont: false
  theme:
    text: "#ABCDEF"
`)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Animations || cfg.NerdFont {
		t.Errorf("file overrides not applied: %+v", cfg)
	}
	if cfg.Theme.Text != "#ABCDEF" {
		t.Errorf("theme text = %q, want #ABCDEF", cfg.Theme.Text)
	}
	// Unset theme values keep their defaults.
	if cfg.Theme.Error != DefaultTheme().Error {
		t.Errorf("theme error = %q, want default", cfg.Theme.Error)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	home := withTempHome(t)
	writeConfig(t, home, "ui: [broken")

	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected error on invalid YAML")
	}
}
