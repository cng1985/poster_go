package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/cng1985/poster_go/internal/poster"
)

// ErrNotFound is returned when a poster or upload does not exist.
var ErrNotFound = errors.New("not found")

// Error is a message that can be shown to the user.
type Error struct {
	Msg    string
	Status int
}

func (e *Error) Error() string { return e.Msg }

func userErr(status int, msg string) error { return &Error{Status: status, Msg: msg} }

// Store persists posters, uploads, and model profiles on disk.
type Store struct {
	dir string
	mu  sync.Mutex
}

// Open creates the data directories and seeds a sample poster once.
func Open(dir string) (*Store, error) {
	s := &Store{dir: dir}
	for _, sub := range []string{"", "posters", "uploads"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return nil, err
		}
	}
	if err := s.maybeSeed(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) maybeSeed() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	marker := filepath.Join(s.dir, ".seeded")
	if _, err := os.Stat(marker); err == nil {
		return nil
	}
	entries, err := os.ReadDir(filepath.Join(s.dir, "posters"))
	if err != nil {
		return err
	}
	hasPoster := false
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			hasPoster = true
			break
		}
	}
	if !hasPoster {
		doc := sampleDocument()
		if err := poster.Normalize(&doc, true); err != nil {
			return err
		}
		if err := s.savePosterLocked(doc); err != nil {
			return err
		}
	}
	return os.WriteFile(marker, []byte("ok\n"), 0o644)
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmp)
		}
	}()
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp, mode); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

func writeJSON(path string, v any, mode os.FileMode) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return writeAtomic(path, b, mode)
}
