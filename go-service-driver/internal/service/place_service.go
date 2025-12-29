package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/dto"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/model"
	"github.com/Aminorsh/sztu-trip-planner-backend/internal/repository"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type PlaceService struct {
	amapAPIKey       string
	amapAPIURL       string
	redisClient      *redis.Client
	placeRepo        repository.PlaceRepository
	amapCacheRepo    repository.AmapPoiCacheRepository
	assistantService *AssistantService
}

func NewPlaceService(
	amapAPIKey, amapAPIURL string,
	redisClient *redis.Client,
	placeRepo repository.PlaceRepository,
	amapCacheRepo repository.AmapPoiCacheRepository,
	assistantService *AssistantService,
) *PlaceService {
	return &PlaceService{
		amapAPIKey:       amapAPIKey,
		amapAPIURL:       amapAPIURL,
		redisClient:      redisClient,
		placeRepo:        placeRepo,
		amapCacheRepo:    amapCacheRepo,
		assistantService: assistantService,
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

	// 过滤掉用户自定义地点（只在搜索结果中显示高德API的地点）
	places = s.filterOutCustomPlaces(places)
	total = len(places)

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

	baseURL := fmt.Sprintf("%s/v3/place/text", s.amapAPIURL)
	params := url.Values{}
	params.Add("key", s.amapAPIKey)
	params.Add("keywords", req.Keyword)
	// params.Add("output", "json")
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
		fmt.Printf("[AMAP API] JSON unmarshal error: %v\n", err)
		fmt.Printf("[AMAP API] Failed to parse body: %s\n", string(body))
		return nil, errors.NewParseResponseError(err)
	}

	fmt.Printf("[AMAP API] Parsed response - Status: %s, Info: %s, Count: %s, POIs: %d\n",
		amapResp.Status, amapResp.Info, amapResp.Count, len(amapResp.Pois))

	if amapResp.Status != "1" {
		fmt.Printf("[AMAP API] API returned error status: %s\n", amapResp.Info)
		return nil, errors.NewAmapAPIError(fmt.Errorf("AMAP API error: %s", amapResp.Info))
	}

	// 调试：打印第一个POI的详细信息
	if config.IsTestMode() && len(amapResp.Pois) > 0 {
		firstPOI, _ := json.MarshalIndent(amapResp.Pois[0], "", "  ")
		fmt.Printf("First POI from Amap API:\n%s\n", string(firstPOI))
	}

	return &amapResp, nil
}

// callAmapDetailAPI 调用高德POI详情API获取详细信息
func (s *PlaceService) callAmapDetailAPI(poiID string) (*dto.POI, error) {
	if s.amapAPIKey == "" || s.amapAPIURL == "" {
		return nil, errors.NewAmapConfigError()
	}

	baseURL := fmt.Sprintf("%s/v3/place/detail", s.amapAPIURL)
	params := url.Values{}
	params.Add("key", s.amapAPIKey)
	params.Add("id", poiID)
	params.Add("output", "json")

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

	var detailResp dto.AmapDetailResponse
	if err := json.Unmarshal(body, &detailResp); err != nil {
		return nil, errors.NewParseResponseError(err)
	}

	if detailResp.Status != "1" {
		return nil, errors.NewAmapAPIError(fmt.Errorf("AMAP Detail API error: %s", detailResp.Info))
	}

	if len(detailResp.Pois) == 0 {
		return nil, errors.NewPlaceNotFoundError()
	}

	return &detailResp.Pois[0], nil
}

