/**
 * 管理员监控大盘页面
 */

import React, { useEffect, useState } from 'react'
import { Row, Col, Card, Statistic, Typography, Spin } from 'antd'
import {
  CloudServerOutlined,
  CheckCircleOutlined,
  LoadingOutlined,
  WarningOutlined,
  UserOutlined,
  FileTextOutlined,
} from '@ant-design/icons'
import { getMetricsSummary } from '@/api/admin'
import type { MetricsSummary } from '@/types'

const { Title } = Typography

const MonitorDashboardPage: React.FC = () => {
  const [summary, setSummary] = useState<MetricsSummary | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const res = await getMetricsSummary()
        if (res.data) setSummary(res.data)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
    // 每 30 秒自动刷新
    const timer = setInterval(fetchData, 30000)
    return () => clearInterval(timer)
  }, [])

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <Title level={4}>监控大盘</Title>

      {/* 实例统计 */}
      <Title level={5} type="secondary">实例概况</Title>
      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="总实例数"
              value={summary?.instances.total || 0}
              prefix={<CloudServerOutlined />}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="运行中"
              value={summary?.instances.running || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="创建中"
              value={summary?.instances.creating || 0}
              prefix={<LoadingOutlined />}
              valueStyle={{ color: '#1677ff' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="异常实例"
              value={summary?.instances.failed || 0}
              prefix={<WarningOutlined />}
              valueStyle={{ color: (summary?.instances.failed || 0) > 0 ? '#ff4d4f' : '#666' }}
            />
          </Card>
        </Col>
      </Row>

      {/* 申请和用户统计 */}
      <Title level={5} type="secondary">申请与用户</Title>
      <Row gutter={16}>
        <Col xs={12} md={8}>
          <Card>
            <Statistic
              title="待审批申请"
              value={summary?.applications.pending || 0}
              prefix={<FileTextOutlined />}
              valueStyle={{ color: (summary?.applications.pending || 0) > 0 ? '#faad14' : '#666' }}
            />
          </Card>
        </Col>
        <Col xs={12} md={8}>
          <Card>
            <Statistic
              title="今日新申请"
              value={summary?.applications.today || 0}
              prefix={<FileTextOutlined />}
            />
          </Card>
        </Col>
        <Col xs={12} md={8}>
          <Card>
            <Statistic
              title="总用户数"
              value={summary?.users.total || 0}
              prefix={<UserOutlined />}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default MonitorDashboardPage
