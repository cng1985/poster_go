package template

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"poster.app/studio/internal/modules/studio/shared"
)

// Seed 只补齐缺失的系统模板，不覆盖已有记录。
func Seed(ctx context.Context, db *gorm.DB) error {
	for _, item := range catalog() {
		raw, err := json.Marshal(item.doc)
		if err != nil {
			return err
		}
		entity := Template{
			Code: item.code, Name: item.name, Category: item.category, Summary: item.summary,
			Width: item.doc.Width, Height: item.doc.Height, CoverColor: item.doc.Background,
			Document: string(raw), SortNo: item.sort, Status: "published",
		}
		err = db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "code"}},
			DoNothing: true,
		}).Create(&entity).Error
		if err != nil {
			return err
		}
	}
	return nil
}

type seedItem struct {
	code, name, category, summary string
	sort                          int
	doc                           shared.Document
}

func catalog() []seedItem {
	return []seedItem{
		{code: "spring-tea", name: "一口新绿", category: "饮品", summary: "春季茶饮竖版，适合新品上架。", sort: 10, doc: springTea()},
		{code: "night-sale", name: "夜场限时", category: "促销", summary: "高对比促销海报，突出折扣时段。", sort: 20, doc: nightSale()},
		{code: "hiring-studio", name: "招设计师", category: "招聘", summary: "工作室招聘，信息层次清楚。", sort: 30, doc: hiring()},
		{code: "live-talk", name: "讲座邀请", category: "活动", summary: "分享会邀请函，留出时间地点。", sort: 40, doc: talk()},
		{code: "festival-lantern", name: "上元灯会", category: "节日", summary: "节日祝福竖版，适合灯会与团圆。", sort: 50, doc: lantern()},
		{code: "product-drop", name: "新品发售", category: "发布", summary: "方图发售海报，适合社交平台。", sort: 60, doc: drop()},
		{code: "food-noon", name: "午市套餐", category: "餐饮", summary: "餐厅午市菜单海报。", sort: 70, doc: noon()},
		{code: "course-open", name: "课程招生", category: "教育", summary: "横版课程封面，适合讲座与网课。", sort: 80, doc: course()},
	}
}

func text(id, content string, x, y, w, h, size float64, weight int, color, align, family string) shared.Layer {
	return shared.Layer{
		ID: id, Type: "text", Name: shortName(content), X: x, Y: y, W: w, H: h, Opacity: 1,
		Text: content, FontSize: size, FontWeight: weight, Color: color, Align: align, FontFamily: family,
	}
}

func rect(id string, x, y, w, h, radius float64, fill string) shared.Layer {
	return shared.Layer{ID: id, Type: "rect", Name: "色块", X: x, Y: y, W: w, H: h, Opacity: 1, Fill: fill, Radius: radius}
}

func ellipse(id string, x, y, w, h float64, fill string) shared.Layer {
	return shared.Layer{ID: id, Type: "ellipse", Name: "圆形", X: x, Y: y, W: w, H: h, Opacity: 1, Fill: fill}
}

func shortName(content string) string {
	content = strings.TrimSpace(strings.ReplaceAll(content, "\n", " "))
	if content == "" {
		return "文字"
	}
	if utf8.RuneCountInString(content) <= 12 {
		return content
	}
	runes := []rune(content)
	return string(runes[:12])
}

func springTea() shared.Document {
	return shared.Document{Width: 1080, Height: 1920, Background: "#14342c", Layers: []shared.Layer{
		ellipse("orb", 680, -80, 520, 520, "#1f6a52"),
		rect("band", 0, 1500, 1080, 420, 0, "#e4572e"),
		text("kicker", "春季限定", 80, 220, 500, 80, 36, 600, "#d4a656", "left", "sans"),
		text("title", "一口新绿", 80, 340, 900, 280, 140, 700, "#f3ecdf", "left", "serif"),
		text("sub", "明前龙井 · 冷泡也好喝", 80, 660, 860, 120, 42, 500, "#d7efe4", "left", "sans"),
		text("foot", "本周到店第二杯半价", 80, 1600, 900, 140, 48, 700, "#fff8ef", "left", "serif"),
	}}
}

func nightSale() shared.Document {
	return shared.Document{Width: 1080, Height: 1920, Background: "#2a0d12", Layers: []shared.Layer{
		rect("panel", 70, 180, 940, 1500, 36, "#6e1520"),
		text("kicker", "TONIGHT ONLY", 120, 280, 820, 70, 28, 700, "#f2c14e", "left", "sans"),
		text("title", "夜场五折", 120, 400, 840, 260, 120, 700, "#fff8ea", "left", "serif"),
		text("sub", "19:00 – 21:00\n全场指定货架", 120, 720, 760, 220, 48, 500, "#ffd9a8", "left", "sans"),
		rect("chip", 120, 1080, 360, 120, 60, "#f2c14e"),
		text("chip-text", "限时 2 小时", 120, 1105, 360, 80, 36, 700, "#2a0d12", "center", "sans"),
	}}
}

