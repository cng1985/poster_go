package httpapi

import (
	"net/http"

	"github.com/cng1985/poster_go/internal/poster"
)

func (s *Server) listPosters(w http.ResponseWriter, r *http.Request) {
	list, err := s.Store.List()
	if err != nil {
		mapErr(w, err, "海报不存在")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) createPoster(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r, 2<<20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "请求体过大或无法读取")
		return
	}
	var doc poster.Document
	if err := decodeJSON(body, &doc); err != nil && err != errEmpty {
		writeError(w, http.StatusBadRequest, "海报数据无法解析")
		return
	}
	if err := s.checkAssets(&doc); err != nil {
		mapErr(w, err, "海报不存在")
		return
	}
	created, err := s.Store.Create(doc)
	if err != nil {
		mapErr(w, err, "海报不存在")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getPoster(w http.ResponseWriter, r *http.Request) {
	doc, err := s.Store.Get(r.PathValue("id"))
	if err != nil {
		mapErr(w, err, "海报不存在")
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (s *Server) savePoster(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r, 2<<20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "请求体过大或无法读取")
		return
	}
	var doc poster.Document
	if err := decodeJSON(body, &doc); err != nil {
		writeError(w, http.StatusBadRequest, "海报数据无法解析")
		return
	}
	doc.ID = r.PathValue("id")
	if err := s.checkAssets(&doc); err != nil {
		mapErr(w, err, "海报不存在")
		return
	}
	saved, err := s.Store.Save(doc)
	if err != nil {
		mapErr(w, err, "海报不存在")
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (s *Server) deletePoster(w http.ResponseWriter, r *http.Request) {
	if err := s.Store.Delete(r.PathValue("id")); err != nil {
		mapErr(w, err, "海报不存在")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) duplicatePoster(w http.ResponseWriter, r *http.Request) {
	doc, err := s.Store.Duplicate(r.PathValue("id"))
	if err != nil {
		mapErr(w, err, "海报不存在")
		return
	}
	writeJSON(w, http.StatusCreated, doc)
}

func (s *Server) templates(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, poster.Templates())
}
