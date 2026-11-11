package poster

import "time"

// Document is a poster stored by the editor.
type Document struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Width           int       `json:"width"`
	Height          int       `json:"height"`
	Background      string    `json:"background"`
	BackgroundImage string    `json:"backgroundImage,omitempty"`
	Layers          []Layer   `json:"layers"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// Layer is one object on the poster, painted from back to front.
type Layer struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Role        string  `json:"role,omitempty"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	W           float64 `json:"w"`
	H           float64 `json:"h"`
	Rotation    float64 `json:"rotation"`
	Opacity     float64 `json:"opacity"`
	Visible     bool    `json:"visible"`
	Locked      bool    `json:"locked"`
	Text        string  `json:"text,omitempty"`
	FontSize    float64 `json:"fontSize,omitempty"`
	FontFamily  string  `json:"fontFamily,omitempty"`
	FontWeight  string  `json:"fontWeight,omitempty"`
	Color       string  `json:"color,omitempty"`
	Align       string  `json:"align,omitempty"`
	LineHeight  float64 `json:"lineHeight,omitempty"`
	Fill        string  `json:"fill,omitempty"`
	Stroke      string  `json:"stroke,omitempty"`
	StrokeWidth float64 `json:"strokeWidth,omitempty"`
	Radius      float64 `json:"radius,omitempty"`
	Src         string  `json:"src,omitempty"`
	Fit         string  `json:"fit,omitempty"`
}

// Summary is a poster without its layers.
type Summary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Template is a starter poster the editor can apply.
type Template struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Accent      string   `json:"accent"`
	Document    Document `json:"document"`
}

// Blank is an empty vertical poster.
func Blank() Document {
	return Document{
		Name:       "未命名海报",
		Width:      1080,
		Height:     1520,
		Background: "#f6f1e8",
		Layers:     []Layer{},
	}
}
