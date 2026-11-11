package analysis

import "encoding/json"

func AnalysisEntityToSimple(entity *Analysis) *AnalysisSimple {
	return &AnalysisSimple{
		ID: entity.ID, PosterID: entity.PosterID, Score: entity.Score,
		Summary: entity.Summary, Provider: entity.Provider, CreatedAt: entity.CreatedAt.UTC(),
	}
}

func AnalysisEntityToResponse(entity *Analysis) (*AnalysisResponse, error) {
	suggestions := make([]Suggestion, 0)
	if entity.Suggestions != "" {
		if err := json.Unmarshal([]byte(entity.Suggestions), &suggestions); err != nil {
			return nil, err
		}
	}
	return &AnalysisResponse{
		ID: entity.ID, PosterID: entity.PosterID, Score: entity.Score, Summary: entity.Summary,
		Suggestions: suggestions, Provider: entity.Provider, CreatedAt: entity.CreatedAt.UTC(),
	}, nil
}
