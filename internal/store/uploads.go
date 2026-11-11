package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cng1985/poster_go/internal/id"
	"github.com/cng1985/poster_go/internal/media"
)

// Upload is a user image in the library.
type Upload struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"createdAt"`
}

var uploadNameRe = regexp.MustCompile(`^[A-Za-z0-9_-]+\.(png|jpg|gif|webp|svg)$`)

func (s *Store) uploadIndexPath() string { return filepath.Join(s.dir, "uploads.json") }

func (s *Store) loadUploadsLocked() ([]Upload, error) {
	b, err := os.ReadFile(s.uploadIndexPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Upload{}, nil
		}
		return nil, err
	}
	var list []Upload
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []Upload{}
	}
	return list, nil
}

func (s *Store) saveUploadsLocked(list []Upload) error {
	if list == nil {
		list = []Upload{}
	}
	return writeJSON(s.uploadIndexPath(), list, 0o644)
}

// AddUpload stores an image after sniffing its type. The original name is only used as a label.
func (s *Store) AddUpload(displayName string, data []byte) (Upload, error) {
	if len(data) == 0 || len(data) > 8<<20 {
		return Upload{}, userErr(400, "图片不能超过 8MB")
	}
	ext, _, ok := media.SniffImage(data)
	if !ok {
		return Upload{}, userErr(400, "只支持 PNG、JPEG、GIF、WEBP 和 SVG")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	list, err := s.loadUploadsLocked()
	if err != nil {
		return Upload{}, err
	}
	if len(list) >= 200 {
		return Upload{}, userErr(400, "图片库最多保存 200 张上传图片")
	}
	up := Upload{
		ID:        id.New("u"),
		Name:      cleanDisplayName(displayName),
		CreatedAt: time.Now().UTC(),
	}
	up.Filename = up.ID + "." + ext
	full := filepath.Join(s.dir, "uploads", up.Filename)
	if err := os.WriteFile(full, data, 0o644); err != nil {
		return Upload{}, err
	}
	list = append(list, up)
	if err := s.saveUploadsLocked(list); err != nil {
		_ = os.Remove(full)
		return Upload{}, err
	}
	return up, nil
}

// Uploads returns user images, newest first.
func (s *Store) Uploads() ([]Upload, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list, err := s.loadUploadsLocked()
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })
	return list, nil
}

// HasUpload reports whether a public /uploads/ file belongs to the library.
func (s *Store) HasUpload(filename string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hasUploadLocked(filename)
}

func (s *Store) hasUploadLocked(filename string) bool {
	if !uploadNameRe.MatchString(filename) {
		return false
	}
	list, err := s.loadUploadsLocked()
	if err != nil {
		return false
	}
	for _, up := range list {
		if up.Filename == filename {
			return true
		}
	}
	return false
}

// UploadPath returns the disk path for a library file, or ok false.
func (s *Store) UploadPath(filename string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.hasUploadLocked(filename) {
		return "", false
	}
	full := filepath.Join(s.dir, "uploads", filename)
	rel, err := filepath.Rel(filepath.Join(s.dir, "uploads"), full)
	if err != nil || strings.HasPrefix(rel, "..") || strings.Contains(rel, string(filepath.Separator)) {
		return "", false
	}
	return full, true
}

// DeleteUpload removes an image that is not used by a poster.
func (s *Store) DeleteUpload(uploadID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	list, err := s.loadUploadsLocked()
	if err != nil {
		return err
	}
	var target *Upload
	next := make([]Upload, 0, len(list))
	for i := range list {
		if list[i].ID == uploadID {
			cp := list[i]
			target = &cp
			continue
		}
		next = append(next, list[i])
	}
	if target == nil {
		return ErrNotFound
	}
	src := "/uploads/" + target.Filename
	used, err := s.assetUsedLocked(src)
	if err != nil {
		return err
	}
	if used {
		return userErr(409, "这张图片还在海报里使用，先换成别的再删除")
	}
	if err := s.saveUploadsLocked(next); err != nil {
		return err
	}
	_ = os.Remove(filepath.Join(s.dir, "uploads", target.Filename))
	return nil
}

func (s *Store) assetUsedLocked(src string) (bool, error) {
	entries, err := os.ReadDir(filepath.Join(s.dir, "posters"))
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		doc, err := s.readPosterLocked(id)
		if err != nil {
			continue
		}
		if doc.BackgroundImage == src {
			return true, nil
		}
		for _, l := range doc.Layers {
			if l.Src == src {
				return true, nil
			}
		}
	}
	return false, nil
}

func cleanDisplayName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, name)
	if name == "" || name == "." || name == ".." {
		return "图片"
	}
	rs := []rune(name)
	if len(rs) > 80 {
		return string(rs[:80])
	}
	return name
}
