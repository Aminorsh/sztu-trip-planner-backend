USE trip_planner;

CREATE INDEX idx_destination_city ON trips(destination_city);
CREATE INDEX idx_visibility ON trips(visibility);

ALTER TABLE trips
MODIFY COLUMN end_date DATE NULL COMMENT '行程结束日期（可选，自动计算或手动填写）',
MODIFY COLUMN days INT NULL COMMENT '行程总天数（自动计算：end_date - start_date + 1）',
MODIFY COLUMN destination_city VARCHAR(100) NULL COMMENT '目的地城市（可选，后期填写）';
-- -- ========== 优化 users 表 ==========
-- ALTER TABLE users
-- ADD COLUMN phone VARCHAR(20) COMMENT '手机号' AFTER email,
-- ADD COLUMN timezone VARCHAR(50) DEFAULT 'Asia/Shanghai' COMMENT '用户时区' AFTER last_login_at,
-- ADD COLUMN language VARCHAR(10) DEFAULT 'zh-CN' COMMENT '首选语言' AFTER timezone,
-- ADD COLUMN bio TEXT COMMENT '个人简介' AFTER language;

-- CREATE INDEX idx_phone ON users(phone);


-- ========== 创建地点评价表 ==========
CREATE TABLE IF NOT EXISTS place_reviews (
    id INT AUTO_INCREMENT PRIMARY KEY,
    place_id INT NOT NULL COMMENT '关联地点ID',
    user_id INT NOT NULL COMMENT '评价用户ID',
    trip_item_id INT COMMENT '关联行程项ID',
    rating DECIMAL(2,1) NOT NULL COMMENT '评分（0-5分）',
    content TEXT COMMENT '评价内容',
    visit_date DATE COMMENT '游玩日期',
    helpful_count INT DEFAULT 0 COMMENT '有帮助数',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (place_id) REFERENCES places(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (trip_item_id) REFERENCES trip_items(id) ON DELETE SET NULL,
    
    INDEX idx_place_id (place_id),
    INDEX idx_user_id (user_id),
    INDEX idx_rating (rating),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='地点评价表';