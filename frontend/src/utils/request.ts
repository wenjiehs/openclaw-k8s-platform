/**
 * HTTP 请求工具
 * 基于 axios 封装，统一处理认证 Token、错误处理和响应格式
 */

import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { message } from 'antd'
import type { ApiResponse } from '@/types'
import { useAuthStore } from '@/store/auth'

// 创建 axios 实例
const instance: AxiosInstance = axios.create({
  // 基础 URL：优先使用环境变量，开发时通过 vite proxy 转发
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 30000, // 30 秒超时
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器：自动附加 JWT Token
instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器：统一处理错误
instance.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    // 检查业务状态码：HTTP 200 但 code !== 200 表示业务处理失败
    // 本项目后端约定：code=200 成功，其他值为业务错误码
    const data = response.data
    if (data && typeof data === 'object' && 'code' in data && data.code !== 200) {
      const errMsg = (data as ApiResponse).message || '业务处理失败'
      message.error(errMsg)
      return Promise.reject(new Error(errMsg))
    }
    return response
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response

      switch (status) {
        case 401:
          // Token 过期或无效：同步清除 localStorage 和 Zustand store，跳转登录页
          try {
            localStorage.removeItem('token')
            localStorage.removeItem('user')
          } catch (_) { /* ignore */ }
          // 同步清空 store 状态（避免 isLoggedIn 仍为 true）
          useAuthStore.getState().logout()
          message.error('登录已过期，请重新登录')
          window.location.href = '/login'
          break

        case 403:
          message.error('权限不足，无法执行此操作')
          break

        case 404:
          message.error('请求的资源不存在')
          break

        case 500:
          message.error('服务器内部错误，请稍后重试')
          break

        default: {
          // data.message 可能为 undefined，兜底提示
          const defaultMsg = (data && typeof data === 'object' && data.message)
            ? data.message
            : `请求失败（${status}）`
          message.error(defaultMsg)
        }
      }
    } else if (error.code === 'ECONNABORTED') {
      message.error('请求超时，请检查网络连接')
    } else {
      message.error('网络连接失败，请检查网络')
    }

    return Promise.reject(error)
  }
)

// 封装请求方法

/** GET 请求 */
export const get = <T = unknown>(url: string, params?: Record<string, unknown>, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
  return instance.get<ApiResponse<T>>(url, { params, ...config }).then(res => res.data)
}

/** POST 请求 */
export const post = <T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
  return instance.post<ApiResponse<T>>(url, data, config).then(res => res.data)
}

/** PUT 请求 */
export const put = <T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
  return instance.put<ApiResponse<T>>(url, data, config).then(res => res.data)
}

/** DELETE 请求 */
export const del = <T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> => {
  return instance.delete<ApiResponse<T>>(url, config).then(res => res.data)
}

export default instance
