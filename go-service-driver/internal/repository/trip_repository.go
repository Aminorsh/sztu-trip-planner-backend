package repository

import (
	"context"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"gorm.io/gorm"
)

type tripRepository struct {
	db *gorm.DB
}

// UpdateCoverImage implements TripRepository.
func (t *tripRepository) UpdateCoverImage(ctx context.Context, tripID uint64, coverImagePath string) error {
	return t.db.WithContext(ctx).
		Model(&model.Trip{}).
		Where("id = ?", tripID).
		Update("cover_image", coverImagePath).Error
}

// Create 创建新行程
func (t *tripRepository) Create(ctx context.Context, trip *model.Trip) error {
	return t.db.WithContext(ctx).Create(trip).Error
}

// FindByID 根据ID查找行程
func (t *tripRepository) FindByID(ctx context.Context, id int) (*model.Trip, error) {
	var trip model.Trip
	err := t.db.WithContext(ctx).
		Preload("Days", func(db *gorm.DB) *gorm.DB {
			return db.Order("day_number ASC")
		}).
		Preload("Days.Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence ASC")
		}).
		Preload("Days.Items.Place").
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
		query = query.Where("title LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%")
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
	return t.db.WithContext(ctx).
		Where("id = ? AND user_id = ? ", id, userID).
		Delete(&model.Trip{}).Error
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

// UpdateFields 更新行程的指定字段
func (t *tripRepository) UpdateFields(ctx context.Context, tripID uint64, updates map[string]any) error {
	return t.db.WithContext(ctx).
		Model(&model.Trip{}).
		Where("id = ?", tripID).
		Updates(updates).Error
}

// CreateTripItem 创建行程项
func (t *tripRepository) CreateTripItem(ctx context.Context, item *model.TripItem) error {
	return t.db.WithContext(ctx).Create(item).Error
}

// FindTripItemByID 根据ID查找行程项
func (t *tripRepository) FindTripItemByID(ctx context.Context, itemID uint64) (*model.TripItem, error) {
	var item model.TripItem
	if err := t.db.WithContext(ctx).Preload("Place").First(&item, itemID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

// UpdateTripItem 更新行程项
func (t *tripRepository) UpdateTripItem(ctx context.Context, item *model.TripItem) error {
	return t.db.WithContext(ctx).Save(item).Error
}

// DeleteTripItem 删除行程项
func (t *tripRepository) DeleteTripItem(ctx context.Context, itemID uint64) error {
	return t.db.WithContext(ctx).Delete(&model.TripItem{}, itemID).Error
}

// DeleteTripItemsByDay 删除指定天数的所有行程项
func (t *tripRepository) DeleteTripItemsByDay(ctx context.Context, tripID uint64, dayNumber int) error {
	return t.db.WithContext(ctx).
		Where("trip_id = ? AND day_number = ?", tripID, dayNumber).
		Delete(&model.TripItem{}).Error
}

// CreateDay 创建新的一天
func (t *tripRepository) CreateDay(ctx context.Context, day *model.Day) error {
	return t.db.WithContext(ctx).Create(day).Error
}

// FindDayByTripAndNumber 根据行程ID和天数查找Day
func (t *tripRepository) FindDayByTripAndNumber(ctx context.Context, tripID uint64, dayNumber int) (*model.Day, error) {
	var day model.Day
	err := t.db.WithContext(ctx).
		Where("trip_id = ? AND day_number = ? AND deleted_at IS NULL", tripID, dayNumber).
		First(&day).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &day, nil
}

// DeleteDay 删除指定的Day（会级联删除关联的TripItems）
func (t *tripRepository) DeleteDay(ctx context.Context, dayID uint64) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先删除该天的所有行程项
		if err := tx.Where("day_id = ?", dayID).Delete(&model.TripItem{}).Error; err != nil {
			return err
		}
		// 再删除Day记录
		return tx.Delete(&model.Day{}, dayID).Error
	})
}

// DeleteDaysByTripID 删除行程的所有Day
func (t *tripRepository) DeleteDaysByTripID(ctx context.Context, tripID uint64) error {
	return t.db.WithContext(ctx).Where("trip_id = ?", tripID).Delete(&model.Day{}).Error
}

func NewTripRepository(db *gorm.DB) TripRepository {
	return &tripRepository{db: db}
}
