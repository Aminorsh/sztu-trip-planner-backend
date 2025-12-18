package database

import (
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Trip{},
		&model.TripItem{},
		&model.Place{},
		&model.TripRoute{},
		&model.AmapPoiCache{},
	)
}
