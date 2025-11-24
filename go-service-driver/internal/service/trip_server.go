package service

import (
	"errors"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"gorm.io/gorm"
)

type TripService struct {
	DB *gorm.DB
}

func NewTripService(db *gorm.DB) *TripService {
	return &TripService{
		DB: db,
	}
}

func (s *TripService) CreateTrip(userID int, req dto.CreateTripRequest) (dto.TripResponse, error) {
	trip := model.Trips{
		UserID:             userID,
		Title:              req.Title,
		Description:        req.Description,
		StartDate:          *req.StartDate,
		EndDate:            *req.EndDate,
		OriginAddress:      req.OriginAddress,
		DestinationAddress: req.DestinationAddress,
		DestinationCity:    req.DestinationCity,
		DestinationCountry: req.DestinationCountry,
	}

	if req.IsPublic != nil {
		trip.IsPublic = *req.IsPublic
	}

	if req.StartDate != nil && req.EndDate != nil {
		days := int(req.EndDate.Sub(*req.StartDate).Hours()/24) + 1
		trip.Days = days
	}

	if err := s.DB.Create(&trip).Error; err != nil {
		return dto.TripResponse{}, err
	}

	response := dto.TripResponse{
		ID:                 trip.ID,
		UserID:             trip.UserID,
		Title:              trip.Title,
		Description:        trip.Description,
		StartDate:          &trip.StartDate,
		EndDate:            &trip.EndDate,
		Days:               &trip.Days,
		IsPublic:           trip.IsPublic,
		DestinationCity:    trip.DestinationCity,
		DestinationCountry: trip.DestinationCountry,
		CreatedAt:          trip.CreatedAt,
		UpdatedAt:          trip.UpdatedAt,
	}

	return response, nil
}

func (s *TripService) ListTrips(userID int, search string) (dto.TripListResponse, error) {
	var trips []model.Trips
	query := s.DB.Where("user_id = ? AND deleted_at IS NULL", userID)

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	if err := query.Find(&trips).Error; err != nil {
		return dto.TripListResponse{}, err
	}

	var tripResponses []dto.TripResponse
	for _, trip := range trips {
		days := trip.Days
		tripResponses = append(tripResponses, dto.TripResponse{
			ID:                 trip.ID,
			UserID:             trip.UserID,
			Title:              trip.Title,
			Description:        trip.Description,
			StartDate:          &trip.StartDate,
			EndDate:            &trip.EndDate,
			Days:               &days,
			IsPublic:           trip.IsPublic,
			DestinationCity:    trip.DestinationCity,
			DestinationCountry: trip.DestinationCountry,
			CreatedAt:          trip.CreatedAt,
			UpdatedAt:          trip.UpdatedAt,
		})
	}

	response := dto.TripListResponse{
		Total: len(tripResponses),
		Data:  tripResponses,
	}

	return response, nil
}

func (s *TripService) DeleteTrip(userID int, tripID int) error {
	var trip model.Trips

	if err := s.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", tripID, userID).First(&trip).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("no permission to delete this trip or trip does not exist")
		}
		return err
	}

	now := time.Now()
	return s.DB.Model(&trip).Update("deleted_at", &now).Error
}
