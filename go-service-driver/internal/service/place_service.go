package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/repository"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type PlaceService struct {
	amapAPIKey    string
	amapAPIURL    string
	redisClient   *redis.Client
	placeRepo     repository.PlaceRepository
	amapCacheRepo repository.AmapPoiCacheRepository
}

func NewPlaceService(
	amapAPIKey, amapAPIURL string,
	redisClient *redis.Client,
	placeRepo repository.PlaceRepository,
	amapCacheRepo repository.AmapPoiCacheRepository,
) *PlaceService {
	return &PlaceService{
		amapAPIKey:    amapAPIKey,
		amapAPIURL:    amapAPIURL,
		redisClient:   redisClient,
		placeRepo:     placeRepo,
		amapCacheRepo: amapCacheRepo,
	}
}

func (s *PlaceService) SearchPlaces(ctx context.Context, req *dto.PlaceSearchRequest) (*dto.PlaceSearchResponse, error) {
	// 缓存key只基于keyword，不包含筛选和排序条件
	cacheKey := s.buildCacheKey(req.Keyword)

	var places []dto.PlaceResponse
	var total int

	// Layer 1: Check Redis cache (5 minutes)
	cacheData, err := s.getFromRedis(cacheKey)
	if err == nil && cacheData != nil {
		places = cacheData.Places
		total = cacheData.Total
	} else {
		// Layer 2: Check database cache (24 hours)
		dbCacheData, err := s.getFromDBCache(ctx, cacheKey)
		if err == nil && dbCacheData != nil {
			places = dbCacheData.Places
			total = dbCacheData.Total
			// 保存到Redis缓存
			go s.saveToRedis(cacheKey, dbCacheData, 5*time.Minute)
		} else {
			// Layer 3: Call Amap API
			amapResponse, err := s.callAmapAPI(req)
			if err != nil {
				return nil, err
			}

			places, total, err = s.parseAndSavePlaces(ctx, amapResponse)
			if err != nil {
				return nil, err
			}

			// 缓存原始未筛选的数据
			originalResponse := &dto.PlaceSearchResponse{
				Places: places,
				Total:  total,
			}

			// Save to DB cache
			go s.saveToDBCache(ctx, cacheKey, req, amapResponse, 24*time.Hour)

			// Save to Redis cache
			go s.saveToRedis(cacheKey, originalResponse, 5*time.Minute)
		}
	}

	// 应用筛选（在获取缓存数据后）
	if req.FilterType != "" {
		places = s.filterPlaces(places, req.FilterType)
		total = len(places)
	}

	// 应用排序（在筛选后）
	if req.SortOrder != "" {
		s.sortPlaces(places, req.SortOrder)
	}

	return &dto.PlaceSearchResponse{
		Places: places,
		Total:  total,
	}, nil
}

