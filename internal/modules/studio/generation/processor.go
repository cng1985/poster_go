package generation

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"poster.app/studio/internal/modules/studio/shared"
	"poster.app/studio/internal/platform/ai"
	apperrors "poster.app/studio/internal/platform/errors"
)

type palette struct {
	background string
	ink        string
	muted      string
	accent     string
}

var presets = map[string][2]int{
	"story":  {1080, 1920},
	"poster": {1080, 1440},
	"square": {1080, 1080},
	"cover":  {1920, 1080},
}

var palettes = map[string]palette{
	"editorial": {background: "#221c17", ink: "#f3ecdf", muted: "#d9cbb8", accent: "#e4572e"},
	"sale":      {background: "#6e1520", ink: "#fff8ea", muted: "#ffd9a8", accent: "#f2c14e"},
	"festive":   {background: "#4a1420", ink: "#f8ecd4", muted: "#f0d7a2", accent: "#e7c27d"},
	"minimal":   {background: "#f4efe6", ink: "#1a1612", muted: "#3d3832", accent: "#e4572e"},
	"bold":      {background: "#101820", ink: "#f4fff8", muted: "#b7f3d4", accent: "#3ddc97"},
}

func knownStyle(style string) bool {
	_, ok := palettes[style]
	return ok
}

func knownPreset(preset string) bool {
	_, ok := presets[preset]
	return ok
}

func Messages(prompt, style, preset string) []ai.Message {
	size := presets[preset]
	system := strings.TrimSpace(`
你是海报版式设计师。只输出一个 JSON 对象，不要使用 Markdown。
字段为 width、height、background、layers。
layer.type 只能是 text、rect、ellipse，不要输出 image，不要输出任何网址。
每个 layer 必须包含 id、type、name、x、y、w、h、rotation、opacity。
text 层还要包含 text、font_size、font_weight、font_family、color、align。
rect 与 ellipse 还要包含 fill，rect 可以包含 radius。
font_family 只能是 serif 或 sans。align 只能是 left、center、right。font_weight 只能是 400、500、600、700、800。
颜色使用 #RRGGBB。opacity 取 0 到 1。文案使用简体中文，并贴合用户主题。
不要生成违法、仇恨或色情内容；如果需求不适合做海报，就改成一张礼貌拒绝并建议换主题的海报。
`)
	user := "尺寸 " + itoa(size[0]) + " x " + itoa(size[1]) + "，风格 " + style + "。需求：" + prompt
	return []ai.Message{{Role: "system", Content: system}, {Role: "user", Content: user}}
}

func ParseModelDocument(text string) (shared.Document, error) {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return shared.Document{}, apperrors.Unavailable("模型没有返回海报")
	}
	var doc shared.Document
	if err := json.Unmarshal([]byte(text[start:end+1]), &doc); err != nil {
		return shared.Document{}, apperrors.Unavailable("模型返回的海报无法解析")
	}
	kept := make([]shared.Layer, 0, len(doc.Layers))
	for _, layer := range doc.Layers {
		if layer.Type == "image" {
			continue
		}
		kept = append(kept, layer)
	}
	doc.Layers = kept
	return doc, nil
}

func ComposeLocal(prompt, style, preset string) shared.Document {
	size := presets[preset]
	colors := palettes[style]
	title := firstClause(prompt)
	subtitle := prompt
	if subtitle == title {
		subtitle = "由本地排版根据你的描述生成"
	}
	width, height := float64(size[0]), float64(size[1])
	margin := width * 0.08
	titleSize := width / 10
	if titleSize > 120 {
		titleSize = 120
	}
	if titleSize < 48 {
		titleSize = 48
	}
	layers := []shared.Layer{
		{ID: "accent", Type: "rect", Name: "accent", X: margin, Y: height * 0.16, W: width * 0.18, H: 10, Opacity: 1, Fill: colors.accent, Radius: 4},
		{ID: "kicker", Type: "text", Name: "kicker", X: margin, Y: height * 0.2, W: width * 0.8, H: 70, Opacity: 1, Text: styleLabel(style), FontSize: 28, FontWeight: 700, FontFamily: "sans", Color: colors.accent, Align: "left"},
		{ID: "title", Type: "text", Name: "title", X: margin, Y: height * 0.28, W: width - margin*2, H: titleSize * 2.4, Opacity: 1, Text: title, FontSize: titleSize, FontWeight: 700, FontFamily: "serif", Color: colors.ink, Align: "left"},
		{ID: "sub", Type: "text", Name: "sub", X: margin, Y: height * 0.52, W: width - margin*2, H: 220, Opacity: 1, Text: trimRunes(subtitle, 80), FontSize: 32, FontWeight: 500, FontFamily: "sans", Color: colors.muted, Align: "left"},
		{ID: "foot", Type: "text", Name: "foot", X: margin, Y: height - 160, W: width - margin*2, H: 60, Opacity: 1, Text: "海报工坊", FontSize: 24, FontWeight: 600, FontFamily: "sans", Color: colors.accent, Align: "left"},
	}
	return shared.Document{Width: size[0], Height: size[1], Background: colors.background, Layers: layers}
}

func styleLabel(style string) string {
	switch style {
	case "sale":
		return "促销"
	case "festive":
		return "节日"
	case "minimal":
		return "留白"
	case "bold":
		return "强对比"
	default:
		return "编辑感"
	}
}

func firstClause(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	for _, sep := range []string{"。", "！", "？", "\n", ".", "!"} {
		if index := strings.Index(prompt, sep); index > 0 {
			prompt = prompt[:index]
			break
		}
	}
	prompt = trimRunes(prompt, 16)
	if prompt == "" {
		return "一张新海报"
	}
	return prompt
}

func trimRunes(value string, max int) string {
	if utf8.RuneCountInString(value) <= max {
		return value
	}
	count := 0
	for index := range value {
		if count == max {
			return value[:index]
		}
		count++
	}
	return value
}

func itoa(value int) string {
	return strings.TrimPrefix(strings.Replace(jsonNumber(value), ",", "", -1), "")
}

func jsonNumber(value int) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}
