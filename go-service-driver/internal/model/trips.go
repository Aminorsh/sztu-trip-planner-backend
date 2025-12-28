package model

import (
	"time"

	"gorm.io/gorm"
)

type Trip struct {
	ID          uint64    `gorm:"primaryKey"`
	UserID      uint64    `gorm:"not null;index"`
	Title       string    `gorm:"not null"`
	Description string    `gorm:"type:text"`
	StartDate   time.Time `gorm:"type:date"`
	EndDate     time.Time `gorm:"type:date"`
	CoverImage  string    `gorm:"type:text"`

	// 简化字段
	IsPublic bool   `gorm:"default:false"`
	Status   string `gorm:"default:'draft'"` // draft/active/completed

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	TotalDistance     float64 `gorm:"-"` // 非数据库字段，行程总距离，单位公里
	EstimatedDuration int     `gorm:"-"` // 非数据库字段，行程总时长，单位分钟

	// 关联关系
	Days  []Day      `gorm:"foreignKey:TripID"`
	Items []TripItem `gorm:"foreignKey:TripID"`
}

func (Trip) TableName() string {
	return "trips"
}
