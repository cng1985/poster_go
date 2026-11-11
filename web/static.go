package web

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/platform/response"
)

func RegisterStatic(engine *gin.Engine) error {
	sub, err := fs.Sub(Dist, "dist")
	if err != nil {
		return err
	}
	files := http.FS(sub)
	engine.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if requestPath == "/api" || strings.HasPrefix(requestPath, "/api/") {
			response.HandleError(c, apperrors.NotFound("接口不存在"))
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Status(http.StatusNotFound)
			return
		}
		name := strings.TrimPrefix(path.Clean("/"+requestPath), "/")
		if name == "" || name == "." {
			writeIndex(c, sub)
			return
		}
		if strings.Contains(name, "..") {
			c.Status(http.StatusNotFound)
			return
		}
		file, err := sub.Open(name)
		if err != nil {
			if filepath.Ext(name) != "" {
				c.Status(http.StatusNotFound)
				return
			}
			writeIndex(c, sub)
			return
		}
		stat, statErr := file.Stat()
		closeErr := file.Close()
		if statErr != nil || closeErr != nil || stat.IsDir() {
			c.Status(http.StatusNotFound)
			return
		}
		if strings.HasPrefix(name, "static/") {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
		c.Request.URL.Path = "/" + name
		http.FileServer(files).ServeHTTP(c.Writer, c.Request)
	})
	return nil
}

func writeIndex(c *gin.Context, sub fs.FS) {
	file, err := sub.Open("index.html")
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("Cache-Control", "no-cache")
	http.ServeContent(c.Writer, c.Request, "index.html", stat.ModTime(), bytes.NewReader(content))
}
