package generation

import (
	"context"
	"log/slog"
	"strconv"
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
	max := cfg.AI.MaxPerMinute
	if max <= 0 {
		max = 10
	}
	return &Resource{
		repo: repo, posters: posters, models: models,
		limit: limit.NewWindow(max, timeMinute),
	}
}

func (r *Resource) Create(ctx context.Context, request *GenerationRequest) (response.Response[GenerationResponse], error) {
	var zero response.Response[GenerationResponse]
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return zero, err
	}
	if request == nil {
		return zero, apperrors.Invalid("请求不能为空")
	}
	request.Prompt = strings.TrimSpace(request.Prompt)
	request.Style = strings.TrimSpace(request.Style)
	request.SizePreset = strings.TrimSpace(request.SizePreset)
	request.Title = strings.TrimSpace(request.Title)
	if request.Style == "" {
		request.Style = "editorial"
	}
	if request.SizePreset == "" {
		request.SizePreset = "poster"
	}
	if request.Prompt == "" || utf8.RuneCountInString(request.Prompt) > 500 {
		return zero, apperrors.Invalid("请用 1 到 500 字描述想要的海报")
	}
	if !knownStyle(request.Style) || !knownPreset(request.SizePreset) {
		return zero, apperrors.Invalid("风格或尺寸不正确")
	}
	if err := r.limit.Allow("generation:" + strconv.FormatInt(owner, 10)); err != nil {
		return zero, err
	}
	doc, provider, modelName, message, err := r.compose(ctx, request)
	if err != nil {
		return zero, err
	}
	size := presets[request.SizePreset]
	if doc.Width == 0 {
		doc.Width = size[0]
	}
	if doc.Height == 0 {
		doc.Height = size[1]
	}
	raw, err := shared.Prepare(&doc)
	if err != nil {
		if r.models.FallbackLocal() && provider != "local" {
			doc = ComposeLocal(request.Prompt, request.Style, request.SizePreset)
			raw, err = shared.Prepare(&doc)
			provider = "local"
			modelName = ""
			message = "模型结果无法排版，已改用本地排版。"
		}
		if err != nil {
			return zero, err
		}
	}
	entity := &Generation{
		OwnerID: owner, Prompt: request.Prompt, Style: request.Style, SizePreset: request.SizePreset,
		Provider: provider, Model: modelName, Document: raw, Message: trimRunes(message, 120),
	}
	created, err := r.repo.Create(ctx, entity)
	if err != nil {
		return zero, err
	}
	if request.SavePoster {
		title := request.Title
		if title == "" {
			title = firstClause(request.Prompt)
		}
		saved, saveErr := r.posters.Create(ctx, &poster.PosterRequest{Title: title, Document: doc})
		if saveErr != nil {
			created.Data.Message = trimRunes(created.Data.Message+" 作品未能保存："+apperrors.PublicMessage(saveErr), 120)
			return created, nil
		}
		if err := r.repo.attachPoster(ctx, created.Data.ID, owner, saved.Data.ID); err != nil {
			slog.Default().Error("attach generated poster", "err", err.Error(), "generation_id", created.Data.ID)
		}
		created.Data.PosterID = saved.Data.ID
	}
	return created, nil
}

func (r *Resource) One(ctx context.Context, request *GenerationRequest) (response.Response[GenerationResponse], error) {
	var zero response.Response[GenerationResponse]
	if _, err := domain.ActorID(ctx); err != nil {
		return zero, err
	}
	if request == nil || request.ID < 1 {
		return zero, apperrors.Invalid("生成记录标识不正确")
	}
	return r.repo.One(ctx, request)
}

func (r *Resource) Search(ctx context.Context, search *GenerationSearch) (response.PageResponse[GenerationSimple], error) {
	var zero response.PageResponse[GenerationSimple]
	if _, err := domain.ActorID(ctx); err != nil {
		return zero, err
	}
	if search == nil {
		search = &GenerationSearch{}
	}
	page, pageSize, err := domain.NormalizePage(search.Page, search.PageSize)
	if err != nil {
		return zero, err
	}
	search.Page, search.PageSize = page, pageSize
	return r.repo.Search(ctx, search)
}

func (r *Resource) compose(ctx context.Context, request *GenerationRequest) (shared.Document, string, string, string, error) {
	if r.models != nil && r.models.Enabled() {
		text, err := r.models.Complete(ctx, Messages(request.Prompt, request.Style, request.SizePreset))
		if err == nil {
			doc, parseErr := ParseModelDocument(text)
			if parseErr == nil {
				return doc, r.models.Provider(), r.models.Model(), "已由大模型生成海报。", nil
			}
			err = parseErr
		}
		if r.models.FallbackLocal() {
			return ComposeLocal(request.Prompt, request.Style, request.SizePreset), "local", "", "大模型不可用，已改用本地排版。", nil
		}
		return shared.Document{}, "", "", "", apperrors.Unavailable("大模型暂时不可用")
	}
	if r.models != nil && r.models.FallbackLocal() {
		return ComposeLocal(request.Prompt, request.Style, request.SizePreset), "local", "", "未配置大模型，已使用本地排版。设置 APP_AI_API_KEY 后可调用模型。", nil
	}
	return shared.Document{}, "", "", "", apperrors.Unavailable("未配置大模型")
}
