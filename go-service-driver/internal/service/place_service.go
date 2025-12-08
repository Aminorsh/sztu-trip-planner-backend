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
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	// Layer 1: Check Redis cache (5 minutes)
	cacheKey := s.buildCacheKey(req.Keyword, req.City, req.Page, req.PageSize)
	cacheData, err := s.getFromRedis(cacheKey)
	if err == nil && cacheData != nil {
		return cacheData, nil
	}

	// Layer 2: Check database cache (24 hours)
	dbCacheData, err := s.getFromDBCache(ctx, cacheKey)
	if err == nil && dbCacheData != nil {
		go s.saveToRedis(cacheKey, dbCacheData, 5*time.Minute)
		return dbCacheData, nil
	}

	// Layer 3: Call Amap API
	amapResponse, err := s.callAmapAPI(req)
	if err != nil {
		return nil, err
	}

	places, total, err := s.parseAndSavePlaces(ctx, amapResponse)
	if err != nil {
		return nil, err
	}

	response := &dto.PlaceSearchResponse{
		Total:    total,
		Page:     req.Page,
		Places:   places,
		PageSize: req.PageSize,
	}

	// Save to DB cache
	go s.saveToDBCache(ctx, cacheKey, req, amapResponse, 24*time.Hour)

	// Save to Redis cache
	go s.saveToRedis(cacheKey, response, 5*time.Minute)

	return response, nil
}

func (s *PlaceService) callAmapAPI(req *dto.PlaceSearchRequest) (*dto.AmapSearchResponse, error) {
	if s.amapAPIKey == "" || s.amapAPIURL == "" {
		return nil, errors.NewAmapConfigError()
	}

	baseURL := fmt.Sprintf("%s/place/text", s.amapAPIURL)
	params := url.Values{}
	params.Add("key", s.amapAPIKey)
	params.Add("keywords", req.Keyword)
	params.Add("output", "json")
	params.Add("page", strconv.Itoa(req.Page))
	params.Add("offset", strconv.Itoa(req.PageSize))
	params.Add("extensions", "all")

	if req.City != "" {
		params.Add("city", req.City)
	}

	if req.Category != "" {
		if types := s.mapCategoryToAmapType(req.Category); types != "" {
			params.Add("types", types)
		}
	}

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
		placeResp := dto.PlaceResponse{
			ID:           place.ID,
			AmapPoiID:    place.AmapPoiID,
			Name:         place.Name,
			NameEn:       place.NameEn,
			Address:      place.Address,
			Location:     dto.Location{Lng: lng, Lat: lat},
			Province:     place.Province,
			City:         place.City,
			District:     place.District,
			Category:     string(place.Category),
			AmapType:     place.AmapType,
			Tel:          place.Phone,
			PriceRange:   place.PriceRange,
			OpeningHours: place.OpeningHours,
			CoverImage:   place.CoverImage,
		}

		if place.Rating != nil {
			placeResp.Rating = *place.Rating
		}
		if place.ReviewCount != nil {
			placeResp.ReviewCount = *place.ReviewCount
		}
		if place.PriceLevel != nil {
			placeResp.PriceLevel = string(*place.PriceLevel)
		}
		if place.Photos != nil {
			placeResp.Photos = place.Photos
		}
		if place.Tags != nil {
			placeResp.Tags = place.Tags
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

func (s *PlaceService) buildCacheKey(keyword, city string, page, pageSize int) string {
	raw := fmt.Sprintf("%s:%s:%d:%d", keyword, city, page, pageSize)
	hash := md5.Sum([]byte(raw))
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

	page := 1
	pageSize := 10
	if result.RequestParams["page"] != nil {
		if p, ok := result.RequestParams["page"].(float64); ok {
			page = int(p)
		}
		if ps, ok := result.RequestParams["page_size"].(float64); ok {
			pageSize = int(ps)
		}
	}

	return &dto.PlaceSearchResponse{
		Places:   places,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *PlaceService) saveToDBCache(ctx context.Context, cacheKey string, req *dto.PlaceSearchRequest, amapResp *dto.AmapSearchResponse, expiration time.Duration) error {
	requestParams := model.JSONObject{
		"keyword":   req.Keyword,
		"city":      req.City,
		"category":  req.Category,
		"page":      req.Page,
		"page_size": req.PageSize,
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