func (s *PlaceService) callAmapAPI(req *dto.PlaceSearchRequest) (*dto.AmapSearchResponse, error) {
	if s.amapAPIKey == "" || s.amapAPIURL == "" {
		return nil, errors.NewAmapConfigError()
	}

	baseURL := fmt.Sprintf("%s/v5/place/text", s.amapAPIURL)
	params := url.Values{}
	params.Add("key", s.amapAPIKey)
	params.Add("keywords", req.Keyword)
	params.Add("output", "json")
	params.Add("page", "1")
	params.Add("offset", "20")
	params.Add("extensions", "all")

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(requestURL)
	if err != nil {
		return nil, errors.NewAmapAPIError(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewReadResponseError(err)
	}

	var amapResp dto.AmapSearchResponse
	if err := json.Unmarshal(body, &amapResp); err != nil {
		return nil, errors.NewParseResponseError(err)
	}

	if amapResp.Status != "1" {
		return nil, errors.NewAmapAPIError(fmt.Errorf("AMAP API error: %s", amapResp.Info))
	}

	return &amapResp, nil
}

func (s *PlaceService) parseAndSavePlaces(ctx context.Context, amapResp *dto.AmapSearchResponse) ([]dto.PlaceResponse, int, error) {
	places := make([]dto.PlaceResponse, 0, len(amapResp.Pois))
	for _, poi := range amapResp.Pois {
		location := strings.Split(poi.Location, ",")
		if len(location) != 2 {
			continue
		}
		lng, err := strconv.ParseFloat(location[0], 64)
		if err != nil {
			continue
		}
		lat, err := strconv.ParseFloat(location[1], 64)
		if err != nil {
			continue
		}

		place, err := s.saveOrUpdatePlace(ctx, &poi, lng, lat)
		if err != nil {
			return nil, 0, err
		}
		// 构建符合前端需求的响应
		rating := 0.0
		if place.Rating != nil {
			rating = *place.Rating
		}

		placeResp := dto.PlaceResponse{
			ID:           fmt.Sprintf("%d", place.ID),
			Name:         place.Name,
			Address:      place.Address,
			Description:  place.Description,
			Image:        place.CoverImage,
			Rating:       rating,
			Distance:     0, // 需要根据用户位置计算
			OpeningHours: place.OpeningHours,
			Added:        false, // 需要根据用户行程判断
			Expanded:     false,
			Category:     string(place.Category),
			Location:     dto.Location{Lng: lng, Lat: lat},
			Tel:          place.Phone,
			PriceRange:   place.PriceRange,
		}

		if place.Photos != nil {
			placeResp.Photos = place.Photos
		}
		places = append(places, placeResp)
	}
	return places, len(places), nil
}

func (s *PlaceService) saveOrUpdatePlace(ctx context.Context, poi *dto.POI, lng, lat float64) (*model.Place, error) {
	var place model.Place

	// Look up existing place with caller's context
	existing, err := s.placeRepo.FindPlaceByPoiID(ctx, poi.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	category := s.mapAmapTypeToCategory(poi.Type)

	// Process photos first
	photos := make([]string, 0)
	for _, photo := range poi.Photos {
		photos = append(photos, photo.URL)
	}

	var coverImage string
	if len(photos) > 0 {
		coverImage = photos[0]
	}

	if existing == nil {
		// Create new place
		place = model.Place{
			Name:           poi.Name,
			Category:       category,
			AmapType:       poi.Type,
			AmapTypecode:   poi.Typecode,
			Province:       poi.Pname,
			City:           poi.Cityname,
			District:       poi.Adname,
			Address:        poi.Address,
			Adcode:         poi.Adcode,
			Latitude:       &lat,
			Longitude:      &lng,
			Phone:          poi.Tel,
			CoverImage:     coverImage,
			Photos:         photos,
			AmapPoiID:      poi.ID,
			DataSource:     model.DataSourceAmapAPI,
			APILastUpdated: &now,
			VisitCount:     1,
		}
		if err := s.placeRepo.Create(ctx, &place); err != nil {
			return nil, err
		}
	} else {
		// Update existing place
		place = *existing
		updates := map[string]any{
			"name":             poi.Name,
			"amap_type":        poi.Type,
			"amap_typecode":    poi.Typecode,
			"address":          poi.Address,
			"latitude":         lat,
			"longitude":        lng,
			"phone":            poi.Tel,
			"cover_image":      coverImage,
			"photos":           photos,
			"api_last_updated": now,
			"visit_count":      gorm.Expr("visit_count + ?", 1),
		}
		if err := s.placeRepo.UpdateFields(ctx, place.ID, updates); err != nil {
			return nil, err
		}
	}

	return &place, nil
}

func (s *PlaceService) mapAmapTypeToCategory(amapType string) model.PlaceCategory {
	if strings.Contains(amapType, "风景名胜") || strings.Contains(amapType, "旅游景点") {
		return model.PlaceCategoryScenic
	}
	if strings.Contains(amapType, "餐饮") {
		return model.PlaceCategoryRestaurant
	}
	if strings.Contains(amapType, "酒店") || strings.Contains(amapType, "宾馆") {
		return model.PlaceCategoryHotel
	}

	return model.PlaceCategoryOther
}

func (s *PlaceService) mapCategoryToAmapType(category string) string {
	mapping := map[string]string{
		"restaurant": "050000",
		"hotel":      "100000",
		"scenic":     "110000",
	}
	return mapping[category]
}

func (s *PlaceService) buildCacheKey(keyword string) string {
	hash := md5.Sum([]byte(keyword))
	return fmt.Sprintf("place_search:%x", hash)
}

func (s *PlaceService) getFromRedis(key string) (*dto.PlaceSearchResponse, error) {
	ctx := context.Background()
	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss is not an error
		}
		return nil, err
	}

	var resp dto.PlaceSearchResponse
	if err := json.Unmarshal([]byte(val), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *PlaceService) saveToRedis(key string, resp *dto.PlaceSearchResponse, expiration time.Duration) error {
	ctx := context.Background()
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	if err := s.redisClient.Set(ctx, key, data, expiration).Err(); err != nil {
		return err
	}
	return nil
}

func (s *PlaceService) getFromDBCache(ctx context.Context, cacheKey string) (*dto.PlaceSearchResponse, error) {
	result, err := s.amapCacheRepo.FindByCacheKey(ctx, cacheKey)
	if err != nil {
		// Cache miss is not an error, just return nil
		return nil, nil
	}

	if result == nil {
		return nil, nil
	}

	now := time.Now()
	s.amapCacheRepo.UpdateFields(ctx, result.ID, map[string]any{
		"hit_count":   gorm.Expr("hit_count + ?", 1),
		"last_hit_at": &now,
	})

	var amapResp dto.AmapSearchResponse
	jsonData, err := json.Marshal(result.ResponseData)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonData, &amapResp); err != nil {
		return nil, err
	}

	places, total, err := s.parseAndSavePlaces(ctx, &amapResp)
	if err != nil {
		return nil, err
	}

	return &dto.PlaceSearchResponse{
		Places: places,
		Total:  total,
	}, nil
}

