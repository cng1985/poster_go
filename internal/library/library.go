package library

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Item is one picture in the built-in library.
type Item struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Placement string `json:"placement"`
	Src       string `json:"src"`
	Builtin   bool   `json:"builtin"`
}

// Catalog is the built-in image library shipped with the editor.
type Catalog struct {
	Items []Item
	bySrc map[string]Item
}

// Load reads web/library/manifest.json and checks that every file exists.
func Load(webDir string) (*Catalog, error) {
	path := filepath.Join(webDir, "library", "manifest.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取图片库清单: %w", err)
	}
	var raw struct {
		Items []Item `json:"items"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("解析图片库清单: %w", err)
	}
	if len(raw.Items) == 0 {
		return nil, fmt.Errorf("图片库是空的")
	}
	cat := &Catalog{bySrc: map[string]Item{}}
	seen := map[string]bool{}
	allowedCat := map[string]bool{"background": true, "texture": true, "illustration": true, "icon": true}
	allowedPlace := map[string]bool{"backdrop": true, "element": true}
	for _, it := range raw.Items {
		if it.ID == "" || seen[it.ID] {
			return nil, fmt.Errorf("图片库编号无效或重复: %s", it.ID)
		}
		seen[it.ID] = true
		if strings.TrimSpace(it.Name) == "" {
			return nil, fmt.Errorf("图片 %s 缺少名称", it.ID)
		}
		if !allowedCat[it.Category] || !allowedPlace[it.Placement] {
			return nil, fmt.Errorf("图片 %s 分类无效", it.ID)
		}
		if !strings.HasPrefix(it.Src, "/library/") || strings.Contains(it.Src, "..") {
			return nil, fmt.Errorf("图片 %s 路径无效", it.ID)
		}
		full := filepath.Join(webDir, filepath.FromSlash(strings.TrimPrefix(it.Src, "/")))
		rel, err := filepath.Rel(filepath.Join(webDir, "library"), full)
		if err != nil || strings.HasPrefix(rel, "..") {
			return nil, fmt.Errorf("图片 %s 路径越界", it.ID)
		}
		fi, err := os.Stat(full)
		if err != nil || fi.IsDir() {
			return nil, fmt.Errorf("图片文件不存在: %s", it.Src)
		}
		it.Builtin = true
		cat.Items = append(cat.Items, it)
		cat.bySrc[it.Src] = it
	}
	return cat, nil
}

// Has reports whether src is a built-in library path.
func (c *Catalog) Has(src string) bool {
	if c == nil {
		return false
	}
	_, ok := c.bySrc[src]
	return ok
}

// Get returns the item for a public src.
func (c *Catalog) Get(src string) (Item, bool) {
	if c == nil {
		return Item{}, false
	}
	it, ok := c.bySrc[src]
	return it, ok
}
