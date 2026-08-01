-- ============================================================
-- 清理 systemcheck 表中的重复记录
-- 保留每个 client_uuid 的最新一条记录（MAX(id)）
-- ============================================================

DELETE FROM systemcheck 
WHERE id NOT IN (
    SELECT MAX(id) 
    FROM systemcheck 
    GROUP BY client_uuid
);
