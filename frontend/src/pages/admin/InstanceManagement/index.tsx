/**
 * 管理员实例管理页面
 */

import React, { useEffect, useState } from 'react'
import { Table, Typography, Tag, Select, Space, Button, Popconfirm, message } from 'antd'
import { DeleteOutlined, LinkOutlined } from '@ant-design/icons'
import { getAdminInstances } from '@/api/admin'
import { del } from '@/utils/request'
import { InstanceStatusBadge } from '@/components/StatusBadge'
import type { Instance } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import { SPEC_CONFIGS } from '@/types'
import dayjs from 'dayjs'

const { Title } = Typography

const InstanceManagementPage: React.FC = () => {
  const [instances, setInstances] = useState<Instance[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [page, setPage] = useState(1)

  const fetchInstances = async () => {
    setLoading(true)
    try {
      const res = await getAdminInstances({
        page,
        size: 20,
        status: statusFilter || undefined,
      })
      if (res.data) {
        setInstances(res.data.list)
        setTotal(res.data.total)
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchInstances() }, [page, statusFilter])

  const handleDelete = async (instance: Instance) => {
    try {
      const res = await del(`/api/v1/instances/${instance.id}`)
      if (res.code === 200) {
        message.success('实例删除中')
        fetchInstances()
      }
    } catch {
      // 错误已在请求拦截器处理
    }
  }

  const columns: ColumnsType<Instance> = [
    {
      title: '实例名称',
      dataIndex: 'name',
    },
    {
      title: '所属用户',
      render: (_, record) => (
        <div>
          <div>{record.user?.username}</div>
          <Tag color="blue" style={{ fontSize: 11 }}>{record.user?.department}</Tag>
        </div>
      ),
    },
    {
      title: '规格',
      dataIndex: 'spec',
      render: (spec: keyof typeof SPEC_CONFIGS) => (
        <Tag>{SPEC_CONFIGS[spec]?.label || spec}</Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status) => <InstanceStatusBadge status={status} showDot />,
    },
    {
      title: 'Namespace',
      dataIndex: 'namespace',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      render: (date) => date && dayjs(date).isValid() ? dayjs(date).format('MM-DD HH:mm') : '-',
    },
    {
      title: '操作',
      render: (_, record) => (
        <Space>
          {record.ingress_url && record.status === 'running' && (
            <Button
              size="small"
              icon={<LinkOutlined />}
              onClick={() => window.open(record.ingress_url, '_blank')}
            >
              访问
            </Button>
          )}
          <Popconfirm
            title="确认删除"
            description="确定要强制删除此实例吗？"
            onConfirm={() => handleDelete(record)}
            okText="确认"
            cancelText="取消"
            okButtonProps={{ danger: true }}
          >
            <Button size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Title level={4}>实例管理</Title>

      {/* 过滤条件 */}
      <Space style={{ marginBottom: 16 }}>
        <Select
          placeholder="按状态过滤"
          allowClear
          style={{ width: 160 }}
          onChange={(value) => {
            setStatusFilter(value || '')
            setPage(1)
          }}
          options={[
            { label: '运行中', value: 'running' },
            { label: '创建中', value: 'creating' },
            { label: '已停止', value: 'stopped' },
            { label: '创建失败', value: 'failed' },
          ]}
        />
      </Space>

      <Table
        columns={columns}
        dataSource={instances}
        loading={loading}
        rowKey="id"
        pagination={{
          current: page,
          total,
          pageSize: 20,
          onChange: setPage,
        }}
      />
    </div>
  )
}

export default InstanceManagementPage
