/**
 * 全局错误边界组件
 * 捕获子组件树中的 JS 运行时异常，防止整个应用白屏崩溃
 */

import React from 'react'
import { Button, Result } from 'antd'

interface Props {
  children: React.ReactNode
  fallback?: React.ReactNode
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error('[ErrorBoundary] Caught error:', error)
    console.error('[ErrorBoundary] Component stack:', info.componentStack)
  }

  handleReset = () => {
    this.setState({ hasError: false, error: undefined })
    window.location.href = '/login'
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }
      return (
        <Result
          status="500"
          title="应用发生错误"
          subTitle={this.state.error?.message || '页面遇到了一些问题，请返回重试'}
          extra={
            <Button type="primary" onClick={this.handleReset}>
              返回登录页
            </Button>
          }
        />
      )
    }
    return this.props.children
  }
}

export default ErrorBoundary
