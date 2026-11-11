package poster

import (
	"context"
	"strings"
	"unicode/utf8"

	"poster.app/studio/internal/modules/studio/asset"
	"poster.app/studio/internal/modules/studio/shared"
	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/platform/response"
)

type Resource struct {
	repo   *Repository
	assets *asset.Resource
}

func NewResource(repo *Repository, assets *asset.Resource) *Resource {
	return &Resource{repo: repo, assets: assets}
}

func (r *Resource) Create(ctx context.Context, request *PosterRequest) (response.Response[PosterResponse], error) {
	var zero response.Response[PosterResponse]
	if err := r.prepareWrite(ctx, request, false); err != nil {
		return zero, err
	}
	return r.repo.Create(ctx, request)
}

func (r *Resource) Update(ctx context.Context, request *PosterRequest) (response.Response[PosterResponse], error) {
	var zero response.Response[PosterResponse]
	if err := r.prepareWrite(ctx, request, true); err != nil {
		return zero, err
	}
	if request.Version < 1 {
		return zero, apperrors.Invalid("缺少海报版本")
	}
	return r.repo.Update(ctx, request)
}

func (r *Resource) One(ctx context.Context, request *PosterRequest) (response.Response[PosterResponse], error) {
	var zero response.Response[PosterResponse]
	if _, err := domain.ActorID(ctx); err != nil {
		return zero, err
	}
	if request == nil || request.ID < 1 {
		return zero, apperrors.Invalid("海报标识不正确")
	}
	return r.repo.One(ctx, request)
}

func (r *Resource) Search(ctx context.Context, search *PosterSearch) (response.PageResponse[PosterSimple], error) {
	var zero response.PageResponse[PosterSimple]
	if _, err := domain.ActorID(ctx); err != nil {
		return zero, err
	}
	if search == nil {
		search = &PosterSearch{}
	}
	search.Title = strings.TrimSpace(search.Title)
	if utf8.RuneCountInString(search.Title) > 80 {
		return zero, apperrors.Invalid("标题筛选过长")
	}
	page, pageSize, err := domain.NormalizePage(search.Page, search.PageSize)
	if err != nil {
		return zero, err
	}
	search.Page, search.PageSize = page, pageSize
	return r.repo.Search(ctx, search)
}

func (r *Resource) Delete(ctx context.Context, request *PosterRequest) (response.Response[PosterResponse], error) {
	var zero response.Response[PosterResponse]
	if _, err := domain.ActorID(ctx); err != nil {
		return zero, err
	}
	if request == nil || request.ID < 1 {
		return zero, apperrors.Invalid("海报标识不正确")
	}
	return r.repo.Delete(ctx, request)
}

func (r *Resource) Duplicate(ctx context.Context, request *PosterDuplicateRequest) (response.Response[PosterResponse], error) {
	var zero response.Response[PosterResponse]
	if request == nil || request.ID < 1 {
		return zero, apperrors.Invalid("海报标识不正确")
	}
	current, err := r.One(ctx, &PosterRequest{ID: request.ID})
	if err != nil {
		return zero, err
	}
	title := strings.TrimSpace(current.Data.Title + " 副本")
	if utf8.RuneCountInString(title) > 80 {
		title = trimRunes(current.Data.Title, 76) + " 副本"
	}
	return r.Create(ctx, &PosterRequest{Title: title, Document: current.Data.Document})
}

func (r *Resource) prepareWrite(ctx context.Context, request *PosterRequest, existing bool) error {
	if _, err := domain.ActorID(ctx); err != nil {
		return err
	}
	if request == nil {
		return apperrors.Invalid("请求不能为空")
	}
	if !existing && request.ID != 0 {
		return apperrors.Invalid("创建不能指定主键")
	}
	if existing && request.ID < 1 {
		return apperrors.Invalid("海报标识不正确")
	}
	request.Title = strings.TrimSpace(request.Title)
	if request.Title == "" {
		request.Title = "未命名海报"
	}
	if utf8.RuneCountInString(request.Title) > 80 {
		return apperrors.Invalid("标题不能超过 80 个字")
	}
	shared.Normalize(&request.Document)
	if err := shared.Validate(&request.Document); err != nil {
		return err
	}
	for _, id := range shared.AssetIDs(request.Document) {
		_, err := r.assets.One(ctx, &asset.AssetRequest{ID: id})
		if err != nil {
			if appErr, ok := apperrors.As(err); ok && appErr.Class == apperrors.ClassNotFound {
				return apperrors.Invalid("素材不存在或无权使用")
			}
			return err
		}
	}
	return nil
}

func trimRunes(value string, max int) string {
	count := 0
	for index := range value {
		if count == max {
			return value[:index]
		}
		count++
	}
	return value
}
