USE trip_planner;

CREATE TABLE IF NOT EXISTS trip_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    trip_id INT NOT NULL COMMENT '关联行程ID（确定行程项归属的行程）',
    place_id INT NOT NULL COMMENT '关联地点ID（确定行程项对应的地点）',
    day_number INT NOT NULL COMMENT '行程项所属天数（如“1”代表行程第1天，关联TripDetail.vue天数切换按钮）',
    start_time TIME COMMENT '行程项开始时间（如“09:00”，关联TripDetail.vue行程项列表、日历视图）',
    end_time TIME COMMENT '行程项结束时间（如“11:30”，关联TripDetail.vue行程项列表、日历视图）',
    sequence INT NOT NULL COMMENT '当日行程项排序序号（用于TripDetail.vue行程项拖拽排序，确定展示顺序）',
    note TEXT COMMENT '行程项备注（如“需提前预约”，关联TripDetail.vue行程项编辑备注功能）',
    priority ENUM('high', 'medium', 'low') DEFAULT 'medium' COMMENT '行程项优先级：high-高，medium-中，low-低（关联TripDetail.vue行程项编辑优先级功能）',
    is_completed BOOLEAN DEFAULT FALSE COMMENT '行程项是否已完成（打卡状态，关联TripDetail.vue行程项打卡按钮）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '行程项创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '行程项信息最后更新时间',
    
    FOREIGN KEY (trip_id) REFERENCES trips(id) ON DELETE CASCADE,
    FOREIGN KEY (place_id) REFERENCES places(id) ON DELETE CASCADE,
    
    INDEX idx_trip_day (trip_id, day_number),
    INDEX idx_place_id (place_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='行程项表：存储每日行程的具体细节，是行程与地点的关联表';
