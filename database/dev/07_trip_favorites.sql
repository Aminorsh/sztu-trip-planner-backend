USE trip_planner;

CREATE TABLE IF NOT EXISTS trip_favorites (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '收藏用户ID（关联users表，确定收藏归属）',
    trip_id BIGINT UNSIGNED NOT NULL COMMENT '被收藏行程ID（关联trips表，确定收藏的行程）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '收藏时间',
    deleted_at TIMESTAMP NULL COMMENT '软删除时间（取消收藏时标记，NULL表示有效收藏）',
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (trip_id) REFERENCES trips(id) ON DELETE CASCADE,
    
    UNIQUE KEY uk_user_trip (user_id, trip_id),
    
    INDEX idx_user_id (user_id),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='行程收藏表：存储用户收藏的行程记录，支撑收藏功能';
