package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/cng1985/poster_go/internal/id"
	"github.com/cng1985/poster_go/internal/llm"
	"github.com/cng1985/poster_go/internal/store"
)

type profileInput struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	BaseURL     string   `json:"baseUrl"`
	APIKey      string   `json:"apiKey"`
	ClearAPIKey bool     `json:"clearApiKey"`
	Model       string   `json:"model"`
	Temperature *float64 `json:"temperature"`
	MaxTokens   *int     `json:"maxTokens"`
	AuthType    string   `json:"authType"`
	TokenParam  string   `json:"tokenParam"`
	Activate    bool     `json:"activate"`
}

type publicProfile struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	BaseURL     string  `json:"baseUrl"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxTokens"`
	AuthType    string  `json:"authType"`
	TokenParam  string  `json:"tokenParam"`
	APIKeySet   bool    `json:"apiKeySet"`
}

func toPublic(p llm.Profile) publicProfile {
	return publicProfile{
		ID: p.ID, Name: p.Name, BaseURL: p.BaseURL, Model: p.Model,
		Temperature: p.Temperature, MaxTokens: p.MaxTokens, AuthType: p.AuthType,
		TokenParam: p.TokenParam, APIKeySet: p.APIKey != "",
	}
}

func (s *Server) listProfiles(w http.ResponseWriter, r *http.Request) {
	file, err := s.Store.LoadLLM()
	if err != nil {
		mapErr(w, err, "配置不存在")
		return
	}
	out := make([]publicProfile, 0, len(file.Profiles))
	for _, p := range file.Profiles {
		out = append(out, toPublic(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"activeProfileId": file.ActiveProfileID,
		"profiles":        out,
	})
}

func (s *Server) createProfile(w http.ResponseWriter, r *http.Request) {
	in, err := readProfileInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	in.ID = ""
	var created publicProfile
	err = s.Store.UpdateLLM(func(file *store.LLMFile) error {
		if len(file.Profiles) >= 20 {
			return &store.Error{Status: 400, Msg: "最多保存 20 个模型配置"}
		}
		profile, err := s.mergeProfile(*file, in, true)
		if err != nil {
			return err
		}
		profile.ID = id.New("m")
		file.Profiles = append(file.Profiles, profile)
		if in.Activate || file.ActiveProfileID == "" {
			file.ActiveProfileID = profile.ID
		}
		created = toPublic(profile)
		return nil
	})
	if err != nil {
		writeProfileErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request) {
	in, err := readProfileInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	in.ID = r.PathValue("id")
	var updated publicProfile
	active := ""
	err = s.Store.UpdateLLM(func(file *store.LLMFile) error {
		profile, err := s.mergeProfile(*file, in, true)
		if err != nil {
			return err
		}
		found := false
		for i := range file.Profiles {
			if file.Profiles[i].ID == in.ID {
				profile.ID = in.ID
				file.Profiles[i] = profile
				found = true
				break
			}
		}
		if !found {
			return store.ErrNotFound
		}
		if in.Activate {
			file.ActiveProfileID = in.ID
		}
		updated = toPublic(profile)
		active = file.ActiveProfileID
		return nil
	})
	if err != nil {
		writeProfileErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"profile": updated, "activeProfileId": active})
}

func (s *Server) deleteProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	active := ""
	err := s.Store.UpdateLLM(func(file *store.LLMFile) error {
		orig := append([]llm.Profile(nil), file.Profiles...)
		next := make([]llm.Profile, 0, len(orig))
		found := false
		for _, p := range orig {
			if p.ID == id {
				found = true
				continue
			}
			next = append(next, p)
		}
		if !found {
			return store.ErrNotFound
		}
		file.Profiles = next
		if file.ActiveProfileID == id {
			file.ActiveProfileID = ""
			if len(next) > 0 {
				file.ActiveProfileID = next[0].ID
			}
		}
		active = file.ActiveProfileID
		return nil
	})
	if err != nil {
		writeProfileErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "activeProfileId": active})
}

func (s *Server) activateProfile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := s.Store.UpdateLLM(func(file *store.LLMFile) error {
		for _, p := range file.Profiles {
			if p.ID == id {
				file.ActiveProfileID = id
				return nil
			}
		}
		return store.ErrNotFound
	})
	if err != nil {
		writeProfileErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"activeProfileId": id})
}

