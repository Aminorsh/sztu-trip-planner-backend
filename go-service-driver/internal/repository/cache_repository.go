package repository

import (
	"context"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"gorm.io/gorm"
)

type amapPoiCacheRepository struct {
	db *gorm.DB
}

// UpdateFields implements AmapPoiCacheRepository.
func (a *amapPoiCacheRepository) UpdateFields(ctx context.Context, cacheID uint64, updates map[string]any) error {
	return a.db.WithContext(ctx).Model(&model.AmapPoiCache{}).Where("id = ?", cacheID).Updates(updates).Error
}

// Create implements AmapPoiCacheRepository.
func (a *amapPoiCacheRepository) Create(ctx context.Context, cache *model.AmapPoiCache) error {
	return a.db.WithContext(ctx).Create(cache).Error
}

// FindByCacheKey implements AmapPoiCacheRepository.
func (a *amapPoiCacheRepository) FindByCacheKey(ctx context.Context, cacheKey string) (*model.AmapPoiCache, error) {
	var cache model.AmapPoiCache
	if err := a.db.WithContext(ctx).Where("cache_key = ? AND expire_at > ?", cacheKey, time.Now()).First(&cache).Error; err != nil {
		return nil, err
	}
	return &cache, nil
}

func NewAmapPoiCacheRepository(db *gorm.DB) AmapPoiCacheRepository {
	return &amapPoiCacheRepository{db: db}
}
