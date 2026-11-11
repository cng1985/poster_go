package shared

import (
	"encoding/json"
	"strconv"
	"unicode/utf8"

	apperrors "poster.app/studio/internal/platform/errors"
)

// Document 是编辑器、模板、生成和分析共用的海报稿。
// 图层整体替换，不拆成独立明细表。
type Document struct {
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Background string  `json:"background"`
	Layers     []Layer `json:"layers"`
}

type Layer struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	W          float64 `json:"w"`
	H          float64 `json:"h"`
	Rotation   float64 `json:"rotation"`
	Opacity    float64 `json:"opacity"`
	Text       string  `json:"text,omitempty"`
	FontSize   float64 `json:"font_size,omitempty"`
	FontWeight int     `json:"font_weight,omitempty"`
	FontFamily string  `json:"font_family,omitempty"`
	Color      string  `json:"color,omitempty"`
	Align      string  `json:"align,omitempty"`
	Fill       string  `json:"fill,omitempty"`
	Radius     float64 `json:"radius,omitempty"`
	AssetID    int64   `json:"asset_id,string,omitempty"`
}

func (d *Document) UnmarshalJSON(data []byte) error {
	var aux struct {
		Width      int               `json:"width"`
		Height     int               `json:"height"`
		Background string            `json:"background"`
		Layers     []json.RawMessage `json:"layers"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	d.Width = aux.Width
	d.Height = aux.Height
	d.Background = aux.Background
	d.Layers = make([]Layer, 0, len(aux.Layers))
	for _, raw := range aux.Layers {
		layer, err := decodeLayer(raw)
		if err != nil {
			return err
		}
		d.Layers = append(d.Layers, layer)
	}
	return nil
}

func decodeLayer(raw []byte) (Layer, error) {
	var aux struct {
		ID         string          `json:"id"`
		Type       string          `json:"type"`
		Name       string          `json:"name"`
		X          float64         `json:"x"`
		Y          float64         `json:"y"`
		W          float64         `json:"w"`
		H          float64         `json:"h"`
		Rotation   float64         `json:"rotation"`
		Opacity    *float64        `json:"opacity"`
		Text       string          `json:"text"`
		FontSize   float64         `json:"font_size"`
		FontWeight int             `json:"font_weight"`
		FontFamily string          `json:"font_family"`
		Color      string          `json:"color"`
		Align      string          `json:"align"`
		Fill       string          `json:"fill"`
		Radius     float64         `json:"radius"`
		AssetID    json.RawMessage `json:"asset_id"`
	}
	if err := json.Unmarshal(raw, &aux); err != nil {
		return Layer{}, err
	}
	layer := Layer{
		ID: aux.ID, Type: aux.Type, Name: aux.Name,
		X: aux.X, Y: aux.Y, W: aux.W, H: aux.H, Rotation: aux.Rotation,
		Text: aux.Text, FontSize: aux.FontSize, FontWeight: aux.FontWeight,
		FontFamily: aux.FontFamily, Color: aux.Color, Align: aux.Align,
		Fill: aux.Fill, Radius: aux.Radius,
	}
	if aux.Opacity == nil {
		layer.Opacity = 1
	} else {
		layer.Opacity = *aux.Opacity
	}
	assetID, err := parseAssetID(aux.AssetID)
	if err != nil {
		return Layer{}, err
	}
	layer.AssetID = assetID
	return layer, nil
}

func parseAssetID(raw json.RawMessage) (int64, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		if text == "" {
			return 0, nil
		}
		id, err := strconv.ParseInt(text, 10, 64)
		if err != nil || id < 1 {
			return 0, errInvalidAsset
		}
		return id, nil
	}
	var number int64
	if err := json.Unmarshal(raw, &number); err != nil || number < 1 {
		return 0, errInvalidAsset
	}
	return number, nil
}

var errInvalidAsset = apperrors.Invalid("素材标识不正确")

func Normalize(doc *Document) {
	if doc.Layers == nil {
		doc.Layers = []Layer{}
	}
	for i := range doc.Layers {
		layer := &doc.Layers[i]
		if layer.Name == "" {
			layer.Name = layer.Type
		}
		if layer.FontFamily == "" {
			layer.FontFamily = "serif"
		}
		if layer.Align == "" {
			layer.Align = "left"
		}
		if layer.Type == "text" && layer.FontSize == 0 {
			layer.FontSize = 64
		}
		if layer.Type == "text" && layer.FontWeight == 0 {
			layer.FontWeight = 600
		}
		if layer.Type == "text" && layer.Color == "" {
			layer.Color = "#f4efe6"
		}
		if (layer.Type == "rect" || layer.Type == "ellipse") && layer.Fill == "" {
			layer.Fill = "#e4572e"
		}
	}
}

func Validate(doc *Document) error {
	if doc == nil {
		return apperrors.Invalid("海报内容不能为空")
	}
	if doc.Width < 320 || doc.Width > 4096 || doc.Height < 320 || doc.Height > 4096 {
		return apperrors.Invalid("海报尺寸需要在 320 到 4096 之间")
	}
	if !colorOK(doc.Background) {
		return apperrors.Invalid("背景色不正确")
	}
	if len(doc.Layers) > 80 {
		return apperrors.Invalid("图层不能超过 80 个")
	}
	seen := map[string]struct{}{}
	for i := range doc.Layers {
		if err := validateLayer(&doc.Layers[i], seen); err != nil {
			return err
		}
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return apperrors.Invalid("海报内容无法保存")
	}
	if len(raw) > 400000 {
		return apperrors.TooLarge("海报内容过大")
	}
	return nil
}

func Prepare(doc *Document) (string, error) {
	Normalize(doc)
	if err := Validate(doc); err != nil {
		return "", err
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return "", apperrors.Invalid("海报内容无法保存")
	}
	return string(raw), nil
}

func Parse(raw string) (Document, error) {
	var doc Document
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		if appErr, ok := apperrors.As(err); ok {
			return Document{}, appErr
		}
		return Document{}, apperrors.Invalid("海报内容格式不正确")
	}
	Normalize(&doc)
	if err := Validate(&doc); err != nil {
		return Document{}, err
	}
	return doc, nil
}

func AssetIDs(doc Document) []int64 {
	ids := make([]int64, 0)
	seen := map[int64]struct{}{}
	for _, layer := range doc.Layers {
		if layer.Type == "image" && layer.AssetID > 0 {
			if _, ok := seen[layer.AssetID]; ok {
				continue
			}
			seen[layer.AssetID] = struct{}{}
			ids = append(ids, layer.AssetID)
		}
	}
	return ids
}

func validateLayer(layer *Layer, seen map[string]struct{}) error {
	if !idOK(layer.ID) {
		return apperrors.Invalid("图层标识不正确")
	}
	if _, ok := seen[layer.ID]; ok {
		return apperrors.Invalid("图层标识重复")
	}
	seen[layer.ID] = struct{}{}
	switch layer.Type {
	case "text", "rect", "ellipse", "image":
	default:
		return apperrors.Invalid("图层类型不正确")
	}
	if layer.X < -20000 || layer.X > 20000 || layer.Y < -20000 || layer.Y > 20000 {
		return apperrors.Invalid("图层位置超出范围")
	}
	if layer.W <= 0 || layer.H <= 0 || layer.W > 8000 || layer.H > 8000 {
		return apperrors.Invalid("图层尺寸不正确")
	}
	if layer.Rotation < -360 || layer.Rotation > 360 {
		return apperrors.Invalid("图层旋转角度不正确")
	}
	if layer.Opacity < 0 || layer.Opacity > 1 {
		return apperrors.Invalid("图层透明度不正确")
	}
	if utf8.RuneCountInString(layer.Name) > 40 {
		return apperrors.Invalid("图层名称过长")
	}
	switch layer.Type {
	case "text":
		if utf8.RuneCountInString(layer.Text) > 500 {
			return apperrors.Invalid("文字不能超过 500 字")
		}
		if layer.FontSize < 8 || layer.FontSize > 400 {
			return apperrors.Invalid("字号不正确")
		}
		switch layer.FontWeight {
		case 400, 500, 600, 700, 800:
		default:
			return apperrors.Invalid("字重不正确")
		}
		if layer.FontFamily != "serif" && layer.FontFamily != "sans" {
			return apperrors.Invalid("字体不正确")
		}
		if layer.Align != "left" && layer.Align != "center" && layer.Align != "right" {
			return apperrors.Invalid("对齐方式不正确")
		}
		if !colorOK(layer.Color) {
			return apperrors.Invalid("文字颜色不正确")
		}
	case "rect", "ellipse":
		if !colorOK(layer.Fill) {
			return apperrors.Invalid("填充色不正确")
		}
		if layer.Radius < 0 || layer.Radius > 2000 {
			return apperrors.Invalid("圆角不正确")
		}
	case "image":
		if layer.AssetID < 1 {
			return apperrors.Invalid("图片图层需要选择素材")
		}
	}
	return nil
}

func idOK(id string) bool {
	if len(id) < 1 || len(id) > 40 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-':
		default:
			return false
		}
	}
	return true
}

func colorOK(value string) bool {
	if len(value) != 4 && len(value) != 7 && len(value) != 9 {
		return false
	}
	if value[0] != '#' {
		return false
	}
	for _, r := range value[1:] {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		case r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
}
