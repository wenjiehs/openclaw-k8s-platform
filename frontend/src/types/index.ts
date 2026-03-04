/**
 * TypeScript 类型定义文件
 * 统一定义前端使用的所有数据类型，与后端 API 保持一致
 */

// ============ 用户相关类型 ============

/** 用户角色 */
export type UserRole = 'employee' | 'admin' | 'super_admin'

/** 用户信息 */
export interface User {
  id: number
  username: string
  email: string
  department: string
  role: UserRole
}

/** 登录请求 */
export interface LoginRequest {
  username: string
  password: string
}

/** 登录响应 */
export interface LoginResponse {
  token: string
  expire_at: string
  user: User
}

// ============ 实例相关类型 ============

/** 实例规格 */
export type InstanceSpec = 'basic' | 'standard' | 'enterprise'

/** 实例状态 */
export type InstanceStatus = 'pending' | 'creating' | 'running' | 'stopped' | 'failed' | 'deleted'

/** 使用时长类型 */
export type DurationType = 'long' | 'temporary'

/** OpenClaw 实例 */
export interface Instance {
  id: number
  name: string
  user_id: number
  user?: User
  spec: InstanceSpec
  status: InstanceStatus
  namespace: string
  ingress_url: string
  duration_type: DurationType
  expire_at?: string
  created_at: string
  updated_at: string
  deleted_at?: string
}

/** 实例规格配置 */
export interface SpecConfig {
  label: string
  cpu: string
  memory: string
  description: string
}

/** 规格配置映射 */
export const SPEC_CONFIGS: Record<InstanceSpec, SpecConfig> = {
  basic: {
    label: '基础版',
    cpu: '1核',
    memory: '2GB',
    description: '适合个人轻量使用',
  },
  standard: {
    label: '标准版',
    cpu: '2核',
    memory: '4GB',
    description: '适合日常开发工作',
  },
  enterprise: {
    label: '企业版',
    cpu: '4核',
    memory: '8GB',
    description: '适合高负载场景',
  },
}

// ============ 申请相关类型 ============

/** 申请状态 */
export type ApplicationStatus = 'pending' | 'approved' | 'rejected' | 'cancelled'

/** 申请记录 */
export interface Application {
  id: number
  user_id: number
  user?: User
  instance_name: string
  spec: InstanceSpec
  duration_type: DurationType
  duration_days: number
  reason: string
  status: ApplicationStatus
  approver_id?: number
  approver?: User
  approve_note?: string
  approved_at?: string
  created_at: string
  updated_at: string
}

/** 创建申请请求 */
export interface CreateApplicationRequest {
  instance_name: string
  spec: InstanceSpec
  duration_type: DurationType
  duration_days?: number // 仅 temporary 类型需要
  reason: string
}

// ============ 审计日志类型 ============

/** 审计日志 */
export interface AuditLog {
  id: number
  user_id: number
  user?: User
  action: string
  resource_type: string
  resource_id: string
  ip: string
  user_agent: string
  result: 'success' | 'failed'
  extra?: string
  created_at: string
}

// ============ API 响应类型 ============

/** 通用 API 响应格式 */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

/** 分页数据 */
export interface PaginatedData<T> {
  list: T[]
  total: number
  page: number
  size: number
}

/** 分页查询参数 */
export interface PaginationParams {
  page?: number
  size?: number
}

// ============ 监控相关类型 ============

/** 监控汇总数据 */
export interface MetricsSummary {
  instances: {
    total: number
    running: number
    creating: number
    failed: number
  }
  applications: {
    pending: number
    today: number
  }
  users: {
    total: number
  }
}

/** 实例状态颜色映射 */
export const INSTANCE_STATUS_COLORS: Record<InstanceStatus, string> = {
  pending: 'default',
  creating: 'processing',
  running: 'success',
  stopped: 'warning',
  failed: 'error',
  deleted: 'default',
}

/** 实例状态文本映射 */
export const INSTANCE_STATUS_TEXT: Record<InstanceStatus, string> = {
  pending: '等待创建',
  creating: '创建中',
  running: '运行中',
  stopped: '已停止',
  failed: '创建失败',
  deleted: '已删除',
}

/** 申请状态颜色映射 */
export const APPLICATION_STATUS_COLORS: Record<ApplicationStatus, string> = {
  pending: 'processing',
  approved: 'success',
  rejected: 'error',
  cancelled: 'default',
}

/** 申请状态文本映射 */
export const APPLICATION_STATUS_TEXT: Record<ApplicationStatus, string> = {
  pending: '待审批',
  approved: '已批准',
  rejected: '已拒绝',
  cancelled: '已撤销',
}
