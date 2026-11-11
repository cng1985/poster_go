package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"poster.app/studio/internal/platform/config"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Completer 是模型调用边界。未配置密钥时 Enabled 为 false，业务可改走本地排版。
type Completer interface {
	Enabled() bool
	FallbackLocal() bool
	Provider() string
	Model() string
	Complete(ctx context.Context, messages []Message) (string, error)
}

type Client struct {
	enabled  bool
	fallback bool
	provider string
	baseURL  string
	apiKey   string
	model    string
	http     *http.Client
}

func NewCompleter(cfg *config.Config) Completer {
	timeout := cfg.Runtime.AITimeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	base := strings.TrimRight(cfg.AI.BaseURL, "/")
	client := &Client{
		enabled:  cfg.AI.Enabled && cfg.AI.APIKey != "",
		fallback: cfg.AI.FallbackLocal,
		provider: cfg.AI.Provider,
		baseURL:  base,
		apiKey:   cfg.AI.APIKey,
		model:    cfg.AI.Model,
		http: &http.Client{
			Timeout: timeout,
		},
	}
	if host := hostOf(base); host != "" {
		client.http.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= 2 || req.URL.Host != host {
				return errors.New("redirect rejected")
			}
			return nil
		}
	}
	return client
}

func (c *Client) Enabled() bool       { return c.enabled }
func (c *Client) FallbackLocal() bool { return c.fallback }
func (c *Client) Provider() string    { return c.provider }
func (c *Client) Model() string       { return c.model }

func (c *Client) Complete(ctx context.Context, messages []Message) (string, error) {
	if !c.enabled {
		return "", errors.New("model disabled")
	}
	payload := map[string]any{
		"model":       c.model,
		"messages":    messages,
		"temperature": 0.7,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("model status " + resp.Status)
	}
	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return "", err
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", errors.New("empty model content")
	}
	return decoded.Choices[0].Message.Content, nil
}

func hostOf(base string) string {
	if !strings.Contains(base, "://") {
		return ""
	}
	rest := strings.SplitN(base, "://", 2)[1]
	host := strings.SplitN(rest, "/", 2)[0]
	return host
}
