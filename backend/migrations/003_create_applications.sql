-- 003_create_applications.sql
-- 创建申请记录表
-- 存储员工申请 OpenClaw 实例的流程记录，包括申请内容和审批结果

CREATE TABLE IF NOT EXISTS applications (
    id            SERIAL PRIMARY KEY,
    user_id       INT NOT NULL REFERENCES users(id),
    -- 申请的实例名称
    instance_name VARCHAR(50) NOT NULL,
    -- 申请规格：basic/standard/enterprise
    spec          VARCHAR(20) NOT NULL CHECK (spec IN ('basic', 'standard', 'enterprise')),
    -- 使用时长类型：long（长期）/ temporary（临时）
    duration_type VARCHAR(20) DEFAULT 'long' CHECK (duration_type IN ('long', 'temporary')),
    -- 临时实例使用天数（仅 temporary 类型有效）
    duration_days INT DEFAULT 0,
    -- 申请理由
    reason        TEXT NOT NULL,
    -- 状态：pending（待审批）/ approved（已批准）/ rejected（已拒绝）/ cancelled（已撤销）
    status        VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    -- 审批人（外键关联 users 表）
    approver_id   INT REFERENCES users(id),
    -- 审批备注
    approve_note  TEXT,
    -- 审批时间
    approved_at   TIMESTAMP,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_applications_user_id ON applications(user_id);
CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);
CREATE INDEX IF NOT EXISTS idx_applications_approver_id ON applications(approver_id);
CREATE INDEX IF NOT EXISTS idx_applications_created_at ON applications(created_at);
-- 部分索引：只索引待审批的申请（提高审批列表查询性能）
CREATE INDEX IF NOT EXISTS idx_applications_pending ON applications(created_at) WHERE status = 'pending';

COMMENT ON TABLE applications IS '申请记录表：员工申请 OpenClaw 实例的流程记录';
COMMENT ON COLUMN applications.status IS '审批状态：pending/approved/rejected/cancelled';
