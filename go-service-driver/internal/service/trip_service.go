package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/repository"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/utils"
)

type TripService struct {
	tripRepo  repository.TripRepository
	placeRepo repository.PlaceRepository
}

func NewTripService(tripRepo repository.TripRepository, placeRepo repository.PlaceRepository) *TripService {
	return &TripService{
		tripRepo:  tripRepo,
		placeRepo: placeRepo,
	}
}

// CreateTrip 创建新行程
func (s *TripService) CreateTrip(ctx context.Context, userID uint64, req *dto.CreateTripRequest) (*dto.TripResponse, error) {
	trip := &model.Trip{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		CoverImage:  "/static/system/trips/joshua-hibbert-gwzj_ftMpWM-unsplash.jpg",
		Status:      "draft",
	}

	if req.StartDate != nil {
		trip.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		trip.EndDate = *req.EndDate
	}
	if req.IsPublic != nil {
		trip.IsPublic = *req.IsPublic
	}

	if err := s.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}

	return s.buildTripResponse(trip), nil
}

// GetTrips 获取用户的行程列表
func (s *TripService) GetTrips(ctx context.Context, userID uint64, search string) (*dto.TripListResponse, error) {
	trips, err := s.tripRepo.FindByUserID(ctx, int(userID), search)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TripResponse, 0, len(trips))
	for _, trip := range trips {
		// 对于列表，只返回基本信息，不加载行程项
		responses = append(responses, dto.TripResponse{
			ID:            fmt.Sprintf("%d", trip.ID),
			Title:         trip.Title,
			Status:        trip.Status,
			CoverImageURL: trip.CoverImage,
			CreatedAt:     trip.CreatedAt,
			LastSaved:     trip.UpdatedAt,
			Description:   trip.Description,
			StartDate:     &trip.StartDate,
			EndDate:       &trip.EndDate,
			Days:          []dto.TripDay{}, // 空数组
		})
	}

	return &dto.TripListResponse{
		Total: len(responses),
		Data:  responses,
	}, nil
}

// GetTrip 获取行程详情
func (s *TripService) GetTrip(ctx context.Context, tripID string) (*dto.TripResponse, error) {
	id, err := strconv.ParseUint(tripID, 10, 64)
	if err != nil {
		return nil, errors.NewInvalidRequestError("invalid trip ID")
	}

	trip, err := s.tripRepo.FindByID(ctx, int(id))
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, errors.NewTripNotFoundError()
	}

	return s.buildTripResponse(trip), nil
}

