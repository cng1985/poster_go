package generation

import (
	"time"

	"poster.app/studio/internal/modules/studio/shared"
)

type GenerationRequest struct {
	ID         int64  `json:"id,string"`
	Prompt     string `json:"prompt"`
	Style      string `json:"style"`
	SizePreset string `json:"size_preset"`
	SavePoster bool   `json:"save_poster"`
	Title      string `json:"title"`
}

type GenerationSearch struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type GenerationSimple struct {
	ID         int64     `json:"id,string"`
	Prompt     string    `json:"prompt"`
	Style      string    `json:"style"`
	SizePreset string    `json:"size_preset"`
	Provider   string    `json:"provider"`
	Model      string    `json:"model"`
	PosterID   int64     `json:"poster_id,string,omitempty"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

type GenerationResponse struct {
	ID         int64           `json:"id,string"`
	Prompt     string          `json:"prompt"`
	Style      string          `json:"style"`
	SizePreset string          `json:"size_preset"`
	Provider   string          `json:"provider"`
	Model      string          `json:"model"`
	PosterID   int64           `json:"poster_id,string,omitempty"`
	Document   shared.Document `json:"document"`
	Message    string          `json:"message"`
	CreatedAt  time.Time       `json:"created_at"`
}
