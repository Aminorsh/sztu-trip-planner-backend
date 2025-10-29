USE trip_planner;

-- 创建行程表
CREATE TABLE IF NOT EXISTS trips (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL COMMENT '创建用户ID',
    title VARCHAR(200) NOT NULL COMMENT '行程标题',
    description TEXT COMMENT '行程描述',
    start_date DATE COMMENT '开始日期',
    end_date DATE COMMENT '结束日期',
    cover_image VARCHAR(255) COMMENT '封面图片URL',
    is_public BOOLEAN DEFAULT FALSE COMMENT '是否公开',
    share_token VARCHAR(32) UNIQUE COMMENT '分享令牌',
    status ENUM('draft', 'active', 'completed', 'archived') DEFAULT 'draft' COMMENT '行程状态',
    
    -- 路线规划相关字段
    origin_address VARCHAR(255) COMMENT '起点地址',
    destination_address VARCHAR(255) COMMENT '终点地址',
    total_distance DECIMAL(10,2) COMMENT '总距离(km)',
    estimated_duration INT COMMENT '预计时长(分钟)',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_dates (start_date, end_date),
    INDEX idx_share_token (share_token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='行程表';
