package util

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// withReleaseServer serves the given tag (or status) as the latest GitHub release.
func withReleaseServer(t *testing.T, status int, tag string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		_, _ = w.Write([]byte(`{"tag_name": "` + tag + `"}`))
	}))
	t.Cleanup(srv.Close)

	orig := latestReleaseURL
	latestReleaseURL = srv.URL
	t.Cleanup(func() { latestReleaseURL = orig })
}

func TestNewerVersionDevSkipsLookup(t *testing.T) {
	// No server override: a network call would fail the test via error.
	got, err := NewerVersion("dev")
	if err != nil || got != "" {
		t.Fatalf("NewerVersion(dev) = %q, %v; want empty, nil", got, err)
	}
}

func TestNewerVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    string
		wantErr bool
	}{
		{"newer available", "v1.0.0", "v1.2.0", "v1.2.0", false},
		{"up to date", "v1.2.0", "v1.2.0", "", false},
		{"ahead of release", "v1.3.0", "v1.2.0", "", false},
		{"invalid latest", "v1.0.0", "not-semver", "", true},
		{"invalid current", "1.0.0", "v1.2.0", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withReleaseServer(t, http.StatusOK, tt.latest)

			got, err := NewerVersion(tt.current)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("NewerVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewerVersionHTTPError(t *testing.T) {
	withReleaseServer(t, http.StatusForbidden, "")

	if _, err := NewerVersion("v1.0.0"); err == nil {
		t.Fatal("expected error on HTTP 403")
	}
}
