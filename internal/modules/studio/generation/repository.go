package generation

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

func (r *Repository) Create(ctx context.Context, entity *Generation) (response.Response[GenerationResponse], error) {
	var result response.Response[GenerationResponse]
	if _, err := domain.ActorID(ctx); err != nil {
		return result, err
	}
	if err := r.db.WithContext(ctx).Omit(clause.Associations).Create(entity).Error; err != nil {
		return result, database.MapError(err)
	}
	item, err := GenerationEntityToResponse(entity)
	if err != nil {
		return result, err
	}
	return response.OK(*item), nil
}

func (r *Repository) One(ctx context.Context, request *GenerationRequest) (response.Response[GenerationResponse], error) {
	var result response.Response[GenerationResponse]
	entity, err := r.find(ctx, request.ID)
	if err != nil {
		return result, err
	}
	item, err := GenerationEntityToResponse(entity)
	if err != nil {
		return result, err
	}
	return response.OK(*item), nil
}

func (r *Repository) Search(ctx context.Context, search *GenerationSearch) (response.PageResponse[GenerationSimple], error) {
	var result response.PageResponse[GenerationSimple]
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
		return r.db.WithContext(ctx).Model(&Generation{}).Where("owner_id = ?", owner)
	}
	var total int64
	if err := makeQuery().Count(&total).Error; err != nil {
		return result, database.MapError(err)
	}
	rows := make([]Generation, 0)
	err = makeQuery().Order("id desc").Offset(domain.Offset(search.Page, search.PageSize)).Limit(search.PageSize).Find(&rows).Error
	if err != nil {
		return result, database.MapError(err)
	}
	data := make([]GenerationSimple, 0, len(rows))
	for i := range rows {
		data = append(data, *GenerationEntityToSimple(&rows[i]))
	}
	return response.PageResponse[GenerationSimple]{Data: data, Total: total, Page: search.Page, PageSize: search.PageSize}, nil
}

func (r *Repository) attachPoster(ctx context.Context, id, owner, posterID int64) error {
	res := r.db.WithContext(ctx).Model(&Generation{}).
		Where("id = ? and owner_id = ?", id, owner).
		Update("poster_id", posterID)
	if res.Error != nil {
		return database.MapError(res.Error)
	}
	if res.RowsAffected == 0 {
		return apperrors.NotFound("生成记录不存在")
	}
	return nil
}

func (r *Repository) find(ctx context.Context, id int64) (*Generation, error) {
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return nil, err
	}
	var entity Generation
	err = r.db.WithContext(ctx).Where("id = ? and owner_id = ?", id, owner).Take(&entity).Error
	if err != nil {
		if appErr, ok := apperrors.As(database.MapError(err)); ok && appErr.Class == apperrors.ClassNotFound {
			return nil, apperrors.NotFound("生成记录不存在")
		}
		return nil, database.MapError(err)
	}
	return &entity, nil
}
