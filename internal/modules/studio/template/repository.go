package template

import (
	"context"
	"strings"

	"gorm.io/gorm"

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

func (r *Repository) One(ctx context.Context, request *TemplateRequest) (response.Response[TemplateResponse], error) {
	var result response.Response[TemplateResponse]
	var entity Template
	err := r.db.WithContext(ctx).Where("id = ? and status = ?", request.ID, "published").Take(&entity).Error
	if err != nil {
		return result, database.MapError(err)
	}
	item, err := TemplateEntityToResponse(&entity)
	if err != nil {
		return result, err
	}
	return response.OK(*item), nil
}

func (r *Repository) Search(ctx context.Context, search *TemplateSearch) (response.PageResponse[TemplateSimple], error) {
	var result response.PageResponse[TemplateSimple]
	page, pageSize, err := domain.NormalizePage(search.Page, search.PageSize)
	if err != nil {
		return result, err
	}
	search.Page, search.PageSize = page, pageSize
	makeQuery := func() *gorm.DB {
		query := r.db.WithContext(ctx).Model(&Template{}).Where("status = ?", "published")
		if category := strings.TrimSpace(search.Category); category != "" {
			query = query.Where("category = ?", category)
		}
		if name := strings.TrimSpace(search.Name); name != "" {
			clause, arg := database.Contains("name", name)
			query = query.Where(clause, arg)
		}
		return query
	}
	var total int64
	if err := makeQuery().Count(&total).Error; err != nil {
		return result, database.MapError(err)
	}
	rows := make([]Template, 0)
	err = makeQuery().
		Select("id", "code", "name", "category", "summary", "width", "height", "cover_color").
		Order("sort_no asc, id asc").
		Offset(domain.Offset(page, pageSize)).
		Limit(pageSize).
		Find(&rows).Error
	if err != nil {
		return result, database.MapError(err)
	}
	data := make([]TemplateSimple, 0, len(rows))
	for i := range rows {
		data = append(data, *TemplateEntityToSimple(&rows[i]))
	}
	return response.PageResponse[TemplateSimple]{Data: data, Total: total, Page: page, PageSize: pageSize}, nil
}

func requirePositive(id int64) error {
	if id < 1 {
		return apperrors.Invalid("模板标识不正确")
	}
	return nil
}
