/**
 * 实例管理 API 接口
 * 封装实例列表、详情、删除等接口
 */

import { get, post, del } from '@/utils/request'
import type { Instance, Application, CreateApplicationRequest, PaginatedData, PaginationParams } from '@/types'

// ============ 员工接口 ============

/** 我的实例列表 */
export const getMyInstances = (params?: PaginationParams & { status?: string }) => {
  return get<PaginatedData<Instance>>('/api/v1/instances', params as Record<string, unknown>)
}

/** 实例详情 */
export const getInstance = (id: number) => {
  return get<Instance>(`/api/v1/instances/${id}`)
}

/** 删除实例 */
export const deleteInstance = (id: number) => {
  return del(`/api/v1/instances/${id}`)
}

// ============ 申请接口 ============

/** 我的申请列表 */
export const getMyApplications = (params?: PaginationParams & { status?: string }) => {
  return get<PaginatedData<Application>>('/api/v1/applications', params as Record<string, unknown>)
}

/** 申请详情 */
export const getApplication = (id: number) => {
  return get<Application>(`/api/v1/applications/${id}`)
}

/** 提交申请 */
export const createApplication = (data: CreateApplicationRequest) => {
  return post<Application>('/api/v1/applications', data)
}

/** 撤销申请 */
export const cancelApplication = (id: number) => {
  return del(`/api/v1/applications/${id}`)
}
