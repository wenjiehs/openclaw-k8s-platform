/**
 * 认证状态管理 Store
 * 使用 Zustand 管理用户登录状态
 */

import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types'

interface AuthState {
  /** 当前登录用户 */
  user: User | null
  /** JWT Token */
  token: string | null
  /** 是否已登录 */
  isLoggedIn: boolean
  /**
   * persist 中间件是否已完成从 localStorage 恢复状态
   * 在 hydrated 为 true 之前，ProtectedRoute 不应做跳转判断，否则刷新页面会闪现登录页
   */
  hydrated: boolean

  /** 登录操作：保存 token 和用户信息 */
  login: (token: string, user: User) => void
  /** 登出操作：清除所有状态 */
  logout: () => void
  /** 更新用户信息 */
  updateUser: (user: User) => void
  /** 手动设置 hydrated（内部使用） */
  setHydrated: () => void
}

/** 认证状态 Store */
export const useAuthStore = create<AuthState>()(
  // persist 中间件：自动持久化到 localStorage
  persist(
    (set) => ({
      user: null,
      token: null,
      isLoggedIn: false,
      hydrated: false,

      login: (token: string, user: User) => {
        // 同步保存到 localStorage（兼容 axios 拦截器直接读取）
        try {
          localStorage.setItem('token', token)
          localStorage.setItem('user', JSON.stringify(user))
        } catch (error) {
          console.error('[AuthStore] Failed to persist auth to localStorage:', error)
        }
        set({ user, token, isLoggedIn: true })
      },

      logout: () => {
        try {
          localStorage.removeItem('token')
          localStorage.removeItem('user')
        } catch (error) {
          console.error('[AuthStore] Failed to clear auth from localStorage:', error)
        }
        set({ user: null, token: null, isLoggedIn: false })
      },

      updateUser: (user: User) => {
        try {
          localStorage.setItem('user', JSON.stringify(user))
        } catch (error) {
          console.error('[AuthStore] Failed to update user in localStorage:', error)
        }
        set({ user })
      },

      setHydrated: () => set({ hydrated: true }),
    }),
    {
      name: 'auth-storage', // localStorage key
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isLoggedIn: state.isLoggedIn,
        // hydrated 不持久化，每次启动重新设置
      }),
      // persist 完成后将 hydrated 置为 true
      onRehydrateStorage: () => (state) => {
        if (state) {
          state.setHydrated()
        }
      },
    }
  )
)
