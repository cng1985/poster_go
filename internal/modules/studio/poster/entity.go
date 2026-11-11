package poster

import "time"

type Poster struct {
	ID         int64  `gorm:"primaryKey;autoIncrement"`
	OwnerID    int64  `gorm:"not null"`
	Title      string `gorm:"size:80;not null"`
	Width      int    `gorm:"not null"`
	Height     int    `gorm:"not null"`
	CoverColor string `gorm:"size:16;not null"`
	Document   string `gorm:"type:text;not null"`
	Version    int64  `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (Poster) TableName() string { return "poster" }
