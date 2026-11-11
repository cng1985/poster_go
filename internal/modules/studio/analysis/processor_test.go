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

func TestMarginDoesNotSkipLaterText(t *testing.T) {
	doc := shared.Document{
		Width: 1080, Height: 1440, Background: "#111111",
		Layers: []shared.Layer{
			{
				ID: "edge", Type: "text", Name: "edge", X: 0, Y: 40, W: 200, H: 40, Opacity: 1,
				Text: "边", FontSize: 16, FontWeight: 400, FontFamily: "sans", Color: "#ffffff", Align: "left",
			},
			{
				ID: "title", Type: "text", Name: "title", X: 120, Y: 200, W: 800, H: 200, Opacity: 1,
				Text: "标题", FontSize: 120, FontWeight: 700, FontFamily: "serif", Color: "#121212", Align: "left",
			},
		},
	}
	report := Evaluate(doc)
	var contrast, tooSmall bool
	for _, item := range report.Suggestions {
		if item.Text == "有文字和背景太接近，缩小屏幕上会看不清。" {
			contrast = true
		}
		if item.Text == "最大的字还不够大，标题可以再突出一点。" {
			tooSmall = true
		}
	}
	if !contrast || tooSmall {
		t.Fatalf("contrast=%v tooSmall=%v %+v", contrast, tooSmall, report)
	}
}
