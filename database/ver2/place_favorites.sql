USE trip_planner;

CREATE TABLE IF NOT EXISTS place_favorites (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL COMMENT '收藏用户ID（关联users表，确定收藏归属）',
    place_id INT NOT NULL COMMENT '被收藏地点ID（关联places表，确定收藏的地点）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '收藏时间',
    deleted_at TIMESTAMP NULL COMMENT '软删除时间（取消收藏时标记，NULL表示有效收藏）',
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (place_id) REFERENCES places(id) ON DELETE CASCADE,
    
    UNIQUE KEY uk_user_place (user_id, place_id),
    
    INDEX idx_user_id (user_id),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='地点收藏表：存储用户收藏的地点记录，支撑收藏功能';
