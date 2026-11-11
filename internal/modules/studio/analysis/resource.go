package analysis

import (
	"context"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"poster.app/studio/internal/modules/studio/poster"
	"poster.app/studio/internal/modules/studio/shared"
	"poster.app/studio/internal/platform/ai"
	"poster.app/studio/internal/platform/config"
	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/platform/limit"
	"poster.app/studio/internal/platform/response"
)

type Resource struct {
	repo    *Repository
	posters *poster.Resource
	models  ai.Completer
	limit   *limit.Window
}

func NewResource(repo *Repository, posters *poster.Resource, models ai.Completer, cfg *config.Config) *Resource {
	return &Resource{
		repo: repo, posters: posters, models: models,
		limit: limit.NewWindow(20, timeMinute),
	}
}

func (r *Resource) Create(ctx context.Context, request *AnalysisRequest) (response.Response[AnalysisResponse], error) {
	var zero response.Response[AnalysisResponse]
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return zero, err
	}
	if request == nil || request.PosterID < 1 {
		return zero, apperrors.Invalid("请选择要分析的海报")
	}
	if err := r.limit.Allow("analysis:" + ownerKey(owner)); err != nil {
		return zero, err
	}
	current, err := r.posters.One(ctx, &poster.PosterRequest{ID: request.PosterID})
	if err != nil {
		return zero, err
	}
	report := Evaluate(current.Data.Document)
	provider := "rules"
	if r.models != nil && r.models.Enabled() {
		if note := r.modelNote(ctx, current.Data.Document); note != "" {
			report.Suggestions = append(report.Suggestions, Suggestion{Level: "info", Text: note})
			provider = r.models.Provider() + "+rules"
		}
	}
	raw, err := json.Marshal(report.Suggestions)
	if err != nil {
		return zero, apperrors.Internal(err)
	}
	return r.repo.Create(ctx, &Analysis{
		OwnerID: owner, PosterID: request.PosterID, Score: report.Score,
		Summary: report.Summary, Suggestions: string(raw), Provider: provider,
	})
}

func (r *Resource) One(ctx context.Context, request *AnalysisRequest) (response.Response[AnalysisResponse], error) {
	var zero response.Response[AnalysisResponse]
	if _, err := domain.ActorID(ctx); err != nil {
		return zero, err
	}
	if request == nil || request.ID < 1 {
		return zero, apperrors.Invalid("分析记录标识不正确")
	}
	return r.repo.One(ctx, request)
}

func (r *Resource) Search(ctx context.Context, search *AnalysisSearch) (response.PageResponse[AnalysisSimple], error) {
	var zero response.PageResponse[AnalysisSimple]
	if _, err := domain.ActorID(ctx); err != nil {
		return zero, err
	}
	if search == nil {
		search = &AnalysisSearch{}
	}
	page, pageSize, err := domain.NormalizePage(search.Page, search.PageSize)
	if err != nil {
		return zero, err
	}
	search.Page, search.PageSize = page, pageSize
	return r.repo.Search(ctx, search)
}

func (r *Resource) modelNote(ctx context.Context, doc shared.Document) string {
	brief := briefDocument(doc)
	messages := []ai.Message{
		{Role: "system", Content: "你是海报审稿人。只用一句中文给出最重要的修改建议，不要输出 JSON，不要超过 80 字。"},
		{Role: "user", Content: brief},
	}
	text, err := r.models.Complete(ctx, messages)
	if err != nil {
		return ""
	}
	text = strings.TrimSpace(text)
	text = trimRunes(text, 80)
	if text == "" {
		return ""
	}
	return "模型补充：" + text
}

func briefDocument(doc shared.Document) string {
	var b strings.Builder
	b.WriteString("尺寸 ")
	b.WriteString(strings.TrimSpace(strings.ReplaceAll(jsonNumber(doc.Width), " ", "")))
	b.WriteString("x")
	b.WriteString(strings.TrimSpace(strings.ReplaceAll(jsonNumber(doc.Height), " ", "")))
	b.WriteString(" 背景 ")
	b.WriteString(doc.Background)
	b.WriteString("。文字：")
	for _, layer := range doc.Layers {
		if layer.Type != "text" {
			continue
		}
		b.WriteString(layer.Text)
		b.WriteString(" / ")
	}
	return trimRunes(b.String(), 800)
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

func ownerKey(id int64) string {
	return strings.TrimSpace(strings.ReplaceAll(jsonNumber(int(id)), " ", ""))
}

func jsonNumber(value int) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}
