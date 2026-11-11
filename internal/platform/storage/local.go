package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/fx"

	"poster.app/studio/internal/platform/config"
)

type Local struct {
	root string
}

func NewLocal(cfg *config.Config) (*Local, error) {
	if err := os.MkdirAll(cfg.Runtime.UploadDir, 0o750); err != nil {
		return nil, err
	}
	return &Local{root: cfg.Runtime.UploadDir}, nil
}

func (l *Local) Save(key string, data []byte) error {
	full, err := l.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		return err
	}
	tmp := full + ".tmp"
	if err := os.WriteFile(tmp, data, 0o640); err != nil {
		return err
	}
	return os.Rename(tmp, full)
}

func (l *Local) Open(key string) (*os.File, error) {
	full, err := l.path(key)
	if err != nil {
		return nil, err
	}
	return os.Open(full)
}

func (l *Local) Remove(key string) error {
	full, err := l.path(key)
	if err != nil {
		return err
	}
	err = os.Remove(full)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (l *Local) path(key string) (string, error) {
	if key == "" || strings.Contains(key, "..") || strings.ContainsAny(key, `\/`) && strings.Contains(key, "\\") {
		return "", errors.New("非法存储键")
	}
	if strings.Contains(key, "\\") || strings.HasPrefix(key, "/") {
		return "", errors.New("非法存储键")
	}
	clean := filepath.Clean(key)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return "", errors.New("非法存储键")
	}
	full := filepath.Join(l.root, clean)
	rel, err := filepath.Rel(l.root, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", errors.New("非法存储键")
	}
	return full, nil
}

var Module = fx.Module(
	"storage",
	fx.Provide(NewLocal),
)
