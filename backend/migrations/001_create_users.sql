-- 001_create_users.sql
-- 创建用户表
-- 存储平台用户信息，包括员工和管理员

CREATE TABLE IF NOT EXISTS users (
    id          SERIAL PRIMARY KEY,
    username    VARCHAR(50) UNIQUE NOT NULL,
    email       VARCHAR(100) NOT NULL,
    department  VARCHAR(100),
    -- 角色：employee（员工）/ admin（管理员）/ super_admin（超级管理员）
    role        VARCHAR(20) NOT NULL DEFAULT 'employee' CHECK (role IN ('employee', 'admin', 'super_admin')),
    -- 密码哈希（使用 bcrypt）
    password_hash VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_department ON users(department);

-- 插入默认超级管理员账号（密码：admin123，实际部署时必须修改）
-- 密码哈希使用 bcrypt，这里是 admin123 的哈希值
INSERT INTO users (username, email, department, role, password_hash)
VALUES 
    ('admin', 'admin@company.com', '技术部', 'super_admin', '$2a$10$GoBD4h8xragGE.3FpWXTdO86jBdIajirRiVWo/R3Pcan.FULOsk1C')
ON CONFLICT (username) DO NOTHING;

COMMENT ON TABLE users IS '用户表：存储平台用户信息';
COMMENT ON COLUMN users.role IS '角色：employee/admin/super_admin';
