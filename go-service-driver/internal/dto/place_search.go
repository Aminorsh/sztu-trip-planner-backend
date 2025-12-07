package dto

type PlaceSearchRequest struct {
	Keyword  string `json:"keyword" binding:"required"`
	City     string `json:"city"`
	Category string `json:"category"`
	Page     int    `json:"page" binding:"min=0"`
	PageSize int    `json:"page_size" binding:"min=1,max=50"`
}

type Location struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

type PlaceResponse struct {
	ID           uint64   `json:"id"`
	AmapPoiID    string   `json:"amap_poi_id,omitempty"`
	Name         string   `json:"name"`
	NameEn       string   `json:"name_en,omitempty"`
	Address      string   `json:"address"`
	Location     Location `json:"location"`
	Province     string   `json:"province,omitempty"`
	City         string   `json:"city"`
	District     string   `json:"district,omitempty"`
	Category     string   `json:"category"`
	AmapType     string   `json:"amap_type,omitempty"`
	Tel          string   `json:"tel,omitempty"`
	Rating       float64  `json:"rating,omitempty"`
	ReviewCount  int      `json:"review_count"`
	PriceRange   string   `json:"price_range,omitempty"`
	PriceLevel   string   `json:"price_level,omitempty"`
	OpeningHours string   `json:"opening_hours,omitempty"`
	CoverImage   string   `json:"cover_image,omitempty"`
	Photos       []string `json:"photos,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type PlaceSearchResponse struct {
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	Places   []PlaceResponse `json:"places"`
}

type AmapSearchResponse struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
	Info   string `json:"info"`
	Pois   []POI  `json:"pois"`
}

type POI struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Typecode string `json:"typecode"`
	Address  string `json:"address"`
	Location string `json:"location"`
	Pcode    string `json:"pcode"`
	Pname    string `json:"pname"`
	Citycode string `json:"citycode"`
	Cityname string `json:"cityname"`
	Adcode   string `json:"adcode"`
	Adname   string `json:"adname"`
	Tel      string `json:"tel"`
	Photos   []struct {
		URL string `json:"url"`
	} `json:"photos"`
}
