package dto

import "encoding/json"

type PlaceSearchRequest struct {
	Keyword string `form:"keyword" binding:"required"`
	// City     string `json:"city"`
	// Category string `json:"category"`
	FilterType string `form:"filter_type"`
	// Page       int    `json:"page" binding:"min=0"`
	// PageSize   int    `json:"page_size" binding:"min=1,max=50"`
	SortOrder string `form:"sort_order"`
}

// FlexibleString 用于处理高德API返回的字段，可能是字符串或空数组
type FlexibleString string

func (fs *FlexibleString) UnmarshalJSON(data []byte) error {
	// 尝试解析为字符串
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fs = FlexibleString(s)
		return nil
	}

	// 如果是数组（通常是空数组），设置为空字符串
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err == nil {
		*fs = FlexibleString("")
		return nil
	}

	// 其他情况设置为空字符串
	*fs = FlexibleString("")
	return nil
}

type Location struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

type PlaceResponse struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Address      string   `json:"address"`
	Description  string   `json:"description"`
	Image        string   `json:"image"`
	Rating       float64  `json:"rating"`
	Distance     float64  `json:"distance"`
	OpeningHours string   `json:"openingHours"`
	Added        bool     `json:"added"`
	Expanded     bool     `json:"expanded"`
	Type         string   `json:"type,omitempty"`
	Location     Location `json:"location,omitempty"`
	Photos       []string `json:"photos,omitempty"`
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
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	Typecode     string         `json:"typecode"`
	Address      string         `json:"address"`
	Location     string         `json:"location"`
	Pcode        string         `json:"pcode"`
	Pname        string         `json:"pname"`
	Citycode     string         `json:"citycode"`
	Cityname     string         `json:"cityname"`
	Adcode       string         `json:"adcode"`
	Adname       string         `json:"adname"`
	Tel          FlexibleString `json:"tel"`          // 可能是字符串或空数组
	Alias        FlexibleString `json:"alias"`        // 可能是字符串或空数组
	Website      FlexibleString `json:"website"`      // 可能是字符串或空数组
	Email        FlexibleString `json:"email"`        // 可能是字符串或空数组
	Postcode     FlexibleString `json:"postcode"`     // 可能是字符串或空数组
	Parking_type FlexibleString `json:"parking_type"` // 可能是字符串或空数组
	Photos       []struct {
		Title FlexibleString `json:"title"`
		URL   string         `json:"url"`
	} `json:"photos"`
	// 额外字段（当extensions=all时返回）
	Biz_ext struct {
		Rating          FlexibleString `json:"rating"`          // 评分
		Cost            FlexibleString `json:"cost"`            // 人均消费
		Meal_ordering   FlexibleString `json:"meal_ordering"`   // 是否可订餐（逐渐废弃）
		Seat_ordering   FlexibleString `json:"seat_ordering"`   // 是否可选座（逐渐废弃）
		Ticket_ordering FlexibleString `json:"ticket_ordering"` // 是否可订票（逐渐废弃）
		Hotel_ordering  FlexibleString `json:"hotel_ordering"`  // 是否可订房（逐渐废弃）
	} `json:"biz_ext"`
	Business_area FlexibleString `json:"business_area"` // 商圈
	Tag           FlexibleString `json:"tag"`           // 特色内容（如特色菜）
	Biz_type      FlexibleString `json:"biz_type"`      // 行业类型
}

// AmapDetailResponse 高德POI详情API响应
type AmapDetailResponse struct {
	Status string `json:"status"`
	Info   string `json:"info"`
	Pois   []POI  `json:"pois"`
}

// WikipediaResponse Wikipedia API响应
type WikipediaResponse struct {
	Type         string `json:"type"`
	Title        string `json:"title"`
	DisplayTitle string `json:"displaytitle"`
	Extract      string `json:"extract"` // 简介
	Description  string `json:"description"`
	Thumbnail    struct {
		Source string `json:"source"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"thumbnail,omitempty"`
	Originalimage struct {
		Source string `json:"source"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"originalimage,omitempty"`
}
