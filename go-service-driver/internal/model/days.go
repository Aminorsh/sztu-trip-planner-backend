package model

import (
	"time"

	"gorm.io/gorm"
)

type Day struct {
	ID        uint64    `gorm:"primaryKey"`
	TripID    uint64    `gorm:"not null;index"`
	DayNumber int       `gorm:"not null"` // 第几天，从1开始
	Date      time.Time `gorm:"type:date"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// 关联关系
	Items []TripItem `gorm:"foreignKey:DayID"`
}

func (Day) TableName() string {
	return "days"
}
