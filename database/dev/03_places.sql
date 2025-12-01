USE trip_planner;

CREATE TABLE IF NOT EXISTS places (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(200) NOT NULL COMMENT '地点名称（如“故宫博物院”，用于搜索结果、详情页展示）',
    category ENUM('scenic', 'restaurant', 'hotel', 'other') NOT NULL COMMENT '地点类别：scenic-景点，restaurant-餐饮，hotel-酒店，other-其他（关联PlaceSearch.vue类别筛选）',
    address VARCHAR(255) NOT NULL COMMENT '地点详细地址（关联PlaceDetail.vue地址展示、地图定位）',
    latitude DECIMAL(10,6) COMMENT '地点纬度（关联TripDetail.vue地图区标记、地点定位）',
    longitude DECIMAL(10,6) COMMENT '地点经度（关联TripDetail.vue地图区标记、地点定位）',
    description TEXT COMMENT '地点简介（关联PlaceDetail.vue简介区，补充地点背景信息）',
    rating DECIMAL(2,1) COMMENT '地点评分（0-5分，关联PlaceDetail.vue评分展示，用于用户参考）',
    opening_hours VARCHAR(255) COMMENT '地点开放/营业时长（如“08:30-17:00”，关联PlaceDetail.vue信息表格）',
    price_range VARCHAR(50) COMMENT '地点价格范围（如“¥50-100”，关联PlaceDetail.vue信息表格）',
    contact_info VARCHAR(100) COMMENT '地点联系方式（如电话、官网，关联PlaceDetail.vue信息表格）',
    cover_image VARCHAR(255) COMMENT '地点封面图片URL（关联PlaceSearch.vue卡片图片、PlaceDetail.vue轮播图）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '地点数据创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '地点数据最后更新时间',
    deleted_at TIMESTAMP NULL COMMENT '软删除时间（用于回收站功能，NULL表示未删除）',
    
    INDEX idx_category (category),
    INDEX idx_location (latitude, longitude),
    INDEX idx_rating (rating),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='地点表：存储行程中涉及的各类地点基础信息，支撑地点搜索与详情展示';