func hiring() shared.Document {
	return shared.Document{Width: 1080, Height: 1440, Background: "#f3ecdf", Layers: []shared.Layer{
		rect("side", 0, 0, 28, 1440, 0, "#e4572e"),
		text("kicker", "工作室招聘", 90, 140, 800, 70, 32, 600, "#e4572e", "left", "sans"),
		text("title", "找一位\n会排版的人", 90, 250, 880, 360, 92, 700, "#1a1612", "left", "serif"),
		text("sub", "平面 / 海报 / 品牌物料\n作品集请发到 hello@studio.test", 90, 700, 860, 220, 36, 500, "#3d3832", "left", "sans"),
		rect("box", 90, 1040, 420, 180, 8, "#1a1612"),
		text("box-text", "全职 · 可远程", 90, 1090, 420, 90, 36, 600, "#f3ecdf", "center", "sans"),
	}}
}

func talk() shared.Document {
	return shared.Document{Width: 1080, Height: 1440, Background: "#e7efe8", Layers: []shared.Layer{
		ellipse("moon", 760, 80, 240, 240, "#d4a656"),
		text("kicker", "周五晚八点", 80, 180, 700, 70, 32, 600, "#3e5c4a", "left", "sans"),
		text("title", "把海报\n讲清楚", 80, 300, 900, 360, 110, 700, "#1a2420", "left", "serif"),
		text("sub", "一场关于网格、字号和留白的分享", 80, 720, 880, 140, 36, 500, "#31463c", "left", "sans"),
		text("meta", "市立美术馆 · 报告厅 B\n免费入场，请提前占座", 80, 1040, 860, 180, 34, 600, "#1a2420", "left", "sans"),
	}}
}

func lantern() shared.Document {
	return shared.Document{Width: 1080, Height: 1920, Background: "#4a1420", Layers: []shared.Layer{
		ellipse("lamp1", 120, 160, 280, 360, "#e7c27d"),
		ellipse("lamp2", 680, 220, 220, 300, "#f3ecdf"),
		rect("ground", 0, 1280, 1080, 640, 0, "#2a0c12"),
		text("title", "灯火可亲", 80, 760, 920, 220, 120, 700, "#f8ecd4", "center", "serif"),
		text("sub", "上元快乐\n愿你今年有一盏一直亮着的灯", 120, 1040, 840, 180, 36, 500, "#f0d7a2", "center", "sans"),
		text("foot", "海报工坊", 80, 1560, 920, 80, 28, 600, "#e7c27d", "center", "sans"),
	}}
}

func drop() shared.Document {
	return shared.Document{Width: 1080, Height: 1080, Background: "#161513", Layers: []shared.Layer{
		rect("card", 80, 80, 920, 920, 28, "#221e1a"),
		text("kicker", "DROP 04", 140, 180, 760, 60, 28, 700, "#d4a656", "left", "sans"),
		text("title", "新的封面\n今天上架", 140, 300, 800, 280, 84, 700, "#f4efe6", "left", "serif"),
		rect("line", 140, 680, 180, 8, 0, "#e4572e"),
		text("sub", "限量 200 份 · 售完即止", 140, 740, 760, 80, 32, 500, "#d9cbb8", "left", "sans"),
	}}
}

func noon() shared.Document {
	return shared.Document{Width: 1080, Height: 1440, Background: "#f6efe4", Layers: []shared.Layer{
		rect("top", 0, 0, 1080, 420, 0, "#3d2a22"),
		text("kicker", "午市", 80, 90, 400, 70, 32, 600, "#e7c27d", "left", "sans"),
		text("title", "两菜一汤", 80, 180, 900, 160, 92, 700, "#f6efe4", "left", "serif"),
		text("a", "番茄炒蛋", 80, 540, 700, 90, 48, 700, "#3d2a22", "left", "serif"),
		text("b", "清炒时蔬", 80, 680, 700, 90, 48, 700, "#3d2a22", "left", "serif"),
		text("c", "紫菜蛋汤", 80, 820, 700, 90, 48, 700, "#3d2a22", "left", "serif"),
		text("price", "38", 760, 1180, 240, 140, 92, 700, "#e4572e", "right", "serif"),
	}}
}

func course() shared.Document {
	return shared.Document{Width: 1920, Height: 1080, Background: "#10243a", Layers: []shared.Layer{
		rect("block", 0, 0, 720, 1080, 0, "#e4572e"),
		text("kicker", "四周小课", 120, 180, 500, 70, 32, 600, "#fff4ea", "left", "sans"),
		text("title", "会做海报\n的人", 120, 300, 520, 360, 92, 700, "#fff8f3", "left", "serif"),
		text("right-title", "从一张白纸到能发出去", 820, 280, 960, 160, 64, 700, "#f4efe6", "left", "serif"),
		text("right-sub", "网格、字号、颜色和导出。\n每周三晚，小班练习。", 820, 500, 900, 180, 36, 500, "#d5e2ef", "left", "sans"),
		text("meta", "4 月开班 · 线上", 820, 820, 700, 80, 32, 700, "#d4a656", "left", "sans"),
	}}
}
