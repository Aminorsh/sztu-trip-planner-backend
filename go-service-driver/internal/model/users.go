package model

import "time"

type User struct {
	ID           uint64 `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex;not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	DisplayName  string
	AvatarURL    string
	Status       string `gorm:"default:'active'"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
	DeletedAt    *time.Time `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}
