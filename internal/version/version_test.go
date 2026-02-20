package version

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInfo_Defaults(t *testing.T) {
	t.Parallel()

	got := Info()
	want := "dev (unknown) built unknown"
	if got != want {
		t.Errorf("Info() = %q, want %q", got, want)
	}
}

func TestInfo_WithValues(t *testing.T) {
	origVersion, origCommit, origDate := Version, Commit, Date
	t.Cleanup(func() {
		Version = origVersion
		Commit = origCommit
		Date = origDate
	})

	Version = "v1.2.3"
	Commit = "abc1234"
	Date = "2025-06-15T10:00:00Z"

	got := Info()
	want := "v1.2.3 (abc1234) built 2025-06-15T10:00:00Z"
	if got != want {
		t.Errorf("Info() = %q, want %q", got, want)
	}
}

func TestHandleVersion_JSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()

	HandleVersion().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	for _, key := range []string{"version", "commit", "date"} {
		if _, ok := body[key]; !ok {
			t.Errorf("missing key %q in response", key)
		}
	}

	if body["version"] != Version {
		t.Errorf("version = %q, want %q", body["version"], Version)
	}
}
