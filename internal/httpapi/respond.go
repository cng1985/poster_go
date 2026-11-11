package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/cng1985/poster_go/internal/poster"
	"github.com/cng1985/poster_go/internal/store"
)

var (
	errTooLarge = errors.New("too large")
	errEmpty    = errors.New("empty")
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readBody(r *http.Request, max int64) ([]byte, error) {
	defer r.Body.Close()
	b, err := io.ReadAll(io.LimitReader(r.Body, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, errTooLarge
	}
	return b, nil
}

func decodeJSON(b []byte, dest any) error {
	if len(bytes.TrimSpace(b)) == 0 {
		return errEmpty
	}
	return json.Unmarshal(b, dest)
}

func mapErr(w http.ResponseWriter, err error, notFound string) {
	var inv *poster.InvalidError
	var se *store.Error
	switch {
	case errors.As(err, &inv):
		writeError(w, 400, inv.Msg)
	case errors.As(err, &se):
		status := se.Status
		if status == 0 {
			status = 400
		}
		writeError(w, status, se.Msg)
	case errors.Is(err, store.ErrNotFound):
		writeError(w, 404, notFound)
	default:
		log.Printf("handler: %v", err)
		writeError(w, 500, "服务器内部错误")
	}
}