// UpdateTrip 更新行程
func (s *TripService) UpdateTrip(ctx context.Context, tripID string, req *dto.UpdateTripRequest) error {
	id, err := strconv.ParseUint(tripID, 10, 64)
	if err != nil {
		return errors.NewInvalidRequestError("invalid trip ID")
	}

	trip, err := s.tripRepo.FindByID(ctx, int(id))
	if err != nil {
		return err
	}
	if trip == nil {
		return errors.NewTripNotFoundError()
	}

	// 更新基本信息
	updates := map[string]any{
		"title":       req.Title,
		"description": req.Description,
	}
	if req.StartDate != nil {
		updates["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		updates["end_date"] = *req.EndDate
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := s.tripRepo.UpdateFields(ctx, id, updates); err != nil {
		return err
	}

	return nil
}

func (s *TripService) getOrCreatePlaceID(ctx context.Context, item *dto.AddTripItemRequest) (uint64, error) {
	if item.ID != "" {
		placeID, err := strconv.ParseUint(item.ID, 10, 64)
		if err != nil {
			return 0, errors.NewInvalidRequestError("invalid place ID")
		}
		// Optionally verify place exists in repository
		return placeID, nil
	}

	// Create new place
	if len(item.Lnglat) != 2 {
		return 0, errors.NewInvalidRequestError("coordinates required for custom place")
	}
	if item.Name == "" {
		item.Name = "未知地点"
	}

	lng := item.Lnglat[0]
	lat := item.Lnglat[1]

	newPlace := &model.Place{
		Name:       "（自定义）" + item.Name,
		Category:   model.PlaceCategoryOther,
		Address:    "自定义地点",
		Longitude:  &lng,
		Latitude:   &lat,
		DataSource: model.DataSourceUserCreated,
		AmapPoiID:  fmt.Sprintf("custom_%d", time.Now().UnixNano()),
	}

	if err := s.placeRepo.Create(ctx, newPlace); err != nil {
		return 0, err
	}

	return newPlace.ID, nil
}

// AddTripItems 添加行程项
func (s *TripService) AddTripItems(ctx context.Context, tripID string, dayNumber int, req *dto.AddTripItemRequest) (string, error) {
	id, err := strconv.ParseUint(tripID, 10, 64)
	if err != nil {
		return "", errors.NewInvalidRequestError("invalid trip ID")
	}

	trip, err := s.tripRepo.FindByID(ctx, int(id))
	if err != nil {
		return "", err
	}
	if trip == nil {
		return "", errors.NewTripNotFoundError()
	}

	// 确保 Day 存在
	day, err := s.tripRepo.FindDayByTripAndNumber(ctx, id, dayNumber)
	if err != nil {
		return "", err
	}
	if day == nil {
		return "", errors.NewInvalidRequestError(fmt.Sprintf("day %d does not exist for this trip", dayNumber))
	}

	// Get place ID (existing or create new)
	placeID, err := s.getOrCreatePlaceID(ctx, req)
	if err != nil {
		return "", err
	}

	// Calculate next sequence number automatically
	maxSeq := 0
	for _, item := range day.Items {
		if item.Sequence > maxSeq {
			maxSeq = item.Sequence
		}
	}

	tripItem := &model.TripItem{
		TripID:    id,
		DayID:     day.ID,
		PlaceID:   placeID,
		DayNumber: dayNumber,
		Sequence:  maxSeq + 1,
		StartTime: extractTime(req.Time),
		Name:      req.Name,
		Note:      req.Note,
	}

	if err := s.tripRepo.CreateTripItem(ctx, tripItem); err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", tripItem.ID), nil
}

// UpdateTripItem 更新行程项
func (s *TripService) UpdateTripItem(ctx context.Context, tripID, dayID, itemID string, req *dto.UpdateTripItemRequest) error {
	id, err := strconv.ParseUint(itemID, 10, 64)
	if err != nil {
		return errors.NewInvalidRequestError("invalid item ID")
	}

	item, err := s.tripRepo.FindTripItemByID(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.NewTripItemNotFoundError()
	}

	// 只在提供了 ID 或 Lnglat 时才更新 placeID
	if req.ID != nil || req.Lnglat != nil {
		// 如果提供了 ID 或 Lnglat，必须至少提供其中一个
		if (req.ID == nil || *req.ID == "") && req.Lnglat == nil {
			return errors.NewInvalidRequestError("either place ID or coordinates must be provided to update place")
		}

		var placeIDStr string
		var lnglat [2]float64
		if req.ID != nil {
			placeIDStr = *req.ID
		}
		if req.Lnglat != nil {
			lnglat = *req.Lnglat
		}

		placeID, err := s.getOrCreatePlaceID(ctx, &dto.AddTripItemRequest{
			ID:     placeIDStr,
			Name:   getStringValue(req.Name),
			Lnglat: lnglat,
		})
		if err != nil {
			return err
		}
		item.PlaceID = placeID
	}

	// 只更新提供的字段
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Time != nil {
		item.StartTime = extractTime(*req.Time)
	}
	if req.Note != nil {
		item.Note = *req.Note
	}

	return s.tripRepo.UpdateTripItem(ctx, item)
}

// getStringValue 获取字符串指针的值，如果为nil则返回空字符串
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// DeleteTripItem 删除行程项
func (s *TripService) DeleteTripItem(ctx context.Context, tripID, dayID, itemID string) error {
	id, err := strconv.ParseUint(itemID, 10, 64)
	if err != nil {
		return errors.NewInvalidRequestError("invalid item ID")
	}

	return s.tripRepo.DeleteTripItem(ctx, id)
}

// DeleteTrip 删除行程
func (s *TripService) DeleteTrip(ctx context.Context, tripID string, userID uint64) error {
	id, err := strconv.ParseUint(tripID, 10, 64)
	if err != nil {
		return errors.NewInvalidRequestError("invalid trip ID")
	}

	return s.tripRepo.SoftDelete(ctx, int(id), int(userID))
}

// AddTripDay 添加新的一天
func (s *TripService) AddTripDay(ctx context.Context, tripID string, req *dto.AddTripDayRequest) (int, error) {
	id, err := strconv.ParseUint(tripID, 10, 64)
	if err != nil {
		return 0, errors.NewInvalidRequestError("invalid trip ID")
	}

	trip, err := s.tripRepo.FindByID(ctx, int(id))
	if err != nil {
		return 0, err
	}
	if trip == nil {
		return 0, errors.NewTripNotFoundError()
	}

	dayNumber := req.Day
	// 如果没有指定天数，找到最大天数+1
	if dayNumber == 0 {
		maxDay := 0
		for _, day := range trip.Days {
			if day.DayNumber > maxDay {
				maxDay = day.DayNumber
			}
		}
		dayNumber = maxDay + 1
	}

	// 检查该天是否已存在
	existingDay, err := s.tripRepo.FindDayByTripAndNumber(ctx, id, dayNumber)
	if err != nil {
		return 0, err
	}
	if existingDay != nil {
		return dayNumber, nil // 已存在，直接返回
	}

	// 创建新的 Day 记录
	date := trip.StartDate.AddDate(0, 0, dayNumber-1)
	newDay := &model.Day{
		TripID:    id,
		DayNumber: dayNumber,
		Date:      date,
	}

	if err := s.tripRepo.CreateDay(ctx, newDay); err != nil {
		return 0, err
	}

	return dayNumber, nil
}

// DeleteTripDay 删除指定天
func (s *TripService) DeleteTripDay(ctx context.Context, tripID string, dayNumber int) error {
	id, err := strconv.ParseUint(tripID, 10, 64)
	if err != nil {
		return errors.NewInvalidRequestError("invalid trip ID")
	}

	// 查找要删除的 Day
	day, err := s.tripRepo.FindDayByTripAndNumber(ctx, id, dayNumber)
	if err != nil {
		return err
	}
	if day == nil {
		return errors.NewInvalidRequestError(fmt.Sprintf("day %d not found", dayNumber))
	}

	// 删除 Day（会级联删除关联的 TripItems）
	return s.tripRepo.DeleteDay(ctx, day.ID)
}

// buildTripResponse 构建行程响应
func (s *TripService) buildTripResponse(trip *model.Trip) *dto.TripResponse {
	// 使用 trip.Days 构建响应
	days := make([]dto.TripDay, 0, len(trip.Days))

	for _, day := range trip.Days {
		items := make([]dto.TripItem, 0, len(day.Items))

		for _, item := range day.Items {
			var lng, lat float64
			if item.Place.Longitude != nil {
				lng = *item.Place.Longitude
			}
			if item.Place.Latitude != nil {
				lat = *item.Place.Latitude
			}

			tripItem := dto.TripItem{
				ID:       fmt.Sprintf("%d", item.ID),
				Name:     item.Name,
				Time:     calculateDateTime(trip.StartDate, day.DayNumber, item.StartTime),
				Note:     item.Note,
				Priority: "medium", // 默认优先级
				Lnglat:   [2]float64{lng, lat},
			}

			items = append(items, tripItem)
		}

		days = append(days, dto.TripDay{
			Day:   day.DayNumber,
			Items: items,
		})
	}

	return &dto.TripResponse{
		ID:            fmt.Sprintf("%d", trip.ID),
		Title:         trip.Title,
		Status:        trip.Status,
		CoverImageURL: trip.CoverImage,
		Days:          days,
		CreatedAt:     trip.CreatedAt,
		LastSaved:     trip.UpdatedAt,
		Description:   trip.Description,
		StartDate:     &trip.StartDate,
		EndDate:       &trip.EndDate,
	}
}

func extractTime(timeStr string) string {
	if timeStr == "" {
		return ""
	}
	// 尝试解析 RFC3339 格式 (e.g., "2025-12-18T09:00:00Z")
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t.Format("15:04:05")
	}
	// 尝试解析 "2006-01-02 15:04" 格式
	if t, err := time.Parse("2006-01-02 15:04", timeStr); err == nil {
		return t.Format("15:04:05")
	}
	// 尝试解析 "2006-01-02 15:04:05" 格式
	if t, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
		return t.Format("15:04:05")
	}
	// 尝试解析 "15:04" 格式
	if t, err := time.Parse("15:04", timeStr); err == nil {
		return t.Format("15:04:05")
	}
	return timeStr
}

