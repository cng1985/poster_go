package generation

import "time"

type Generation struct {
	ID         int64 `gorm:"primaryKey;autoIncrement"`
	OwnerID    int64 `gorm:"not null"`
	PosterID   *int64
	Prompt     string `gorm:"size:500;not null"`
	Style      string `gorm:"size:20;not null"`
	SizePreset string `gorm:"size:20;not null"`
	Provider   string `gorm:"size:40;not null"`
	Model      string `gorm:"size:80;not null"`
	Document   string `gorm:"type:text;not null"`
	Message    string `gorm:"size:300;not null"`
	CreatedAt  time.Time
}

func (Generation) TableName() string { return "poster_generation" }