func (s *Server) testProfile(w http.ResponseWriter, r *http.Request) {
	in, err := readProfileInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	file, err := s.Store.LoadLLM()
	if err != nil {
		mapErr(w, err, "配置不存在")
		return
	}
	profile, err := s.mergeProfile(file, in, false)
	if err != nil {
		writeProfileErr(w, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	reply, err := s.LLM.Chat(ctx, profile, []llm.Message{
		{Role: "system", Content: "You are a health check. Reply with the single word OK."},
		{Role: "user", Content: "Reply with the single word OK."},
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "reply": clipRunes(reply, 200)})
}

func (s *Server) listModels(w http.ResponseWriter, r *http.Request) {
	in, err := readProfileInput(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	file, err := s.Store.LoadLLM()
	if err != nil {
		mapErr(w, err, "配置不存在")
		return
	}
	profile, err := s.mergeProfile(file, in, false)
	if err != nil {
		writeProfileErr(w, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	models, err := s.LLM.ListModels(ctx, profile)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if models == nil {
		models = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"models": models})
}

func (s *Server) generate(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r, 1<<20)
	if err != nil {
		writeError(w, http.StatusBadRequest, "请求体过大或无法读取")
		return
	}
	var req struct {
		Mode   string   `json:"mode"`
		Prompt string   `json:"prompt"`
		Width  int      `json:"width"`
		Height int      `json:"height"`
		Texts  []string `json:"texts"`
	}
	if err := decodeJSON(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求无法解析")
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, http.StatusBadRequest, "请先描述海报内容")
		return
	}
	file, err := s.Store.LoadLLM()
	if err != nil {
		mapErr(w, err, "配置不存在")
		return
	}
	var profile llm.Profile
	found := false
	for _, p := range file.Profiles {
		if p.ID == file.ActiveProfileID {
			profile = p
			found = true
			break
		}
	}
	if !found {
		writeError(w, http.StatusBadRequest, "请先配置并启用一个大模型")
		return
	}
	in := llm.Input{Prompt: req.Prompt, Width: req.Width, Height: req.Height, Texts: req.Texts}
	ctx, cancel := context.WithTimeout(r.Context(), 80*time.Second)
	defer cancel()
	switch req.Mode {
	case "copy":
		copy, err := llm.GenerateCopy(ctx, s.LLM, profile, in)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"mode": "copy", "model": profile.Model, "profileName": profile.Name, "copy": copy,
		})
	case "layout":
		layout, err := llm.GenerateLayout(ctx, s.LLM, profile, in, s.assets())
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"mode": "layout", "model": profile.Model, "profileName": profile.Name, "layout": layout,
		})
	default:
		writeError(w, http.StatusBadRequest, "不支持的生成方式")
	}
}

func (s *Server) assets() []llm.Asset {
	out := make([]llm.Asset, 0, len(s.Catalog.Items))
	for _, it := range s.Catalog.Items {
		out = append(out, llm.Asset{Name: it.Name, Src: it.Src, Category: it.Category})
	}
	uploads, err := s.Store.Uploads()
	if err != nil {
		return out
	}
	for _, up := range uploads {
		out = append(out, llm.Asset{Name: up.Name, Src: "/uploads/" + up.Filename, Category: "upload"})
	}
	return out
}

func (s *Server) mergeProfile(file store.LLMFile, in profileInput, overwrite bool) (llm.Profile, error) {
	base := llm.Profile{Temperature: 0.7, MaxTokens: 2048, AuthType: "bearer", TokenParam: "max_tokens"}
	if in.ID != "" {
		found := false
		for _, p := range file.Profiles {
			if p.ID == in.ID {
				base = p
				found = true
				break
			}
		}
		if !found {
			return llm.Profile{}, store.ErrNotFound
		}
	}
	if overwrite || strings.TrimSpace(in.Name) != "" {
		if overwrite || strings.TrimSpace(in.Name) != "" {
			base.Name = strings.TrimSpace(in.Name)
		}
	}
	if !overwrite && base.Name == "" {
		base.Name = "未命名模型"
	}
	if overwrite || strings.TrimSpace(in.BaseURL) != "" {
		base.BaseURL = strings.TrimSpace(in.BaseURL)
	}
	if overwrite || strings.TrimSpace(in.Model) != "" {
		base.Model = strings.TrimSpace(in.Model)
	}
	if in.Temperature != nil {
		base.Temperature = *in.Temperature
	}
	if in.MaxTokens != nil {
		base.MaxTokens = *in.MaxTokens
	}
	if in.AuthType != "" {
		base.AuthType = in.AuthType
	}
	if in.TokenParam != "" {
		base.TokenParam = in.TokenParam
	}
	if in.ClearAPIKey {
		base.APIKey = ""
	}
	key := strings.TrimSpace(in.APIKey)
	if key != "" && !strings.Contains(key, "•") {
		base.APIKey = key
	}
	return llm.Normalize(base)
}

func readProfileInput(r *http.Request) (profileInput, error) {
	body, err := readBody(r, 1<<20)
	if err != nil {
		return profileInput{}, errString("请求体过大或无法读取")
	}
	var in profileInput
	if err := decodeJSON(body, &in); err != nil && err != errEmpty {
		return profileInput{}, errString("配置数据无法解析")
	}
	return in, nil
}

type errString string

func (e errString) Error() string { return string(e) }

func writeProfileErr(w http.ResponseWriter, err error) {
	if _, ok := err.(errString); ok {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err.Error() == "请填写配置名称" || err.Error() == "请填写模型名称" ||
		strings.Contains(err.Error(), "接口") || strings.Contains(err.Error(), "地址") ||
		strings.Contains(err.Error(), "温度") || strings.Contains(err.Error(), "最大输出") ||
		strings.Contains(err.Error(), "认证") || strings.Contains(err.Error(), "密钥") ||
		strings.Contains(err.Error(), "输出长度") || strings.Contains(err.Error(), "配置名称") ||
		strings.Contains(err.Error(), "模型") {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	mapErr(w, err, "配置不存在")
}

func clipRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
