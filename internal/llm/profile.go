package llm

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Profile is one user-configured OpenAI-compatible model.
type Profile struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	BaseURL     string  `json:"baseUrl"`
	APIKey      string  `json:"apiKey,omitempty"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxTokens"`
	AuthType    string  `json:"authType"`
	TokenParam  string  `json:"tokenParam"`
}

// Normalize checks a profile and fills defaults. The API key may be empty for local servers.
func Normalize(p Profile) (Profile, error) {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return p, fmt.Errorf("请填写配置名称")
	}
	if utf8.RuneCountInString(p.Name) > 40 {
		return p, fmt.Errorf("配置名称不能超过 40 个字")
	}
	p.BaseURL = strings.TrimSpace(p.BaseURL)
	if _, err := Resolve(p.BaseURL); err != nil {
		return p, err
	}
	p.Model = strings.TrimSpace(p.Model)
	if p.Model == "" || len(p.Model) > 128 {
		return p, fmt.Errorf("请填写模型名称")
	}
	if p.Temperature < 0 || p.Temperature > 2 {
		return p, fmt.Errorf("温度需在 0 到 2 之间")
	}
	if p.MaxTokens == 0 {
		p.MaxTokens = 2048
	}
	if p.MaxTokens < 1 || p.MaxTokens > 8192 {
		return p, fmt.Errorf("最大输出需在 1 到 8192 之间")
	}
	if p.AuthType == "" {
		p.AuthType = "bearer"
	}
	if p.AuthType != "bearer" && p.AuthType != "api-key" {
		return p, fmt.Errorf("认证方式无效")
	}
	if p.TokenParam == "" {
		p.TokenParam = "max_tokens"
	}
	if p.TokenParam != "max_tokens" && p.TokenParam != "max_completion_tokens" {
		return p, fmt.Errorf("输出长度参数无效")
	}
	p.APIKey = strings.TrimSpace(p.APIKey)
	if len(p.APIKey) > 512 {
		return p, fmt.Errorf("密钥过长")
	}
	return p, nil
}
