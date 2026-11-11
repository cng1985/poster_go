package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cng1985/poster_go/internal/library"
	"github.com/cng1985/poster_go/internal/store"
)

func TestServerFlows(t *testing.T) {
	web := filepath.Join("..", "..", "web")
	cat, err := library.Load(web)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var auth string
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "health check") {
			io.WriteString(w, `{"choices":[{"message":{"content":"OK"}}]}`)
			return
		}
		io.WriteString(w, `{"choices":[{"message":{"content":"{\"name\":\"生成\",\"background\":\"#ffffff\",\"layers\":[{\"type\":\"text\",\"text\":\"标题\",\"x\":20,\"y\":20,\"w\":400,\"h\":80,\"fontSize\":42,\"color\":\"#111111\"},{\"type\":\"image\",\"src\":\"/nope.png\",\"x\":0,\"y\":0,\"w\":80,\"h\":80},{\"type\":\"image\",\"src\":\"/library/backgrounds/dusk.svg\",\"x\":0,\"y\":200,\"w\":300,\"h\":300}]}"}}]}`)
	}))
	defer mock.Close()

	handler := New(st, cat, web).Handler()
	call := func(method, path, contentType string, body io.Reader) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, body)
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	res := call(http.MethodGet, "/", "", nil)
	if res.Code != 200 || !strings.Contains(res.Body.String(), "图片库") {
		t.Fatalf("index %d %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Header().Get("Content-Security-Policy"), "script-src 'self'") {
		t.Fatal(res.Header().Get("Content-Security-Policy"))
	}

	raw, _ := json.Marshal(map[string]any{
		"name": "测试模型", "baseUrl": mock.URL + "/v1", "model": "demo",
		"apiKey": "sk-super-secret-xyz", "activate": true,
	})
	created := call(http.MethodPost, "/api/llm/profiles", "application/json", bytes.NewReader(raw))
	if created.Code != 201 || strings.Contains(created.Body.String(), "sk-super-secret-xyz") {
		t.Fatalf("create leaked %d %s", created.Code, created.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &profile); err != nil || profile.ID == "" {
		t.Fatal(err, created.Body.String())
	}
	listed := call(http.MethodGet, "/api/llm/profiles", "", nil)
	if strings.Contains(listed.Body.String(), "sk-super-secret-xyz") || !strings.Contains(listed.Body.String(), `"apiKeySet":true`) {
		t.Fatalf("list leaked %s", listed.Body.String())
	}
	raw, _ = json.Marshal(map[string]any{"id": profile.ID})
	tested := call(http.MethodPost, "/api/llm/test", "application/json", bytes.NewReader(raw))
	if tested.Code != 200 || auth != "Bearer sk-super-secret-xyz" {
		t.Fatalf("test %d %s auth %s", tested.Code, tested.Body.String(), auth)
	}
	raw, _ = json.Marshal(map[string]any{
		"mode": "layout", "prompt": "一张春天海报", "width": 1080, "height": 1520,
	})
	gen := call(http.MethodPost, "/api/llm/generate", "application/json", bytes.NewReader(raw))
	if gen.Code != 200 || strings.Contains(gen.Body.String(), "/nope.png") || !strings.Contains(gen.Body.String(), "/library/backgrounds/dusk.svg") {
		t.Fatalf("generate %d %s", gen.Code, gen.Body.String())
	}

	posters := call(http.MethodGet, "/api/posters", "", nil)
	if posters.Code != 200 || !strings.Contains(posters.Body.String(), "sample-spring") {
		t.Fatalf("posters %s", posters.Body.String())
	}
	lib := call(http.MethodGet, "/api/library", "", nil)
	if lib.Code != 200 || !strings.Contains(lib.Body.String(), "暮色") {
		t.Fatalf("library %s", lib.Body.String())
	}

	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("file", "dot.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("\x89PNG\r\n\x1a\nxx")); err != nil {
		t.Fatal(err)
	}
	form.Close()
	up := call(http.MethodPost, "/api/library", form.FormDataContentType(), &body)
	if up.Code != 201 || !strings.Contains(up.Body.String(), "/uploads/") {
		t.Fatalf("upload %d %s", up.Code, up.Body.String())
	}

	leak := call(http.MethodGet, "/uploads/../llm.json", "", nil)
	if leak.Code == 200 || strings.Contains(leak.Body.String(), "sk-super-secret-xyz") {
		t.Fatalf("leak %d %s", leak.Code, leak.Body.String())
	}
}

func TestStoredKeyDoesNotFollowANewOrigin(t *testing.T) {
	web := filepath.Join("..", "..", "web")
	cat, err := library.Load(web)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var seen string
	home := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("Authorization")
		io.WriteString(w, `{"data":[{"id":"demo"}]}`)
	}))
	defer home.Close()
	var stolen string
	thief := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stolen = r.Header.Get("Authorization")
		io.WriteString(w, `{"data":[{"id":"stolen"}]}`)
	}))
	defer thief.Close()

	handler := New(st, cat, web).Handler()
	call := func(path string, payload any) *httptest.ResponseRecorder {
		t.Helper()
		b, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	created := call("/api/llm/profiles", map[string]any{
		"name": "本地", "baseUrl": home.URL + "/v1", "model": "demo",
		"apiKey": "sk-super-secret-xyz", "activate": true,
	})
	if created.Code != 201 {
		t.Fatal(created.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &profile); err != nil {
		t.Fatal(err)
	}
	moved := call("/api/llm/models", map[string]any{"id": profile.ID, "baseUrl": thief.URL + "/v1"})
	if moved.Code != 400 || stolen != "" || strings.Contains(moved.Body.String(), "sk-super-secret-xyz") {
		t.Fatalf("moved %d stolen %q body %s", moved.Code, stolen, moved.Body.String())
	}
	listed := call("/api/llm/models", map[string]any{"id": profile.ID, "baseUrl": home.URL + "/other"})
	if listed.Code != 200 || seen != "Bearer sk-super-secret-xyz" {
		t.Fatalf("same origin %d auth %q %s", listed.Code, seen, listed.Body.String())
	}
	fresh := call("/api/llm/models", map[string]any{
		"baseUrl": thief.URL + "/v1", "name": "发现", "apiKey": "sk-new",
	})
	if fresh.Code != 200 || stolen != "Bearer sk-new" || !strings.Contains(fresh.Body.String(), "stolen") {
		t.Fatalf("discover %d auth %q %s", fresh.Code, stolen, fresh.Body.String())
	}
}
