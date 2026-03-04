-- 004_create_cost_records.sql
-- 创建成本记录表
-- 按月统计每个实例的资源成本，支持部门成本分摊报表

CREATE TABLE IF NOT EXISTS cost_records (
    id          SERIAL PRIMARY KEY,
    instance_id INT NOT NULL REFERENCES instances(id),
    user_id     INT NOT NULL REFERENCES users(id),
    -- 部门（冗余存储，方便按部门汇总统计）
    department  VARCHAR(100),
    -- 统计月份，格式 YYYY-MM，如 2026-03
    month       VARCHAR(7) NOT NULL,
    -- 规格费用（按实例规格计算的固定月费）
    spec_cost   DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    -- API 调用次数（该月累计）
    api_calls   INT DEFAULT 0,
    -- API 超额费用（超出免费额度的费用）
    api_cost    DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    -- 总费用 = spec_cost + api_cost
    total_cost  DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_cost_records_instance_id ON cost_records(instance_id);
CREATE INDEX IF NOT EXISTS idx_cost_records_user_id ON cost_records(user_id);
CREATE INDEX IF NOT EXISTS idx_cost_records_department ON cost_records(department);
CREATE INDEX IF NOT EXISTS idx_cost_records_month ON cost_records(month);
-- 联合唯一索引：每个实例每月只有一条记录
CREATE UNIQUE INDEX IF NOT EXISTS idx_cost_records_instance_month ON cost_records(instance_id, month);

COMMENT ON TABLE cost_records IS '成本记录表：按月统计每个实例的资源成本';
COMMENT ON COLUMN cost_records.month IS '统计月份，格式 YYYY-MM';
COMMENT ON COLUMN cost_records.spec_cost IS '规格基础费用';
COMMENT ON COLUMN cost_records.api_cost IS 'API 调用超额费用';
