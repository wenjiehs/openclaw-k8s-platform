/**
 * 管理员审计日志页面
 * 展示所有用户的操作记录，支持多维度过滤
 */

import React, { useEffect, useState, useCallback } from 'react'
import {
  Table,
  Card,
  Typography,
  Tag,
  Space,
  Select,
  DatePicker,
  Button,
  Row,
  Col,
  Tooltip,
} from 'antd'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { getAdminAuditLogs } from '@/api/admin'
import type { AuditLog } from '@/types'

const { Title } = Typography
const { RangePicker } = DatePicker
const { Option } = Select

// 操作类型选项
const ACTION_OPTIONS = [
  { value: 'login', label: '登录' },
  { value: 'logout', label: '退出' },
  { value: 'create', label: '创建' },
  { value: 'delete', label: '删除' },
  { value: 'approve', label: '审批通过' },
  { value: 'reject', label: '拒绝申请' },
  { value: 'cancel', label: '撤销' },
  { value: 'view', label: '查看' },
]

// 操作类型标签颜色
const ACTION_COLOR: Record<string, string> = {
  login: 'blue',
  logout: 'default',
  create: 'green',
  delete: 'red',
  approve: 'success',
  reject: 'error',
  cancel: 'warning',
  view: 'default',
}

const AuditLogPage: React.FC = () => {
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(20)

  // 筛选状态
  const [filterAction, setFilterAction] = useState<string | undefined>()
  const [filterResult, setFilterResult] = useState<string | undefined>()
  const [filterTimeRange, setFilterTimeRange] = useState<[string, string] | undefined>()

  const fetchLogs = useCallback(async (currentPage = page) => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = {
        page: currentPage,
        size: pageSize,
      }
      if (filterAction) params.action = filterAction
      if (filterResult) params.result = filterResult
      if (filterTimeRange) {
        params.start_time = filterTimeRange[0]
        params.end_time = filterTimeRange[1]
      }

      const res = await getAdminAuditLogs(params as Parameters<typeof getAdminAuditLogs>[0])
      if (res.data) {
        setLogs(res.data.list)
        setTotal(res.data.total)
      }
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, filterAction, filterResult, filterTimeRange])

  useEffect(() => {
    fetchLogs(page)
  }, [page, fetchLogs])

  // 点击搜索按钮时重置到第一页并搜索
  const handleSearch = () => {
    setPage(1)
    fetchLogs(1)
  }

  // 重置筛选条件
  const handleReset = () => {
    setFilterAction(undefined)
    setFilterResult(undefined)
    setFilterTimeRange(undefined)
    setPage(1)
    // 延迟执行，等状态更新后再 fetch
    setTimeout(() => fetchLogs(1), 0)
  }

  const columns: ColumnsType<AuditLog> = [
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (val: string) => (
        <span style={{ fontSize: 12 }}>
          {val && dayjs(val).isValid() ? dayjs(val).format('YYYY-MM-DD HH:mm:ss') : '-'}
        </span>
      ),
    },
    {
      title: '操作用户',
      dataIndex: 'user',
      key: 'user',
      width: 120,
      render: (user: AuditLog['user']) =>
        user ? (
          <span>
            <strong>{user.username}</strong>
            <br />
            <span style={{ fontSize: 11, color: '#999' }}>{user.department}</span>
          </span>
        ) : (
          <span style={{ color: '#ccc' }}>-</span>
        ),
    },
    {
      title: '操作类型',
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (action: string) => (
        <Tag color={ACTION_COLOR[action] || 'default'}>
          {ACTION_OPTIONS.find(o => o.value === action)?.label || action}
        </Tag>
      ),
    },
    {
      title: '资源类型',
      dataIndex: 'resource_type',
      key: 'resource_type',
      width: 100,
      render: (type: string) => <Tag>{type}</Tag>,
    },
    {
      title: '资源 ID',
      dataIndex: 'resource_id',
      key: 'resource_id',
      width: 100,
    },
    {
      title: '结果',
      dataIndex: 'result',
      key: 'result',
      width: 80,
      render: (result: string) => (
        <Tag color={result === 'success' ? 'success' : 'error'}>
          {result === 'success' ? '成功' : '失败'}
        </Tag>
      ),
    },
    {
      title: 'IP 地址',
      dataIndex: 'ip',
      key: 'ip',
      width: 130,
      render: (ip: string) => <span style={{ fontSize: 12 }}>{ip || '-'}</span>,
    },
    {
      title: '详情',
      dataIndex: 'extra',
      key: 'extra',
      ellipsis: true,
      render: (extra: string) =>
        extra ? (
          <Tooltip title={extra} placement="topLeft">
            <span style={{ fontSize: 12, color: '#666', cursor: 'pointer' }}>
              {extra.length > 30 ? extra.slice(0, 30) + '...' : extra}
            </span>
          </Tooltip>
        ) : (
          <span style={{ color: '#ccc' }}>-</span>
        ),
    },
  ]

  return (
    <div>
      <Title level={4} style={{ marginBottom: 16 }}>审计日志</Title>

      {/* 筛选区 */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={12} align="middle">
          <Col>
            <Select
              placeholder="操作类型"
              allowClear
              style={{ width: 130 }}
              value={filterAction}
              onChange={setFilterAction}
            >
              {ACTION_OPTIONS.map(o => (
                <Option key={o.value} value={o.value}>{o.label}</Option>
              ))}
            </Select>
          </Col>
          <Col>
            <Select
              placeholder="操作结果"
              allowClear
              style={{ width: 110 }}
              value={filterResult}
              onChange={setFilterResult}
            >
              <Option value="success">成功</Option>
              <Option value="failed">失败</Option>
            </Select>
          </Col>
          <Col>
            <RangePicker
              showTime
              format="YYYY-MM-DD HH:mm"
              onChange={(_, dateStrings) => {
                if (dateStrings[0] && dateStrings[1]) {
                  setFilterTimeRange([dateStrings[0], dateStrings[1]])
                } else {
                  setFilterTimeRange(undefined)
                }
              }}
            />
          </Col>
          <Col>
            <Space>
              <Button
                type="primary"
                icon={<SearchOutlined />}
                onClick={handleSearch}
              >
                搜索
              </Button>
              <Button onClick={handleReset}>重置</Button>
              <Button
                icon={<ReloadOutlined />}
                onClick={() => fetchLogs(page)}
                loading={loading}
              >
                刷新
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* 日志表格 */}
      <Card>
        <Table<AuditLog>
          rowKey="id"
          columns={columns}
          dataSource={logs}
          loading={loading}
          size="small"
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: false,
            showTotal: (t) => `共 ${t} 条记录`,
            onChange: (p) => setPage(p),
          }}
        />
      </Card>
    </div>
  )
}

export default AuditLogPage
