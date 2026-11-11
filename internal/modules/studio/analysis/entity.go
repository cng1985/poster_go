package analysis

import "time"

type Analysis struct {
	ID          int64  `gorm:"primaryKey;autoIncrement"`
	OwnerID     int64  `gorm:"not null"`
	PosterID    int64  `gorm:"not null"`
	Score       int    `gorm:"not null"`
	Summary     string `gorm:"size:300;not null"`
	Suggestions string `gorm:"type:text;not null"`
	Provider    string `gorm:"size:40;not null"`
	CreatedAt   time.Time
}

func (Analysis) TableName() string { return "poster_analysis" }
