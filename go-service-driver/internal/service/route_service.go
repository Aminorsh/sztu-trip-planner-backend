package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	apperrors "github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/repository"
)

type RouteService struct {
	routeRepo  repository.RouteRepository
	tripRepo   repository.TripRepository
	amapAPIKey string
	amapAPIURL string
}

func NewRouteService(
	routeRepo repository.RouteRepository,
	tripRepo repository.TripRepository,
	amapAPIKey string,
	amapAPIURL string,
) *RouteService {
	return &RouteService{
		routeRepo:  routeRepo,
		tripRepo:   tripRepo,
		amapAPIKey: amapAPIKey,
		amapAPIURL: amapAPIURL,
	}
}

// PlanRoute 规划路线
func (s *RouteService) PlanRoute(ctx context.Context, req *dto.RoutePlanRequest) (*dto.RoutePlanResponse, error) {
	// 1. 验证行程是否存在
	trip, err := s.tripRepo.FindByID(ctx, int(req.TripID))
	if err != nil {
		return nil, err
	}
	if trip == nil {
		return nil, apperrors.NewTripNotFoundError()
	}

	// 2. 调用高德地图 API
	amapResp, err := s.callAmapRouteAPI(req)
	if err != nil {
		return nil, err
	}

	// 3. 解析响应数据
	routeData, err := s.parseAmapResponse(amapResp)
	if err != nil {
		return nil, err
	}

	// 4. 保存到数据库
	tripRoute, err := s.saveRouteToDatabase(ctx, req, routeData, amapResp)
	if err != nil {
		return nil, err
	}

	// 5. 更新 Trip 表的距离和时长
	err = s.updateTripStats(ctx, req.TripID, routeData.TotalDistance, routeData.TotalDuration)
	if err != nil {
		return nil, err
	}

	// 6. 构造响应
	response := &dto.RoutePlanResponse{
		RouteID:        tripRoute.ID,
		TripID:         tripRoute.TripID,
		TotalDistance:  routeData.TotalDistance,
		TotalDuration:  routeData.TotalDuration,
		PathCount:      len(routeData.Paths),
		Paths:          routeData.Paths,
		OptimizedOrder: s.extractOptimizedOrder(req),
		CreatedAt:      tripRoute.CreatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// callAmapRouteAPI 调用高德地图路径规划 API
func (s *RouteService) callAmapRouteAPI(req *dto.RoutePlanRequest) (*dto.AmapRouteResponse, error) {
	if s.amapAPIKey == "" || s.amapAPIURL == "" {
		return nil, apperrors.NewAmapConfigError()
	}

	var apiURL string

	// 根据出行方式选择不同的 API 端点
	switch req.TravelMode {
	case "driving":
		apiURL = s.amapAPIURL + "/v3/direction/driving"
	case "walking":
		apiURL = s.amapAPIURL + "/v3/direction/walking"
	case "bicycling":
		apiURL = s.amapAPIURL + "/v4/direction/bicycling"
	case "transit":
		apiURL = s.amapAPIURL + "/v3/direction/transit/integrated"
	default:
		return nil, apperrors.NewInvalidRequestError("不支持的出行方式")
	}

	// 构造起点终点坐标
	origin := fmt.Sprintf("%.6f,%.6f", req.Origin.Lng, req.Origin.Lat)
	destination := fmt.Sprintf("%.6f,%.6f", req.Destination.Lng, req.Destination.Lat)

	// 构造途经点（高德最多支持16个途经点）
	waypoints := ""
	if len(req.Waypoints) > 0 {
		waypointStrs := make([]string, 0, len(req.Waypoints))
		for _, wp := range req.Waypoints {
			waypointStrs = append(waypointStrs, fmt.Sprintf("%.6f,%.6f", wp.Lng, wp.Lat))
		}
		waypoints = strings.Join(waypointStrs, ";")
	}

	// 构造请求参数
	httpReq, err := http.NewRequestWithContext(context.Background(), "GET", apiURL, nil)
	if err != nil {
		return nil, apperrors.NewInternalServerError(err)
	}

	q := httpReq.URL.Query()
	q.Add("key", s.amapAPIKey)
	q.Add("origin", origin)
	q.Add("destination", destination)

	if waypoints != "" {
		q.Add("waypoints", waypoints)
	}

	if req.Strategy != "" {
		q.Add("strategy", req.Strategy)
	} else {
		q.Add("strategy", "0") // 默认速度优先
	}

	if req.Extensions != "" {
		q.Add("extensions", req.Extensions)
	} else {
		q.Add("extensions", "all") // 返回详细信息
	}

	httpReq.URL.RawQuery = q.Encode()

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, apperrors.NewAmapAPIError(err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.NewReadResponseError(err)
	}

	// 解析 JSON
	var amapResp dto.AmapRouteResponse
	if err := json.Unmarshal(body, &amapResp); err != nil {
		return nil, apperrors.NewParseResponseError(err)
	}

	// 检查高德 API 状态
	if amapResp.Status != "1" {
		return nil, apperrors.NewAmapAPIError(fmt.Errorf("高德API错误: %s", amapResp.Info))
	}

	return &amapResp, nil
}

// parseAmapResponse 解析高德地图响应
func (s *RouteService) parseAmapResponse(amapResp *dto.AmapRouteResponse) (*dto.RoutePlanResponse, error) {
	if len(amapResp.Route.Paths) == 0 {
		return nil, apperrors.NewInvalidRequestError("未找到可用路线")
	}

	paths := make([]dto.PathInfo, 0, len(amapResp.Route.Paths))
	var totalDistance float64
	var totalDuration int

	for _, path := range amapResp.Route.Paths {
		// 解析距离（米转公里）
		distance, _ := strconv.ParseFloat(path.Distance, 64)
		distanceKM := distance / 1000

		// 解析时长（秒转分钟）
		duration, _ := strconv.Atoi(path.Duration)
		durationMin := duration / 60

		// 解析步骤
		steps := make([]dto.StepInfo, 0, len(path.Steps))
		for _, step := range path.Steps {
			stepDist, _ := strconv.ParseFloat(step.Distance, 64)
			stepDur, _ := strconv.Atoi(step.Duration)

			steps = append(steps, dto.StepInfo{
				Instruction: step.Instruction,
				Road:        step.Road,
				Distance:    stepDist,
				Duration:    stepDur,
				Action:      step.Action,
			})
		}

		pathInfo := dto.PathInfo{
			Distance: distanceKM,
			Duration: durationMin,
			Strategy: s.getStrategyName(path.Strategy),
			Steps:    steps,
			Polyline: path.Polyline,
		}

		paths = append(paths, pathInfo)

		// 累加总距离和总时长（取第一条路线）
		if len(paths) == 1 {
			totalDistance = distanceKM
			totalDuration = durationMin
		}
	}

	return &dto.RoutePlanResponse{
		TotalDistance: totalDistance,
		TotalDuration: totalDuration,
		PathCount:     len(paths),
		Paths:         paths,
	}, nil
}

// saveRouteToDatabase 保存路线到数据库
func (s *RouteService) saveRouteToDatabase(
	ctx context.Context,
	req *dto.RoutePlanRequest,
	routeData *dto.RoutePlanResponse,
	amapResp *dto.AmapRouteResponse,
) (*model.TripRoute, error) {
	// 将途经点转换为 JSON
	waypointsJSON := make(map[string]any)
	for i, wp := range req.Waypoints {
		waypointsJSON[fmt.Sprintf("waypoint_%d", i)] = map[string]any{
			"place_id": wp.PlaceID,
			"lng":      wp.Lng,
			"lat":      wp.Lat,
		}
	}

	// 缓存完整的 API 响应
	apiCache, _ := json.Marshal(amapResp)
	var apiCacheJSON map[string]any
	json.Unmarshal(apiCache, &apiCacheJSON)

	// 缓存路径数据
	pathDataBytes, _ := json.Marshal(routeData.Paths)
	var pathDataJSON map[string]any
	json.Unmarshal(pathDataBytes, &pathDataJSON)

	// 创建路线记录
	tripRoute := &model.TripRoute{
		TripID:             req.TripID,
		OriginPlaceID:      &req.Origin.PlaceID,
		OriginLng:          req.Origin.Lng,
		OriginLat:          req.Origin.Lat,
		DestinationPlaceID: &req.Destination.PlaceID,
		DestinationLng:     req.Destination.Lng,
		DestinationLat:     req.Destination.Lat,
		Waypoints:          waypointsJSON,
		TravelMode:         req.TravelMode,
		Strategy:           req.Strategy,
		TotalDistance:      routeData.TotalDistance,
		TotalDuration:      routeData.TotalDuration,
		PathData:           pathDataJSON,
		Polyline:           routeData.Paths[0].Polyline,
		APIResponseCache:   apiCacheJSON,
		IsActive:           true,
	}

	// 保存到数据库
	if err := s.routeRepo.Create(ctx, tripRoute); err != nil {
		return nil, err
	}

	// 停用该行程的其他路线
	if err := s.routeRepo.DeactivateOtherRoutes(ctx, req.TripID, tripRoute.ID); err != nil {
		return nil, err
	}

	return tripRoute, nil
}

// updateTripStats 更新行程统计信息
func (s *RouteService) updateTripStats(ctx context.Context, tripID uint64, distance float64, duration int) error {
	trip, err := s.tripRepo.FindByID(ctx, int(tripID))
	if err != nil {
		return err
	}
	if trip == nil {
		return apperrors.NewTripNotFoundError()
	}

	trip.TotalDistance = distance
	trip.EstimatedDuration = duration

	return s.tripRepo.Update(ctx, trip)
}

// extractOptimizedOrder 提取优化后的地点顺序
func (s *RouteService) extractOptimizedOrder(req *dto.RoutePlanRequest) []uint64 {
	order := []uint64{req.Origin.PlaceID}
	for _, wp := range req.Waypoints {
		order = append(order, wp.PlaceID)
	}
	order = append(order, req.Destination.PlaceID)
	return order
}

// getStrategyName 获取策略名称
func (s *RouteService) getStrategyName(strategy string) string {
	strategies := map[string]string{
		"0":  "速度优先（时间最短）",
		"1":  "费用优先（不走收费路段的最快路线）",
		"2":  "距离优先",
		"3":  "不走高速",
		"4":  "躲避拥堵",
		"5":  "多策略（同时使用速度优先、费用优先、距离优先）",
		"6":  "高速优先",
		"7":  "不走高速且避免收费",
		"8":  "躲避收费和拥堵",
		"9":  "不走高速且躲避收费和拥堵",
		"10": "不走高速且躲避拥堵",
	}

	if name, ok := strategies[strategy]; ok {
		return name
	}
	return "未知策略"
}

// GetRouteByID 获取路线详情
func (s *RouteService) GetRouteByID(ctx context.Context, routeID uint64) (*model.TripRoute, error) {
	return s.routeRepo.FindByID(ctx, routeID)
}

// GetRoutesByTripID 获取行程的所有路线
func (s *RouteService) GetRoutesByTripID(ctx context.Context, tripID uint64) ([]model.TripRoute, error) {
	return s.routeRepo.FindByTripID(ctx, tripID)
}
