package dto

// RoutePlanRequest 路线规划请求
type RoutePlanRequest struct {
	TripID      uint64       `json:"trip_id" binding:"required"`
	Origin      RoutePoint   `json:"origin" binding:"required"`
	Destination RoutePoint   `json:"destination" binding:"required"`
	Waypoints   []RoutePoint `json:"waypoints"`
	Strategy    string       `json:"strategy"` // 0-速度优先 1-费用优先 2-距离优先 3-不走高速等
	TravelMode  string       `json:"travel_mode" binding:"required,oneof=driving walking bicycling transit"`
	Extensions  string       `json:"extensions"` // base或all，默认base
}

type RoutePoint struct {
	PlaceID uint64  `json:"place_id"`
	Lng     float64 `json:"lng" binding:"required"`
	Lat     float64 `json:"lat" binding:"required"`
}

// RoutePlanResponse 路线规划响应
type RoutePlanResponse struct {
	RouteID        uint64     `json:"route_id"`
	TripID         uint64     `json:"trip_id"`
	TotalDistance  float64    `json:"total_distance"` // 公里
	TotalDuration  int        `json:"total_duration"` // 分钟
	PathCount      int        `json:"path_count"`
	Paths          []PathInfo `json:"paths"`
	OptimizedOrder []uint64   `json:"optimized_order"` // 优化后的地点顺序
	CreatedAt      string     `json:"created_at"`
}

type PathInfo struct {
	Distance float64    `json:"distance"` // 公里
	Duration int        `json:"duration"` // 分钟
	Strategy string     `json:"strategy"`
	Steps    []StepInfo `json:"steps"`
	Polyline string     `json:"polyline"`
}

type StepInfo struct {
	Instruction string  `json:"instruction"`
	Road        string  `json:"road"`
	Distance    float64 `json:"distance"` // 米
	Duration    int     `json:"duration"` // 秒
	Action      string  `json:"action"`
}

// 高德地图 API 响应结构
type AmapRouteResponse struct {
	Status string `json:"status"`
	Info   string `json:"info"`
	Count  string `json:"count"`
	Route  Route  `json:"route"`
}

type Route struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Paths       []Path `json:"paths"`
}

type Path struct {
	Distance string `json:"distance"` // 米
	Duration string `json:"duration"` // 秒
	Strategy string `json:"strategy"`
	Steps    []Step `json:"steps"`
	Polyline string `json:"polyline"`
}

type Step struct {
	Instruction      string      `json:"instruction"`
	Road             string      `json:"road"`
	Distance         string      `json:"distance"`
	Duration         string      `json:"duration"`
	Polyline         string      `json:"polyline"`
	Action           interface{} `json:"action"`           // Can be string or array
	Assistant_action interface{} `json:"assistant_action"` // Can be string or array
}
