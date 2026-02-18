package favorites

import (
	"encoding/json"
	"os"
	"sync"
)

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) All() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids, _ := s.load()
	return ids
}

func (s *Store) AllSet() map[string]bool {
	out := make(map[string]bool)
	for _, id := range s.All() {
		out[id] = true
	}
	return out
}

func (s *Store) Toggle(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ids, err := s.load()
	if err != nil {
		return false, err
	}

	found := false
	out := make([]string, 0, len(ids))
	for _, x := range ids {
		if x == id {
			found = true
			continue
		}
		out = append(out, x)
	}

	isFav := false
	if !found {
		out = append(out, id)
		isFav = true
	}

	if err := s.save(out); err != nil {
		return false, err
	}
	return isFav, nil
}

func (s *Store) load() ([]string, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return []string{}, nil
	}
	if len(b) == 0 {
		return []string{}, nil
	}
	var ids []string
	if err := json.Unmarshal(b, &ids); err != nil {
		return []string{}, nil
	}
	return ids, nil
}

func (s *Store) save(ids []string) error {
	b, err := json.MarshalIndent(ids, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0644)
}
