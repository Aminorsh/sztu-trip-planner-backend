package repository

import (
	"context"

	apperrors "github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"gorm.io/gorm"
)

type routeRepository struct {
	db *gorm.DB
}

func (r *routeRepository) Create(ctx context.Context, route *model.TripRoute) error {
	if err := r.db.WithContext(ctx).Create(route).Error; err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

func (r *routeRepository) FindByID(ctx context.Context, id uint64) (*model.TripRoute, error) {
	var route model.TripRoute
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&route).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	return &route, nil
}

func (r *routeRepository) FindByTripID(ctx context.Context, tripID uint64) ([]model.TripRoute, error) {
	var routes []model.TripRoute
	if err := r.db.WithContext(ctx).
		Where("trip_id = ? ", tripID).
		Order("created_at DESC").
		Find(&routes).Error; err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	return routes, nil
}

func (r *routeRepository) FindActiveByTripID(ctx context.Context, tripID uint64) (*model.TripRoute, error) {
	var route model.TripRoute
	if err := r.db.WithContext(ctx).
		Where("trip_id = ? AND is_active = ? ", tripID, true).
		First(&route).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	return &route, nil
}

func (r *routeRepository) Update(ctx context.Context, route *model.TripRoute) error {
	if err := r.db.WithContext(ctx).Save(route).Error; err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

func (r *routeRepository) DeactivateOtherRoutes(ctx context.Context, tripID uint64, currentRouteID uint64) error {
	if err := r.db.WithContext(ctx).
		Model(&model.TripRoute{}).
		Where("trip_id = ? AND id != ?", tripID, currentRouteID).
		Update("is_active", false).Error; err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

func NewRouteRepository(db *gorm.DB) RouteRepository {
	return &routeRepository{db: db}
}
