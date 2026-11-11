package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"poster.app/studio/internal/modules/account"
	"poster.app/studio/internal/modules/studio/analysis"
	"poster.app/studio/internal/modules/studio/asset"
	"poster.app/studio/internal/modules/studio/generation"
	"poster.app/studio/internal/modules/studio/poster"
	"poster.app/studio/internal/modules/studio/template"
	"poster.app/studio/internal/platform/ai"
	"poster.app/studio/internal/platform/logging"
	"poster.app/studio/internal/platform/middleware"
	"poster.app/studio/internal/platform/routing"
	"poster.app/studio/internal/platform/server"
	"poster.app/studio/internal/platform/storage"
	"poster.app/studio/internal/testkit"
	"poster.app/studio/web"
)

func TestHTTPFlow(t *testing.T) {
	db, cfg := testkit.DB(t)
	if err := template.Seed(t.Context(), db); err != nil {
		t.Fatal(err)
	}
	accountRepo, err := account.NewRepository(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	accountResource := account.NewResource(accountRepo, cfg)
	files, err := storage.NewLocal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	assetResource := asset.NewResource(asset.NewRepository(db), files, cfg)
	posterResource := poster.NewResource(poster.NewRepository(db), assetResource)
	models := ai.NewCompleter(cfg)
	routes := []routing.Registrar{
		account.NewController(accountResource),
		template.NewController(template.NewResource(template.NewRepository(db))),
		asset.NewController(assetResource),
		poster.NewController(posterResource),
		generation.NewController(generation.NewResource(generation.NewRepository(db), posterResource, models, cfg)),
		analysis.NewController(analysis.NewResource(analysis.NewRepository(db), posterResource, models, cfg)),
	}
	state := server.NewState()
	state.MarkReady()
	engine, err := server.NewEngine(cfg, logging.New("error", "text"), routes, middleware.NewAuth(account.NewSessionResolver(db)), state, db)
	if err != nil {
		t.Fatal(err)
	}
	if err := web.RegisterStatic(engine); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(engine)
	defer ts.Close()

	live, err := http.Get(ts.URL + "/health/live")
	if err != nil || live.StatusCode != 200 {
		t.Fatalf("live %v %v", live, err)
	}
	live.Body.Close()

	registered := postJSON(t, ts.URL+"/api/account/register", "", map[string]any{
		"username": "FlowUser", "password": "secret123", "display_name": "流程",
	})
	token := registered.Data["token"].(string)
	page := postJSON(t, ts.URL+"/api/template/search", token, map[string]any{})
	if page.Total < 8 {
		t.Fatalf("templates = %d", page.Total)
	}
	created := postJSON(t, ts.URL+"/api/poster/create", token, map[string]any{
		"title": "流程海报",
		"document": map[string]any{
			"width": 1080, "height": 1440, "background": "#112233",
			"layers": []any{map[string]any{
				"id": "t1", "type": "text", "name": "title", "x": 40, "y": 40, "w": 600, "h": 120,
				"opacity": 1, "text": "流程", "font_size": 64, "font_weight": 700,
				"font_family": "serif", "color": "#ffffff", "align": "left",
			}},
		},
	})
	posterID := created.Data["id"].(string)
	other := postJSON(t, ts.URL+"/api/account/register", "", map[string]any{
		"username": "OtherUser", "password": "secret123",
	})
	denied := postRaw(t, ts.URL+"/api/poster/one", other.Data["token"].(string), map[string]any{"id": posterID})
	if denied["code"].(float64) != 404 {
		t.Fatalf("cross user = %v", denied)
	}
	missing := postRaw(t, ts.URL+"/api/missing", token, map[string]any{})
	if missing["code"].(float64) != 404 {
		t.Fatalf("missing api = %v", missing)
	}
	method, err := http.Get(ts.URL + "/api/account/login")
	if err != nil {
		t.Fatal(err)
	}
	defer method.Body.Close()
	var methodBody map[string]any
	if err := json.NewDecoder(method.Body).Decode(&methodBody); err != nil {
		t.Fatal(err)
	}
	if method.StatusCode != 200 || methodBody["code"].(float64) != 405 {
		t.Fatalf("method = %d %v", method.StatusCode, methodBody)
	}
	pageRes, err := http.Get(ts.URL + "/editor/1")
	if err != nil {
		t.Fatal(err)
	}
	defer pageRes.Body.Close()
	html, _ := io.ReadAll(pageRes.Body)
	if pageRes.StatusCode != 200 || !bytes.Contains(html, []byte("海报工坊")) {
		t.Fatalf("spa status %d", pageRes.StatusCode)
	}
	library, err := http.Get(ts.URL + "/assets")
	if err != nil {
		t.Fatal(err)
	}
	defer library.Body.Close()
	libraryHTML, _ := io.ReadAll(library.Body)
	if library.StatusCode != 200 || !bytes.Contains(libraryHTML, []byte("海报工坊")) {
		t.Fatalf("asset page status %d", library.StatusCode)
	}
	assetRes, err := http.Get(ts.URL + "/static/missing-file.js")
	if err != nil {
		t.Fatal(err)
	}
	assetRes.Body.Close()
	if assetRes.StatusCode != 404 {
		t.Fatalf("missing asset status %d", assetRes.StatusCode)
	}
}

type envelope struct {
	Code  int
	Data  map[string]any
	Total int64
}

func postJSON(t *testing.T, url, token string, body any) envelope {
	t.Helper()
	payload := postRaw(t, url, token, body)
	if code, ok := payload["code"].(float64); ok && code != 200 {
		t.Fatalf("%s code %v msg %v", url, code, payload["msg"])
	}
	out := envelope{Total: int64(number(payload["total"]))}
	if data, ok := payload["data"].(map[string]any); ok {
		out.Data = data
		out.Code = 200
	}
	return out
}

func postRaw(t *testing.T, url, token string, body any) map[string]any {
	t.Helper()
	raw, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func number(value any) float64 {
	n, _ := value.(float64)
	return n
}
