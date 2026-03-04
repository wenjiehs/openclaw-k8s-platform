-- 005_create_audit_logs.sql
-- 创建审计日志表
-- 记录所有关键操作，满足等保三级合规要求，日志保留 180 天

CREATE TABLE IF NOT EXISTS audit_logs (
    id            SERIAL PRIMARY KEY,
    user_id       INT REFERENCES users(id),
    -- 操作类型：login/logout/create/delete/approve/reject/view 等
    action        VARCHAR(50) NOT NULL,
    -- 资源类型：instance/application/user 等
    resource_type VARCHAR(50),
    -- 资源 ID（字符串格式，兼容不同类型的 ID）
    resource_id   VARCHAR(100),
    -- 操作来源 IP
    ip            VARCHAR(50),
    -- 浏览器 User-Agent
    user_agent    TEXT,
    -- 操作结果：success（成功）/ failed（失败）
    result        VARCHAR(20) DEFAULT 'success' CHECK (result IN ('success', 'failed')),
    -- 额外信息（JSON 格式，存储操作的详细参数）
    extra         TEXT,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引（查询性能优化）
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- 创建自动清理过期日志的函数（保留 180 天）
-- 需要在 PostgreSQL 中配置定时任务执行此函数
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs()
RETURNS void AS $$
BEGIN
    DELETE FROM audit_logs 
    WHERE created_at < NOW() - INTERVAL '180 days';
    
    RAISE NOTICE '已清理 % 天前的审计日志', 180;
END;
$$ LANGUAGE plpgsql;

COMMENT ON TABLE audit_logs IS '审计日志表：记录所有关键操作，满足合规要求，保留 180 天';
COMMENT ON COLUMN audit_logs.action IS '操作类型：login/logout/create/delete/approve/reject 等';
COMMENT ON COLUMN audit_logs.result IS '操作结果：success/failed';
