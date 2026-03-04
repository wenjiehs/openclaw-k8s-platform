/**
 * 管理员 API 接口
 * 封装审批管理、实例管理、监控等管理员接口
 */

import { get, post } from '@/utils/request'
import type { Application, AuditLog, Instance, MetricsSummary, PaginatedData, PaginationParams } from '@/types'

/** 获取待审批申请列表 */
export const getAdminApplications = (params?: PaginationParams & { status?: string }) => {
  return get<PaginatedData<Application>>('/api/v1/admin/applications', params as Record<string, unknown>)
}

/** 批准申请 */
export const approveApplication = (id: number, note?: string) => {
  return post(`/api/v1/admin/applications/${id}/approve`, { note })
}

/** 拒绝申请 */
export const rejectApplication = (id: number, note: string) => {
  return post(`/api/v1/admin/applications/${id}/reject`, { note })
}

/** 获取所有实例列表 */
export const getAdminInstances = (params?: PaginationParams & { status?: string; department?: string }) => {
  return get<PaginatedData<Instance>>('/api/v1/admin/instances', params as Record<string, unknown>)
}

/** 获取监控汇总数据 */
export const getMetricsSummary = () => {
  return get<MetricsSummary>('/api/v1/admin/metrics/summary')
}

/** 获取审计日志列表 */
export const getAdminAuditLogs = (params?: PaginationParams & {
  user_id?: number
  action?: string
  result?: string
  start_time?: string
  end_time?: string
}) => {
  return get<PaginatedData<AuditLog>>('/api/v1/admin/audit-logs', params as Record<string, unknown>)
}
