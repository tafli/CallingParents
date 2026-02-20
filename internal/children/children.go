package children

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
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

// ServeHTTP handles GET /children — returns the names as a JSON array.
func (s *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	names := s.names
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(names); err != nil {
		http.Error(w, "encoding error", http.StatusInternalServerError)
	}
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
