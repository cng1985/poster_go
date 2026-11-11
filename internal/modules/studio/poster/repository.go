package poster

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"poster.app/studio/internal/modules/studio/shared"
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

func (r *Repository) Create(ctx context.Context, request *PosterRequest) (response.Response[PosterResponse], error) {
	var result response.Response[PosterResponse]
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return result, err
	}
	raw, err := shared.Prepare(&request.Document)
	if err != nil {
		return result, err
	}
	entity := &Poster{
		OwnerID: owner, Title: request.Title, Width: request.Document.Width, Height: request.Document.Height,
		CoverColor: request.Document.Background, Document: raw, Version: 1,
	}
	if err := r.db.WithContext(ctx).Omit(clause.Associations).Create(entity).Error; err != nil {
		return result, database.MapError(err)
	}
	item, err := PosterEntityToResponse(entity)
	if err != nil {
		return result, err
	}
	return response.OK(*item), nil
}

func (r *Repository) Update(ctx context.Context, request *PosterRequest) (response.Response[PosterResponse], error) {
	var result response.Response[PosterResponse]
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return result, err
	}
	raw, err := shared.Prepare(&request.Document)
	if err != nil {
		return result, err
	}
	now := time.Now().UTC()
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Poster{}).
			Where("id = ? and owner_id = ? and version = ?", request.ID, owner, request.Version).
			Updates(map[string]any{
				"title":       request.Title,
				"width":       request.Document.Width,
				"height":      request.Document.Height,
				"cover_color": request.Document.Background,
				"document":    raw,
				"version":     request.Version + 1,
				"updated_at":  now,
			})
		if res.Error != nil {
			return database.MapError(res.Error)
		}
		if res.RowsAffected == 0 {
			var count int64
			if err := tx.Model(&Poster{}).Where("id = ? and owner_id = ?", request.ID, owner).Count(&count).Error; err != nil {
				return database.MapError(err)
			}
			if count == 0 {
				return apperrors.NotFound("海报不存在")
			}
			return apperrors.Conflict(10004, "海报已被更新，请刷新后再保存")
		}
		var stored Poster
		if err := tx.Where("id = ? and owner_id = ?", request.ID, owner).Take(&stored).Error; err != nil {
			return database.MapError(err)
		}
		item, err := PosterEntityToResponse(&stored)
		if err != nil {
			return err
		}
		result = response.OK(*item)
		return nil
	})
	if err != nil {
		return response.Response[PosterResponse]{}, err
	}
	return result, nil
}

func (r *Repository) One(ctx context.Context, request *PosterRequest) (response.Response[PosterResponse], error) {
	var result response.Response[PosterResponse]
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return result, err
	}
	var entity Poster
	err = r.db.WithContext(ctx).Where("id = ? and owner_id = ?", request.ID, owner).Take(&entity).Error
	if err != nil {
		if appErr, ok := apperrors.As(database.MapError(err)); ok && appErr.Class == apperrors.ClassNotFound {
			return result, apperrors.NotFound("海报不存在")
		}
		return result, database.MapError(err)
	}
	item, err := PosterEntityToResponse(&entity)
	if err != nil {
		return result, err
	}
	return response.OK(*item), nil
}

func (r *Repository) Search(ctx context.Context, search *PosterSearch) (response.PageResponse[PosterSimple], error) {
	var result response.PageResponse[PosterSimple]
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return result, err
	}
	page, pageSize, err := domain.NormalizePage(search.Page, search.PageSize)
	if err != nil {
		return result, err
	}
	search.Page, search.PageSize = page, pageSize
	order, err := domain.OrderBy(search.Sort, search.Order, map[string]string{
		"updated_at": "updated_at",
		"created_at": "created_at",
		"title":      "title",
	}, "updated_at", "desc")
	if err != nil {
		return result, err
	}
	makeQuery := func() *gorm.DB {
		query := r.db.WithContext(ctx).Model(&Poster{}).Where("owner_id = ?", owner)
		if title := strings.TrimSpace(search.Title); title != "" {
			clause, arg := database.Contains("title", title)
			query = query.Where(clause, arg)
		}
		return query
	}
	var total int64
	if err := makeQuery().Count(&total).Error; err != nil {
		return result, database.MapError(err)
	}
	rows := make([]Poster, 0)
	err = makeQuery().
		Select("id", "title", "width", "height", "cover_color", "version", "updated_at").
		Order(order).
		Offset(domain.Offset(search.Page, search.PageSize)).
		Limit(search.PageSize).
		Find(&rows).Error
	if err != nil {
		return result, database.MapError(err)
	}
	data := make([]PosterSimple, 0, len(rows))
	for i := range rows {
		data = append(data, *PosterEntityToSimple(&rows[i]))
	}
	return response.PageResponse[PosterSimple]{Data: data, Total: total, Page: search.Page, PageSize: search.PageSize}, nil
}

func (r *Repository) Delete(ctx context.Context, request *PosterRequest) (response.Response[PosterResponse], error) {
	var result response.Response[PosterResponse]
	owner, err := domain.ActorID(ctx)
	if err != nil {
		return result, err
	}
	var entity Poster
	err = r.db.WithContext(ctx).Where("id = ? and owner_id = ?", request.ID, owner).Take(&entity).Error
	if err != nil {
		if appErr, ok := apperrors.As(database.MapError(err)); ok && appErr.Class == apperrors.ClassNotFound {
			return result, apperrors.NotFound("海报不存在")
		}
		return result, database.MapError(err)
	}
	res := r.db.WithContext(ctx).Where("id = ? and owner_id = ?", request.ID, owner).Delete(&Poster{})
	if res.Error != nil {
		return result, database.MapError(res.Error)
	}
	if res.RowsAffected == 0 {
		return result, apperrors.NotFound("海报不存在")
	}
	item, err := PosterEntityToResponse(&entity)
	if err != nil {
		return result, err
	}
	return response.OK(*item), nil
}
