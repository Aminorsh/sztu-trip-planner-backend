package repository

import (
	"context"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByEmailOrUsername(ctx context.Context, email, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	UpdateAvatarURL(ctx context.Context, userID uint, avatarURL string) error
	UpdatePassword(ctx context.Context, userID uint, newPasswordHash string) error
	UpdateStatus(ctx context.Context, userID uint, status string) error
	UpdateLastLogin(ctx context.Context, userID uint) error
	SoftDelete(ctx context.Context, userID uint) error
	Restore(ctx context.Context, userID uint) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

type PlaceRepository interface {
	Create(ctx context.Context, place *model.Place) error
	FindByID(ctx context.Context, id uint) (*model.Place, error)
	FindPlaceByPoiID(ctx context.Context, poiID string) (*model.Place, error)
	// Update(ctx context.Context, place *model.Place) error
	UpdateFields(ctx context.Context, placeID uint64, updates map[string]any) error
}

type AmapPoiCacheRepository interface {
	Create(ctx context.Context, cache *model.AmapPoiCache) error
	FindByCacheKey(ctx context.Context, cacheKey string) (*model.AmapPoiCache, error)
	UpdateFields(ctx context.Context, cacheID uint64, updates map[string]any) error
}

type RouteRepository interface {
	Create(ctx context.Context, route *model.TripRoute) error
	FindByID(ctx context.Context, id uint64) (*model.TripRoute, error)
	FindByTripID(ctx context.Context, tripID uint64) ([]model.TripRoute, error)
	FindActiveByTripID(ctx context.Context, tripID uint64) (*model.TripRoute, error)
	Update(ctx context.Context, route *model.TripRoute) error
	DeactivateOtherRoutes(ctx context.Context, tripID uint64, currentRouteID uint64) error
}

// TripRepository 行程数据访问接口
type TripRepository interface {
	Create(ctx context.Context, trip *model.Trip) error
	FindByID(ctx context.Context, id int) (*model.Trip, error)
	FindByUserID(ctx context.Context, userID int, search string) ([]model.Trip, error)
	Update(ctx context.Context, trip *model.Trip) error
	SoftDelete(ctx context.Context, id int, userID int) error
	UpdateStats(ctx context.Context, tripID uint64, totalDistance float64, totalDuration int) error
	UpdateFields(ctx context.Context, tripID uint64, updates map[string]any) error
	CreateTripItem(ctx context.Context, item *model.TripItem) error
	FindTripItemByID(ctx context.Context, itemID uint64) (*model.TripItem, error)
	UpdateTripItem(ctx context.Context, item *model.TripItem) error
	DeleteTripItem(ctx context.Context, itemID uint64) error
	DeleteTripItemsByDay(ctx context.Context, tripID uint64, dayNumber int) error
	// DeleteTripDay(ctx context.Context, tripID uint64, dayNumber int) error
	UploadCoverImage(ctx context.Context, tripID uint64, coverImagePath string) error
}

type AssistantRepository interface {
	GetSummary(ctx context.Context, userID uint64) (string, error)
	SetSummary(ctx context.Context, userID uint64, summary string) error
}
