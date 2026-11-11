package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/cng1985/poster_go/internal/id"
	"github.com/cng1985/poster_go/internal/poster"
)

// Asset is a picture the model is allowed to place.
type Asset struct {
	Name     string
	Src      string
	Category string
}

// Input is a generation request from the editor.
type Input struct {
	Prompt string
	Width  int
	Height int
	Texts  []string
}

// Layout is a sanitized poster composition.
type Layout struct {
	Name            string         `json:"name,omitempty"`
	Background      string         `json:"background,omitempty"`
	BackgroundImage string         `json:"backgroundImage,omitempty"`
	Layers          []poster.Layer `json:"layers"`
}

// Copy is poster writing the model produced.
type Copy struct {
	Title    string   `json:"title"`
	Subtitle string   `json:"subtitle"`
	Body     string   `json:"body"`
	Bullets  []string `json:"bullets"`
}

type aiLayer struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Role        string   `json:"role"`
	X           float64  `json:"x"`
	Y           float64  `json:"y"`
	W           float64  `json:"w"`
	H           float64  `json:"h"`
	Rotation    float64  `json:"rotation"`
	Opacity     *float64 `json:"opacity"`
	Text        string   `json:"text"`
	FontSize    float64  `json:"fontSize"`
	FontFamily  string   `json:"fontFamily"`
	FontWeight  string   `json:"fontWeight"`
	Color       string   `json:"color"`
	Align       string   `json:"align"`
	LineHeight  float64  `json:"lineHeight"`
	Fill        string   `json:"fill"`
	Stroke      string   `json:"stroke"`
	StrokeWidth float64  `json:"strokeWidth"`
	Radius      float64  `json:"radius"`
	Src         string   `json:"src"`
	Fit         string   `json:"fit"`
}

type aiLayout struct {
	Name            string    `json:"name"`
	Background      string    `json:"background"`
	BackgroundImage string    `json:"backgroundImage"`
	Layers          []aiLayer `json:"layers"`
}

type aiCopy struct {
	Title    string   `json:"title"`
	Subtitle string   `json:"subtitle"`
	Body     string   `json:"body"`
	Bullets  []string `json:"bullets"`
}

// GenerateCopy asks the model for poster writing.
func GenerateCopy(ctx context.Context, chat Chatter, profile Profile, in Input) (Copy, error) {
	in = cleanInput(in)
	user := "用户描述：\n" + in.Prompt
	if extra := formatTexts(in.Texts); extra != "" {
		user += "\n\n当前画布上的文字：\n" + extra
	}
	msgs := []Message{
		{Role: "system", Content: copySystem},
		{Role: "user", Content: user},
	}
	raw, err := chatJSON(ctx, chat, profile, msgs)
	if err != nil {
		return Copy{}, err
	}
	var parsed aiCopy
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Copy{}, fmt.Errorf("模型没有返回可用的文案")
	}
	out := Copy{
		Title:    clip(parsed.Title, 24),
		Subtitle: clip(parsed.Subtitle, 36),
		Body:     clip(parsed.Body, 80),
		Bullets:  []string{},
	}
	for _, b := range parsed.Bullets {
		b = clip(b, 20)
		if b == "" {
			continue
		}
		out.Bullets = append(out.Bullets, b)
		if len(out.Bullets) == 4 {
			break
		}
	}
	if out.Title == "" && out.Subtitle == "" && out.Body == "" && len(out.Bullets) == 0 {
		return Copy{}, fmt.Errorf("模型没有返回可用的文案")
	}
	return out, nil
}

// GenerateLayout asks the model for a poster composition on the current canvas.
func GenerateLayout(ctx context.Context, chat Chatter, profile Profile, in Input, assets []Asset) (Layout, error) {
	in = cleanInput(in)
	user := fmt.Sprintf("画布宽 %d，高 %d。\n用户描述：\n%s", in.Width, in.Height, in.Prompt)
	if extra := formatTexts(in.Texts); extra != "" {
		user += "\n\n尽量保留这些已有文案：\n" + extra
	}
	msgs := []Message{
		{Role: "system", Content: layoutSystem(assets)},
		{Role: "user", Content: user},
	}
	raw, err := chatJSON(ctx, chat, profile, msgs)
	if err != nil {
		return Layout{}, err
	}
	var parsed aiLayout
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Layout{}, fmt.Errorf("模型没有返回可用的排版")
	}
	layout, err := sanitizeLayout(parsed, in.Width, in.Height, assets)
	if err != nil {
		return Layout{}, err
	}
	return layout, nil
}

