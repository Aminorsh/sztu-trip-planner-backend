USE trip_planner;

-- ========== 优化 places 表 ==========
ALTER TABLE places
-- 添加高德地图字段
ADD COLUMN name_en VARCHAR(200) COMMENT '地点英文名称' AFTER name,
ADD COLUMN amap_type VARCHAR(100) COMMENT '高德地图分类' AFTER category,
ADD COLUMN amap_typecode VARCHAR(20) COMMENT '高德分类编码' AFTER amap_type,
ADD COLUMN tags JSON COMMENT '地点标签数组' AFTER amap_typecode,

-- 添加地址组件
ADD COLUMN province VARCHAR(50) COMMENT '省份' AFTER tags,
ADD COLUMN city VARCHAR(50) COMMENT '城市' AFTER province,
ADD COLUMN district VARCHAR(50) COMMENT '区/县' AFTER city,
ADD COLUMN street VARCHAR(100) COMMENT '街道' AFTER district,
ADD COLUMN adcode VARCHAR(10) COMMENT '行政区划代码' AFTER address,

-- 修改坐标精度
MODIFY COLUMN latitude DECIMAL(11,8) COMMENT '纬度（精确到1mm）',
MODIFY COLUMN longitude DECIMAL(12,8) COMMENT '经度（精确到1mm）',

-- 添加评论数量
ADD COLUMN review_count INT DEFAULT 0 COMMENT '评论数量' AFTER rating,

-- 优化营业时间
ADD COLUMN opening_hours_json JSON COMMENT '营业时间（JSON格式）' AFTER opening_hours,
ADD COLUMN is_open_24h BOOLEAN DEFAULT FALSE COMMENT '是否24小时营业' AFTER opening_hours_json,

-- 添加价格字段
ADD COLUMN price_min DECIMAL(10,2) COMMENT '最低价格（元）' AFTER price_range,
ADD COLUMN price_max DECIMAL(10,2) COMMENT '最高价格（元）' AFTER price_min,
ADD COLUMN price_level ENUM('free', 'budget', 'moderate', 'expensive', 'luxury') COMMENT '价格等级' AFTER price_max,

-- 添加联系方式
ADD COLUMN phone VARCHAR(50) COMMENT '联系电话' AFTER price_level,
ADD COLUMN website VARCHAR(255) COMMENT '官方网站' AFTER phone,
ADD COLUMN email VARCHAR(100) COMMENT '邮箱' AFTER website,

-- 添加照片数组
ADD COLUMN photos JSON COMMENT '图片数组' AFTER cover_image,

-- 添加高德API字段
ADD COLUMN amap_poi_id VARCHAR(50) UNIQUE COMMENT '高德地图POI ID' AFTER photos,
ADD COLUMN data_source ENUM('user_created', 'amap_api', 'imported') DEFAULT 'user_created' COMMENT '数据来源' AFTER amap_poi_id,
ADD COLUMN api_response_cache JSON COMMENT '原始API响应缓存' AFTER data_source,
ADD COLUMN api_last_updated TIMESTAMP NULL COMMENT 'API数据最后更新时间' AFTER api_response_cache,

-- 添加推荐算法字段
ADD COLUMN popularity_score INT DEFAULT 0 COMMENT '热度分数' AFTER api_last_updated,
ADD COLUMN visit_count INT DEFAULT 0 COMMENT '被访问次数' AFTER popularity_score;

-- 添加新索引
CREATE INDEX idx_amap_poi_id ON places(amap_poi_id);
CREATE INDEX idx_adcode ON places(adcode);
CREATE INDEX idx_city_district ON places(city, district);
CREATE INDEX idx_popularity ON places(popularity_score);
CREATE FULLTEXT INDEX ft_name ON places(name, address);


-- ========== 创建高德地图API缓存表 ==========
CREATE TABLE IF NOT EXISTS amap_poi_cache (
    id INT AUTO_INCREMENT PRIMARY KEY,
    cache_key VARCHAR(255) UNIQUE NOT NULL COMMENT '缓存键',
    request_params JSON COMMENT '请求参数',
    response_data JSON NOT NULL COMMENT '完整API响应数据',
    poi_count INT COMMENT 'POI数量',
    hit_count INT DEFAULT 0 COMMENT '缓存命中次数',
    expire_at TIMESTAMP NOT NULL COMMENT '缓存过期时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_hit_at TIMESTAMP NULL COMMENT '最后命中时间',
    
    INDEX idx_cache_key (cache_key),
    INDEX idx_expire_at (expire_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='高德地图API缓存表';


-- ========== 优化 trip_items 表 ==========
ALTER TABLE trip_items
-- 修改 place_id 为可选
DROP FOREIGN KEY trip_items_ibfk_2,
MODIFY COLUMN place_id INT COMMENT '关联地点ID（可为空）',
ADD COLUMN custom_place_name VARCHAR(200) COMMENT '自定义地点名称' AFTER place_id,

-- 添加时长字段
ADD COLUMN duration_minutes INT COMMENT '持续时长（分钟）' AFTER end_time,

-- 添加交通信息
ADD COLUMN transport_mode ENUM('walk', 'car', 'bus', 'subway', 'taxi', 'train', 'plane', 'other') COMMENT '交通方式' AFTER is_completed,
ADD COLUMN transport_distance DECIMAL(10,2) COMMENT '交通距离（km）' AFTER transport_mode,
ADD COLUMN transport_duration INT COMMENT '交通时长（分钟）' AFTER transport_distance,

-- 添加预算字段
ADD COLUMN budget_amount DECIMAL(10,2) COMMENT '预算金额（元）' AFTER transport_duration,
ADD COLUMN actual_amount DECIMAL(10,2) COMMENT '实际花费（元）' AFTER budget_amount,
ADD COLUMN currency VARCHAR(3) DEFAULT 'CNY' COMMENT '货币代码' AFTER actual_amount,

-- 添加天气字段
ADD COLUMN weather_condition VARCHAR(50) COMMENT '天气状况' AFTER currency,
ADD COLUMN temperature_high INT COMMENT '最高气温（℃）' AFTER weather_condition,
ADD COLUMN temperature_low INT COMMENT '最低气温（℃）' AFTER temperature_high,

-- 添加提醒时间
ADD COLUMN reminder_time TIMESTAMP NULL COMMENT '提醒时间' AFTER note;

-- 重新添加外键
ALTER TABLE trip_items
ADD FOREIGN KEY (place_id) REFERENCES places(id) ON DELETE SET NULL;


-- ========== 优化 trips 表 ==========
ALTER TABLE trips
ADD COLUMN cover_image_url VARCHAR(255) COMMENT '行程封面图URL' AFTER status,
ADD COLUMN budget_total DECIMAL(10,2) COMMENT '总预算（元）' AFTER cover_image_url,
ADD COLUMN actual_total DECIMAL(10,2) COMMENT '实际总花费（元）' AFTER budget_total,
ADD COLUMN weather_checked BOOLEAN DEFAULT FALSE COMMENT '是否已查询天气' AFTER actual_total,
ADD COLUMN visibility ENUM('private', 'friends', 'public') DEFAULT 'private' COMMENT '可见性' AFTER is_public;
