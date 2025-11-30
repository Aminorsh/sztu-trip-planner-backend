package dto

import "time"

type CreateTripRequest struct {
	Title              string     `json:"title" binding:"required"`
	Description        string     `json:"description"`
	StartDate          *time.Time `json:"start_date" binding:"required"`
	EndDate            *time.Time `json:"end_date" binding:"required"`
	IsPublic           *bool      `json:"is_public"`
	OriginAddress      string     `json:"origin_address"`
	DestinationAddress string     `json:"destination_address"`
	DestinationCity    string     `json:"destination_city"`
	DestinationCountry string     `json:"destination_country"`
}

type TripResponse struct {
	ID                 int        `json:"id"`
	UserID             int        `json:"user_id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	StartDate          *time.Time `json:"start_date"`
	EndDate            *time.Time `json:"end_date"`
	Days               *int       `json:"days"`
	IsPublic           bool       `json:"is_public"`
	DestinationCity    string     `json:"destination_city"`
	DestinationCountry string     `json:"destination_country"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type TripListResponse struct {
	Total int            `json:"total"`
	Data  []TripResponse `json:"data"`
}
