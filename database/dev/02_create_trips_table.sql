USE trip_planner;

CREATE TABLE IF NOT EXISTS trips (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL COMMENT '行程创建者ID（关联users表，确定行程归属）',
    title VARCHAR(200) NOT NULL COMMENT '行程标题（如“2024国庆北京5日游”，用于列表展示）',
    description TEXT COMMENT '行程描述（补充行程背景、目的等信息，关联TripDetail.vue概览区）',
    start_date DATE COMMENT '行程开始日期（关联TripDetail.vue日历视图）',
    end_date DATE COMMENT '行程结束日期（关联TripDetail.vue日历视图，计算行程天数）',
    days INT COMMENT '行程总天数（用于前端天数切换，可选：手动输入或应用层计算）',
    trip_items_count INT DEFAULT 0 COMMENT '行程项总数（Dashboard性能优化字段）',
    photos_count INT DEFAULT 0 COMMENT '图片总数（Dashboard性能优化字段）',
    unique_places_count INT DEFAULT 0 COMMENT '独特地点数（Dashboard性能优化字段）',
    is_public BOOLEAN DEFAULT FALSE COMMENT '行程是否公开：true-公开，false-私有（关联分享功能权限控制）',
    share_token VARCHAR(32) UNIQUE COMMENT '行程分享令牌（公开分享时生成，用于通过链接访问行程）',
    status ENUM('draft', 'active', 'completed', 'archived') DEFAULT 'draft' COMMENT '行程状态：draft-草稿，active-进行中，completed-已完成，archived-已归档',
    
    -- 路线规划相关字段（关联TripDetail.vue地图区、总距离/总时间展示）
    origin_address VARCHAR(255) COMMENT '行程起点地址',
    destination_address VARCHAR(255) COMMENT '行程终点地址',
    destination_city VARCHAR(100) COMMENT '目的地城市（如“北京”，用于Dashboard列表展示）',
    destination_country VARCHAR(100) COMMENT '目的地国家（如“中国”，用于推荐算法）',
    total_distance DECIMAL(10,2) COMMENT '行程总距离（单位：km，关联TripDetail.vue本日概览摘要）',
    estimated_duration INT COMMENT '行程预计总时长（单位：分钟，关联TripDetail.vue本日概览摘要）',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '行程创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '行程信息最后更新时间',
    deleted_at TIMESTAMP NULL COMMENT '软删除时间（用于回收站功能，NULL表示未删除）',
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_dates (start_date, end_date),
    INDEX idx_share_token (share_token),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='行程表：存储行程整体信息，是所有行程关联数据的核心表';
