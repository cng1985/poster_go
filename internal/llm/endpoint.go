package llm

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// Endpoint is a resolved OpenAI-compatible API root.
type Endpoint struct {
	Chat   string
	Models string
}

// Resolve turns a user-supplied base URL into chat and model-list endpoints.
// The base may be an API root (https://host/v1) or a full /chat/completions URL.
func Resolve(base string) (Endpoint, error) {
	raw := strings.TrimSpace(base)
	if raw == "" || len(raw) > 500 {
		return Endpoint{}, fmt.Errorf("请填写有效的接口地址")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return Endpoint{}, fmt.Errorf("接口地址需要以 http:// 或 https:// 开头")
	}
	if u.User != nil {
		return Endpoint{}, fmt.Errorf("地址里不要包含用户名或密码")
	}
	if blockedHost(u.Hostname()) {
		return Endpoint{}, fmt.Errorf("不允许使用该地址")
	}
	path := strings.TrimRight(u.EscapedPath(), "/")
	if strings.HasSuffix(path, "/chat/completions") {
		path = strings.TrimSuffix(path, "/chat/completions")
		path = strings.TrimRight(path, "/")
	}
	chat := *u
	chat.Path = path + "/chat/completions"
	chat.RawPath = ""
	models := *u
	models.Path = path + "/models"
	models.RawPath = ""
	return Endpoint{Chat: chat.String(), Models: models.String()}, nil
}

func blockedHost(host string) bool {
	h := strings.ToLower(strings.TrimSuffix(host, "."))
	switch h {
	case "metadata.google.internal", "metadata.goog":
		return true
	}
	ip := net.ParseIP(h)
	if ip == nil {
		return false
	}
	if ip.Equal(net.ParseIP("169.254.169.254")) || ip.Equal(net.ParseIP("fd00:ec2::254")) {
		return true
	}
	return false
}