func calculateDateTime(startDate time.Time, dayNumber int, timeStr string) string {
	if timeStr == "" {
		return ""
	}
	if startDate.IsZero() {
		return timeStr
	}

	// 解析时间部分
	t, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		t, err = time.Parse("15:04", timeStr)
		if err != nil {
			// 尝试解析完整日期时间，以防万一
			if fullT, err := time.Parse(time.RFC3339, timeStr); err == nil {
				t = fullT
			} else if fullT, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
				t = fullT
			} else {
				return timeStr
			}
		}
	}

	// 计算日期: StartDate + (DayNumber - 1)
	date := startDate.AddDate(0, 0, dayNumber-1)

	// 组合日期和时间
	fullDateTime := time.Date(
		date.Year(), date.Month(), date.Day(),
		t.Hour(), t.Minute(), t.Second(), 0,
		startDate.Location(),
	)

	return fullDateTime.Format(time.RFC3339)
}

func (s *TripService) UpdateTripCoverImage(ctx context.Context, tripID string, file *multipart.FileHeader) (string, error) {
	id, err := strconv.ParseUint(tripID, 10, 64)
	if err != nil {
		return "", errors.NewInvalidRequestError("invalid trip ID")
	}

	trip, err := s.tripRepo.FindByID(ctx, int(id))
	if err != nil {
		return "", err
	}
	if trip == nil {
		return "", errors.NewTripNotFoundError()
	}

	oldCover := trip.CoverImage

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", errors.NewInvalidRequestError("unsupported image format")
	}

	coverImagePath := fmt.Sprintf("uploads/trips/trip_%s_%d%s", tripID, time.Now().Unix(), ext)

	if err := utils.SaveUploadedFile(file, coverImagePath); err != nil {
		return "", errors.NewInternalServerError(err)
	}

	coverImageURL := fmt.Sprintf("/static/trips/trip_%s_%d%s", tripID, time.Now().Unix(), ext)
	if err := s.tripRepo.UpdateCoverImage(ctx, id, coverImagePath); err != nil {
		_ = os.Remove(coverImagePath)
		return "", err
	}

	s.deleteOldCoverFile(oldCover)

	return coverImageURL, nil
}

