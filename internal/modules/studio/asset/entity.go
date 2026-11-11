package asset

import "time"

type Asset struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`
	OwnerID      int64  `gorm:"not null"`
	StorageKey   string `gorm:"size:200;not null"`
	OriginalName string `gorm:"size:120;not null"`
	ContentType  string `gorm:"size:40;not null"`
	ByteSize     int64  `gorm:"not null"`
	Width        int    `gorm:"not null"`
	Height       int    `gorm:"not null"`
	CreatedAt    time.Time
}

func (Asset) TableName() string { return "asset" }
