package template

import "poster.app/studio/internal/modules/studio/shared"

type TemplateRequest struct {
	ID int64 `json:"id,string"`
}

type TemplateSearch struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type TemplateSimple struct {
	ID         int64  `json:"id,string"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	Summary    string `json:"summary"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	CoverColor string `json:"cover_color"`
}

type TemplateResponse struct {
	ID         int64           `json:"id,string"`
	Code       string          `json:"code"`
	Name       string          `json:"name"`
	Category   string          `json:"category"`
	Summary    string          `json:"summary"`
	Width      int             `json:"width"`
	Height     int             `json:"height"`
	CoverColor string          `json:"cover_color"`
	Document   shared.Document `json:"document"`
}
