package message

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleSendSuccess(t *testing.T) {
	t.Parallel()

	var receivedPath, receivedBody string
	pp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		b := make([]byte, 1024)
		n, _ := r.Body.Read(b)
		receivedBody = string(b[:n])
		w.WriteHeader(http.StatusOK)
	}))
	defer pp.Close()

	h := New(pp.URL, "Eltern rufen")

	body := strings.NewReader(`{"name":"Paul"}`)
	req := httptest.NewRequest(http.MethodPost, "/message/send", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.HandleSend(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	expectedPath := "/v1/message/Eltern rufen/trigger"
	if receivedPath != expectedPath {
		t.Errorf("expected PP path %q, got %q", expectedPath, receivedPath)
	}

	if !strings.Contains(receivedBody, `"text":"Paul"`) {
		t.Errorf("expected body to contain Paul, got %q", receivedBody)
	}
}

func TestHandleSendEmptyName(t *testing.T) {
	t.Parallel()

	h := New("http://localhost:1", "Eltern rufen")

	body := strings.NewReader(`{"name":"  "}`)
	req := httptest.NewRequest(http.MethodPost, "/message/send", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.HandleSend(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty name, got %d", rec.Code)
	}
}

func TestHandleSendRejectsGet(t *testing.T) {
	t.Parallel()

	h := New("http://localhost:1", "Eltern rufen")

	req := httptest.NewRequest(http.MethodGet, "/message/send", nil)
	rec := httptest.NewRecorder()

	h.HandleSend(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleSendProPresenterDown(t *testing.T) {
	t.Parallel()

	h := New("http://127.0.0.1:1", "Eltern rufen")

	body := strings.NewReader(`{"name":"Paul"}`)
	req := httptest.NewRequest(http.MethodPost, "/message/send", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.HandleSend(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502 when PP is down, got %d", rec.Code)
	}
}

func TestHandleClearSuccess(t *testing.T) {
	t.Parallel()

	var receivedPath string
	pp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer pp.Close()

	h := New(pp.URL, "Eltern rufen")

	req := httptest.NewRequest(http.MethodPost, "/message/clear", nil)
	rec := httptest.NewRecorder()

	h.HandleClear(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	expectedPath := "/v1/message/Eltern rufen/clear"
	if receivedPath != expectedPath {
		t.Errorf("expected PP path %q, got %q", expectedPath, receivedPath)
	}
}

func TestHandleClearRejectsGet(t *testing.T) {
	t.Parallel()

	h := New("http://localhost:1", "Eltern rufen")

	req := httptest.NewRequest(http.MethodGet, "/message/clear", nil)
	rec := httptest.NewRecorder()

	h.HandleClear(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleTestSuccess(t *testing.T) {
	t.Parallel()

	pp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": map[string]any{"uuid": "abc"}, "name": "Eltern rufen"},
		})
	}))
	defer pp.Close()

	h := New(pp.URL, "Eltern rufen")

	req := httptest.NewRequest(http.MethodGet, "/message/test", nil)
	rec := httptest.NewRecorder()

	h.HandleTest(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestHandleTestProPresenterDown(t *testing.T) {
	t.Parallel()

	h := New("http://127.0.0.1:1", "Eltern rufen")

	req := httptest.NewRequest(http.MethodGet, "/message/test", nil)
	rec := httptest.NewRecorder()

	h.HandleTest(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502 when PP is down, got %d", rec.Code)
	}
}

func TestEscapeJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"Paul", "Paul"},
		{`O'Brien`, `O'Brien`},
		{`He said "hi"`, `He said \"hi\"`},
	}

	for _, tc := range tests {
		got := escapeJSON(tc.input)
		if got != tc.expected {
			t.Errorf("escapeJSON(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