func chatJSON(ctx context.Context, chat Chatter, profile Profile, msgs []Message) (json.RawMessage, error) {
	text, err := chat.Chat(ctx, profile, msgs)
	if err != nil {
		return nil, err
	}
	raw, err := ExtractJSON(text)
	if err == nil {
		return raw, nil
	}
	retry := append(append([]Message{}, msgs...),
		Message{Role: "assistant", Content: text},
		Message{Role: "user", Content: "刚才的输出无法解析。请只输出一个 JSON 对象，不要 Markdown，不要解释。"},
	)
	text, err = chat.Chat(ctx, profile, retry)
	if err != nil {
		return nil, err
	}
	raw, err = ExtractJSON(text)
	if err != nil {
		return nil, fmt.Errorf("模型没有返回可解析的 JSON")
	}
	return raw, nil
}

// ExtractJSON finds the first JSON object in a model response.
func ExtractJSON(s string) (json.RawMessage, error) {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "```"); i >= 0 {
		rest := s[i+3:]
		rest = strings.TrimPrefix(rest, "json")
		rest = strings.TrimPrefix(rest, "JSON")
		if j := strings.Index(rest, "```"); j >= 0 {
			s = rest[:j]
		} else {
			s = rest
		}
	}
	start := strings.Index(s, "{")
	if start < 0 {
		return nil, fmt.Errorf("没有 JSON")
	}
	dec := json.NewDecoder(strings.NewReader(s[start:]))
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw) == 0 || raw[0] != '{' {
		return nil, fmt.Errorf("没有 JSON 对象")
	}
	return raw, nil
}

func sanitizeLayout(in aiLayout, width, height int, assets []Asset) (Layout, error) {
	out := Layout{Layers: []poster.Layer{}}
	name := clip(in.Name, 80)
	if name != "" {
		out.Name = name
	}
	if isHexColor(in.Background) {
		out.Background = in.Background
	}
	if src := resolveSrc(in.BackgroundImage, assets); src != "" {
		out.BackgroundImage = src
	}
	seen := map[string]bool{}
	for _, raw := range in.Layers {
		if len(out.Layers) >= 12 {
			break
		}
		layer, ok := sanitizeLayer(raw, assets)
		if !ok {
			continue
		}
		if layer.ID == "" || seen[layer.ID] || !validLayerID(layer.ID) {
			layer.ID = id.New("l")
		}
		seen[layer.ID] = true
		doc := poster.Document{
			Name: "校验", Width: width, Height: height, Background: "#ffffff",
			Layers: []poster.Layer{layer},
		}
		if err := poster.Normalize(&doc, false); err != nil {
			continue
		}
		out.Layers = append(out.Layers, doc.Layers[0])
	}
	if len(out.Layers) == 0 {
		return Layout{}, fmt.Errorf("模型没有返回可用的排版")
	}
	return out, nil
}

func sanitizeLayer(raw aiLayer, assets []Asset) (poster.Layer, bool) {
	l := poster.Layer{
		Name: raw.Name, Type: strings.TrimSpace(raw.Type), Role: raw.Role,
		X: raw.X, Y: raw.Y, W: raw.W, H: raw.H, Rotation: raw.Rotation,
		Visible: true, Opacity: 1,
		Text: clip(raw.Text, 200), FontSize: raw.FontSize, FontFamily: raw.FontFamily,
		FontWeight: raw.FontWeight, Color: raw.Color, Align: raw.Align, LineHeight: raw.LineHeight,
		Fill: raw.Fill, Stroke: raw.Stroke, StrokeWidth: raw.StrokeWidth, Radius: raw.Radius,
		Src: resolveSrc(raw.Src, assets), Fit: raw.Fit,
	}
	if raw.Opacity != nil {
		l.Opacity = *raw.Opacity
	}
	l.FontFamily = mapFamily(l.FontFamily)
	switch l.Type {
	case "text":
		if strings.TrimSpace(l.Text) == "" {
			return l, false
		}
	case "image":
		if l.Src == "" {
			return l, false
		}
		if l.Role != "backdrop" {
			l.Role = "image"
		}
	case "rect", "ellipse", "line":
	default:
		return l, false
	}
	if l.W < 4 {
		l.W = 4
	}
	if l.H < 4 {
		l.H = 4
	}
	return l, true
}

