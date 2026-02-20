package children

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewStoreLoadsFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "children.json")
	os.WriteFile(path, []byte(`["Clara","Anna","Ben"]`), 0644)

	s, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}

	names := s.Names()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}
	expected := []string{"Anna", "Ben", "Clara"}
	for i, want := range expected {
		if names[i] != want {
			t.Errorf("names[%d] = %q, want %q", i, names[i], want)
		}
	}
}

func TestNewStoreEmptyWhenFileNotExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.json")

	s, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}

	names := s.Names()
	if len(names) != 0 {
		t.Errorf("expected 0 names, got %d", len(names))
	}
}

func TestNewStoreInvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte(`not json`), 0644)

	_, err := NewStore(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestServeHTTPReturnsJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "children.json")
	os.WriteFile(path, []byte(`["David","Emma"]`), 0644)

	s, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/children", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var names []string
	if err := json.NewDecoder(rec.Body).Decode(&names); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	if names[0] != "David" || names[1] != "Emma" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestServeHTTPPostAddsName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "children.json")
	os.WriteFile(path, []byte(`["Anna","Ben"]`), 0644)

	s, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}

	body := strings.NewReader(`{"name":"Clara"}`)
	req := httptest.NewRequest(http.MethodPost, "/children", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	// Response should contain all names sorted.
	var names []string
	json.NewDecoder(rec.Body).Decode(&names)
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d: %v", len(names), names)
	}
	if names[0] != "Anna" || names[1] != "Ben" || names[2] != "Clara" {
		t.Errorf("unexpected names: %v", names)
	}

	// Verify file was updated on disk.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	var persisted []string
	json.Unmarshal(data, &persisted)
	if len(persisted) != 3 {
		t.Fatalf("expected 3 persisted names, got %d", len(persisted))
	}
}

func TestServeHTTPPostDuplicate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "children.json")
	os.WriteFile(path, []byte(`["Anna","Ben"]`), 0644)

	s, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}

	body := strings.NewReader(`{"name":"Anna"}`)
	req := httptest.NewRequest(http.MethodPost, "/children", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for duplicate, got %d", rec.Code)
	}

	// Store should still have 2 names.
	if len(s.Names()) != 2 {
		t.Errorf("expected 2 names after duplicate add, got %d", len(s.Names()))
	}
}

func TestServeHTTPPostEmptyName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "children.json")
	os.WriteFile(path, []byte(`[]`), 0644)

	s, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}

	body := strings.NewReader(`{"name":"  "}`)
	req := httptest.NewRequest(http.MethodPost, "/children", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty name, got %d", rec.Code)
	}
}

func TestServeHTTPRejectsDelete(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "children.json")
	os.WriteFile(path, []byte(`[]`), 0644)

	s, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/children", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rec.Code)
	}
}

func TestServeHTTPPostCreatesFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "children.json")
	// File does not exist initially.

	s, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}

	body := strings.NewReader(`{"name":"Finn"}`)
	req := httptest.NewRequest(http.MethodPost, "/children", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}

	// File should now exist with the name.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	var persisted []string
	json.Unmarshal(data, &persisted)
	if len(persisted) != 1 || persisted[0] != "Finn" {
		t.Errorf("unexpected persisted content: %v", persisted)
	}
}

func TestNamesReturnsCopy(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "children.json")
	os.WriteFile(path, []byte(`["Anna"]`), 0644)

	s, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}

	names := s.Names()
	names[0] = "Modified"

	original := s.Names()
	if original[0] != "Anna" {
		t.Error("Names() did not return a copy — mutation affected the store")
	}
}
