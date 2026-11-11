package template

import (
	"context"
	"strings"
	"unicode/utf8"

	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/platform/response"
)

type Resource struct {
	repo *Repository
}

func NewResource(repo *Repository) *Resource {
	return &Resource{repo: repo}
}

func (r *Resource) One(ctx context.Context, request *TemplateRequest) (response.Response[TemplateResponse], error) {
	var zero response.Response[TemplateResponse]
	if _, err := domain.ActorID(ctx); err != nil {
		return zero, err
	}
	if request == nil {
		return zero, apperrors.Invalid("请求不能为空")
	}
	if err := requirePositive(request.ID); err != nil {
		return zero, err
	}
	return r.repo.One(ctx, request)
}

func (r *Resource) Search(ctx context.Context, search *TemplateSearch) (response.PageResponse[TemplateSimple], error) {
	var zero response.PageResponse[TemplateSimple]
	if _, err := domain.ActorID(ctx); err != nil {
		return zero, err
	}
	if search == nil {
		search = &TemplateSearch{}
	}
	search.Category = strings.TrimSpace(search.Category)
	search.Name = strings.TrimSpace(search.Name)
	if utf8.RuneCountInString(search.Category) > 20 || utf8.RuneCountInString(search.Name) > 80 {
		return zero, apperrors.Invalid("筛选条件过长")
	}
	page, pageSize, err := domain.NormalizePage(search.Page, search.PageSize)
	if err != nil {
		return zero, err
	}
	search.Page, search.PageSize = page, pageSize
	return r.repo.Search(ctx, search)
}
