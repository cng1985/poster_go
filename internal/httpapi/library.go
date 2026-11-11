package httpapi

import (
	"io"
	"net/http"
	"strings"

	"github.com/cng1985/poster_go/internal/library"
)

func (s *Server) listLibrary(w http.ResponseWriter, r *http.Request) {
	items := append([]library.Item{}, s.Catalog.Items...)
	uploads, err := s.Store.Uploads()
	if err != nil {
		mapErr(w, err, "图片不存在")
		return
	}
	for _, up := range uploads {
		items = append(items, library.Item{
			ID: up.ID, Name: up.Name, Category: "upload", Placement: "element",
			Src: "/uploads/" + up.Filename, Builtin: false,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) uploadLibrary(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20+1024)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "图片不能超过 8MB")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "请选择一张图片")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 8<<20+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "无法读取图片")
		return
	}
	up, err := s.Store.AddUpload(header.Filename, data)
	if err != nil {
		mapErr(w, err, "图片不存在")
		return
	}
	writeJSON(w, http.StatusCreated, library.Item{
		ID: up.ID, Name: up.Name, Category: "upload", Placement: "element",
		Src: "/uploads/" + up.Filename, Builtin: false,
	})
}

func (s *Server) deleteLibrary(w http.ResponseWriter, r *http.Request) {
	if err := s.Store.DeleteUpload(r.PathValue("id")); err != nil {
		mapErr(w, err, "图片不存在")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) serveUpload(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if strings.Contains(name, "/") || strings.Contains(name, "..") {
		http.NotFound(w, r)
		return
	}
	path, ok := s.Store.UploadPath(name)
	if !ok {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}
