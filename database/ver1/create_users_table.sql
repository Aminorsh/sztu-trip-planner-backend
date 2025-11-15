CREATE DATABASE IF NOT EXISTS trip_planner CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE trip_planner;

CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名（唯一标识，用于登录）',
    email VARCHAR(100) UNIQUE NOT NULL COMMENT '用户邮箱（唯一，用于登录、找回密码）',
    password_hash VARCHAR(255) NOT NULL COMMENT '加密后的密码（不存储明文，保障安全）',
    display_name VARCHAR(100) COMMENT '用户显示名称（用于界面展示，可与用户名不同）',
    avatar_url VARCHAR(255) COMMENT '用户头像URL（关联Profile.vue页面的头像上传功能）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '用户账号创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '用户信息最后更新时间',
    last_login_at TIMESTAMP NULL COMMENT '用户最后登录时间（用于账号状态追踪）',
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active' COMMENT '用户账号状态：active-正常，inactive-未激活，suspended-封禁',
    
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表：存储行程规划系统所有注册用户信息';
