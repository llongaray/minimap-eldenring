package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	mu    sync.RWMutex
	file  string
	value Settings
}

func NewStore(path string) (*Store, error) {
	s := &Store{file: path, value: Defaults()}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	if err = json.Unmarshal(raw, &s.value); err != nil {
		return nil, fmt.Errorf("configuração JSON: %w", err)
	}
	if err = s.value.Validate(); err != nil {
		return nil, err
	}
	if s.value.Markers == nil {
		s.value.Markers = map[string]Flags{}
	}
	if s.value.HiddenCategories == nil {
		s.value.HiddenCategories = map[string]bool{}
	}
	return s, nil
}
func (s *Store) Get() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v := s.value
	v.Markers = make(map[string]Flags, len(s.value.Markers))
	for k, z := range s.value.Markers {
		v.Markers[k] = z
	}
	v.HiddenCategories = make(map[string]bool, len(s.value.HiddenCategories))
	for k, z := range s.value.HiddenCategories {
		v.HiddenCategories[k] = z
	}
	return v
}
func (s *Store) Put(v Settings) error {
	if err := v.Validate(); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err = os.MkdirAll(filepath.Dir(s.file), 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.file), ".aray-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(raw); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	// Windows Rename does not replace an existing destination: preserve prior state
	// before writing a new copy; rollback on failure.
	old := s.file + ".bak"
	if _, err = os.Stat(s.file); err == nil {
		_ = os.Remove(old)
		if err = os.Rename(s.file, old); err != nil {
			return err
		}
	}
	if err = os.Rename(tmp.Name(), s.file); err != nil {
		_ = os.Rename(old, s.file)
		return err
	}
	s.value = v
	return nil
}
