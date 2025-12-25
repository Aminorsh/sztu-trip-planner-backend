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
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Status        string     `json:"status"`
	CoverImageURL string     `json:"cover_image_url,omitempty"`
	Days          []TripDay  `json:"days"`
	CreatedAt     time.Time  `json:"createdAt"`
	LastSaved     time.Time  `json:"lastSaved"`
	Description   string     `json:"description,omitempty"`
	StartDate     *time.Time `json:"startDate,omitempty"`
	EndDate       *time.Time `json:"endDate,omitempty"`
}

type TripDay struct {
	Day   int        `json:"day"`
	Items []TripItem `json:"items"`
}

type TripItem struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Time     string     `json:"time"` // 格式: "15:04" 或 RFC3339
	EndTime  string     `json:"end_time,omitempty"`
	Note     string     `json:"note,omitempty"`
	Priority string     `json:"priority,omitempty"`
	Lnglat   [2]float64 `json:"lnglat,omitempty"`
}

// 更新行程请求
type UpdateTripRequest struct {
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	Description string     `json:"description"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	// Days   []TripDay `json:"days"`
}

// 添加行程项请求
type AddTripItemRequest struct {
	Items []TripItem `json:"items" binding:"required"`
}

// 更新行程项请求
type UpdateTripItemRequest struct {
	Name     string `json:"name"`
	Time     string `json:"time"`
	EndTime  string `json:"end_time"`
	Note     string `json:"note"`
	Priority string `json:"priority"`
}

// 添加天数请求
type AddTripDayRequest struct {
	Day int `json:"day"`
}

type TripListResponse struct {
	Total int            `json:"total"`
	Data  []TripResponse `json:"data"`
}

type UpdateTripCoverImageRequest struct {
	CoverImageURL string `json:"cover_image_url" binding:"required"`
}
