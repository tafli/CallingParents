// Package version holds build-time version information injected via ldflags.
package version

import (
	"encoding/json"
	"net/http"
)

// These variables are set at build time via -ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// Info returns a human-readable version string.
func Info() string {
	return Version + " (" + Commit + ") built " + Date
}

// HandleVersion returns an http.HandlerFunc that responds with version info as JSON.
func HandleVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"version": Version,
			"commit":  Commit,
			"date":    Date,
		})
	}
}
