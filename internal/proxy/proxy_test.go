package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyStripsAPIPrefix(t *testing.T) {
	// Create a fake ProPresenter backend.
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back the received path so we can verify the prefix was stripped.
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.URL.Path))
	}))
	defer backend.Close()

	proxy := New(backend.URL)

	tests := []struct {
		requestPath  string
		expectedPath string
	}{
		{"/api/v1/messages", "/v1/messages"},
		{"/api/v1/message/0/trigger", "/v1/message/0/trigger"},
		{"/api/v1/message/Eltern%20rufen/clear", "/v1/message/Eltern rufen/clear"},
		{"/api/v1/clear/layer/messages", "/v1/clear/layer/messages"},
	}

	for _, tt := range tests {
		t.Run(tt.requestPath, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rec := httptest.NewRecorder()

			proxy.ServeHTTP(rec, req)

			resp := rec.Result()
			body, _ := io.ReadAll(resp.Body)
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected status 200, got %d", resp.StatusCode)
			}
			if string(body) != tt.expectedPath {
				t.Errorf("expected path %q, got %q", tt.expectedPath, string(body))
			}
		})
	}
}

func TestProxyReturns502WhenBackendDown(t *testing.T) {
	// Point the proxy at a URL that is not listening.
	proxy := New("http://127.0.0.1:1")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/messages", nil)
	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status 502, got %d", rec.Code)
	}
}
