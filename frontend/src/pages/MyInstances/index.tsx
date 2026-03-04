/**
 * 我的实例列表页面
 */

import React, { useEffect, useState } from 'react'
import { Typography, Spin, Empty, Row, Col, message } from 'antd'
import { getMyInstances, deleteInstance } from '@/api/instances'
import InstanceCard from '@/components/InstanceCard'
import type { Instance } from '@/types'

const { Title } = Typography

const MyInstancesPage: React.FC = () => {
  const [instances, setInstances] = useState<Instance[]>([])
  const [loading, setLoading] = useState(true)

  const fetchInstances = async () => {
    try {
      const res = await getMyInstances({ page: 1, size: 50 })
      if (res.data) setInstances(res.data.list)
    } catch {
      // 错误已在请求拦截器处理
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchInstances() }, [])

  const handleDelete = async (instance: Instance) => {
    try {
      const res = await deleteInstance(instance.id)
      if (res.code === 200) {
        message.success('实例删除中，K8s 资源将在后台清理')
        fetchInstances() // 刷新列表
      }
    } catch {
      // 错误已在请求拦截器处理
    }
  }

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px 0' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <Title level={4}>我的实例</Title>

      {instances.length === 0 ? (
        <Empty description="还没有实例" />
      ) : (
        <Row gutter={16}>
          {instances.map(instance => (
            <Col xs={24} md={12} xl={8} key={instance.id}>
              <InstanceCard
                instance={instance}
                onDelete={handleDelete}
                onRefresh={() => fetchInstances()}
              />
            </Col>
          ))}
        </Row>
      )}
    </div>
  )
}

export default MyInstancesPage
