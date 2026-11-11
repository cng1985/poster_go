package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cng1985/poster_go/internal/id"
	"github.com/cng1985/poster_go/internal/poster"
)

var posterIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}$`)

func (s *Store) posterPath(id string) (string, error) {
	if !posterIDRe.MatchString(id) {
		return "", ErrNotFound
	}
	return filepath.Join(s.dir, "posters", id+".json"), nil
}

// List returns posters newest first.
func (s *Store) List() ([]poster.Summary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listLocked()
}

func (s *Store) listLocked() ([]poster.Summary, error) {
	entries, err := os.ReadDir(filepath.Join(s.dir, "posters"))
	if err != nil {
		return nil, err
	}
	var out []poster.Summary
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		doc, err := s.readPosterLocked(id)
		if err != nil {
			continue
		}
		out = append(out, poster.Summary{
			ID: doc.ID, Name: doc.Name, Width: doc.Width, Height: doc.Height, UpdatedAt: doc.UpdatedAt,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	if out == nil {
		out = []poster.Summary{}
	}
	return out, nil
}

// Get loads one poster.
func (s *Store) Get(id string) (poster.Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readPosterLocked(id)
}

func (s *Store) readPosterLocked(id string) (poster.Document, error) {
	path, err := s.posterPath(id)
	if err != nil {
		return poster.Document{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return poster.Document{}, ErrNotFound
		}
		return poster.Document{}, err
	}
	var doc poster.Document
	if err := json.Unmarshal(b, &doc); err != nil {
		return poster.Document{}, err
	}
	if doc.Layers == nil {
		doc.Layers = []poster.Layer{}
	}
	return doc, nil
}

// Create stores a new poster and assigns an id.
func (s *Store) Create(doc poster.Document) (poster.Document, error) {
	if doc.Width == 0 && doc.Height == 0 && strings.TrimSpace(doc.Name) == "" && len(doc.Layers) == 0 && doc.Background == "" {
		doc = poster.Blank()
	}
	if doc.Width == 0 {
		doc.Width = 1080
	}
	if doc.Height == 0 {
		doc.Height = 1520
	}
	if strings.TrimSpace(doc.Name) == "" {
		doc.Name = "未命名海报"
	}
	if doc.Background == "" {
		doc.Background = "#f6f1e8"
	}
	doc.ID = id.New("p")
	now := time.Now().UTC()
	doc.CreatedAt = now
	doc.UpdatedAt = now
	if err := poster.Normalize(&doc, true); err != nil {
		return poster.Document{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.savePosterLocked(doc); err != nil {
		return poster.Document{}, err
	}
	return doc, nil
}

// Save replaces an existing poster.
func (s *Store) Save(doc poster.Document) (poster.Document, error) {
	if err := poster.Normalize(&doc, true); err != nil {
		return poster.Document{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.readPosterLocked(doc.ID)
	if err != nil {
		return poster.Document{}, err
	}
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = old.CreatedAt
	}
	doc.UpdatedAt = time.Now().UTC()
	if err := s.savePosterLocked(doc); err != nil {
		return poster.Document{}, err
	}
	return doc, nil
}

func (s *Store) savePosterLocked(doc poster.Document) error {
	if doc.Layers == nil {
		doc.Layers = []poster.Layer{}
	}
	path, err := s.posterPath(doc.ID)
	if err != nil {
		return err
	}
	return writeJSON(path, doc, 0o644)
}

// Delete removes a poster.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	path, err := s.posterPath(id)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// Duplicate copies a poster.
func (s *Store) Duplicate(id string) (poster.Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc, err := s.readPosterLocked(id)
	if err != nil {
		return poster.Document{}, err
	}
	doc.ID = idpkg()
	name := strings.TrimSpace(doc.Name) + " 副本"
	if utf8.RuneCountInString(name) > 80 {
		name = string([]rune(name)[:80])
	}
	doc.Name = name
	now := time.Now().UTC()
	doc.CreatedAt = now
	doc.UpdatedAt = now
	if err := poster.Normalize(&doc, true); err != nil {
		return poster.Document{}, err
	}
	if err := s.savePosterLocked(doc); err != nil {
		return poster.Document{}, err
	}
	return doc, nil
}

func idpkg() string { return id.New("p") }

func sampleDocument() poster.Document {
	doc := poster.Sample()
	now := time.Now().UTC()
	doc.CreatedAt = now
	doc.UpdatedAt = now
	return doc
}
