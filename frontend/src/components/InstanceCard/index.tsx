/**
 * 实例卡片组件
 * 展示单个 OpenClaw 实例的基本信息
 */

import React from 'react'
import { Card, Button, Space, Typography, Tooltip, Popconfirm } from 'antd'
import { LinkOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons'
import type { Instance } from '@/types'
import { SPEC_CONFIGS } from '@/types'
import { InstanceStatusBadge } from '@/components/StatusBadge'
import dayjs from 'dayjs'

const { Text, Paragraph } = Typography

interface InstanceCardProps {
  instance: Instance
  onDelete?: (instance: Instance) => void
  onRefresh?: (instance: Instance) => void
  showDeleteButton?: boolean
}

/** 实例卡片 */
const InstanceCard: React.FC<InstanceCardProps> = ({
  instance,
  onDelete,
  onRefresh,
  showDeleteButton = true,
}) => {
  // 安全访问 specConfig：规格 key 可能不在配置表中
  const specConfig = SPEC_CONFIGS[instance.spec] || null

  // 安全格式化日期：避免 dayjs(null/undefined) 渲染 "Invalid Date"
  const formatDate = (date?: string | null) =>
    date && dayjs(date).isValid() ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-'

  return (
    <Card
      hoverable
      title={
        <Space>
          <span>{instance.name}</span>
          <InstanceStatusBadge status={instance.status} showDot />
        </Space>
      }
      extra={
        <Space>
          {/* 访问实例按钮（仅运行中时可用） */}
          {instance.ingress_url && instance.status === 'running' && (
            <Tooltip title="访问实例">
              <Button
                type="primary"
                size="small"
                icon={<LinkOutlined />}
                onClick={() => window.open(instance.ingress_url, '_blank')}
              >
                访问
              </Button>
            </Tooltip>
          )}
          {/* 刷新状态 */}
          {onRefresh && (
            <Tooltip title="刷新状态">
              <Button
                size="small"
                icon={<ReloadOutlined />}
                onClick={() => onRefresh(instance)}
              />
            </Tooltip>
          )}
          {/* 删除按钮 */}
          {showDeleteButton && onDelete && (
            <Popconfirm
              title="确认删除"
              description={`确定要删除实例 "${instance.name}" 吗？删除后无法恢复。`}
              onConfirm={() => onDelete(instance)}
              okText="确认删除"
              cancelText="取消"
              okButtonProps={{ danger: true }}
            >
              <Tooltip title="删除实例">
                <Button
                  size="small"
                  danger
                  icon={<DeleteOutlined />}
                  disabled={instance.status === 'creating'}
                />
              </Tooltip>
            </Popconfirm>
          )}
        </Space>
      }
      style={{ marginBottom: 16 }}
    >
      {/* 实例基本信息 */}
      <Space direction="vertical" style={{ width: '100%' }}>
        <Space wrap>
          <Text type="secondary">规格：</Text>
          {specConfig ? (
            <Text>{specConfig.label}（{specConfig.cpu}，{specConfig.memory}）</Text>
          ) : (
            <Text>{instance.spec}</Text>
          )}
        </Space>

        {instance.ingress_url && (
          <Space>
            <Text type="secondary">访问地址：</Text>
            <Paragraph
              copyable={{ text: instance.ingress_url }}
              style={{ margin: 0 }}
            >
              <a href={instance.ingress_url} target="_blank" rel="noopener noreferrer">
                {instance.ingress_url}
              </a>
            </Paragraph>
          </Space>
        )}

        <Space wrap>
          <Text type="secondary">创建时间：</Text>
          <Text>{formatDate(instance.created_at)}</Text>
        </Space>

        {/* 临时实例显示到期时间 */}
        {instance.duration_type === 'temporary' && instance.expire_at && (
          <Space>
            <Text type="secondary">到期时间：</Text>
            <Text type={dayjs(instance.expire_at).isBefore(dayjs()) ? 'danger' : 'warning'}>
              {formatDate(instance.expire_at)}
            </Text>
          </Space>
        )}
      </Space>
    </Card>
  )
}

export default InstanceCard
