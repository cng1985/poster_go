package analysis

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"poster.app/studio/internal/platform/database"
	"poster.app/studio/internal/platform/domain"
	apperrors "poster.app/studio/internal/platform/errors"
	"poster.app/studio/internal/platform/response"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, entity *Analysis) (response.Response[AnalysisResponse], error) {
	var result response.Response[AnalysisResponse]
	if _, err := domain.ActorID(ctx); err != nil {
		return result, err
	}
	if err := r.db.WithContext(ctx).Omit(clause.Associations).Create(entity).Error; err != nil {
		return result, database.MapError(err)
	}
	item, err := AnalysisEntityToResponse(entity)
	if err != nil {
		return result, apperrors.Internal(err)
	}
	return response.OK(*item), nil
}

func (r *Repository) One(ctx context.Context, request *AnalysisRequest) (response.Response[AnalysisResponse], error) {
	var result response.Response[AnalysisResponse]
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return result, err
	}
	var entity Analysis
	err = r.db.WithContext(ctx).Where("id = ? and owner_id = ?", request.ID, owner).Take(&entity).Error
	if err != nil {
		if appErr, ok := apperrors.As(database.MapError(err)); ok && appErr.Class == apperrors.ClassNotFound {
			return result, apperrors.NotFound("分析记录不存在")
		}
		return result, database.MapError(err)
	}
	item, err := AnalysisEntityToResponse(&entity)
	if err != nil {
		return result, apperrors.Internal(err)
	}
	return response.OK(*item), nil
}

func (r *Repository) Search(ctx context.Context, search *AnalysisSearch) (response.PageResponse[AnalysisSimple], error) {
	var result response.PageResponse[AnalysisSimple]
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return result, err
	}
	page, pageSize, err := domain.NormalizePage(search.Page, search.PageSize)
	if err != nil {
		return result, err
	}
	search.Page, search.PageSize = page, pageSize
	makeQuery := func() *gorm.DB {
		query := r.db.WithContext(ctx).Model(&Analysis{}).Where("owner_id = ?", owner)
		if search.PosterID > 0 {
			query = query.Where("poster_id = ?", search.PosterID)
		}
		return query
	}
	var total int64
	if err := makeQuery().Count(&total).Error; err != nil {
		return result, database.MapError(err)
	}
	rows := make([]Analysis, 0)
	err = makeQuery().Order("id desc").Offset(domain.Offset(search.Page, search.PageSize)).Limit(search.PageSize).Find(&rows).Error
	if err != nil {
		return result, database.MapError(err)
	}
	data := make([]AnalysisSimple, 0, len(rows))
	for i := range rows {
		data = append(data, *AnalysisEntityToSimple(&rows[i]))
	}
	return response.PageResponse[AnalysisSimple]{Data: data, Total: total, Page: search.Page, PageSize: search.PageSize}, nil
}
