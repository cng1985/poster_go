package template

import "time"

type Template struct {
	ID         int64  `gorm:"primaryKey;autoIncrement"`
	Code       string `gorm:"size:40;not null"`
	Name       string `gorm:"size:80;not null"`
	Category   string `gorm:"size:20;not null"`
	Summary    string `gorm:"size:200;not null"`
	Width      int    `gorm:"not null"`
	Height     int    `gorm:"not null"`
	CoverColor string `gorm:"size:16;not null"`
	Document   string `gorm:"type:text;not null"`
	SortNo     int    `gorm:"not null"`
	Status     string `gorm:"size:16;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (Template) TableName() string { return "poster_template" }
