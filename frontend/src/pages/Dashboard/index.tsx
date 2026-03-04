/**
 * 控制台首页
 * 展示用户的实例状态概览
 */

import React, { useEffect, useState } from 'react'
import { Row, Col, Card, Statistic, Typography, Spin, Button } from 'antd'
import {
  CloudServerOutlined,
  CheckCircleOutlined,
  FileTextOutlined,
  PlusCircleOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { getMyInstances, getMyApplications } from '@/api/instances'
import { useAuthStore } from '@/store/auth'
import type { Instance, Application } from '@/types'

const { Title, Text } = Typography

const DashboardPage: React.FC = () => {
  const navigate = useNavigate()
  const { user } = useAuthStore()
  const [instances, setInstances] = useState<Instance[]>([])
  const [applications, setApplications] = useState<Application[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [instancesRes, applicationsRes] = await Promise.all([
          getMyInstances({ page: 1, size: 100 }),
          getMyApplications({ page: 1, size: 100 }),
        ])

        if (instancesRes.data) setInstances(instancesRes.data.list ?? [])
        if (applicationsRes.data) setApplications(applicationsRes.data.list ?? [])
      } catch {
        // 错误已在请求拦截器处理
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [])

  const runningInstances = instances.filter(i => i.status === 'running').length
  const pendingApplications = applications.filter(a => a.status === 'pending').length

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <div style={{ marginBottom: 24 }}>
        <Title level={4}>欢迎回来，{user?.username || '用户'} 👋</Title>
        <Text type="secondary">{user?.department || ''}</Text>
      </div>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="我的实例"
              value={instances.length}
              prefix={<CloudServerOutlined />}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="运行中"
              value={runningInstances}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="待审批申请"
              value={pendingApplications}
              prefix={<FileTextOutlined />}
              valueStyle={{ color: pendingApplications > 0 ? '#faad14' : '#666' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <div style={{ textAlign: 'center', padding: '8px 0' }}>
              <Button
                type="primary"
                icon={<PlusCircleOutlined />}
                size="large"
                onClick={() => navigate('/apply')}
                block
              >
                申请新实例
              </Button>
            </div>
          </Card>
        </Col>
      </Row>

      {/* 最近实例 */}
      {instances.length > 0 && (
        <Card
          title="最近的实例"
          extra={<Button type="link" onClick={() => navigate('/instances')}>查看全部</Button>}
        >
          {instances.slice(0, 3).map(instance => (
            <div key={instance.id} style={{
              padding: '12px 0',
              borderBottom: '1px solid #f0f0f0',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}>
              <div>
                <Text strong>{instance.name}</Text>
                <Text type="secondary" style={{ marginLeft: 8 }}>{instance.spec}</Text>
              </div>
              <Button
                type="link"
                size="small"
                onClick={() => navigate(`/instances/${instance.id}`)}
              >
                查看详情
              </Button>
            </div>
          ))}
        </Card>
      )}

      {/* 无实例时的引导 */}
      {instances.length === 0 && (
        <Card>
          <div style={{ textAlign: 'center', padding: '40px 0' }}>
            <CloudServerOutlined style={{ fontSize: 48, color: '#d9d9d9' }} />
            <Title level={5} type="secondary" style={{ marginTop: 16 }}>还没有实例</Title>
            <Text type="secondary">点击下方按钮申请你的第一个 OpenClaw 实例</Text>
            <div style={{ marginTop: 16 }}>
              <Button type="primary" icon={<PlusCircleOutlined />} onClick={() => navigate('/apply')}>
                申请实例
              </Button>
            </div>
          </div>
        </Card>
      )}
    </div>
  )
}

export default DashboardPage