func mapFamily(family string) string {
	switch strings.ToLower(strings.TrimSpace(family)) {
	case "", "sans", "sans-serif", "arial", "helvetica", "inter":
		return "sans"
	case "serif", "songti", "song", "times", "times new roman":
		return "serif"
	case "mono", "monospace", "courier":
		return "mono"
	case "kai", "kaiti", "楷体":
		return "kai"
	case "hei", "heiti", "simhei", "黑体":
		return "hei"
	default:
		return family
	}
}

func resolveSrc(src string, assets []Asset) string {
	src = strings.TrimSpace(src)
	if src == "" {
		return ""
	}
	byBase := map[string]string{}
	dup := map[string]bool{}
	for _, a := range assets {
		base := path.Base(a.Src)
		if _, ok := byBase[base]; ok {
			dup[base] = true
		}
		byBase[base] = a.Src
		if src == a.Src {
			return a.Src
		}
	}
	base := path.Base(src)
	if dup[base] {
		return ""
	}
	if full, ok := byBase[base]; ok && (base == src || strings.HasSuffix(src, "/"+base)) {
		return full
	}
	return ""
}

func cleanInput(in Input) Input {
	in.Prompt = clip(strings.TrimSpace(in.Prompt), 2000)
	if in.Width < 240 {
		in.Width = 240
	}
	if in.Width > 4096 {
		in.Width = 4096
	}
	if in.Height < 240 {
		in.Height = 240
	}
	if in.Height > 4096 {
		in.Height = 4096
	}
	if len(in.Texts) > 20 {
		in.Texts = in.Texts[:20]
	}
	for i := range in.Texts {
		in.Texts[i] = clip(in.Texts[i], 200)
	}
	return in
}

func formatTexts(texts []string) string {
	var b strings.Builder
	for _, t := range texts {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("- ")
		b.WriteString(t)
	}
	return b.String()
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return strings.TrimSpace(string(r[:n]))
}

func isHexColor(s string) bool {
	if !strings.HasPrefix(s, "#") {
		return false
	}
	if len(s) != 4 && len(s) != 7 && len(s) != 9 {
		return false
	}
	for _, r := range s[1:] {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f', r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
}

func validLayerID(id string) bool {
	if id == "" || len(id) > 40 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
		default:
			return false
		}
	}
	return true
}

const copySystem = `你是中文海报文案。只输出一个 JSON 对象，不要 Markdown，不要解释。
格式：{"title":"","subtitle":"","body":"","bullets":["",""]}
要求：title 不超过 14 个汉字，subtitle 不超过 20 个汉字，body 不超过 60 个汉字，bullets 最多 4 条、每条不超过 16 个汉字。文案适合印在海报上，具体、好读、不要空话。`

func layoutSystem(assets []Asset) string {
	var b strings.Builder
	b.WriteString(`你是海报排版助手。只输出一个 JSON 对象，不要 Markdown，不要解释。
字段：
{"name":"海报名","background":"#112233","backgroundImage":"","layers":[{"name":"","type":"text|rect|ellipse|line|image","x":0,"y":0,"w":100,"h":100,"rotation":0,"opacity":1,"text":"","fontSize":64,"fontFamily":"sans|serif|mono|kai|hei","fontWeight":"400|500|600|700|800","color":"#ffffff","align":"left|center|right","lineHeight":1.2,"fill":"#000000","stroke":"","strokeWidth":0,"radius":0,"src":"","fit":"cover|contain","role":"title|subtitle|body|kicker|price|decoration|image|backdrop"}]}
规则：
- 图层不超过 12 个，全部落在画布内并留出边距
- 文字用中文，标题醒目，层级清楚
- image 的 src 和 backgroundImage 只能从下面的列表原样复制，不确定就留空
- 背景图用 type=image、role=backdrop，铺满画布
可用图片：
`)
	for _, a := range assets {
		fmt.Fprintf(&b, "- %s （%s，%s）\n", a.Src, a.Name, a.Category)
	}
	return b.String()
}
