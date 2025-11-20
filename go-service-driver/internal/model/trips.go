package model

import "time"

type Trip struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement;comment:行程ID"`
	UserID      int       `json:"user_id" gorm:"not null;comment:创建用户ID"`
	Title       string    `json:"title" gorm:"not null;comment:行程标题"`
	Description string    `json:"description,omitempty" gorm:"type:text;comment:行程描述"`
	StartDate   time.Time `json:"start_date,omitempty" gorm:"type:date;comment:开始日期"`
	EndDate     time.Time `json:"end_date,omitempty" gorm:"type:date;comment:结束日期"`

	Days              int `json:"days,omitempty" gorm:"comment:行程天数"`
	TripItemsCount    int `json:"trip_items_count" gorm:"default:0;comment:行程项总数"`
	PhotosCount       int `json:"photos_count" gorm:"default:0;comment:图片总数"`
	UniquePlacesCount int `json:"unique_places_count" gorm:"default:0;comment:独特地点数"`

	IsPublic           bool    `json:"is_public" gorm:"default:false;comment:是否公开"`
	ShareToken         string  `json:"share_token,omitempty" gorm:"unique;comment:分享令牌"`
	Status             string  `json:"status" gorm:"type:enum('draft','active','completed','archived');default:'draft';comment:行程状态"`
	OriginAddress      string  `json:"origin_address,omitempty" gorm:"comment:起点地址"`
	DestinationAddress string  `json:"destination_address,omitempty" gorm:"comment:终点地址"`
	DestinationCity    string  `json:"destination_city,omitempty" gorm:"comment:目的地城市"`
	DestinationCountry string  `json:"destination_country,omitempty" gorm:"comment:目的地国家"`
	TotalDistance      float64 `json:"total_distance,omitempty" gorm:"type:decimal(10,2);comment:总距离(km)"`
	EstimatedDuration  int     `json:"estimated_duration,omitempty" gorm:"comment:预计时长(分钟)"`

	CreatedAt time.Time  `json:"created_at" gorm:"default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"comment:软删除时间"`
}

func (Trip) TableName() string {
	return "trips"
}
