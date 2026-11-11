package httpapi

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cng1985/poster_go/internal/library"
	"github.com/cng1985/poster_go/internal/llm"
	"github.com/cng1985/poster_go/internal/poster"
	"github.com/cng1985/poster_go/internal/store"
)

// Server serves the editor and its API.
type Server struct {
	Store   *store.Store
	Catalog *library.Catalog
	WebDir  string
	LLM     *llm.Client
}

// New builds a server. LLM may be nil.
func New(st *store.Store, cat *library.Catalog, webDir string) *Server {
	return &Server{Store: st, Catalog: cat, WebDir: webDir, LLM: &llm.Client{}}
}

// Handler returns the root HTTP handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.api(s.health))
	mux.HandleFunc("GET /api/posters", s.api(s.listPosters))
	mux.HandleFunc("POST /api/posters", s.api(s.createPoster))
	mux.HandleFunc("GET /api/posters/{id}", s.api(s.getPoster))
	mux.HandleFunc("PUT /api/posters/{id}", s.api(s.savePoster))
	mux.HandleFunc("DELETE /api/posters/{id}", s.api(s.deletePoster))
	mux.HandleFunc("POST /api/posters/{id}/duplicate", s.api(s.duplicatePoster))
	mux.HandleFunc("GET /api/templates", s.api(s.templates))
	mux.HandleFunc("GET /api/library", s.api(s.listLibrary))
	mux.HandleFunc("POST /api/library", s.api(s.uploadLibrary))
	mux.HandleFunc("DELETE /api/library/{id}", s.api(s.deleteLibrary))
	mux.HandleFunc("GET /api/llm/profiles", s.api(s.listProfiles))
	mux.HandleFunc("POST /api/llm/profiles", s.api(s.createProfile))
	mux.HandleFunc("PUT /api/llm/profiles/{id}", s.api(s.updateProfile))
	mux.HandleFunc("DELETE /api/llm/profiles/{id}", s.api(s.deleteProfile))
	mux.HandleFunc("POST /api/llm/profiles/{id}/activate", s.api(s.activateProfile))
	mux.HandleFunc("POST /api/llm/test", s.api(s.testProfile))
	mux.HandleFunc("POST /api/llm/models", s.api(s.listModels))
	mux.HandleFunc("POST /api/llm/generate", s.api(s.generate))

	mux.Handle("GET /css/", fileHandler(filepath.Join(s.WebDir, "css"), "/css/"))
	mux.Handle("GET /js/", fileHandler(filepath.Join(s.WebDir, "js"), "/js/"))
	mux.Handle("GET /library/", fileHandler(filepath.Join(s.WebDir, "library"), "/library/"))
	mux.HandleFunc("GET /uploads/{name}", s.serveUpload)
	mux.HandleFunc("GET /favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(s.WebDir, "favicon.svg"))
	})
	mux.HandleFunc("GET /{$}", s.index)
	return withSecurity(mux)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.WebDir, "index.html"))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) api(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				if !sw.wrote {
					writeError(sw, http.StatusInternalServerError, "服务器内部错误")
				}
			}
			log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Millisecond))
		}()
		fn(sw, r)
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (s *statusWriter) WriteHeader(code int) {
	if s.wrote {
		return
	}
	s.wrote = true
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusWriter) Write(b []byte) (int, error) {
	if !s.wrote {
		s.WriteHeader(http.StatusOK)
	}
	return s.ResponseWriter.Write(b)
}

func withSecurity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Cache-Control", "no-cache")
		p := r.URL.Path
		if strings.HasPrefix(p, "/library/") || strings.HasPrefix(p, "/uploads/") || p == "/favicon.svg" {
			w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
		} else {
			w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob:; style-src 'self'; script-src 'self'; connect-src 'self'; base-uri 'none'; object-src 'none'; frame-ancestors 'none'")
		}
		next.ServeHTTP(w, r)
	})
}

func fileHandler(root, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, prefix)
		full, ok := safeJoin(root, rel)
		if !ok {
			http.NotFound(w, r)
			return
		}
		fi, err := os.Stat(full)
		if err != nil || fi.IsDir() {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, full)
	})
}

func safeJoin(root, rel string) (string, bool) {
	rel = strings.TrimPrefix(filepath.Clean("/"+rel), "/")
	if rel == "" || rel == "." {
		return "", false
	}
	full := filepath.Clean(filepath.Join(root, rel))
	root = filepath.Clean(root)
	if full != root && !strings.HasPrefix(full, root+string(os.PathSeparator)) {
		return "", false
	}
	return full, true
}

func (s *Server) knownAsset(src string) bool {
	if src == "" {
		return true
	}
	if strings.HasPrefix(src, "/library/") {
		return s.Catalog.Has(src)
	}
	if strings.HasPrefix(src, "/uploads/") {
		name := strings.TrimPrefix(src, "/uploads/")
		if strings.Contains(name, "/") {
			return false
		}
		return s.Store.HasUpload(name)
	}
	return false
}

func (s *Server) checkAssets(doc *poster.Document) error {
	if !s.knownAsset(doc.BackgroundImage) {
		return &poster.InvalidError{Msg: "背景图不存在"}
	}
	for _, l := range doc.Layers {
		if l.Type == "image" && !s.knownAsset(l.Src) {
			return &poster.InvalidError{Msg: "图片不存在或已删除"}
		}
	}
	return nil
}