// callWikipediaAPI 调用Wikipedia API获取景点信息
func (s *PlaceService) callWikipediaAPI(placeName string) (*dto.WikipediaResponse, error) {
	// URL编码地点名称
	encodedName := url.QueryEscape(placeName)
	apiURL := fmt.Sprintf("https://zh.wikipedia.org/api/rest_v1/page/summary/%s", encodedName)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 404表示Wikipedia没有该词条
	if resp.StatusCode == 404 {
		return nil, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wikipedia API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wikiResp dto.WikipediaResponse
	if err := json.Unmarshal(body, &wikiResp); err != nil {
		return nil, err
	}

	return &wikiResp, nil
}

func (s *PlaceService) parseAndSavePlaces(ctx context.Context, amapResp *dto.AmapSearchResponse) ([]dto.PlaceResponse, int, error) {
	fmt.Printf("[parseAndSavePlaces] Starting to parse %d POIs\n", len(amapResp.Pois))
	places := make([]dto.PlaceResponse, 0, len(amapResp.Pois))
	for i, poi := range amapResp.Pois {
		fmt.Printf("[parseAndSavePlaces] Processing POI %d: ID=%s, Name=%s, Location=%s\n",
			i, poi.ID, poi.Name, poi.Location)

		location := strings.Split(poi.Location, ",")
		if len(location) != 2 {
			fmt.Printf("[parseAndSavePlaces] POI %d: Invalid location format, skipping\n", i)
			continue
		}
		lng, err := strconv.ParseFloat(location[0], 64)
		if err != nil {
			fmt.Printf("[parseAndSavePlaces] POI %d: Failed to parse longitude: %v\n", i, err)
			continue
		}
		lat, err := strconv.ParseFloat(location[1], 64)
		if err != nil {
			fmt.Printf("[parseAndSavePlaces] POI %d: Failed to parse latitude: %v\n", i, err)
			continue
		}

		place, err := s.saveOrUpdatePlace(ctx, &poi, lng, lat)
		if err != nil {
			fmt.Printf("[parseAndSavePlaces] POI %d: Failed to save/update place: %v\n", i, err)
			return nil, 0, err
		}
		fmt.Printf("[parseAndSavePlaces] POI %d: Successfully saved place ID=%d\n", i, place.ID)

		// 提取评分
		rating := 0.0
		if place.Rating != nil {
			rating = *place.Rating
		} else if string(poi.Biz_ext.Rating) != "" {
			// 尝试从高德API响应中解析评分
			if r, err := strconv.ParseFloat(string(poi.Biz_ext.Rating), 64); err == nil {
				rating = r
			}
		}

		// 提取营业时间（v3 API不返回，只使用数据库中的数据）
		openingHours := place.OpeningHours

		// 提取简介（v3 API不返回，只使用数据库中的数据）
		description := place.Description

		placeResp := dto.PlaceResponse{
			ID:           fmt.Sprintf("%d", place.ID),
			Name:         place.Name,
			Address:      place.Address,
			Description:  description,
			Image:        place.CoverImage,
			Rating:       rating,
			Distance:     0, // 需要根据用户位置计算
			OpeningHours: openingHours,
			// Added:        false, // 需要根据用户行程判断
			Expanded: false,
			// Category:   string(place.Category),
			Location: dto.Location{Lng: lng, Lat: lat},
			// Tel:        place.Phone,
			// PriceRange: place.PriceRange,
		}

		if place.Photos != nil {
			placeResp.Photos = place.Photos
		}
		places = append(places, placeResp)
	}
	fmt.Printf("[parseAndSavePlaces] Completed: parsed %d places\n", len(places))
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

	// 解析评分
	var rating *float64
	if string(poi.Biz_ext.Rating) != "" {
		if r, err := strconv.ParseFloat(string(poi.Biz_ext.Rating), 64); err == nil {
			rating = &r
		}
	}

	// v3 API不返回营业时间和简介，这些字段保持空
	var openingHours string
	var description string

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
			Phone:          string(poi.Tel),
			CoverImage:     coverImage,
			Photos:         photos,
			AmapPoiID:      poi.ID,
			DataSource:     model.DataSourceAmapAPI,
			APILastUpdated: &now,
			VisitCount:     1,
			Rating:         rating,
			OpeningHours:   openingHours,
			Description:    description,
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
			"phone":            string(poi.Tel),
			"cover_image":      coverImage,
			"photos":           photos,
			"api_last_updated": now,
			"visit_count":      gorm.Expr("visit_count + ?", 1),
		}
		// 只在有新数据时更新
		if rating != nil {
			updates["rating"] = *rating
		}
		// v3 API不返回营业时间和简介，不更新这些字段
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

// func (s *PlaceService) mapCategoryToAmapType(category string) string {
// 	mapping := map[string]string{
// 		"restaurant": "050000",
// 		"hotel":      "100000",
// 		"scenic":     "110000",
// 	}
// 	return mapping[category]
// }

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
		if place.Type == filterType {
			filtered = append(filtered, place)
		}
	}
	return filtered
}