func (s *TripService) deleteOldCoverFile(oldCover string) {
	if oldCover == "" || strings.HasPrefix(oldCover, "/static/system/") {
		return
	}

	if !strings.HasPrefix(oldCover, "/static/trips/") {
		return
	}

	oldCoverPath := strings.Replace(oldCover, "/static/trips/", "/uploads/trips/", 1)

	_ = os.Remove(oldCoverPath)
}

// CheckoutTripItem 打卡行程项（标记为已完成）
func (s *TripService) CheckoutTripItem(ctx context.Context, tripID, dayID, itemID string) error {
	id, err := strconv.ParseUint(itemID, 10, 64)
	if err != nil {
		return errors.NewInvalidRequestError("invalid item ID")
	}

	item, err := s.tripRepo.FindTripItemByID(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.NewTripItemNotFoundError()
	}

	// 验证参数一致性
	tripIDUint, err := strconv.ParseUint(tripID, 10, 64)
	if err != nil {
		return errors.NewInvalidRequestError("invalid trip ID")
	}
	if item.TripID != tripIDUint {
		return errors.NewInvalidRequestError("item does not belong to this trip")
	}

	dayIDUint, err := strconv.ParseUint(dayID, 10, 64)
	if err != nil {
		return errors.NewInvalidRequestError("invalid day ID")
	}
	if item.DayID != dayIDUint {
		return errors.NewInvalidRequestError("item does not belong to this day")
	}

	// 标记为已完成
	item.IsChecked = true

	return s.tripRepo.UpdateTripItem(ctx, item)
}

// UncheckoutTripItem 取消打卡行程项（标记为未完成）
func (s *TripService) UncheckoutTripItem(ctx context.Context, tripID, dayID, itemID string) error {
	id, err := strconv.ParseUint(itemID, 10, 64)
	if err != nil {
		return errors.NewInvalidRequestError("invalid item ID")
	}

	item, err := s.tripRepo.FindTripItemByID(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.NewTripItemNotFoundError()
	}

	// 验证参数一致性
	tripIDUint, err := strconv.ParseUint(tripID, 10, 64)
	if err != nil {
		return errors.NewInvalidRequestError("invalid trip ID")
	}
	if item.TripID != tripIDUint {
		return errors.NewInvalidRequestError("item does not belong to this trip")
	}

	dayIDUint, err := strconv.ParseUint(dayID, 10, 64)
	if err != nil {
		return errors.NewInvalidRequestError("invalid day ID")
	}
	if item.DayID != dayIDUint {
		return errors.NewInvalidRequestError("item does not belong to this day")
	}

	// 标记为未完成
	item.IsChecked = false

	return s.tripRepo.UpdateTripItem(ctx, item)
}
