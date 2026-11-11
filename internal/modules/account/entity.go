package account

import "time"

type Account struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`
	Username     string `gorm:"size:32;not null"`
	PasswordHash string `gorm:"size:100;not null"`
	DisplayName  string `gorm:"size:40;not null"`
	Status       string `gorm:"size:16;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Account) TableName() string { return "account" }

type Session struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`
	AccountID int64  `gorm:"not null"`
	TokenHash string `gorm:"size:64;not null"`
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (Session) TableName() string { return "session" }
