import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router'
import { Form, Input, Button, Card, Typography, message } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import type { LoginRequest } from '@/types'

const { Title, Text } = Typography

export default function LoginPage() {
  const navigate = useNavigate()
  const { login, loading, user, restore } = useAuthStore()
  const [error, setError] = useState('')

  // 确保 store 从 localStorage 恢复（LoginPage 不走 AppLayout，需自行恢复）
  useEffect(() => {
    restore()
  }, [restore])

  // 已登录用户直接跳转
  useEffect(() => {
    if (user) {
      navigate('/dashboard', { replace: true })
    }
  }, [user, navigate])

  const onFinish = async (values: LoginRequest) => {
    setError('')
    try {
      await login(values)
      message.success('登录成功')
      navigate('/dashboard', { replace: true })
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : '登录失败')
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <Card className="w-96 shadow-md">
        <div className="text-center mb-6">
          <Title level={3} className="mb-2">教培物资管理系统</Title>
          <Text type="secondary">请使用您的账户登录</Text>
        </div>

        {error && (
          <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded text-red-600 text-sm">
            {error}
          </div>
        )}

        <Form
          name="login"
          onFinish={onFinish}
          autoComplete="off"
          size="large"
        >
          <Form.Item
            name="username"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}
