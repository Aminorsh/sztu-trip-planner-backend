package repository

import (
	"context"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"gorm.io/gorm"
)

type tripRepository struct {
	db *gorm.DB
}

// Create implements TripRepository.
func (t *tripRepository) Create(ctx context.Context, trip *model.Trips) error {
	panic("unimplemented")
}

// FindByID implements TripRepository.
func (t *tripRepository) FindByID(ctx context.Context, id int) (*model.Trips, error) {
	panic("unimplemented")
}

// FindByUserID implements TripRepository.
func (t *tripRepository) FindByUserID(ctx context.Context, userID int, search string) ([]model.Trips, error) {
	panic("unimplemented")
}

// SoftDelete implements TripRepository.
func (t *tripRepository) SoftDelete(ctx context.Context, id int, userID int) error {
	panic("unimplemented")
}

// Update implements TripRepository.
func (t *tripRepository) Update(ctx context.Context, trip *model.Trips) error {
	panic("unimplemented")
}

func NewTripRepository(db *gorm.DB) TripRepository {
	return &tripRepository{db: db}
}