func (s *PlaceService) saveToDBCache(ctx context.Context, cacheKey string, req *dto.PlaceSearchRequest, amapResp *dto.AmapSearchResponse, expiration time.Duration) error {
	requestParams := model.JSONObject{
		"keyword": req.Keyword,
	}

	responseData := model.JSONObject{}
	jsonData, err := json.Marshal(amapResp)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(jsonData, &responseData); err != nil {
		return err
	}

	poiCount, _ := strconv.Atoi(amapResp.Count)

	cache := model.AmapPoiCache{
		CacheKey:      cacheKey,
		RequestParams: requestParams,
		ResponseData:  responseData,
		PoiCount:      poiCount,
		ExpireAt:      time.Now().Add(expiration),
	}

	return s.amapCacheRepo.Create(ctx, &cache)
}

// filterPlaces 根据筛选类型过滤地点
func (s *PlaceService) filterPlaces(places []dto.PlaceResponse, filterType string) []dto.PlaceResponse {
	if filterType == "" {
		return places
	}

	filtered := make([]dto.PlaceResponse, 0)
	for _, place := range places {
		if place.Category == filterType {
			filtered = append(filtered, place)
		}
	}
	return filtered
}

// sortPlaces 根据排序方式排序地点
func (s *PlaceService) sortPlaces(places []dto.PlaceResponse, sortOrder string) {
	switch sortOrder {
	case "rating":
		// 按评分降序
		for i := 0; i < len(places)-1; i++ {
			for j := i + 1; j < len(places); j++ {
				if places[i].Rating < places[j].Rating {
					places[i], places[j] = places[j], places[i]
				}
			}
		}
	case "distance":
		// 按距离升序
		for i := 0; i < len(places)-1; i++ {
			for j := i + 1; j < len(places); j++ {
				if places[i].Distance > places[j].Distance {
					places[i], places[j] = places[j], places[i]
				}
			}
		}
	case "name":
		// 按名称字母顺序
		for i := 0; i < len(places)-1; i++ {
			for j := i + 1; j < len(places); j++ {
				if places[i].Name > places[j].Name {
					places[i], places[j] = places[j], places[i]
				}
			}
		}
	}
}
