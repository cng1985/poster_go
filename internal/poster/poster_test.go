package poster

import "testing"

func TestTemplatesNormalize(t *testing.T) {
	for _, tpl := range Templates() {
		doc := tpl.Document
		if err := Normalize(&doc, false); err != nil {
			t.Fatalf("%s: %v", tpl.ID, err)
		}
		if len(doc.Layers) == 0 {
			t.Fatalf("%s has no layers", tpl.ID)
		}
	}
	sample := Sample()
	if err := Normalize(&sample, true); err != nil {
		t.Fatal(err)
	}
	if sample.ID != "sample-spring" {
		t.Fatalf("sample id %s", sample.ID)
	}
}

func TestNormalizeRejects(t *testing.T) {
	base := Blank()
	base.ID = "p1"
	cases := []struct {
		name string
		edit func(*Document)
	}{
		{"color", func(d *Document) { d.Background = "red" }},
		{"src", func(d *Document) {
			d.Layers = []Layer{{ID: "a", Type: "image", Name: "图", W: 100, H: 100, Src: "/etc/passwd", Visible: true, Opacity: 1}}
		}},
		{"size", func(d *Document) { d.Width = 10 }},
		{"dup", func(d *Document) {
			d.Layers = []Layer{
				{ID: "a", Type: "rect", Name: "一", W: 10, H: 10, Visible: true, Opacity: 1},
				{ID: "a", Type: "rect", Name: "二", W: 10, H: 10, Visible: true, Opacity: 1},
			}
		}},
	}
	for _, tc := range cases {
		doc := base
		doc.Layers = nil
		tc.edit(&doc)
		if err := Normalize(&doc, true); err == nil {
			t.Fatalf("%s was accepted", tc.name)
		}
	}
}
