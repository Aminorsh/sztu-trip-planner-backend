// go-service-driver/internal/model/trip_item.go
package model

import "time"

type TripItem struct {
	ID      uint64 `gorm:"primaryKey"`
	TripID  uint64 `gorm:"not null;index: idx_trip_day"`
	PlaceID uint64 `gorm:"not null"`

	DayNumber int `gorm:"not null;index:idx_trip_day"` // 第几天
	Sequence  int `gorm:"not null"`                    // 当天第几个

	StartTime string `gorm:"type:time"` // "09:00"
	EndTime   string `gorm:"type:time"` // "11:00"
	Note      string `gorm:"type:text"`

	ItemType string `gorm:"type:varchar(20)"` // scenic/restaurant/hotel

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	// 关联
	Place Place `gorm:"foreignKey:PlaceID"`
}

func (TripItem) TableName() string {
	return "trip_items"
}
