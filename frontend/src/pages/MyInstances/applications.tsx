/**
 * 我的申请列表页面
 */

import React, { useEffect, useState } from 'react'
import { Table, Typography, Tag, Button, Popconfirm, message } from 'antd'
import { getMyApplications, cancelApplication } from '@/api/instances'
import { ApplicationStatusBadge } from '@/components/StatusBadge'
import type { Application } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import { SPEC_CONFIGS } from '@/types'
import dayjs from 'dayjs'

const { Title } = Typography

const MyApplicationsPage: React.FC = () => {
  const [applications, setApplications] = useState<Application[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)

  const fetchApplications = async () => {
    setLoading(true)
    try {
      const res = await getMyApplications({ page, size: 10 })
      if (res.data) {
        setApplications(res.data.list)
        setTotal(res.data.total)
      }
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchApplications() }, [page])

  const handleCancel = async (id: number) => {
    try {
      const res = await cancelApplication(id)
      if (res.code === 200) {
        message.success('申请已撤销')
        fetchApplications()
      }
    } catch {
      // 错误已在请求拦截器处理
    }
  }

  const columns: ColumnsType<Application> = [
    {
      title: '实例名称',
      dataIndex: 'instance_name',
    },
    {
      title: '规格',
      dataIndex: 'spec',
      render: (spec: keyof typeof SPEC_CONFIGS) => (
        <Tag>{SPEC_CONFIGS[spec]?.label || spec}</Tag>
      ),
    },
    {
      title: '申请状态',
      dataIndex: 'status',
      render: (status) => <ApplicationStatusBadge status={status} />,
    },
    {
      title: '申请理由',
      dataIndex: 'reason',
      ellipsis: true,
    },
    {
      title: '审批备注',
      dataIndex: 'approve_note',
      ellipsis: true,
    },
    {
      title: '申请时间',
      dataIndex: 'created_at',
      render: (date) => date && dayjs(date).isValid() ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: '操作',
      render: (_, record) => {
        if (record.status !== 'pending') return null
        return (
          <Popconfirm
            title="确认撤销"
            description="确定要撤销此申请吗？"
            onConfirm={() => handleCancel(record.id)}
            okText="确认撤销"
            cancelText="取消"
          >
            <Button size="small" danger>撤销</Button>
          </Popconfirm>
        )
      },
    },
  ]

  return (
    <div>
      <Title level={4}>我的申请</Title>

      <Table
        columns={columns}
        dataSource={applications}
        loading={loading}
        rowKey="id"
        pagination={{
          current: page,
          total,
          pageSize: 10,
          onChange: setPage,
        }}
        expandable={{
          expandedRowRender: (record) => (
            <div style={{ padding: '8px 0' }}>
              <strong>申请理由：</strong>
              <p>{record.reason}</p>
              {record.approve_note && (
                <>
                  <strong>审批备注：</strong>
                  <p>{record.approve_note}</p>
                </>
              )}
            </div>
          ),
        }}
      />
    </div>
  )
}

export default MyApplicationsPage
