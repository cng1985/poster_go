package poster

import (
	"time"

	"poster.app/studio/internal/modules/studio/shared"
)

type PosterRequest struct {
	ID       int64           `json:"id,string"`
	Title    string          `json:"title"`
	Document shared.Document `json:"document"`
	Version  int64           `json:"version"`
}

type PosterDuplicateRequest struct {
	ID int64 `json:"id,string"`
}

type PosterSearch struct {
	Title    string `json:"title"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Sort     string `json:"sort"`
	Order    string `json:"order"`
}

type PosterSimple struct {
	ID         int64     `json:"id,string"`
	Title      string    `json:"title"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	CoverColor string    `json:"cover_color"`
	Version    int64     `json:"version"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type PosterResponse struct {
	ID         int64           `json:"id,string"`
	Title      string          `json:"title"`
	Width      int             `json:"width"`
	Height     int             `json:"height"`
	CoverColor string          `json:"cover_color"`
	Document   shared.Document `json:"document"`
	Version    int64           `json:"version"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
