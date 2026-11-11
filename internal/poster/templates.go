package poster

func layer(id, name, typ string, x, y, w, h float64) Layer {
	return Layer{
		ID: id, Name: name, Type: typ,
		X: x, Y: y, W: w, H: h,
		Opacity: 1, Visible: true,
		FontFamily: "sans", FontWeight: "400", Align: "left",
		Color: "#1a1a1a", LineHeight: 1.25, Fit: "cover",
	}
}

func text(id, name, role, text string, x, y, w, h, size float64, weight, family, color, align string) Layer {
	l := layer(id, name, "text", x, y, w, h)
	l.Role = role
	l.Text = text
	l.FontSize = size
	l.FontWeight = weight
	l.FontFamily = family
	l.Color = color
	l.Align = align
	l.Fit = ""
	return l
}

func shape(id, name, typ, role, fill string, x, y, w, h, radius float64) Layer {
	l := layer(id, name, typ, x, y, w, h)
	l.Role = role
	l.Fill = fill
	l.Radius = radius
	l.Fit = ""
	return l
}

func image(id, name, role, src string, x, y, w, h float64) Layer {
	l := layer(id, name, "image", x, y, w, h)
	l.Role = role
	l.Src = src
	l.Fit = "cover"
	return l
}

// Templates returns the built-in starter posters.
func Templates() []Template {
	spring := Document{
		Name: "春季上新", Width: 1080, Height: 1520, Background: "#1d2348",
		Layers: []Layer{
			image("sp-bg", "暮色背景", "backdrop", "/library/backgrounds/dusk.svg", 0, 0, 1080, 1520),
			shape("sp-scrim", "底部遮罩", "rect", "decoration", "#120f0ccc", 0, 980, 1080, 540, 0),
			shape("sp-badge", "角标底", "rect", "decoration", "#e23d2b", 72, 72, 168, 52, 999),
			text("sp-badge-tx", "角标", "kicker", "NEW", 72, 72, 168, 52, 22, "700", "sans", "#ffffff", "center"),
			text("sp-kicker", "眉题", "kicker", "SPRING / 春季系列", 72, 1040, 936, 48, 26, "600", "sans", "#f6d3b0", "left"),
			text("sp-title", "主标题", "title", "把春天\n穿在身上", 72, 1096, 936, 230, 92, "800", "sans", "#fffaf3", "left"),
			text("sp-sub", "副标题", "subtitle", "限时首发 · 全场 8 折", 72, 1360, 936, 70, 30, "500", "sans", "#ffe6c7", "left"),
		},
	}
	spring.Layers[5].LineHeight = 1.05

	invite := Document{
		Name: "活动邀请", Width: 1080, Height: 1520, Background: "#f6efe6",
		Layers: []Layer{
			image("iv-bg", "晨雾背景", "backdrop", "/library/backgrounds/mist.svg", 0, 0, 1080, 1520),
			shape("iv-card", "卡片", "rect", "decoration", "#fffcf8", 80, 160, 920, 1160, 40),
			text("iv-kicker", "眉题", "kicker", "邀请函", 140, 250, 800, 50, 26, "600", "sans", "#a2674e", "center"),
			text("iv-title", "主标题", "title", "本周六\n一起来", 120, 330, 840, 210, 88, "700", "serif", "#2c241c", "center"),
			shape("iv-line", "分隔线", "line", "decoration", "#e23d2b", 470, 570, 140, 6, 999),
			text("iv-body", "正文", "body", "看一场关于春天的展览，\n把下午留给光和朋友。", 160, 620, 760, 180, 32, "400", "sans", "#3a332c", "center"),
			text("iv-meta", "信息", "subtitle", "4 月 18 日 · 15:00 · 城西美术馆", 160, 840, 760, 80, 26, "500", "sans", "#8a5a44", "center"),
			shape("iv-btn", "按钮", "rect", "decoration", "#2c241c", 390, 1040, 300, 70, 999),
			text("iv-btn-tx", "按钮文字", "price", "确认出席", 390, 1040, 300, 70, 26, "700", "sans", "#ffffff", "center"),
		},
	}
	invite.Layers[3].LineHeight = 1.08

	sale := Document{
		Name: "限时折扣", Width: 1080, Height: 1080, Background: "#160818",
		Layers: []Layer{
			image("sale-bg", "霓虹背景", "backdrop", "/library/backgrounds/neon.svg", 0, 0, 1080, 1080),
			text("sale-kicker", "眉题", "kicker", "WEEKEND SALE", 80, 150, 920, 50, 28, "700", "sans", "#ff9ad6", "center"),
			text("sale-price", "折扣", "price", "5 折", 80, 250, 920, 280, 200, "800", "sans", "#ffffff", "center"),
			text("sale-sub", "说明", "subtitle", "全场精选，本周末截止", 140, 560, 800, 80, 32, "500", "sans", "#f7d5ee", "center"),
			shape("sale-btn", "按钮", "rect", "decoration", "#ffffff", 340, 760, 400, 80, 999),
			text("sale-btn-tx", "按钮文字", "body", "立即抢购", 340, 760, 400, 80, 30, "700", "sans", "#160818", "center"),
		},
	}

	job := Document{
		Name: "岗位招聘", Width: 1080, Height: 1520, Background: "#f3f0e8",
		Layers: []Layer{
			shape("job-bar", "色条", "rect", "decoration", "#1f6b4a", 0, 0, 26, 1520, 0),
			text("job-kicker", "眉题", "kicker", "WE'RE HIRING", 80, 110, 620, 40, 24, "700", "sans", "#1f6b4a", "left"),
			text("job-title", "职位", "title", "品牌\n设计师", 80, 170, 620, 210, 84, "800", "sans", "#1c1c1c", "left"),
			image("job-photo", "叶片", "image", "/library/illustrations/leaf.svg", 740, 150, 260, 260),
			text("job-body", "正文", "body", "我们在找一位把版式、色彩和文字都当真的设计师。", 80, 480, 860, 120, 30, "400", "sans", "#2a2722", "left"),
			text("job-list", "要点", "body", "· 负责活动与电商海报\n· 能独立完成从构图到导出\n· 喜欢清晰的网格和大胆的留白", 80, 660, 860, 280, 28, "500", "sans", "#1c1c1c", "left"),
			shape("job-line", "分隔线", "line", "decoration", "#1f6b4a", 80, 1240, 180, 6, 999),
			text("job-mail", "联系方式", "subtitle", "投递 hello@studio.test", 80, 1320, 860, 60, 26, "600", "sans", "#1f6b4a", "left"),
		},
	}
	job.Layers[2].LineHeight = 1.05
	job.Layers[5].LineHeight = 1.55

	return []Template{
		{ID: "spring", Name: "春季上新", Description: "竖版大图，适合新品发售", Accent: "#e23d2b", Document: spring},
		{ID: "invite", Name: "活动邀请", Description: "留白卡片，适合展览和沙龙", Accent: "#a2674e", Document: invite},
		{ID: "sale", Name: "限时折扣", Description: "方形促销，价格是主角", Accent: "#ff9ad6", Document: sale},
		{ID: "job", Name: "岗位招聘", Description: "竖版文字海报", Accent: "#1f6b4a", Document: job},
	}
}

// Sample is the poster created the first time a data directory is opened.
func Sample() Document {
	d := Templates()[0].Document
	d.ID = "sample-spring"
	return d
}
