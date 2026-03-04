/**
 * 实例状态徽章组件
 * 根据实例状态显示对应的颜色和文字
 */

import React from 'react'
import { Badge, Tag } from 'antd'
import type { InstanceStatus, ApplicationStatus } from '@/types'
import { INSTANCE_STATUS_COLORS, INSTANCE_STATUS_TEXT, APPLICATION_STATUS_COLORS, APPLICATION_STATUS_TEXT } from '@/types'

interface InstanceStatusBadgeProps {
  status: InstanceStatus
  showDot?: boolean // 是否显示圆点动画（运行中时）
}

/** 实例状态徽章 */
export const InstanceStatusBadge: React.FC<InstanceStatusBadgeProps> = ({ status, showDot = false }) => {
  // 安全访问：未知 status 时提供兜底值，防止渲染 undefined
  const statusText = INSTANCE_STATUS_TEXT[status] || '未知状态'
  const statusColor = INSTANCE_STATUS_COLORS[status] || 'default'

  if (showDot && status === 'running') {
    return (
      <Badge
        status="processing"
        text={statusText}
        style={{ color: '#52c41a' }}
      />
    )
  }

  return (
    <Tag color={statusColor}>
      {statusText}
    </Tag>
  )
}

interface ApplicationStatusBadgeProps {
  status: ApplicationStatus
}

/** 申请状态徽章 */
export const ApplicationStatusBadge: React.FC<ApplicationStatusBadgeProps> = ({ status }) => {
  // 安全访问：未知 status 时提供兜底值，防止渲染 undefined
  const statusText = APPLICATION_STATUS_TEXT[status] || '未知状态'
  const statusColor = APPLICATION_STATUS_COLORS[status] || 'default'

  return (
    <Tag color={statusColor}>
      {statusText}
    </Tag>
  )
}

export default InstanceStatusBadge
