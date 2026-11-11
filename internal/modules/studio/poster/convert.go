package poster

import "poster.app/studio/internal/modules/studio/shared"

func PosterEntityToSimple(entity *Poster) *PosterSimple {
	return &PosterSimple{
		ID: entity.ID, Title: entity.Title, Width: entity.Width, Height: entity.Height,
		CoverColor: entity.CoverColor, Version: entity.Version, UpdatedAt: entity.UpdatedAt.UTC(),
	}
}

func PosterEntityToResponse(entity *Poster) (*PosterResponse, error) {
	doc, err := shared.Parse(entity.Document)
	if err != nil {
		return nil, err
	}
	return &PosterResponse{
		ID: entity.ID, Title: entity.Title, Width: entity.Width, Height: entity.Height,
		CoverColor: entity.CoverColor, Document: doc, Version: entity.Version,
		CreatedAt: entity.CreatedAt.UTC(), UpdatedAt: entity.UpdatedAt.UTC(),
	}, nil
}
