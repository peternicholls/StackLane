// State store: per-project JSON files, atomic writes, registry projection.
//
// Atomicity is enforced via temp-file + os.Rename (FR-008). Concurrent access
// at the same project slug is serialised by the caller.
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const SchemaVersion = 1

// ErrNotFound is returned when a slug has no recorded state.
var ErrNotFound = errors.New("state: project not found")

// Store is the default StateStore implementation.
type Store struct {
	stateDir string
	mu       sync.Mutex
}

// NewStore returns a Store rooted at stateDir. It ensures stateDir/projects exists.
func NewStore(stateDir string) (*Store, error) {
	if stateDir == "" {
		return nil, errors.New("state: empty state dir")
	}
	if err := os.MkdirAll(filepath.Join(stateDir, "projects"), 0o755); err != nil {
		return nil, err
	}
	return &Store{stateDir: stateDir}, nil
}

// StateDir returns the configured state directory.
func (s *Store) StateDir() string { return s.stateDir }

func (s *Store) projectFile(slug string) string {
	return filepath.Join(s.stateDir, "projects", slug+".json")
}

// Save writes a project record to disk atomically.
func (s *Store) Save(rec Record) error {
	if rec.Project.Slug == "" {
		return errors.New("state: cannot save record with empty slug")
	}
	rec.SchemaVersion = SchemaVersion
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Join(s.stateDir, "projects"), 0o755); err != nil {
		return err
	}
	target := s.projectFile(rec.Project.Slug)

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("state: marshal record: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), "."+rec.Project.Slug+".*.json.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, target)
}

// Load reads the recorded state for slug.
func (s *Store) Load(slug string) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadFile(s.projectFile(slug))
}

func (s *Store) loadFile(path string) (Record, error) {
	var rec Record
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return rec, ErrNotFound
		}
		return rec, err
	}
	if err := json.Unmarshal(data, &rec); err != nil {
		return rec, fmt.Errorf("state: parse %s: %w", path, err)
	}
	return rec, nil
}

// Remove deletes the state record for slug. Idempotent.
func (s *Store) Remove(slug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(s.projectFile(slug))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// listFiles returns every per-project state file (sorted for determinism).
// Includes .json files only.
func (s *Store) listFiles() ([]string, error) {
	dir := filepath.Join(s.stateDir, "projects")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".json") {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(out)
	return out, nil
}

// StateFileForSelector resolves a selector against the recorded projects.
// Selectors match against slug, name, hostname, or project dir (mirrors
// stageserve_state_file_for_selector).
func (s *Store) StateFileForSelector(selector string) (Record, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	files, err := s.listFiles()
	if err != nil {
		return Record{}, "", err
	}
	for _, f := range files {
		rec, err := s.loadFile(f)
		if err != nil {
			continue
		}
		p := rec.Project
		if selector == p.Slug || selector == p.Name || selector == p.Hostname || selector == p.Dir {
			return rec, f, nil
		}
	}
	return Record{}, "", ErrNotFound
}
