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
	UpdatePassword(ctx context.Context, userID uint, newPasswordHash string) error
	UpdateStatus(ctx context.Context, userID uint, status string) error
	UpdateLastLogin(ctx context.Context, userID uint) error
	SoftDelete(ctx context.Context, userID uint) error
	Restore(ctx context.Context, userID uint) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}

// TripRepository 行程数据访问接口
type TripRepository interface {
	Create(ctx context.Context, trip *model.Trips) error
	FindByID(ctx context.Context, id int) (*model.Trips, error)
	FindByUserID(ctx context.Context, userID int, search string) ([]model.Trips, error)
	Update(ctx context.Context, trip *model.Trips) error
	SoftDelete(ctx context.Context, id, userID int) error
}
