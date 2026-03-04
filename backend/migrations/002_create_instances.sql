-- 002_create_instances.sql
-- 创建实例表
-- 存储 OpenClaw 实例信息，每个实例对应一个 K8s Namespace

CREATE TABLE IF NOT EXISTS instances (
    id            SERIAL PRIMARY KEY,
    -- 实例名称，如 openclaw-zhangsan，全局唯一
    name          VARCHAR(50) UNIQUE NOT NULL,
    user_id       INT NOT NULL REFERENCES users(id),
    -- 规格：basic（基础版：1核2G）/ standard（标准版：2核4G）/ enterprise（企业版：4核8G）
    spec          VARCHAR(20) NOT NULL DEFAULT 'standard' CHECK (spec IN ('basic', 'standard', 'enterprise')),
    -- 状态：pending/creating/running/stopped/failed/deleted
    status        VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'creating', 'running', 'stopped', 'failed', 'deleted')),
    -- K8s Namespace，如 openclaw-zhangsan
    namespace     VARCHAR(100),
    -- 外部访问地址，如 https://openclaw-zhangsan.tke-cloud.com
    ingress_url   VARCHAR(200),
    -- 使用时长类型：long（长期）/ temporary（临时）
    duration_type VARCHAR(20) DEFAULT 'long' CHECK (duration_type IN ('long', 'temporary')),
    -- 临时实例到期时间（长期实例为 NULL）
    expire_at     TIMESTAMP,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- 软删除时间（NULL 表示未删除）
    deleted_at    TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_instances_user_id ON instances(user_id);
CREATE INDEX IF NOT EXISTS idx_instances_status ON instances(status);
CREATE INDEX IF NOT EXISTS idx_instances_namespace ON instances(namespace);
CREATE INDEX IF NOT EXISTS idx_instances_deleted_at ON instances(deleted_at);
CREATE INDEX IF NOT EXISTS idx_instances_expire_at ON instances(expire_at);

COMMENT ON TABLE instances IS '实例表：存储 OpenClaw K8s 实例信息';
COMMENT ON COLUMN instances.spec IS '规格：basic/standard/enterprise';
COMMENT ON COLUMN instances.status IS '状态：pending/creating/running/stopped/failed/deleted';
COMMENT ON COLUMN instances.namespace IS 'K8s Namespace 名称';
