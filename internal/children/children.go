package children

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
)

// Store loads and serves a list of children's names from a JSON file.
type Store struct {
	mu       sync.RWMutex
	names    []string
	filePath string
}

// NewStore creates a Store that reads names from the given JSON file.
// The file must contain a JSON array of strings, e.g. ["Anna","Ben","Clara"].
// If the file does not exist, the store starts with an empty list.
func NewStore(filePath string) (*Store, error) {
	s := &Store{filePath: filePath}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// Names returns a copy of the current children list.
func (s *Store) Names() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, len(s.names))
	copy(out, s.names)
	return out
}

// ServeHTTP handles GET, POST, and DELETE /children.
// GET returns the names as a JSON array.
// POST accepts {"name":"..."} and adds the name to the list, persisting to disk.
// DELETE accepts {"name":"..."} and removes the name from the list, persisting to disk.
func (s *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodDelete:
		s.handleDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Store) handleGet(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	names := s.names
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(names); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
}

// addRequest is the expected JSON body for POST /children.
type addRequest struct {
	Name string `json:"name"`
}

func (s *Store) handlePost(w http.ResponseWriter, r *http.Request) {
	var req addRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "name must not be empty", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate (case-sensitive).
	for _, existing := range s.names {
		if existing == name {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(s.names)
			return
		}
	}

	s.names = append(s.names, name)
	sort.Slice(s.names, func(i, j int) bool {
		return s.names[i] < s.names[j]
	})

	if err := s.save(); err != nil {
		// Roll back the append on save failure.
		s.load()
		http.Error(w, "failed to persist name", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s.names)
}

// deleteRequest is the expected JSON body for DELETE /children.
type deleteRequest struct {
	Name string `json:"name"`
}

func (s *Store) handleDelete(w http.ResponseWriter, r *http.Request) {
	var req deleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "name must not be empty", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	idx := -1
	for i, existing := range s.names {
		if existing == name {
			idx = i
			break
		}
	}

	if idx == -1 {
		// Name not found — return current list.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(s.names)
		return
	}

	s.names = append(s.names[:idx], s.names[idx+1:]...)

	if err := s.save(); err != nil {
		s.load()
		http.Error(w, "failed to persist deletion", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s.names)
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.names, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling children: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("writing children file %q: %w", s.filePath, err)
	}
	return nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.names = []string{}
			return nil
		}
		return fmt.Errorf("reading children file %q: %w", s.filePath, err)
	}

	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return fmt.Errorf("parsing children file %q: %w", s.filePath, err)
	}

	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})

	s.names = names
	return nil
}
