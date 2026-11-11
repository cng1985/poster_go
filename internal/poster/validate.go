package poster

import (
	"math"
	"regexp"
	"strings"
	"unicode/utf8"
)

// InvalidError is a user-facing document error.
type InvalidError struct {
	Msg string
}

func (e *InvalidError) Error() string { return e.Msg }

func invalid(msg string) error { return &InvalidError{Msg: msg} }

var (
	colorRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
	assetRe = regexp.MustCompile(`^/(library|uploads)/[A-Za-z0-9_./-]+$`)
	idRe    = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}$`)
)

var (
	fontFamilies = map[string]bool{"sans": true, "serif": true, "mono": true, "kai": true, "hei": true}
	fontWeights  = map[string]string{
		"400": "400", "500": "500", "600": "600", "700": "700", "800": "800",
		"normal": "400", "bold": "700",
	}
	aligns = map[string]bool{"left": true, "center": true, "right": true}
	fits   = map[string]bool{"cover": true, "contain": true}
	types  = map[string]bool{"text": true, "image": true, "rect": true, "ellipse": true, "line": true}
	roles  = map[string]bool{
		"": true, "title": true, "subtitle": true, "body": true, "decoration": true,
		"image": true, "price": true, "kicker": true, "backdrop": true,
	}
)

// Normalize checks a document and fills safe defaults.
// When requireID is false, an empty document id is allowed (templates).
func Normalize(doc *Document, requireID bool) error {
	if doc == nil {
		return invalid("海报内容为空")
	}
	if requireID && !idRe.MatchString(doc.ID) {
		return invalid("海报编号无效")
	}
	if !requireID && doc.ID != "" && !idRe.MatchString(doc.ID) {
		return invalid("海报编号无效")
	}
	doc.Name = strings.TrimSpace(doc.Name)
	if doc.Name == "" {
		doc.Name = "未命名海报"
	}
	if utf8.RuneCountInString(doc.Name) > 80 {
		return invalid("海报名称不能超过 80 个字")
	}
	if doc.Width < 240 || doc.Width > 4096 || doc.Height < 240 || doc.Height > 4096 {
		return invalid("画布尺寸需在 240 到 4096 之间")
	}
	if !colorRe.MatchString(doc.Background) {
		return invalid("背景色格式无效")
	}
	if err := checkAsset(doc.BackgroundImage); err != nil {
		return err
	}
	if doc.Layers == nil {
		doc.Layers = []Layer{}
	}
	if len(doc.Layers) > 80 {
		return invalid("图层不能超过 80 个")
	}
	seen := map[string]bool{}
	for i := range doc.Layers {
		if err := normalizeLayer(&doc.Layers[i]); err != nil {
			return err
		}
		if seen[doc.Layers[i].ID] {
			return invalid("图层编号重复")
		}
		seen[doc.Layers[i].ID] = true
	}
	return nil
}

func normalizeLayer(l *Layer) error {
	if !idRe.MatchString(l.ID) {
		return invalid("图层编号无效")
	}
	if !types[l.Type] {
		return invalid("不支持的图层类型")
	}
	l.Name = strings.TrimSpace(l.Name)
	if l.Name == "" {
		l.Name = defaultLayerName(l.Type)
	}
	if utf8.RuneCountInString(l.Name) > 40 {
		return invalid("图层名称不能超过 40 个字")
	}
	if !roles[l.Role] {
		return invalid("图层角色无效")
	}
	if !finite(l.X) || !finite(l.Y) || !finite(l.W) || !finite(l.H) || !finite(l.Rotation) || !finite(l.Opacity) {
		return invalid("图层数值无效")
	}
	if l.X < -2000 || l.X > 8000 || l.Y < -2000 || l.Y > 8000 {
		return invalid("图层位置超出范围")
	}
	if l.W < 1 || l.W > 4096 || l.H < 1 || l.H > 4096 {
		return invalid("图层尺寸超出范围")
	}
	l.Rotation = clamp(l.Rotation, -360, 360)
	l.Opacity = clamp(l.Opacity, 0, 1)
	l.StrokeWidth = clamp(l.StrokeWidth, 0, 64)
	l.Radius = clamp(l.Radius, 0, 2048)
	if l.Color != "" && !validColor(l.Color) {
		return invalid("文字颜色无效")
	}
	if l.Fill != "" && l.Fill != "transparent" && !validColor(l.Fill) {
		return invalid("填充色无效")
	}
	if l.Stroke != "" && !validColor(l.Stroke) {
		return invalid("描边颜色无效")
	}
	if err := checkAsset(l.Src); err != nil {
		return err
	}
	switch l.Type {
	case "text":
		l.Text = strings.ReplaceAll(l.Text, "\r\n", "\n")
		if utf8.RuneCountInString(l.Text) > 2000 {
			return invalid("文字不能超过 2000 个字")
		}
		if l.FontSize == 0 {
			l.FontSize = 32
		}
		l.FontSize = clamp(l.FontSize, 8, 400)
		if l.FontFamily == "" {
			l.FontFamily = "sans"
		}
		if !fontFamilies[l.FontFamily] {
			return invalid("字体无效")
		}
		if l.FontWeight == "" {
			l.FontWeight = "400"
		}
		mapped, ok := fontWeights[l.FontWeight]
		if !ok {
			return invalid("字重无效")
		}
		l.FontWeight = mapped
		if l.Align == "" {
			l.Align = "left"
		}
		if !aligns[l.Align] {
			return invalid("对齐方式无效")
		}
		if l.Color == "" {
			l.Color = "#1a1a1a"
		}
		if l.LineHeight == 0 {
			l.LineHeight = 1.25
		}
		l.LineHeight = clamp(l.LineHeight, 0.8, 3)
		l.Src = ""
		l.Fit = ""
	case "image":
		if l.Fit == "" {
			l.Fit = "cover"
		}
		if !fits[l.Fit] {
			return invalid("图片适配方式无效")
		}
		l.Text = ""
	default:
		l.Src = ""
		l.Fit = ""
		l.Text = ""
	}
	return nil
}

func defaultLayerName(typ string) string {
	switch typ {
	case "text":
		return "文本"
	case "image":
		return "图片"
	case "ellipse":
		return "圆形"
	case "line":
		return "直线"
	default:
		return "矩形"
	}
}

func checkAsset(src string) error {
	if src == "" {
		return nil
	}
	if strings.Contains(src, "..") || strings.Contains(src, "//") || strings.Contains(src, "\\") {
		return invalid("图片地址无效")
	}
	if !assetRe.MatchString(src) {
		return invalid("图片地址无效")
	}
	return nil
}

func validColor(c string) bool { return colorRe.MatchString(c) }

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ValidAssetPath reports whether src is an empty or local library/upload path.
func ValidAssetPath(src string) bool { return checkAsset(src) == nil }
