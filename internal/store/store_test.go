package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cng1985/poster_go/internal/llm"
	"github.com/cng1985/poster_go/internal/poster"
)

func TestSeedOnce(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	list, err := s.List()
	if err != nil || len(list) != 1 || list[0].ID != "sample-spring" {
		t.Fatalf("%v %v", list, err)
	}
	if _, err := Open(dir); err != nil {
		t.Fatal(err)
	}
	list, err = s.List()
	if err != nil || len(list) != 1 {
		t.Fatalf("seeded again: %v %v", list, err)
	}
}

func TestCRUDUploadAndSecret(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := s.Create(poster.Blank())
	if err != nil {
		t.Fatal(err)
	}
	doc.Name = "改名"
	doc.Layers = []poster.Layer{{
		ID: "t1", Name: "标题", Type: "text", Role: "title",
		X: 10, Y: 10, W: 200, H: 80, Opacity: 1, Visible: true,
		Text: "你好", FontSize: 40, FontFamily: "sans", FontWeight: "700", Color: "#111111", Align: "center",
	}}
	saved, err := s.Save(doc)
	if err != nil || saved.Name != "改名" || saved.CreatedAt.IsZero() {
		t.Fatalf("%+v %v", saved, err)
	}
	copy, err := s.Duplicate(saved.ID)
	if err != nil || copy.ID == saved.ID || copy.Name == "" {
		t.Fatal(err, copy.ID, copy.Name)
	}
	if err := s.Delete(copy.ID); err != nil {
		t.Fatal(err)
	}

	up, err := s.AddUpload("海报.png", []byte("\x89PNG\r\n\x1a\nrest"))
	if err != nil || up.Filename == "" {
		t.Fatal(err)
	}
	if _, err := s.AddUpload("x.svg", []byte(`<svg><script>alert(1)</script></svg>`)); err == nil {
		t.Fatal("script svg accepted")
	}
	saved.Layers = append(saved.Layers, poster.Layer{
		ID: "img1", Name: "图", Type: "image", Role: "image",
		X: 0, Y: 0, W: 100, H: 100, Opacity: 1, Visible: true,
		Src: "/uploads/" + up.Filename, Fit: "cover",
	})
	if _, err := s.Save(saved); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteUpload(up.ID); err == nil || !isStatus(err, 409) {
		t.Fatalf("delete in use: %v", err)
	}
	if _, ok := s.UploadPath("../llm.json"); ok {
		t.Fatal("traversal accepted")
	}

	err = s.UpdateLLM(func(file *LLMFile) error {
		file.ActiveProfileID = "m1"
		file.Profiles = []llm.Profile{{
			ID: "m1", Name: "本地", BaseURL: "http://127.0.0.1:11434/v1",
			APIKey: "sk-super-secret-xyz", Model: "llama",
		}}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dir, "llm.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %o", info.Mode().Perm())
	}
}

func isStatus(err error, status int) bool {
	e, ok := err.(*Error)
	return ok && e.Status == status
}
