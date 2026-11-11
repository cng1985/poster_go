package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/cng1985/poster_go/internal/llm"
)

// LLMFile is the on-disk model configuration. It contains API keys.
type LLMFile struct {
	ActiveProfileID string        `json:"activeProfileId"`
	Profiles        []llm.Profile `json:"profiles"`
}

func (s *Store) llmPath() string { return filepath.Join(s.dir, "llm.json") }

// UpdateLLM atomically reads, mutates, and writes the model configuration.
func (s *Store) UpdateLLM(fn func(*LLMFile) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := s.loadLLMLocked()
	if err != nil {
		return err
	}
	if err := fn(&file); err != nil {
		return err
	}
	return s.saveLLMLocked(file)
}

// LoadLLM returns a copy of the model configuration.
func (s *Store) LoadLLM() (LLMFile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLLMLocked()
}

func (s *Store) loadLLMLocked() (LLMFile, error) {
	b, err := os.ReadFile(s.llmPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return LLMFile{Profiles: []llm.Profile{}}, nil
		}
		return LLMFile{}, err
	}
	var file LLMFile
	if err := json.Unmarshal(b, &file); err != nil {
		return LLMFile{}, err
	}
	if file.Profiles == nil {
		file.Profiles = []llm.Profile{}
	}
	return file, nil
}

func (s *Store) saveLLMLocked(file LLMFile) error {
	if file.Profiles == nil {
		file.Profiles = []llm.Profile{}
	}
	return writeJSON(s.llmPath(), file, 0o600)
}
