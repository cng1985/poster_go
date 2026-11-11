package analysis

import "time"

type Suggestion struct {
	Level string `json:"level"`
	Text  string `json:"text"`
}

type AnalysisRequest struct {
	ID       int64 `json:"id,string"`
	PosterID int64 `json:"poster_id,string"`
}

type AnalysisSearch struct {
	PosterID int64 `json:"poster_id,string"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type AnalysisSimple struct {
	ID        int64     `json:"id,string"`
	PosterID  int64     `json:"poster_id,string"`
	Score     int       `json:"score"`
	Summary   string    `json:"summary"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
}

type AnalysisResponse struct {
	ID          int64        `json:"id,string"`
	PosterID    int64        `json:"poster_id,string"`
	Score       int          `json:"score"`
	Summary     string       `json:"summary"`
	Suggestions []Suggestion `json:"suggestions"`
	Provider    string       `json:"provider"`
	CreatedAt   time.Time    `json:"created_at"`
}
