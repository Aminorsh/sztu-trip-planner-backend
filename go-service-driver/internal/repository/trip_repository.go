package repository

import (
	"context"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"gorm.io/gorm"
)

type tripRepository struct {
	db *gorm.DB
}

func NewTripRepository(db *gorm.DB) TripRepository {
	return &tripRepository{db: db}
}

// Create 创建新行程
func (t *tripRepository) Create(ctx context.Context, trip *model.Trip) error {
	return t.db.WithContext(ctx).Create(trip).Error
}

// FindByID 根据ID查找行程
func (t *tripRepository) FindByID(ctx context.Context, id int) (*model.Trip, error) {
	var trip model.Trip
	err := t.db.WithContext(ctx).
		Where("id = ?  AND deleted_at IS NULL", id).
		First(&trip).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 返回 nil 表示未找到
		}
		return nil, err
	}
	return &trip, nil
}

// FindByUserID 根据用户ID查找行程（支持搜索）
func (t *tripRepository) FindByUserID(ctx context.Context, userID int, search string) ([]model.Trip, error) {
	var trips []model.Trip
	query := t.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID)

	if search != "" {
		query = query.Where("title LIKE ? OR description LIKE ?  OR destination_city LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	err := query.Order("created_at DESC").Find(&trips).Error
	return trips, err
}

// Update 更新行程信息
func (t *tripRepository) Update(ctx context.Context, trip *model.Trip) error {
	return t.db.WithContext(ctx).
		Model(trip).
		Updates(trip).Error
}

// SoftDelete 软删除行程
func (t *tripRepository) SoftDelete(ctx context.Context, id int, userID int) error {
	now := time.Now()
	return t.db.WithContext(ctx).
		Model(&model.Trip{}).
		Where("id = ? AND user_id = ? ", id, userID).
		Update("deleted_at", now).Error
}

// UpdateStats 更新行程的距离和时长统计
func (t *tripRepository) UpdateStats(ctx context.Context, tripID uint64, totalDistance float64, totalDuration int) error {
	return t.db.WithContext(ctx).
		Model(&model.Trip{}).
		Where("id = ?", tripID).
		Updates(map[string]interface{}{
			"total_distance":     totalDistance,
			"estimated_duration": totalDuration,
		}).Error
}
