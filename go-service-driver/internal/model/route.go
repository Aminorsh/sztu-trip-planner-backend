package model

import "time"

// TripRoute 行程路线表
type TripRoute struct {
	ID     uint64 `json:"id" gorm:"primaryKey;autoIncrement;comment:路线ID"`
	TripID uint64 `json:"trip_id" gorm:"not null;index;comment:关联行程ID"`

	OriginPlaceID *uint64 `json:"origin_place_id" gorm:"comment:起点地点ID"`
	OriginLng     float64 `json:"origin_lng" gorm:"type:decimal(10,6);not null;comment:起点经度"`
	OriginLat     float64 `json:"origin_lat" gorm:"type:decimal(10,6);not null;comment:起点纬度"`

	DestinationPlaceID *uint64 `json:"destination_place_id" gorm:"comment:终点地点ID"`
	DestinationLng     float64 `json:"destination_lng" gorm:"type:decimal(10,6);not null;comment:终点经度"`
	DestinationLat     float64 `json:"destination_lat" gorm:"type:decimal(10,6);not null;comment:终点纬度"`

	Waypoints      JSONObject `json:"waypoints" gorm:"type:json;comment:途经点信息"`
	OptimizedOrder JSONArray  `json:"optimized_order" gorm:"type:json;comment:优化后的顺序"`

	TravelMode string `json:"travel_mode" gorm:"type:enum('driving','walking','bicycling','transit');not null;comment:出行方式"`
	Strategy   string `json:"strategy" gorm:"type:varchar(10);comment:路线策略"`

	TotalDistance float64 `json:"total_distance" gorm:"type:decimal(10,2);comment:总距离(公里)"`
	TotalDuration int     `json:"total_duration" gorm:"comment:总时长(分钟)"`

	PathData JSONObject `json:"path_data" gorm:"type:json;comment:路径详细数据"`
	Polyline string     `json:"polyline" gorm:"type:text;comment:路线坐标串"`

	AmapRouteID      string     `json:"amap_route_id" gorm:"type:varchar(100);comment:高德路线ID"`
	APIResponseCache JSONObject `json:"api_response_cache" gorm:"type:json;comment:API完整响应缓存"`

	IsActive bool `json:"is_active" gorm:"default:true;comment:是否当前激活路线"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (TripRoute) TableName() string {
	return "trip_routes"
}
