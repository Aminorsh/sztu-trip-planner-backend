package dto

type PlaceSearchRequest struct {
	Keyword string `form:"keyword" binding:"required"`
	// City     string `json:"city"`
	// Category string `json:"category"`
	FilterType string `form:"filter_type"`
	// Page       int    `json:"page" binding:"min=0"`
	// PageSize   int    `json:"page_size" binding:"min=1,max=50"`
	SortOrder string `form:"sort_order"`
}

type Location struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

type PlaceResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Address      string  `json:"address"`
	Description  string  `json:"description"`
	Image        string  `json:"image"`
	Rating       float64 `json:"rating"`
	Distance     float64 `json:"distance"`
	OpeningHours string  `json:"openingHours"`
	Added        bool    `json:"added"`
	Expanded     bool    `json:"expanded"`
	// 保留额外信息供后续使用
	Category   string   `json:"category,omitempty"`
	Location   Location `json:"location,omitempty"`
	Tel        string   `json:"tel,omitempty"`
	PriceRange string   `json:"priceRange,omitempty"`
	Photos     []string `json:"photos,omitempty"`
}

type PlaceSearchResponse struct {
	Places []PlaceResponse `json:"places"`
	Total  int             `json:"total"`
}

type AmapSearchResponse struct {
	Status string `json:"status"`
	Count  string `json:"count"` // Amap returns count as string
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
