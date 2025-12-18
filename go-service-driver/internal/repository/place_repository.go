package repository

import (
	"context"
	"errors"

	apperrors "github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"gorm.io/gorm"
)

type placeRepository struct {
	db *gorm.DB
}

// UpdateFields implements PlaceRepository.
func (p *placeRepository) UpdateFields(ctx context.Context, placeID uint64, updates map[string]any) error {
	if err := p.db.WithContext(ctx).Model(&model.Place{}).Where("id = ?", placeID).Updates(updates).Error; err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// Update implements PlaceRepository.
func (p *placeRepository) Update(ctx context.Context, place *model.Place) error {
	if err := p.db.WithContext(ctx).Save(place).Error; err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// Create implements PlaceRepository.
func (p *placeRepository) Create(ctx context.Context, place *model.Place) error {
	if err := p.db.WithContext(ctx).Create(place).Error; err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// FindPlaceByPoiID implements PlaceRepository.
func (p *placeRepository) FindPlaceByPoiID(ctx context.Context, poiID string) (*model.Place, error) {
	var place model.Place
	if err := p.db.WithContext(ctx).Where("amap_poi_id = ?", poiID).First(&place).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	return &place, nil
}

// FindByID implements PlaceRepository.
func (p *placeRepository) FindByID(ctx context.Context, id uint) (*model.Place, error) {
	var place model.Place
	if err := p.db.WithContext(ctx).Where("id = ?", id).First(&place).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	return &place, nil
}

// NewPlaceRepository creates a new instance of PlaceRepository
func NewPlaceRepository(db *gorm.DB) PlaceRepository {
	return &placeRepository{db: db}
}
