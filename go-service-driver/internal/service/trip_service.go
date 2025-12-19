package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/repository"
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
			ID:          fmt.Sprintf("%d", trip.ID),
			Title:       trip.Title,
			Status:      trip.Status,
			CreatedAt:   trip.CreatedAt,
			LastSaved:   trip.UpdatedAt,
			Description: trip.Description,
			StartDate:   &trip.StartDate,
			EndDate:     &trip.EndDate,
			Days:        []dto.TripDay{}, // 空数组
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
		"title":  req.Title,
		"status": req.Status,
	}

	if err := s.tripRepo.UpdateFields(ctx, id, updates); err != nil {
		return err
	}

	// 更新行程项
	// 先删除所有旧的行程项，再创建新的
	for _, day := range req.Days {
		// 删除该天的所有行程项
		if err := s.tripRepo.DeleteTripItemsByDay(ctx, id, day.Day); err != nil {
			return err
		}

		// 创建新的行程项
		for i, item := range day.Items {
			placeID, err := s.getOrCreatePlaceID(ctx, item)
			if err != nil {
				return err
			}

			tripItem := &model.TripItem{
				TripID:    id,
				PlaceID:   placeID,
				DayNumber: day.Day,
				Sequence:  i + 1,
				StartTime: extractTime(item.Time),
				EndTime:   extractTime(item.EndTime),
				Note:      item.Note,
			}

			if err := s.tripRepo.CreateTripItem(ctx, tripItem); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *TripService) getOrCreatePlaceID(ctx context.Context, item dto.TripItem) (uint64, error) {
	if item.ID != "" {
		id, err := strconv.ParseUint(item.ID, 10, 64)
		if err == nil && id > 0 {
			return id, nil
		}
	}

	// Create new place
	if item.Name == "" {
		return 0, errors.NewInvalidRequestError("place name is required")
	}

	lng := item.Lnglat[0]
	lat := item.Lnglat[1]

	newPlace := &model.Place{
		Name:       item.Name,
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
func (s *TripService) AddTripItems(ctx context.Context, tripID string, dayNumber int, req *dto.AddTripItemRequest) ([]string, error) {
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

	itemIDs := make([]string, 0, len(req.Items))
	for i, item := range req.Items {
		placeID, err := s.getOrCreatePlaceID(ctx, item)
		if err != nil {
			return nil, err
		}

		tripItem := &model.TripItem{
			TripID:    id,
			PlaceID:   placeID,
			DayNumber: dayNumber,
			Sequence:  i + 1,
			StartTime: extractTime(item.Time),
			EndTime:   extractTime(item.EndTime),
			Note:      item.Note,
		}

		if err := s.tripRepo.CreateTripItem(ctx, tripItem); err != nil {
			return nil, err
		}

		itemIDs = append(itemIDs, fmt.Sprintf("%d", tripItem.ID))
	}

	return itemIDs, nil
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

	// 更新字段
	if req.Name != nil {
		placeID, err := s.getOrCreatePlaceID(ctx, dto.TripItem{
			Name:   *req.Name,
			Lnglat: [2]float64{},
		})
		if err != nil {
			return err
		}
		item.PlaceID = placeID
	}
	if req.Time != nil {
		item.StartTime = extractTime(*req.Time)
	}
	if req.EndTime != nil {
		item.EndTime = extractTime(*req.EndTime)
	}
	if req.Note != nil {
		item.Note = *req.Note
	}
	if req.Priority != nil {
		// 优先级暂时不存储
	}
	log.Printf("req: %v", req)
	log.Printf("updated item: %v", item)

	return s.tripRepo.UpdateTripItem(ctx, item)
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
		for _, item := range trip.Items {
			if item.DayNumber > maxDay {
				maxDay = item.DayNumber
			}
		}
		dayNumber = maxDay + 1
	}

	// 创建行程项
	for i, item := range req.Items {
		placeID, err := s.getOrCreatePlaceID(ctx, item)
		if err != nil {
			return 0, err
		}

		tripItem := &model.TripItem{
			TripID:    id,
			PlaceID:   placeID,
			DayNumber: dayNumber,
			Sequence:  i + 1,
			StartTime: extractTime(item.Time),
			EndTime:   extractTime(item.EndTime),
			Note:      item.Note,
		}

		if err := s.tripRepo.CreateTripItem(ctx, tripItem); err != nil {
			return 0, err
		}
	}

	return dayNumber, nil
}

// DeleteTripDay 删除指定天
func (s *TripService) DeleteTripDay(ctx context.Context, tripID string, dayNumber int) error {
	id, err := strconv.ParseUint(tripID, 10, 64)
	if err != nil {
		return errors.NewInvalidRequestError("invalid trip ID")
	}

	return s.tripRepo.DeleteTripItemsByDay(ctx, id, dayNumber)
}

// buildTripResponse 构建行程响应
func (s *TripService) buildTripResponse(trip *model.Trip) *dto.TripResponse {
	// 组织按天分组的行程项
	dayMap := make(map[int][]dto.TripItem)
	for _, item := range trip.Items {
		var lng, lat float64
		if item.Place.Longitude != nil {
			lng = *item.Place.Longitude
		}
		if item.Place.Latitude != nil {
			lat = *item.Place.Latitude
		}

		tripItem := dto.TripItem{
			ID:       fmt.Sprintf("%d", item.ID),
			Name:     item.Place.Name,
			Time:     calculateDateTime(trip.StartDate, item.DayNumber, item.StartTime),
			EndTime:  calculateDateTime(trip.StartDate, item.DayNumber, item.EndTime),
			Note:     item.Note,
			Priority: "medium", // 默认优先级
			Lnglat:   [2]float64{lng, lat},
		}

		dayMap[item.DayNumber] = append(dayMap[item.DayNumber], tripItem)
	}

	// 转换为有序的天数数组
	days := make([]dto.TripDay, 0)
	for dayNum := 1; dayNum <= len(dayMap); dayNum++ {
		if items, ok := dayMap[dayNum]; ok {
			days = append(days, dto.TripDay{
				Day:   dayNum,
				Items: items,
			})
		}
	}

	return &dto.TripResponse{
		ID:          fmt.Sprintf("%d", trip.ID),
		Title:       trip.Title,
		Status:      trip.Status,
		Days:        days,
		CreatedAt:   trip.CreatedAt,
		LastSaved:   trip.UpdatedAt,
		Description: trip.Description,
		StartDate:   &trip.StartDate,
		EndDate:     &trip.EndDate,
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
