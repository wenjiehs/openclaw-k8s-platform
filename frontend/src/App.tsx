/**
 * 应用根组件
 * 配置路由和全局状态
 */

import React from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider, Spin } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'

import { useAuthStore } from '@/store/auth'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import MainLayout from '@/components/Layout'
import LoginPage from '@/pages/Login'
import DashboardPage from '@/pages/Dashboard'
import ApplyInstancePage from '@/pages/ApplyInstance'
import MyInstancesPage from '@/pages/MyInstances'
import InstanceDetailPage from '@/pages/InstanceDetail'
import MyApplicationsPage from '@/pages/MyInstances/applications'
import ApprovalListPage from '@/pages/admin/ApprovalList'
import MonitorDashboardPage from '@/pages/admin/MonitorDashboard'
import InstanceManagementPage from '@/pages/admin/InstanceManagement'
import AuditLogPage from '@/pages/admin/AuditLog'

// 设置 dayjs 中文语言
dayjs.locale('zh-cn')

/** 加载中占位符 */
const LoadingFallback: React.FC = () => (
  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>
    <Spin size="large" tip="加载中..." />
  </div>
)

/** 需要认证才能访问的路由守卫 */
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isLoggedIn, hydrated } = useAuthStore()
  // 等待 persist 从 localStorage 恢复完毕，避免刷新页面时因初始值 false 闪现登录页
  if (!hydrated) {
    return <LoadingFallback />
  }
  if (!isLoggedIn) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

/** 管理员路由守卫 */
const AdminRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isLoggedIn, user, hydrated } = useAuthStore()
  // 等待 persist 从 localStorage 恢复完毕
  if (!hydrated) {
    return <LoadingFallback />
  }
  if (!isLoggedIn) {
    return <Navigate to="/login" replace />
  }
  if (!user || (user.role !== 'admin' && user.role !== 'super_admin')) {
    return <Navigate to="/dashboard" replace />
  }
  return <>{children}</>
}

const App: React.FC = () => {
  return (
    // ErrorBoundary 包裹整个应用，防止未捕获异常导致白屏
    <ErrorBoundary>
      {/* ConfigProvider：全局配置 Ant Design（中文语言包） */}
      <ConfigProvider locale={zhCN}>
        <BrowserRouter>
          <Routes>
            {/* 登录页 */}
            <Route path="/login" element={<LoginPage />} />

            {/* 需要认证的页面 */}
            <Route
              path="/"
              element={
                <ProtectedRoute>
                  <MainLayout />
                </ProtectedRoute>
              }
            >
              {/* 默认重定向到控制台 */}
              <Route index element={<Navigate to="/dashboard" replace />} />

              {/* 员工页面 */}
              <Route path="dashboard" element={<DashboardPage />} />
              <Route path="apply" element={<ApplyInstancePage />} />
              <Route path="instances" element={<MyInstancesPage />} />
              <Route path="instances/:id" element={<InstanceDetailPage />} />
              <Route path="applications" element={<MyApplicationsPage />} />

              {/* 管理员页面 */}
              <Route
                path="admin/approvals"
                element={
                  <AdminRoute>
                    <ApprovalListPage />
                  </AdminRoute>
                }
              />
              <Route
                path="admin/instances"
                element={
                  <AdminRoute>
                    <InstanceManagementPage />
                  </AdminRoute>
                }
              />
              <Route
                path="admin/monitor"
                element={
                  <AdminRoute>
                    <MonitorDashboardPage />
                  </AdminRoute>
                }
              />
              <Route
                path="admin/audit-logs"
                element={
                  <AdminRoute>
                    <AuditLogPage />
                  </AdminRoute>
                }
              />
            </Route>

            {/* 404 页面：已登录跳转控制台，未登录跳转登录页 */}
            <Route
              path="*"
              element={
                <ProtectedRoute>
                  <Navigate to="/dashboard" replace />
                </ProtectedRoute>
              }
            />
          </Routes>
        </BrowserRouter>
      </ConfigProvider>
    </ErrorBoundary>
  )
}

export default App
