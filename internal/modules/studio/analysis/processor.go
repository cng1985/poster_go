package analysis

import (
	"math"
	"strconv"
	"strings"

	"poster.app/studio/internal/modules/studio/shared"
)

type outcome struct {
	Score       int
	Summary     string
	Suggestions []Suggestion
}

func Evaluate(doc shared.Document) outcome {
	score := 100
	suggestions := make([]Suggestion, 0)
	textLayers := make([]shared.Layer, 0)
	for _, layer := range doc.Layers {
		if layer.Type == "text" {
			textLayers = append(textLayers, layer)
		}
	}
	if len(textLayers) == 0 {
		score -= 25
		suggestions = append(suggestions, Suggestion{Level: "warn", Text: "海报上还没有文字，观众不知道主题。"})
	}
	if len(textLayers) > 12 {
		score -= 10
		suggestions = append(suggestions, Suggestion{Level: "warn", Text: "文字图层超过 12 个，信息有些挤。"})
	}
	if len(doc.Layers) < 2 {
		score -= 5
		suggestions = append(suggestions, Suggestion{Level: "info", Text: "图层很少，可以加一条色带或短句来建立层次。"})
	}
	largest := 0.0
	contrastHits := 0
	marginNoted := false
	margin := float64(doc.Width) * 0.04
	for _, layer := range textLayers {
		if strings.TrimSpace(layer.Text) == "" {
			score -= 5
			suggestions = append(suggestions, Suggestion{Level: "info", Text: "有一个文字图层是空的。"})
		}
		if layer.FontSize > largest {
			largest = layer.FontSize
		}
		if layer.X < -1 || layer.Y < -1 || layer.X+layer.W > float64(doc.Width)+1 || layer.Y+layer.H > float64(doc.Height)+1 {
			score -= 8
			suggestions = append(suggestions, Suggestion{Level: "warn", Text: "有文字超出画面，导出时会被裁掉。"})
		}
		if contrastRatio(layer.Color, doc.Background) < 3 {
			contrastHits++
		}
		if !marginNoted && (layer.X < margin || layer.Y < margin) {
			marginNoted = true
			score -= 3
			suggestions = append(suggestions, Suggestion{Level: "info", Text: "文字贴边了，留出一点安全边更稳。"})
		}
	}
	if contrastHits > 0 {
		deduct := 15
		if contrastHits > 1 {
			deduct = 30
		}
		score -= deduct
		suggestions = append(suggestions, Suggestion{Level: "warn", Text: "有文字和背景太接近，缩小屏幕上会看不清。"})
	}
	if len(textLayers) > 0 && largest < float64(doc.Width)/18 {
		score -= 10
		suggestions = append(suggestions, Suggestion{Level: "warn", Text: "最大的字还不够大，标题可以再突出一点。"})
	}
	if overlappingText(textLayers) {
		score -= 10
		suggestions = append(suggestions, Suggestion{Level: "warn", Text: "有文字区域互相重叠，阅读顺序会乱。"})
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	summary := "这张海报结构清楚，可以直接继续微调。"
	switch {
	case score < 60:
		summary = "这张海报还有几处会挡住阅读，建议先改对比和标题。"
	case score < 85:
		summary = "整体可用，按下面的提示收一收会更稳。"
	}
	if len(suggestions) == 0 {
		suggestions = append(suggestions, Suggestion{Level: "info", Text: "没有发现明显的排版问题。"})
	}
	return outcome{Score: score, Summary: summary, Suggestions: suggestions}
}

func overlappingText(layers []shared.Layer) bool {
	for i := 0; i < len(layers); i++ {
		for j := i + 1; j < len(layers); j++ {
			if strings.TrimSpace(layers[i].Text) == "" || strings.TrimSpace(layers[j].Text) == "" {
				continue
			}
			if intersect(layers[i], layers[j]) {
				return true
			}
		}
	}
	return false
}

func intersect(a, b shared.Layer) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X && a.Y < b.Y+b.H && a.Y+a.H > b.Y
}

func contrastRatio(foreground, background string) float64 {
	l1 := relativeLuminance(foreground)
	l2 := relativeLuminance(background)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func relativeLuminance(color string) float64 {
	r, g, b := parseHex(color)
	channel := func(value float64) float64 {
		value = value / 255
		if value <= 0.03928 {
			return value / 12.92
		}
		return math.Pow((value+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(r) + 0.7152*channel(g) + 0.0722*channel(b)
}

func parseHex(color string) (float64, float64, float64) {
	value := strings.TrimPrefix(color, "#")
	if len(value) == 3 {
		value = string([]byte{value[0], value[0], value[1], value[1], value[2], value[2]})
	}
	if len(value) < 6 {
		return 0, 0, 0
	}
	r, _ := strconv.ParseInt(value[0:2], 16, 64)
	g, _ := strconv.ParseInt(value[2:4], 16, 64)
	b, _ := strconv.ParseInt(value[4:6], 16, 64)
	return float64(r), float64(g), float64(b)
}
