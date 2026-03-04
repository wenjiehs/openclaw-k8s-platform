/**
 * 实例详情页面
 */

import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Descriptions, Card, Button, Spin, Typography, Space, Popconfirm, message } from 'antd'
import { ArrowLeftOutlined, LinkOutlined, DeleteOutlined } from '@ant-design/icons'
import { getInstance, deleteInstance } from '@/api/instances'
import { InstanceStatusBadge } from '@/components/StatusBadge'
import type { Instance } from '@/types'
import { SPEC_CONFIGS } from '@/types'
import dayjs from 'dayjs'

const { Title } = Typography

const InstanceDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [instance, setInstance] = useState<Instance | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!id) return
    getInstance(Number(id))
      .then(res => {
        if (res.data) setInstance(res.data)
      })
      .catch(() => {
        // 错误已在请求拦截器处理，这里确保 loading 在失败时也清除
      })
      .finally(() => setLoading(false))
  }, [id])

  const handleDelete = async () => {
    if (!instance) return
    try {
      const res = await deleteInstance(instance.id)
      if (res.code === 200) {
        message.success('实例删除中')
        navigate('/instances')
      }
    } catch {
      // 错误已在请求拦截器处理
    }
  }

  if (loading) {
    return <Spin size="large" style={{ display: 'block', textAlign: 'center', marginTop: 100 }} />
  }

  if (!instance) {
    return <div>实例不存在</div>
  }

  // 安全访问 specConfig：规格 key 可能不在配置表中
  const specConfig = SPEC_CONFIGS[instance.spec] || null
  // 安全格式化日期
  const formatDate = (date?: string | null) =>
    date && dayjs(date).isValid() ? dayjs(date).format('YYYY-MM-DD HH:mm:ss') : '-'

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 24 }}>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(-1)}>返回</Button>
          <Title level={4} style={{ margin: 0 }}>{instance.name}</Title>
          <InstanceStatusBadge status={instance.status} showDot />
        </Space>
        <Space>
          {instance.ingress_url && instance.status === 'running' && (
            <Button
              type="primary"
              icon={<LinkOutlined />}
              onClick={() => window.open(instance.ingress_url, '_blank')}
            >
              访问实例
            </Button>
          )}
          <Popconfirm
            title="确认删除"
            description="确定要删除此实例吗？删除后无法恢复。"
            onConfirm={handleDelete}
            okText="确认删除"
            cancelText="取消"
            okButtonProps={{ danger: true }}
          >
            <Button danger icon={<DeleteOutlined />}>删除实例</Button>
          </Popconfirm>
        </Space>
      </div>

      <Card>
        <Descriptions title="实例信息" bordered column={2}>
          <Descriptions.Item label="实例名称">{instance.name}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <InstanceStatusBadge status={instance.status} showDot />
          </Descriptions.Item>
          <Descriptions.Item label="规格">
            {specConfig
              ? `${specConfig.label}（${specConfig.cpu}，${specConfig.memory}）`
              : instance.spec}
          </Descriptions.Item>
          <Descriptions.Item label="K8s Namespace">{instance.namespace}</Descriptions.Item>
          <Descriptions.Item label="访问地址" span={2}>
            {instance.ingress_url ? (
              <a href={instance.ingress_url} target="_blank" rel="noopener noreferrer">
                {instance.ingress_url}
              </a>
            ) : '等待分配'}
          </Descriptions.Item>
          <Descriptions.Item label="使用时长">
            {instance.duration_type === 'long' ? '长期' : '临时'}
          </Descriptions.Item>
          {instance.expire_at && (
            <Descriptions.Item label="到期时间">
              {formatDate(instance.expire_at)}
            </Descriptions.Item>
          )}
          <Descriptions.Item label="创建时间">
            {formatDate(instance.created_at)}
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </div>
  )
}

export default InstanceDetailPage
