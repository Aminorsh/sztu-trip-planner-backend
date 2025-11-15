USE trip_planner;

CREATE TABLE IF NOT EXISTS trip_photos (
    id INT AUTO_INCREMENT PRIMARY KEY,
    trip_id INT NOT NULL COMMENT '关联行程ID（确定图片归属的行程）',
    trip_item_id INT COMMENT '关联行程项ID（可选，如“故宫游览”行程项的照片，可为空）',
    place_id INT COMMENT '关联地点ID（可选，如“故宫博物院”地点的照片，可为空）',
    photo_url VARCHAR(255) NOT NULL COMMENT '图片URL（关联TripDetail.vue照片展示、PlaceDetail.vue图片轮播）',
    caption VARCHAR(200) COMMENT '图片描述（补充图片背景信息，如“2024.10.01故宫午门合影”）',
    upload_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '图片上传时间',
    is_cover BOOLEAN DEFAULT FALSE COMMENT '是否为行程封面图：true-是，false-否（统一控制行程封面）',
    uploader_id INT NOT NULL COMMENT '上传者用户ID（用于多人协作时追踪图片归属）',
    deleted_at TIMESTAMP NULL COMMENT '软删除时间（用于回收站功能，NULL表示未删除）',

    FOREIGN KEY (trip_id) REFERENCES trips(id) ON DELETE CASCADE,
    FOREIGN KEY (trip_item_id) REFERENCES trip_items(id) ON DELETE SET NULL,
    FOREIGN KEY (place_id) REFERENCES places(id) ON DELETE SET NULL,
    FOREIGN KEY (uploader_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_trip_id (trip_id),
    INDEX idx_trip_item_id (trip_item_id),
    INDEX idx_uploader_id (uploader_id),
    INDEX idx_is_cover (is_cover),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='行程图片表：存储行程、行程项、地点相关的图片，支撑图片展示与回忆功能';