// filterOutCustomPlaces 过滤掉用户自定义的地点（只保留高德API数据）
func (s *PlaceService) filterOutCustomPlaces(places []dto.PlaceResponse) []dto.PlaceResponse {
	filtered := make([]dto.PlaceResponse, 0, len(places))
	for _, place := range places {
		// 过滤掉ID以 "custom_" 开头的自定义地点
		// 自定义地点只在用户的行程中显示，不在搜索结果中出现
		placeID, err := strconv.Atoi(place.ID)
		if err != nil {
			continue
		}

		// 从数据库验证地点来源
		ctx := context.Background()
		dbPlace, err := s.placeRepo.FindByID(ctx, uint(placeID))
		if err != nil || dbPlace == nil {
			continue
		}

		// 只保留非用户创建的地点
		if dbPlace.DataSource != model.DataSourceUserCreated {
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
		sort.SliceStable(places, func(i, j int) bool {
			return places[i].Rating > places[j].Rating
		})
	case "distance":
		// 按距离升序
		sort.SliceStable(places, func(i, j int) bool {
			return places[i].Distance < places[j].Distance
		})
	case "name":
		// 按名称字母顺序
		sort.SliceStable(places, func(i, j int) bool {
			return places[i].Name < places[j].Name
		})
	default:
		// 默认不排序
	}
}

func (s *PlaceService) GetPlaceDetail(ctx context.Context, placeID string) (*dto.PlaceResponse, error) {
	id, err := strconv.Atoi(placeID)
	if err != nil {
		return nil, errors.NewInvalidRequestError("invalid place ID")
	}

	place, err := s.placeRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	if place == nil {
		return nil, errors.NewPlaceNotFoundError()
	}

	// 检查是否需要从高德API获取详细信息
	needDetailAPI := (place.Rating == nil || *place.Rating == 0) &&
		place.Description == "" &&
		place.OpeningHours == "" &&
		len(place.Photos) == 0

	// 如果有AmapPoiID且信息不完整，尝试调用详情API
	if needDetailAPI && place.AmapPoiID != "" {
		if poi, err := s.callAmapDetailAPI(place.AmapPoiID); err == nil {
			// 更新地点信息
			updates := make(map[string]any)

			// 更新评分
			if string(poi.Biz_ext.Rating) != "" {
				if r, err := strconv.ParseFloat(string(poi.Biz_ext.Rating), 64); err == nil {
					updates["rating"] = r
					place.Rating = &r
				}
			}

			// v3 API不返回简介和营业时间字段

			// 更新照片
			if len(poi.Photos) > 0 {
				photos := make([]string, 0, len(poi.Photos))
				for _, photo := range poi.Photos {
					photos = append(photos, photo.URL)
				}
				updates["photos"] = photos
				updates["cover_image"] = photos[0]
				place.Photos = photos
				place.CoverImage = photos[0]
			}

			// 更新电话
			if string(poi.Tel) != "" {
				updates["phone"] = string(poi.Tel)
				place.Phone = string(poi.Tel)
			}

			// 保存更新到数据库
			if len(updates) > 0 {
				now := time.Now()
				updates["api_last_updated"] = now
				s.placeRepo.UpdateFields(ctx, place.ID, updates)
			}
		}
		// 即使API调用失败也继续返回现有数据
	}

	// Deactive Wikipedia data fetching for now
	wikipediaDataEnabled := false

	if wikipediaDataEnabled {
		// 如果仍然缺少信息，尝试从Wikipedia获取
		if (place.Description == "" || len(place.Photos) == 0) && place.Name != "" {
			// 尝试多个名称变体
			nameVariants := s.getNameVariants(place.Name)
			fmt.Printf("[DEBUG] Trying Wikipedia for place: %s, variants: %v\n", place.Name, nameVariants)

			for _, name := range nameVariants {
				fmt.Printf("[DEBUG] Calling Wikipedia API with name: %s\n", name)
				if wikiData, err := s.callWikipediaAPI(name); err == nil && wikiData != nil {
					fmt.Printf("[DEBUG] Wikipedia response: title=%s, extract_len=%d, has_image=%v\n",
						wikiData.Title, len(wikiData.Extract), wikiData.Originalimage.Source != "")

					updates := make(map[string]any)

					// 更新简介
					if place.Description == "" && wikiData.Extract != "" {
						updates["description"] = wikiData.Extract
						place.Description = wikiData.Extract
					}

					// 更新封面图片
					if place.CoverImage == "" {
						var imageURL string
						// 优先使用原图
						if wikiData.Originalimage.Source != "" {
							imageURL = wikiData.Originalimage.Source
						} else if wikiData.Thumbnail.Source != "" {
							imageURL = wikiData.Thumbnail.Source
						}

						if imageURL != "" {
							updates["cover_image"] = imageURL
							place.CoverImage = imageURL
							// 如果photos为空，也添加到photos
							if len(place.Photos) == 0 {
								updates["photos"] = []string{imageURL}
								place.Photos = []string{imageURL}
							}
						}
					}

					// 保存Wikipedia数据到数据库
					if len(updates) > 0 {
						now := time.Now()
						updates["api_last_updated"] = now
						fmt.Printf("[DEBUG] Saving Wikipedia data: %+v\n", updates)
						s.placeRepo.UpdateFields(ctx, place.ID, updates)
						break // 成功获取数据后退出循环
					}
				} else if err != nil {
					fmt.Printf("[DEBUG] Wikipedia API error for %s: %v\n", name, err)
				}
			}
		}
	}

	// 如果仍然缺少信息，应用默认值
	s.applyDefaultPlaceData(place)

	var rating float64
	if place.Rating != nil {
		rating = *place.Rating
	}

	var lng, lat float64
	if place.Longitude != nil {
		lng = *place.Longitude
	}
	if place.Latitude != nil {
		lat = *place.Latitude
	}

	resp := &dto.PlaceResponse{
		ID:           fmt.Sprintf("%d", place.ID),
		Name:         place.Name,
		Address:      place.Address,
		Description:  place.Description,
		Image:        place.CoverImage,
		Rating:       rating,
		Distance:     0, // 需要根据用户位置计算
		OpeningHours: place.OpeningHours,
		Expanded:     true,
		Type:         string(place.Category),
		Location:     dto.Location{Lng: lng, Lat: lat},
		Photos:       place.Photos,
	}

	return resp, nil
}

// applyDefaultPlaceData 为知名景点应用默认数据
func (s *PlaceService) applyDefaultPlaceData(place *model.Place) {
	// 定义一些知名景点的默认数据
	defaultData := map[string]struct {
		rating       float64
		description  string
		openingHours string
	}{
		// "广州塔": {
		// 	rating:       4.5,
		// 	description:  "广州地标性建筑，中国第一高塔，可俯瞰珠江两岸美景",
		// 	openingHours: "09:30-22:30",
		// },
		"东方明珠": {
			rating:       4.6,
			description:  "上海标志性建筑，集观光、餐饮、购物于一体",
			openingHours: "08:00-21:30",
		},
		"外滩": {
			rating:       4.7,
			description:  "上海最著名的观光区，万国建筑博览群",
			openingHours: "全天开放",
		},
		"天安门广场": {
			rating:       4.8,
			description:  "世界最大的城市广场，中国的象征",
			openingHours: "05:00-22:00",
		},
		"故宫": {
			rating:       4.8,
			description:  "中国明清两代的皇家宫殿，世界文化遗产",
			openingHours: "08:30-17:00",
		},
	}

	if data, exists := defaultData[place.Name]; exists {
		if place.Rating == nil || *place.Rating == 0 {
			place.Rating = &data.rating
		}
		if place.Description == "" {
			place.Description = data.description
		}
		if place.OpeningHours == "" {
			place.OpeningHours = data.openingHours
		}
	}
}

// getNameVariants 生成地点名称的多个变体用于搜索
func (s *PlaceService) getNameVariants(name string) []string {
	variants := []string{name} // 原始名称

	// 常见后缀映射
	suffixMap := map[string][]string{
		"广播电视塔": {"塔", ""},
		"电视塔":   {"塔", ""},
		"博物馆":   {""},
		"纪念馆":   {""},
		"公园":    {""},
		"广场":    {""},
		"大厦":    {"塔", ""},
	}

	// 尝试移除后缀
	for suffix, replacements := range suffixMap {
		if strings.HasSuffix(name, suffix) {
			base := strings.TrimSuffix(name, suffix)
			for _, repl := range replacements {
				variants = append(variants, base+repl)
			}
		}
	}

	// 去重
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, v := range variants {
		if !seen[v] && v != "" {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}

// GenerateAIDescription 使用 AI 为地点生成描述并保存到数据库
func (s *PlaceService) GenerateAIDescription(ctx context.Context, placeID string) (string, error) {
	id, err := strconv.Atoi(placeID)
	if err != nil {
		return "", errors.NewInvalidRequestError("invalid place ID")
	}

	place, err := s.placeRepo.FindByID(ctx, uint(id))
	if err != nil {
		return "", err
	}
	if place == nil {
		return "", errors.NewPlaceNotFoundError()
	}

	// 如果已经有描述，直接返回（可选：或者重新生成）
	if place.Description != "" {
		return place.Description, nil
	}

	// 构建提示词
	prompt := fmt.Sprintf(
		"请为以下景点生成一段简洁的介绍（100-150字）：\n名称：%s\n地址：%s\n类型：%s\n\n要求：\n1. 突出景点特色\n2. 包含历史文化或建筑特点\n3. 提及适合的游览人群\n4. 语言精炼，吸引人",
		place.Name,
		place.Address,
		place.Category,
	)

	// 调用 AI 服务生成描述
	messages := []dto.AssistantMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	description, err := s.assistantService.Chat(ctx, 0, messages)
	if err != nil {
		return "", err
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return "", errors.NewInternalServerError(fmt.Errorf("AI generated empty description"))
	}

	// 保存到数据库
	updates := map[string]any{
		"description": description,
	}
	if err := s.placeRepo.UpdateFields(ctx, place.ID, updates); err != nil {
		return "", err
	}

	return description, nil
}
