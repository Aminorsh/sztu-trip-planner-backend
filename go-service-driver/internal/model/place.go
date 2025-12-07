package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
)

type PlaceCategory string

const (
	PlaceCategoryRestaurant PlaceCategory = "restaurant"
	PlaceCategoryHotel      PlaceCategory = "hotel"
	PlaceCategoryScenic     PlaceCategory = "scenic"
	PlaceCategoryOther      PlaceCategory = "other"
	// PlaceCategoryShopping      PlaceCategory = "shopping"
	// PlaceCategoryEntertainment PlaceCategory = "entertainment"
)

type DataSource string

const (
	DataSourceUserCreated DataSource = "user_created"
	DataSourceAmapAPI     DataSource = "amap_api"
	DataSourceImported    DataSource = "imported"
)

type PriceLevel string

const (
	PriceLevelFree      PriceLevel = "free"
	PriceLevelBudget    PriceLevel = "budget"
	PriceLevelModerate  PriceLevel = "moderate"
	PriceLevelExpensive PriceLevel = "expensive"
	PriceLevelLuxury    PriceLevel = "luxury"
)

type JSONArray []string

func (j *JSONArray) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.NewJSONArrayScanError()
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

type JSONObject map[string]any

func (j *JSONObject) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.NewJSONArrayScanError()
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONObject) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

type Place struct {
	ID       uint64        `json:"id" gorm:"primaryKey;autoIncrement;comment:地点ID"`
	Name     string        `json:"name" gorm:"type:varchar(200);not null;index:ft_name,class:FULLTEXT;comment:地点名称"`
	NameEn   string        `json:"name_en" gorm:"type:varchar(200);comment:地点英文名称"`
	Category PlaceCategory `json:"category" gorm:"type:enum('restaurant','hotel','scenic','other');not null;comment:地点类别"`

	AmapType     string    `json:"amap_type" gorm:"type:varchar(100);comment:高德地点类型"`
	AmapTypecode string    `json:"amap_typecode" gorm:"type:varchar(20);comment:高德地点类型编码"`
	Tags         JSONArray `json:"tags" gorm:"type:json;comment:地点标签"`

	Province string `json:"province" gorm:"type:varchar(50);comment:省份"`
	City     string `json:"city" gorm:"type:varchar(50);index:idx_city_district;comment:城市"`
	District string `json:"district" gorm:"type:varchar(100);index:idx_city_district;comment:区县"`
	Street   string `json:"street" gorm:"type:varchar(100);comment:街道"`
	Address  string `json:"address" gorm:"type:varchar(255);not null;index:ft_name,class:FULLTEXT;comment:地址"`
	Adcode   string `json:"adcode" gorm:"type:varchar(10);index;comment:区域编码"`

	Latitude  *float64 `json:"latitude" gorm:"type:decimal(11,8);index:idx_location;not null;comment:纬度"`
	Longitude *float64 `json:"longitude" gorm:"type:decimal(12,8);index:idx_location;not null;comment:经度"`

	Description string   `json:"description" gorm:"type:text;comment:地点描述"`
	Rating      *float64 `json:"rating" gorm:"type:decimal(2,1);index;comment:地点评分"`
	ReviewCount *int     `json:"review_count" gorm:"default:0;comment:点评数量"`

	OpeningHours     string     `json:"opening_hours" gorm:"type:varchar(255);comment:营业时间"`
	OpeningHoursJson JSONObject `json:"opening_hours_json" gorm:"type:json;comment:营业时间结构化数据"`
	IsOpen24h        bool       `json:"is_open_24h" gorm:"column:is_open_24h;default:false;comment:是否全天营业"`

	PriceRange string      `json:"price_range" gorm:"type:varchar(50);comment:价格范围描述"`
	PriceMin   *float64    `json:"price_min" gorm:"type:decimal(10,2);comment:最低价格"`
	PriceMax   *float64    `json:"price_max" gorm:"type:decimal(10,2);comment:最高价格"`
	PriceLevel *PriceLevel `json:"price_level" gorm:"type:enum('free','budget','moderate','expensive','luxury');comment:价格等级"`

	Phone       string `json:"phone" gorm:"type:varchar(50);comment:联系电话"`
	Website     string `json:"website" gorm:"type:varchar(255);comment:官方网站"`
	Email       string `json:"email" gorm:"type:varchar(100);comment:电子邮箱"`
	ContactInfo string `json:"contact_info" gorm:"type:varchar(100)"`

	CoverImage string    `json:"cover_image" gorm:"type:varchar(255);comment:封面图片URL"`
	Photos     JSONArray `json:"photos" gorm:"type:json;comment:地点图片URL列表"`

	AmapPoiID        string     `json:"amap_poi_id" gorm:"type:varchar(50);unique;comment:高德POI ID"`
	DataSource       DataSource `json:"data_source" gorm:"type:enum('user_created','amap_api','imported');not null;comment:数据来源"`
	APIResponseCache JSONObject `json:"api_response_cache" gorm:"type:json;comment:API响应缓存"`
	APILastUpdated   *time.Time `json:"api_last_updated" gorm:"comment:API数据最后更新时间"`

	PopularityScore int `json:"popularity_score" gorm:"default:0;index;comment:热门评分"`
	VisitCount      int `json:"visit_count" gorm:"default:0;comment:访问次数"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (Place) TableName() string {
	return "places"
}

type AmapPoiCache struct {
	ID            uint64     `json:"id" gorm:"primaryKey;autoIncrement;comment:缓存ID"`
	CacheKey      string     `json:"cache_key" gorm:"type:varchar(255);uniqueIndex;not null;comment:缓存键"`
	RequestParams JSONObject `json:"request_params" gorm:"type:json;comment:请求参数"`
	ResponseData  JSONObject `json:"response_data" gorm:"type:json;comment:响应数据"`
	PoiCount      int        `json:"poi_count"`
	HitCount      int        `json:"hit_count" gorm:"default:0;comment:命中次数"`
	ExpireAt      time.Time  `json:"expire_at" gorm:"index;not null;comment:过期时间"`
	CreatedAt     time.Time  `json:"created_at"`
	LastHitAt     *time.Time `json:"last_hit_at"`
}

func (AmapPoiCache) TableName() string {
	return "amap_poi_cache"
}
