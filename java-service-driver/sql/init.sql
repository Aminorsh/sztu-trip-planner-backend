-- 创建users表（如果还没有）
CREATE TABLE IF NOT EXISTS users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(50) NOT NULL UNIQUE,
  email VARCHAR(100) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  avatar VARCHAR(255) DEFAULT '',
  role ENUM('user', 'admin') DEFAULT 'user',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 创建trips表
CREATE TABLE IF NOT EXISTS trips (
  id INT AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(100) NOT NULL,
  description TEXT,
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  is_public BOOLEAN DEFAULT FALSE,
  share_token VARCHAR(100) UNIQUE,
  cover_image VARCHAR(255) DEFAULT '',
  status ENUM('planning', 'ongoing', 'completed', 'cancelled') DEFAULT 'planning',
  user_id INT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  INDEX idx_user_id (user_id),
  INDEX idx_share_token (share_token),
  INDEX idx_start_date (start_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 创建trip_items表
CREATE TABLE IF NOT EXISTS trip_items (
  id INT AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(100) NOT NULL,
  description TEXT,
  location VARCHAR(100) NOT NULL,
  address VARCHAR(255),
  latitude DECIMAL(10, 8),
  longitude DECIMAL(11, 8),
  start_time DATETIME,
  end_time DATETIME,
  sort_order INT DEFAULT 0,
  visited BOOLEAN DEFAULT FALSE,
  visited_at DATETIME,
  category ENUM('attraction', 'food', 'accommodation', 'transport', 'shopping', 'entertainment', 'other') DEFAULT 'other',
  cost DECIMAL(10, 2) DEFAULT 0,
  notes TEXT,
  trip_id INT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  INDEX idx_trip_id_sort (trip_id, sort_order),
  INDEX idx_trip_id_visited (trip_id, visited),
  INDEX idx_category (category),
  FOREIGN KEY (trip_id) REFERENCES trips(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;