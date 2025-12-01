package repository

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create implements UserRepository.
// If a user with the same email/username exists but is soft-deleted, it will be restored and updated.
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	// Check if a soft-deleted user exists with the same email or username
	var existingUser model.User
	err := r.db.WithContext(ctx).Unscoped().
		Where("(email = ? OR username = ?) AND deleted_at IS NOT NULL", user.Email, user.Username).
		First(&existingUser).Error

	if err == nil {
		// Soft-deleted user found, restore and update it
		existingUser.Username = user.Username
		existingUser.Email = user.Email
		existingUser.PasswordHash = user.PasswordHash
		existingUser.DisplayName = user.DisplayName
		existingUser.AvatarURL = user.AvatarURL
		existingUser.Status = user.Status
		existingUser.DeletedAt = nil
		existingUser.UpdatedAt = time.Now()

		if err := r.db.WithContext(ctx).Unscoped().Save(&existingUser).Error; err != nil {
			return apperrors.NewDatabaseError(err)
		}
		// Copy the restored user's ID back to the input user
		user.ID = existingUser.ID
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.NewDatabaseError(err)
	}

	// No soft-deleted user found, create new one
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// FindByEmail implements UserRepository.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewUserNotFoundError()
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	return &user, nil
}

// FindByEmailOrUsername implements UserRepository.
func (r *userRepository) FindByEmailOrUsername(ctx context.Context, email, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Where("(email = ? OR username = ?) AND deleted_at IS NULL", email, username).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // It is okay if user not found
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	return &user, nil
}

// FindByID implements UserRepository.
func (r *userRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewUserNotFoundError()
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	return &user, nil
}

// FindByUsername implements UserRepository.
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).
		Where("username = ? AND deleted_at IS NULL", username).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewUserNotFoundError()
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	return &user, nil
}

// SoftDelete implements UserRepository.
func (r *userRepository) SoftDelete(ctx context.Context, userID uint) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("deleted_at", &now)
	if result.Error != nil {
		return apperrors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewUserNotFoundError()
	}
	return nil
}

// Update implements UserRepository.
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", user.ID).
		Updates(user)
	if result.Error != nil {
		return apperrors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewUserNotFoundError()
	}
	return nil
}

// UpdatePassword implements UserRepository.
func (r *userRepository) UpdatePassword(ctx context.Context, userID uint, newPasswordHash string) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("password_hash", newPasswordHash)
	if result.Error != nil {
		return apperrors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewUserNotFoundError()
	}
	return nil
}

// UpdateStatus implements UserRepository.
func (r *userRepository) UpdateStatus(ctx context.Context, userID uint, status string) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("status", status)
	if result.Error != nil {
		return apperrors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewUserNotFoundError()
	}
	return nil
}

// UpdateLastLogin implements UserRepository.
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("last_login_at", &now)
	if result.Error != nil {
		return apperrors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewUserNotFoundError()
	}
	return nil
}

// Restore implements UserRepository.
// Restores a soft-deleted user by setting deleted_at to NULL.
func (r *userRepository) Restore(ctx context.Context, userID uint) error {
	result := r.db.WithContext(ctx).Unscoped().
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NOT NULL", userID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return apperrors.NewDatabaseError(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewUserNotFoundError()
	}
	return nil
}

// ExistsByEmail implements UserRepository.
// Checks if an active (non-deleted) user exists with the given email.
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error
	if err != nil {
		return false, apperrors.NewDatabaseError(err)
	}
	return count > 0, nil
}

// ExistsByUsername implements UserRepository.
// Checks if an active (non-deleted) user exists with the given username.
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("username = ? AND deleted_at IS NULL", username).
		Count(&count).Error
	if err != nil {
		return false, apperrors.NewDatabaseError(err)
	}
	return count > 0, nil
}
