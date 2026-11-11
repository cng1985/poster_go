package asset

import "time"

type AssetUploadRequest struct {
	FileName string `json:"-"`
	Bytes    []byte `json:"-"`
}

type AssetRequest struct {
	ID int64 `json:"id,string"`
}

type AssetSearch struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type AssetSimple struct {
	ID          int64     `json:"id,string"`
	Name        string    `json:"name"`
	ContentType string    `json:"content_type"`
	ByteSize    int64     `json:"byte_size"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	CreatedAt   time.Time `json:"created_at"`
}

type AssetResponse struct {
	ID          int64     `json:"id,string"`
	Name        string    `json:"name"`
	ContentType string    `json:"content_type"`
	ByteSize    int64     `json:"byte_size"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	CreatedAt   time.Time `json:"created_at"`
}
