/**
 * 认证 API 接口
 * 封装登录、获取用户信息等接口
 */

import { get, post } from '@/utils/request'
import type { LoginRequest, LoginResponse, User } from '@/types'

/** 用户登录 */
export const login = (data: LoginRequest) => {
  return post<LoginResponse>('/api/v1/auth/login', data)
}

/** 获取当前登录用户信息 */
export const getCurrentUser = () => {
  return get<User>('/api/v1/auth/me')
}
