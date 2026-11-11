package analysis

import (
	"testing"

	"poster.app/studio/internal/modules/studio/shared"
)

func TestLowContrastIsWarned(t *testing.T) {
	doc := shared.Document{
		Width: 1080, Height: 1440, Background: "#111111",
		Layers: []shared.Layer{{
			ID: "t", Type: "text", Name: "t", X: 10, Y: 10, W: 400, H: 80, Opacity: 1,
			Text: "看不清", FontSize: 20, FontWeight: 400, FontFamily: "sans", Color: "#222222", Align: "left",
		}},
	}
	report := Evaluate(doc)
	if report.Score >= 85 {
		t.Fatalf("score = %d", report.Score)
	}
	found := false
	for _, item := range report.Suggestions {
		if item.Level == "warn" {
			found = true
		}
	}
	if !found {
		t.Fatalf("%+v", report)
	}
}
