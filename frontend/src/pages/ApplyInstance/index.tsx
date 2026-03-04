/**
 * 申请实例页面
 */

import React, { useState } from 'react'
import { Form, Input, Radio, InputNumber, Button, Card, Typography, Alert, Row, Col, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import { SPEC_CONFIGS } from '@/types'
import type { CreateApplicationRequest } from '@/types'
import { post } from '@/utils/request'

const { Title, Text } = Typography
const { TextArea } = Input

const ApplyInstancePage: React.FC = () => {
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [durationType, setDurationType] = useState<'long' | 'temporary'>('long')

  const handleSubmit = async (values: CreateApplicationRequest) => {
    setLoading(true)
    try {
      const res = await post<{ id: number }>('/api/v1/applications', values)
      if (res.code === 200) {
        message.success('申请提交成功！管理员将在 1-2 个工作日内审批')
        navigate('/applications')
      } else {
        message.error(res.message || '提交失败')
      }
    } catch {
      // 错误已在请求拦截器处理
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ maxWidth: 720 }}>
      <Title level={4}>申请 OpenClaw 实例</Title>

      <Alert
        message="申请说明"
        description="填写申请表单后，管理员将在 1-2 个工作日内完成审批。审批通过后，系统将自动创建实例（约 5 分钟）。"
        type="info"
        showIcon
        style={{ marginBottom: 24 }}
      />

      <Card>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{ spec: 'standard', duration_type: 'long' }}
        >
          {/* 实例名称 */}
          <Form.Item
            label="实例名称"
            name="instance_name"
            rules={[
              { required: true, message: '请输入实例名称' },
              { min: 3, message: '实例名称至少 3 个字符' },
              { max: 20, message: '实例名称最多 20 个字符' },
              { pattern: /^[a-z0-9-]+$/, message: '只能包含小写字母、数字和连字符' },
            ]}
            extra="将用于 K8s Namespace 命名，建议使用你的英文名，如 zhangsan"
          >
            <Input placeholder="例如：zhangsan" />
          </Form.Item>

          {/* 规格选择 */}
          <Form.Item label="实例规格" name="spec" rules={[{ required: true }]}>
            <Radio.Group style={{ width: '100%' }}>
              <Row gutter={16}>
                {(Object.entries(SPEC_CONFIGS) as [keyof typeof SPEC_CONFIGS, typeof SPEC_CONFIGS[keyof typeof SPEC_CONFIGS]][]).map(([key, config]) => (
                  <Col span={8} key={key}>
                    <Radio.Button
                      value={key}
                      style={{ width: '100%', height: 'auto', padding: '12px 8px', textAlign: 'center' }}
                    >
                      <div>
                        <Text strong>{config.label}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 12 }}>{config.cpu}，{config.memory}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 11 }}>{config.description}</Text>
                      </div>
                    </Radio.Button>
                  </Col>
                ))}
              </Row>
            </Radio.Group>
          </Form.Item>

          {/* 使用时长 */}
          <Form.Item label="使用时长" name="duration_type" rules={[{ required: true }]}>
            <Radio.Group onChange={(e) => setDurationType(e.target.value)}>
              <Radio value="long">长期使用</Radio>
              <Radio value="temporary">临时使用</Radio>
            </Radio.Group>
          </Form.Item>

          {/* 临时使用天数 */}
          {durationType === 'temporary' && (
            <Form.Item
              label="使用天数"
              name="duration_days"
              rules={[
                { required: true, message: '请输入使用天数' },
                { type: 'number', min: 1, max: 90, message: '使用天数范围 1-90 天' },
              ]}
            >
              <InputNumber
                min={1}
                max={90}
                addonAfter="天"
                style={{ width: 200 }}
                placeholder="1-90"
              />
            </Form.Item>
          )}

          {/* 申请理由 */}
          <Form.Item
            label="申请理由"
            name="reason"
            rules={[
              { required: true, message: '请填写申请理由' },
              { min: 10, message: '申请理由至少 10 个字符' },
            ]}
            extra="请简要说明使用目的，例如：用于 TKE 集群日常运维和自动化脚本开发"
          >
            <TextArea
              rows={4}
              placeholder="请说明申请目的和使用场景..."
              maxLength={500}
              showCount
            />
          </Form.Item>

          {/* 提交按钮 */}
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} style={{ marginRight: 8 }}>
              提交申请
            </Button>
            <Button onClick={() => navigate(-1)}>取消</Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

export default ApplyInstancePage
