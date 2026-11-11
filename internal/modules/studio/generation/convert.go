package generation

import "poster.app/studio/internal/modules/studio/shared"

func posterID(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func GenerationEntityToSimple(entity *Generation) *GenerationSimple {
	return &GenerationSimple{
		ID: entity.ID, Prompt: entity.Prompt, Style: entity.Style, SizePreset: entity.SizePreset,
		Provider: entity.Provider, Model: entity.Model, PosterID: posterID(entity.PosterID),
		Message: entity.Message, CreatedAt: entity.CreatedAt.UTC(),
	}
}

func GenerationEntityToResponse(entity *Generation) (*GenerationResponse, error) {
	doc, err := shared.Parse(entity.Document)
	if err != nil {
		return nil, err
	}
	return &GenerationResponse{
		ID: entity.ID, Prompt: entity.Prompt, Style: entity.Style, SizePreset: entity.SizePreset,
		Provider: entity.Provider, Model: entity.Model, PosterID: posterID(entity.PosterID),
		Document: doc, Message: entity.Message, CreatedAt: entity.CreatedAt.UTC(),
	}, nil
}
