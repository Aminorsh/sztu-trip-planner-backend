package model

import "time"

type User struct {
	ID           uint64 `gorm:"primaryKey"`
	Username     string `gorm:"size:100;uniqueIndex;not null"`
	Email        string `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	DisplayName  string
	AvatarURL    string
	Status       string `gorm:"default:'active'"`
	Bio          string `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
	DeletedAt    *time.Time `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}
