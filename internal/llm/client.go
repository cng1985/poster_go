package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

// Message is one chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Chatter sends a chat completion request.
type Chatter interface {
	Chat(ctx context.Context, profile Profile, messages []Message) (string, error)
}

// Client talks to an OpenAI-compatible HTTP endpoint.
type Client struct {
	HTTP *http.Client
}

func (c *Client) httpClient() *http.Client {
	if c != nil && c.HTTP != nil {
		return c.HTTP
	}
	return defaultHTTP
}

var defaultHTTP = &http.Client{
	Timeout: 100 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type chatBody struct {
	Model               string    `json:"model"`
	Messages            []Message `json:"messages"`
	Temperature         float64   `json:"temperature"`
	Stream              bool      `json:"stream"`
	MaxTokens           int       `json:"max_tokens,omitempty"`
	MaxCompletionTokens int       `json:"max_completion_tokens,omitempty"`
}

// Chat calls POST {base}/chat/completions and returns the assistant text.
func (c *Client) Chat(ctx context.Context, profile Profile, messages []Message) (string, error) {
	ep, err := Resolve(profile.BaseURL)
	if err != nil {
		return "", err
	}
	body := chatBody{
		Model:       profile.Model,
		Messages:    messages,
		Temperature: profile.Temperature,
		Stream:      false,
	}
	switch profile.TokenParam {
	case "max_completion_tokens":
		body.MaxCompletionTokens = profile.MaxTokens
	default:
		body.MaxTokens = profile.MaxTokens
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error json.RawMessage `json:"error"`
	}
	if err := c.do(ctx, profile, http.MethodPost, ep.Chat, body, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("模型没有返回内容")
	}
	text, err := parseContent(parsed.Choices[0].Message.Content)
	if err != nil {
		return "", err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("模型没有返回内容")
	}
	return text, nil
}

// ListModels calls GET {base}/models.
func (c *Client) ListModels(ctx context.Context, profile Profile) ([]string, error) {
	ep, err := Resolve(profile.BaseURL)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		Data   json.RawMessage `json:"data"`
		Models json.RawMessage `json:"models"`
	}
	if err := c.do(ctx, profile, http.MethodGet, ep.Models, nil, &parsed); err != nil {
		return nil, err
	}
	ids := append(readModelIDs(parsed.Data), readModelIDs(parsed.Models)...)
	seen := map[string]bool{}
	var out []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
		if len(out) >= 200 {
			break
		}
	}
	return out, nil
}

func (c *Client) do(ctx context.Context, profile Profile, method, endpoint string, payload any, dest any) error {
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return fmt.Errorf("无法创建模型请求")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "PosterGo/0.1")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	applyAuth(req, profile)
	resp, err := c.httpClient().Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errors.New("模型响应超时")
		}
		return fmt.Errorf("无法连接模型接口：%s", redact(err.Error(), profile.APIKey))
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return fmt.Errorf("读取模型响应失败")
	}
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return errors.New("接口返回了重定向，请填写最终地址")
	}
	if resp.StatusCode >= 300 {
		msg := errorMessage(raw)
		if msg == "" {
			msg = "请求失败"
		}
		return fmt.Errorf("模型接口返回 %d：%s", resp.StatusCode, truncate(redact(msg, profile.APIKey), 300))
	}
	if dest == nil {
		return nil
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return errors.New("模型接口没有返回 JSON")
	}
	return nil
}

func applyAuth(req *http.Request, profile Profile) {
	key := strings.TrimSpace(profile.APIKey)
	if key == "" {
		return
	}
	if profile.AuthType == "api-key" {
		req.Header.Set("api-key", key)
		return
	}
	req.Header.Set("Authorization", "Bearer "+key)
}

func parseContent(raw json.RawMessage) (string, error) {
	if len(bytes.TrimSpace(raw)) == 0 || string(raw) == "null" {
		return "", nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text, nil
	}
	var parts []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		var b strings.Builder
		for _, p := range parts {
			b.WriteString(p.Text)
		}
		return b.String(), nil
	}
	return "", errors.New("无法解析模型返回内容")
}

func errorMessage(raw []byte) string {
	var wrap struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(raw, &wrap) == nil && len(wrap.Error) > 0 {
		var asString string
		if json.Unmarshal(wrap.Error, &asString) == nil {
			return asString
		}
		var obj struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(wrap.Error, &obj) == nil && obj.Message != "" {
			return obj.Message
		}
	}
	s := strings.TrimSpace(string(raw))
	if strings.HasPrefix(s, "<") {
		return "服务返回了网页而不是 JSON"
	}
	return s
}

func readModelIDs(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var strs []string
	if json.Unmarshal(raw, &strs) == nil && len(strs) > 0 {
		return strs
	}
	var objs []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if json.Unmarshal(raw, &objs) == nil {
		out := make([]string, 0, len(objs))
		for _, o := range objs {
			if o.ID != "" {
				out = append(out, o.ID)
			} else if o.Name != "" {
				out = append(out, o.Name)
			}
		}
		return out
	}
	return nil
}

func redact(msg, key string) string {
	if key != "" {
		msg = strings.ReplaceAll(msg, key, "***")
	}
	return msg
}

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max]) + "…"
}
