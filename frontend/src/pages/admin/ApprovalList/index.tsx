/**
 * 管理员审批列表页面
 */

import React, { useEffect, useState } from 'react'
import { Table, Button, Typography, Space, Tag, Modal, Form, Input, Tooltip, message } from 'antd'
import { CheckCircleOutlined, CloseCircleOutlined } from '@ant-design/icons'
import { getAdminApplications, approveApplication, rejectApplication } from '@/api/admin'
import { ApplicationStatusBadge } from '@/components/StatusBadge'
import type { Application } from '@/types'
import type { ColumnsType } from 'antd/es/table'
import { SPEC_CONFIGS } from '@/types'
import dayjs from 'dayjs'

const { Title } = Typography
const { TextArea } = Input

const ApprovalListPage: React.FC = () => {
  const [applications, setApplications] = useState<Application[]>([])
  const [loading, setLoading] = useState(true)
  const [rejectModalVisible, setRejectModalVisible] = useState(false)
  const [selectedApp, setSelectedApp] = useState<Application | null>(null)
  const [rejectForm] = Form.useForm()
  const [actionLoading, setActionLoading] = useState(false)

  const fetchApplications = async () => {
    setLoading(true)
    try {
      const res = await getAdminApplications({ status: 'pending', page: 1, size: 50 })
      if (res.data) setApplications(res.data.list)
    } catch {
      // 错误已在请求拦截器处理
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchApplications() }, [])

  const handleApprove = async (application: Application) => {
    setActionLoading(true)
    try {
      const res = await approveApplication(application.id)
      if (res.code === 200) {
        message.success('申请已批准，系统正在创建实例')
        fetchApplications()
      }
    } finally {
      setActionLoading(false)
    }
  }

  const handleRejectConfirm = async (values: { note: string }) => {
    if (!selectedApp) return
    setActionLoading(true)
    try {
      const res = await rejectApplication(selectedApp.id, values.note)
      if (res.code === 200) {
        message.success('申请已拒绝')
        setRejectModalVisible(false)
        rejectForm.resetFields()
        fetchApplications()
      }
    } finally {
      setActionLoading(false)
    }
  }

  const columns: ColumnsType<Application> = [
    {
      title: '申请人',
      dataIndex: ['user', 'username'],
      render: (_, record) => (
        <div>
          <div>{record.user?.username}</div>
          <Tag color="blue">{record.user?.department}</Tag>
        </div>
      ),
    },
    {
      title: '实例名称',
      dataIndex: 'instance_name',
    },
    {
      title: '规格',
      dataIndex: 'spec',
      render: (spec) => <Tag>{SPEC_CONFIGS[spec as keyof typeof SPEC_CONFIGS]?.label || spec}</Tag>,
    },
    {
      title: '申请理由',
      dataIndex: 'reason',
      ellipsis: true,
      width: 200,
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status) => <ApplicationStatusBadge status={status} />,
    },
    {
      title: '申请时间',
      dataIndex: 'created_at',
      render: (date) => date && dayjs(date).isValid() ? dayjs(date).format('MM-DD HH:mm') : '-',
    },
    {
      title: '操作',
      render: (_, record) => {
        if (record.status !== 'pending') return null
        return (
          <Space>
            <Tooltip title="批准申请">
              <Button
                type="primary"
                size="small"
                icon={<CheckCircleOutlined />}
                loading={actionLoading}
                onClick={() => handleApprove(record)}
              >
                批准
              </Button>
            </Tooltip>
            <Tooltip title="拒绝申请">
              <Button
                danger
                size="small"
                icon={<CloseCircleOutlined />}
                onClick={() => {
                  setSelectedApp(record)
                  setRejectModalVisible(true)
                }}
              >
                拒绝
              </Button>
            </Tooltip>
          </Space>
        )
      },
    },
  ]

  return (
    <div>
      <Title level={4}>审批管理</Title>

      <Table
        columns={columns}
        dataSource={applications}
        loading={loading}
        rowKey="id"
        expandable={{
          expandedRowRender: (record) => (
            <div style={{ padding: '8px 0' }}>
              <strong>完整申请理由：</strong>
              <p>{record.reason}</p>
            </div>
          ),
        }}
      />

      {/* 拒绝原因填写弹窗 */}
      <Modal
        title="填写拒绝原因"
        open={rejectModalVisible}
        onCancel={() => {
          setRejectModalVisible(false)
          rejectForm.resetFields()
        }}
        footer={null}
      >
        <Form form={rejectForm} onFinish={handleRejectConfirm}>
          <Form.Item
            name="note"
            rules={[{ required: true, message: '请填写拒绝原因' }]}
          >
            <TextArea rows={4} placeholder="请填写拒绝原因，会通知到申请人" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" danger htmlType="submit" loading={actionLoading}>
                确认拒绝
              </Button>
              <Button onClick={() => setRejectModalVisible(false)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default ApprovalListPage
