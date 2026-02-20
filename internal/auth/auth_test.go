package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	t.Parallel()

	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	if len(token) != 64 {
		t.Errorf("expected 64-char hex token, got %d chars: %s", len(token), token)
	}

	// Two tokens should differ.
	token2, _ := GenerateToken()
	if token == token2 {
		t.Error("two generated tokens should not be identical")
	}
}

func TestMiddlewareAllowsUnprotectedPaths(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware("secret", []string{"/api/", "/children"})
	wrapped := mw(handler)

	tests := []struct {
		name string
		path string
	}{
		{"root", "/"},
		{"index", "/index.html"},
		{"css", "/style.css"},
		{"js", "/app.js"},
		{"manifest", "/manifest.json"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected 200 for %s, got %d", tc.path, rec.Code)
			}
		})
	}
}

func TestMiddlewareRejectsWithoutToken(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware("secret", []string{"/api/", "/children"})
	wrapped := mw(handler)

	tests := []struct {
		name string
		path string
	}{
		{"api", "/api/v1/messages"},
		{"children", "/children"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected 401 for %s without token, got %d", tc.path, rec.Code)
			}
		})
	}
}

func TestMiddlewareRejectsWrongToken(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware("correct-token", []string{"/api/"})
	wrapped := mw(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong token, got %d", rec.Code)
	}
}

func TestMiddlewareAcceptsValidToken(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware("correct-token", []string{"/api/", "/children"})
	wrapped := mw(handler)

	tests := []struct {
		name string
		path string
	}{
		{"api", "/api/v1/messages"},
		{"children", "/children"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Authorization", "Bearer correct-token")
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected 200 for %s with valid token, got %d", tc.path, rec.Code)
			}
		})
	}
}
