package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSameOrigin(t *testing.T) {
	if !SameOrigin("https://api.example.com/v1", "https://api.example.com/v1/chat/completions") {
		t.Fatal("same host")
	}
	if !SameOrigin("http://127.0.0.1/v1", "http://127.0.0.1:80/other") {
		t.Fatal("default port")
	}
	if SameOrigin("https://api.example.com/v1", "https://evil.example/v1") {
		t.Fatal("different host")
	}
	if SameOrigin("https://api.example.com/v1", "http://api.example.com/v1") {
		t.Fatal("different scheme")
	}
}

func TestDiscoveryAllowsEmptyModel(t *testing.T) {
	p, err := NormalizeDiscovery(Profile{Name: "本地", BaseURL: "http://127.0.0.1:1234/v1"})
	if err != nil || p.Model != "" {
		t.Fatalf("%+v %v", p, err)
	}
	if _, err := Normalize(Profile{Name: "本地", BaseURL: "http://127.0.0.1:1234/v1"}); err == nil {
		t.Fatal("chat profile without a model was accepted")
	}
}

func TestResolve(t *testing.T) {
	ep, err := Resolve("https://api.example.com/v1/")
	if err != nil {
		t.Fatal(err)
	}
	if ep.Chat != "https://api.example.com/v1/chat/completions" || ep.Models != "https://api.example.com/v1/models" {
		t.Fatalf("%+v", ep)
	}
	ep, err = Resolve("https://api.example.com/v1/chat/completions?x=1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ep.Chat, "/v1/chat/completions?x=1") || !strings.Contains(ep.Models, "/v1/models?x=1") {
		t.Fatalf("%+v", ep)
	}
	for _, raw := range []string{
		"http://169.254.169.254/latest",
		"http://[::ffff:169.254.169.254]/latest",
		"http://metadata.google.internal/computeMetadata/v1/",
		"http://user:pass@127.0.0.1:11434/v1",
		"ftp://127.0.0.1/v1",
	} {
		if _, err := Resolve(raw); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	if _, err := Resolve("http://127.0.0.1:11434/v1"); err != nil {
		t.Fatal(err)
	}
}

func TestChatAndModels(t *testing.T) {
	var sawAuth, sawKey, sawMax, sawCompletion string
	var redirects int
	var srvLater string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("redirect") == "1" {
			redirects++
			http.Redirect(w, r, srvLater, http.StatusFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		sawAuth = r.Header.Get("Authorization")
		sawKey = r.Header.Get("api-key")
		if strings.Contains(string(body), `"max_tokens"`) {
			sawMax = string(body)
		}
		if strings.Contains(string(body), `"max_completion_tokens"`) {
			sawCompletion = string(body)
		}
		if r.URL.Path == "/v1/models" {
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"data":[{"id":"demo"},{"id":"other"}]}`)
			return
		}
		if strings.Contains(string(body), `"model":"array"`) {
			io.WriteString(w, `{"choices":[{"message":{"content":[{"text":"你"},{"text":"好"}]}}]}`)
			return
		}
		if strings.Contains(string(body), `"model":"leak"`) {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, `{"error":{"message":"bad key sk-super-secret-xyz"}}`)
			return
		}
		io.WriteString(w, `{"choices":[{"message":{"content":"OK"}}]}`)
	}))
	defer srv.Close()
	srvLater = srv.URL + "/v1/chat/completions"

	client := &Client{}
	profile := Profile{
		Name: "t", BaseURL: srv.URL + "/v1", Model: "demo", APIKey: "sk-super-secret-xyz",
		Temperature: 0.2, MaxTokens: 128, AuthType: "bearer", TokenParam: "max_tokens",
	}
	text, err := client.Chat(context.Background(), profile, []Message{{Role: "user", Content: "hi"}})
	if err != nil || text != "OK" {
		t.Fatalf("%q %v", text, err)
	}
	if sawAuth != "Bearer sk-super-secret-xyz" || sawMax == "" {
		t.Fatalf("auth %q body %s", sawAuth, sawMax)
	}
	profile.AuthType = "api-key"
	profile.TokenParam = "max_completion_tokens"
	profile.Model = "array"
	text, err = client.Chat(context.Background(), profile, []Message{{Role: "user", Content: "hi"}})
	if err != nil || text != "你好" || sawKey != profile.APIKey || sawCompletion == "" {
		t.Fatalf("array %q key %q completion %s err %v", text, sawKey, sawCompletion, err)
	}
	models, err := client.ListModels(context.Background(), profile)
	if err != nil || len(models) != 2 || models[0] != "demo" {
		t.Fatalf("%v %v", models, err)
	}
	profile.Model = "leak"
	_, err = client.Chat(context.Background(), profile, []Message{{Role: "user", Content: "hi"}})
	if err == nil || strings.Contains(err.Error(), "sk-super-secret-xyz") {
		t.Fatalf("key leaked: %v", err)
	}
	profile.BaseURL = srv.URL + "/v1?redirect=1"
	profile.Model = "demo"
	_, err = client.Chat(context.Background(), profile, []Message{{Role: "user", Content: "hi"}})
	if err == nil || !strings.Contains(err.Error(), "重定向") || redirects != 1 {
		t.Fatalf("redirect err=%v count=%d", err, redirects)
	}
}

func TestExtractAndGenerate(t *testing.T) {
	raw, err := ExtractJSON("说明\n```json\n{\"title\":\"春\"}\n```")
	if err != nil || !strings.Contains(string(raw), "春") {
		t.Fatal(err, string(raw))
	}
	chat := &scripted{replies: []string{
		"不是 JSON",
		`{"title":"` + strings.Repeat("标", 30) + `","subtitle":"副","body":"正","bullets":["一","二","三","四","五"]}`,
	}}
	copy, err := GenerateCopy(context.Background(), chat, Profile{}, Input{Prompt: "春天", Width: 1080, Height: 1520})
	if err != nil {
		t.Fatal(err)
	}
	if runeLen(copy.Title) != 24 || len(copy.Bullets) != 4 {
		t.Fatalf("%+v", copy)
	}
	if chat.calls != 2 {
		t.Fatalf("calls %d", chat.calls)
	}
	assets := []Asset{{Name: "暮色", Src: "/library/backgrounds/dusk.svg", Category: "background"}}
	chat = &scripted{replies: []string{`{
		"background":"#112233",
		"layers":[
			{"type":"text","text":"标题","x":10,"y":10,"w":400,"h":80,"fontSize":40,"color":"#111111","fontFamily":"Inter"},
			{"type":"image","src":"/nope.png","x":0,"y":0,"w":100,"h":100},
			{"type":"image","src":"dusk.svg","x":0,"y":0,"w":200,"h":300}
		]
	}`}}
	layout, err := GenerateLayout(context.Background(), chat, Profile{}, Input{Prompt: "海报", Width: 1080, Height: 1520}, assets)
	if err != nil {
		t.Fatal(err)
	}
	if layout.Background != "#112233" || len(layout.Layers) != 2 {
		t.Fatalf("%+v", layout)
	}
	if layout.Layers[1].Src != "/library/backgrounds/dusk.svg" || layout.Layers[0].FontFamily != "sans" {
		b, _ := json.Marshal(layout.Layers)
		t.Fatalf("%s", b)
	}
}

func runeLen(s string) int { return len([]rune(s)) }

type scripted struct {
	replies []string
	calls   int
}

func (s *scripted) Chat(ctx context.Context, profile Profile, messages []Message) (string, error) {
	i := s.calls
	s.calls++
	if i >= len(s.replies) {
		return s.replies[len(s.replies)-1], nil
	}
	return s.replies[i], nil
}
