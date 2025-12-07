USE trip_planner;

-- 确保 places 表已经创建并优化（已在 03_places. sql 和 09_ver3. sql 中完成）

-- 为地点搜索添加额外索引
CREATE INDEX IF NOT EXISTS idx_data_source ON places(data_source);
CREATE INDEX IF NOT EXISTS idx_visit_count ON places(visit_count);

-- 优化全文搜索索引（如果不存在）
-- ALTER TABLE places ADD FULLTEXT INDEX ft_search (name, address);

-- 清理过期的 amap_poi_cache 数据的定时任务（可选）
-- 创建事件调度器来定期清理
SET GLOBAL event_scheduler = ON;

DELIMITER $$
CREATE EVENT IF NOT EXISTS cleanup_expired_amap_cache
ON SCHEDULE EVERY 1 DAY
DO
BEGIN
    DELETE FROM amap_poi_cache WHERE expire_at < NOW();
END$$
DELIMITER ;