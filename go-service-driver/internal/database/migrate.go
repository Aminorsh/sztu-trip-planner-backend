package database

import (
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	// 首先迁移基础表（不包含外键约束的部分）
	err := db.AutoMigrate(
		&model.User{},
		&model.Trip{},
		&model.Place{},
		&model.TripRoute{},
		&model.AmapPoiCache{},
	)
	if err != nil {
		return err
	}

	// 迁移 Day 表
	if err := db.AutoMigrate(&model.Day{}); err != nil {
		return err
	}

	// 迁移历史数据：为已存在的 trip_items 创建对应的 Day 记录
	if err := migrateExistingTripItemsToDays(db); err != nil {
		return err
	}

	// 最后迁移 TripItem（会创建外键约束）
	if err := db.AutoMigrate(&model.TripItem{}); err != nil {
		return err
	}

	// 强制修正列类型为 TIME (MySQL)
	// GORM 有时会将 string + type:time 映射为其他类型，导致 '09:00:00' 被错误解析
	db.Exec("ALTER TABLE trip_items MODIFY COLUMN start_time TIME")
	db.Exec("ALTER TABLE trip_items MODIFY COLUMN end_time TIME")

	return nil
}

// migrateExistingTripItemsToDays 为历史数据创建 Day 记录
func migrateExistingTripItemsToDays(db *gorm.DB) error {
	// 查询所有 trip_items 的唯一 (trip_id, day_number) 组合
	var results []struct {
		TripID    uint64
		DayNumber int
	}

	err := db.Table("trip_items").
		Select("DISTINCT trip_id, day_number").
		Where("day_number > 0").
		Scan(&results).Error
	if err != nil {
		return err
	}

	// 为每个唯一的 (trip_id, day_number) 创建 Day 记录
	for _, result := range results {
		// 检查是否已存在
		var count int64
		db.Model(&model.Day{}).
			Where("trip_id = ? AND day_number = ?", result.TripID, result.DayNumber).
			Count(&count)

		if count == 0 {
			// 获取 trip 的 start_date
			var trip model.Trip
			if err := db.First(&trip, result.TripID).Error; err != nil {
				// 如果找不到 trip，跳过
				continue
			}

			// 计算日期
			date := trip.StartDate.AddDate(0, 0, result.DayNumber-1)

			day := model.Day{
				TripID:    result.TripID,
				DayNumber: result.DayNumber,
				Date:      date,
			}

			if err := db.Create(&day).Error; err != nil {
				return err
			}

			// 更新该 trip_id 和 day_number 的所有 trip_items 的 day_id
			db.Table("trip_items").
				Where("trip_id = ? AND day_number = ?", result.TripID, result.DayNumber).
				Update("day_id", day.ID)
		}
	}

	return nil
}
