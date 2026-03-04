/**
 * 主布局组件
 * 包含侧边栏导航、顶部栏和内容区域
 */

import React, { useState } from 'react'
import { Layout, Menu, Avatar, Dropdown, Typography, Space } from 'antd'
import {
  DashboardOutlined,
  PlusCircleOutlined,
  CloudServerOutlined,
  FileTextOutlined,
  CheckCircleOutlined,
  MonitorOutlined,
  SettingOutlined,
  LogoutOutlined,
  UserOutlined,
  AuditOutlined,
} from '@ant-design/icons'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/store/auth'
import type { MenuProps } from 'antd'

const { Header, Sider, Content } = Layout
const { Text } = Typography

/** 主布局组件 */
const MainLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()

  const isAdmin = user?.role === 'admin' || user?.role === 'super_admin'

  // 员工菜单
  const employeeMenuItems: MenuProps['items'] = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: '控制台',
    },
    {
      key: '/apply',
      icon: <PlusCircleOutlined />,
      label: '申请实例',
    },
    {
      key: '/instances',
      icon: <CloudServerOutlined />,
      label: '我的实例',
    },
    {
      key: '/applications',
      icon: <FileTextOutlined />,
      label: '我的申请',
    },
  ]

  // 管理员额外菜单
  const adminMenuItems: MenuProps['items'] = [
    {
      type: 'divider',
    },
    {
      key: 'admin',
      icon: <SettingOutlined />,
      label: '管理员功能',
      children: [
        {
          key: '/admin/approvals',
          icon: <CheckCircleOutlined />,
          label: '审批管理',
        },
        {
          key: '/admin/instances',
          icon: <CloudServerOutlined />,
          label: '实例管理',
        },
        {
          key: '/admin/monitor',
          icon: <MonitorOutlined />,
          label: '监控大盘',
        },
        {
          key: '/admin/audit-logs',
          icon: <AuditOutlined />,
          label: '审计日志',
        },
      ],
    },
  ]

  // 根据角色合并菜单
  const menuItems = isAdmin ? [...employeeMenuItems, ...adminMenuItems] : employeeMenuItems

  // 用户下拉菜单
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'info',
      icon: <UserOutlined />,
      label: (
        <div>
          <div>{user?.username || '用户'}</div>
          <Text type="secondary" style={{ fontSize: 12 }}>{user?.department || ''}</Text>
        </div>
      ),
      disabled: true,
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => {
        logout()
        navigate('/login')
      },
    },
  ]

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* 左侧边栏 */}
      <Sider
        collapsible
        collapsed={collapsed}
        onCollapse={setCollapsed}
        style={{
          background: '#001529',
          overflow: 'auto',
          height: '100vh',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
        }}
      >
        {/* Logo */}
        <div style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'rgba(255, 255, 255, 0.1)',
          margin: '8px 12px',
          borderRadius: 8,
        }}>
          <Text style={{ color: '#fff', fontWeight: 'bold', fontSize: collapsed ? 14 : 16 }}>
            {collapsed ? 'OC' : 'OpenClaw'}
          </Text>
        </div>

        {/* 导航菜单 */}
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          defaultOpenKeys={['admin']}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>

      {/* 主内容区域 */}
      <Layout style={{ marginLeft: collapsed ? 80 : 200, transition: 'all 0.2s' }}>
        {/* 顶部导航栏 */}
        <Header style={{
          background: '#fff',
          padding: '0 24px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'flex-end',
          boxShadow: '0 1px 4px rgba(0,21,41,.08)',
        }}>
          {/* 用户信息和下拉菜单 */}
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <Space style={{ cursor: 'pointer' }}>
              <Avatar icon={<UserOutlined />} />
              <span>{user?.username || '用户'}</span>
            </Space>
          </Dropdown>
        </Header>

        {/* 页面内容 */}
        <Content style={{
          margin: '24px 16px',
          padding: 24,
          background: '#f5f5f5',
          minHeight: 'calc(100vh - 64px)',
        }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}

export default MainLayout
