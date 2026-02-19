package favorites

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	mu   sync.Mutex
	path string
	set  map[string]bool
}

func NewStore(path string) *Store {
	s := &Store{
		path: path,
		set:  map[string]bool{},
	}
	_ = s.load()
	return s
}

func (s *Store) All() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]string, 0, len(s.set))
	for id := range s.set {
		out = append(out, id)
	}
	return out
}

func (s *Store) AllSet() map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make(map[string]bool, len(s.set))
	for k, v := range s.set {
		out[k] = v
	}
	return out
}

func (s *Store) Toggle(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.set[id] {
		delete(s.set, id)
	} else {
		s.set[id] = true
	}
	return s.set[id], s.save()
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := os.ReadFile(s.path)
	if err != nil {
		return nil
	}

	var ids []string
	if err := json.Unmarshal(b, &ids); err != nil {
		return err
	}

	for _, id := range ids {
		s.set[id] = true
	}
	return nil
}

func (s *Store) save() error {
	dir := filepath.Dir(s.path)
	_ = os.MkdirAll(dir, 0755)

	ids := make([]string, 0, len(s.set))
	for id := range s.set {
		ids = append(ids, id)
	}

	b, err := json.MarshalIndent(ids, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, b, 0644)
}
