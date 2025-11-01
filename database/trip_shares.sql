USE trip_planner;

CREATE TABLE IF NOT EXISTS trip_shares (
    id INT AUTO_INCREMENT PRIMARY KEY,
    trip_id INT NOT NULL COMMENT '关联行程ID（确定分享的行程）',
    share_token VARCHAR(32) UNIQUE NOT NULL COMMENT '分享令牌（与trips表share_token关联，用于通过链接验证访问权限）',
    shared_by INT NOT NULL COMMENT '分享者用户ID（关联users表，记录分享操作人）',
    shared_to_email VARCHAR(100) COMMENT '分享目标邮箱（可选，定向分享时填写；公开分享时为空）',
    permission ENUM('view', 'edit') DEFAULT 'view' COMMENT '分享权限：view-仅查看，edit-可编辑（控制被分享者操作范围）',
    expire_time TIMESTAMP COMMENT '分享链接过期时间（可选，如“2024-11-01 23:59:59”，过期后链接失效；永久有效时为空）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '分享创建时间',
    is_active BOOLEAN DEFAULT TRUE COMMENT '分享是否有效：true-有效，false-已失效（支持手动取消分享）',
    
    FOREIGN KEY (trip_id) REFERENCES trips(id) ON DELETE CASCADE,
    FOREIGN KEY (shared_by) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (share_token) REFERENCES trips(share_token) ON DELETE CASCADE,
    
    INDEX idx_trip_id (trip_id),
    INDEX idx_share_token (share_token),
    INDEX idx_shared_to_email (shared_to_email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='行程分享记录表：存储行程分享的详细信息，支撑分享权限控制与分享记录管理';
