package asset

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

func (r *Repository) Create(ctx context.Context, entity *Asset) (response.Response[AssetResponse], error) {
	var result response.Response[AssetResponse]
	if _, err := domain.ActorID(ctx); err != nil {
		return result, err
	}
	if err := r.db.WithContext(ctx).Omit(clause.Associations).Create(entity).Error; err != nil {
		return result, database.MapError(err)
	}
	return response.OK(*AssetEntityToResponse(entity)), nil
}

func (r *Repository) One(ctx context.Context, request *AssetRequest) (response.Response[AssetResponse], error) {
	var result response.Response[AssetResponse]
	entity, err := r.find(ctx, request.ID)
	if err != nil {
		return result, err
	}
	return response.OK(*AssetEntityToResponse(entity)), nil
}

func (r *Repository) Search(ctx context.Context, search *AssetSearch) (response.PageResponse[AssetSimple], error) {
	var result response.PageResponse[AssetSimple]
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
		return r.db.WithContext(ctx).Model(&Asset{}).Where("owner_id = ?", owner)
	}
	var total int64
	if err := makeQuery().Count(&total).Error; err != nil {
		return result, database.MapError(err)
	}
	rows := make([]Asset, 0)
	err = makeQuery().
		Order("id desc").
		Offset(domain.Offset(search.Page, search.PageSize)).
		Limit(search.PageSize).
		Find(&rows).Error
	if err != nil {
		return result, database.MapError(err)
	}
	data := make([]AssetSimple, 0, len(rows))
	for i := range rows {
		data = append(data, *AssetEntityToSimple(&rows[i]))
	}
	return response.PageResponse[AssetSimple]{Data: data, Total: total, Page: search.Page, PageSize: search.PageSize}, nil
}

func (r *Repository) Delete(ctx context.Context, request *AssetRequest) (response.Response[AssetResponse], error) {
	var result response.Response[AssetResponse]
	entity, err := r.find(ctx, request.ID)
	if err != nil {
		return result, err
	}
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return result, err
	}
	res := r.db.WithContext(ctx).Where("id = ? and owner_id = ?", entity.ID, owner).Delete(&Asset{})
	if res.Error != nil {
		return result, database.MapError(res.Error)
	}
	if res.RowsAffected == 0 {
		return result, apperrors.NotFound("素材不存在")
	}
	return response.OK(*AssetEntityToResponse(entity)), nil
}

func (r *Repository) find(ctx context.Context, id int64) (*Asset, error) {
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return nil, err
	}
	var entity Asset
	err = r.db.WithContext(ctx).Where("id = ? and owner_id = ?", id, owner).Take(&entity).Error
	if err != nil {
		if appErr, ok := apperrors.As(database.MapError(err)); ok && appErr.Class == apperrors.ClassNotFound {
			return nil, apperrors.NotFound("素材不存在")
		}
		return nil, database.MapError(err)
	}
	return &entity, nil
}
