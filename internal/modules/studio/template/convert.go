package template

import "poster.app/studio/internal/modules/studio/shared"

func TemplateEntityToSimple(entity *Template) *TemplateSimple {
	return &TemplateSimple{
		ID: entity.ID, Code: entity.Code, Name: entity.Name, Category: entity.Category,
		Summary: entity.Summary, Width: entity.Width, Height: entity.Height, CoverColor: entity.CoverColor,
	}
}

func TemplateEntityToResponse(entity *Template) (*TemplateResponse, error) {
	doc, err := shared.Parse(entity.Document)
	if err != nil {
		return nil, err
	}
	return &TemplateResponse{
		ID: entity.ID, Code: entity.Code, Name: entity.Name, Category: entity.Category,
		Summary: entity.Summary, Width: entity.Width, Height: entity.Height,
		CoverColor: entity.CoverColor, Document: doc,
	}, nil
}
