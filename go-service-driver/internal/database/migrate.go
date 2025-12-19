package database

import (
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&model.User{},
		&model.Trip{},
		&model.TripItem{},
		&model.Place{},
		&model.TripRoute{},
		&model.AmapPoiCache{},
	)
	if err != nil {
		return err
	}

	// 强制修正列类型为 TIME (MySQL)
	// GORM 有时会将 string + type:time 映射为其他类型，导致 '09:00:00' 被错误解析
	db.Exec("ALTER TABLE trip_items MODIFY COLUMN start_time TIME")
	db.Exec("ALTER TABLE trip_items MODIFY COLUMN end_time TIME")

	return nil
}
