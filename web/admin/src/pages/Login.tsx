import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { Form, Input, Button, Card, Typography, App as AntdApp } from 'antd'
import { LockOutlined, UserOutlined } from '@ant-design/icons'
import { useAuth } from '../hooks/auth'
import { useBrand } from '../hooks/brand'

export default function Login() {
  const { login, user } = useAuth()
  const nav = useNavigate()
  const loc = useLocation()
  const brand = useBrand()
  const { message } = AntdApp.useApp()
  const [busy, setBusy] = useState(false)

  if (user) { nav('/admin'); return null }

  async function submit(values: { username: string; password: string }) {
    setBusy(true)
    try {
      await login(values.username, values.password)
      const from = (loc.state as any)?.from?.pathname || '/admin'
      nav(from, { replace: true })
    } catch (e: any) {
      message.error(e.message || '登录失败')
    } finally { setBusy(false) }
  }

  return (
    <div style={{ minHeight: '100dvh', display: 'grid', placeItems: 'center', padding: 16, background: 'var(--bg)' }}>
      <Card style={{ width: '100%', maxWidth: 380 }}>
        <div style={{ marginBottom: 24, textAlign: 'center' }}>
          <Typography.Text mono style={{ color: 'var(--accent)', display: 'block', marginBottom: 6 }}>
            $ ./login.sh
          </Typography.Text>
          <Typography.Title level={4} style={{ margin: 0 }}>{brand} 管理后台</Typography.Title>
        </div>
        <Form onFinish={submit} size="large" autoComplete="off">
          <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<UserOutlined />} placeholder="用户名" autoFocus />
          </Form.Item>
          <Form.Item name="password" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit" block loading={busy}>
              登录
            </Button>
          </Form.Item>
          <Button type="link" block onClick={() => nav('/admin/forgot-password')} style={{ padding: 0 }}>
            忘记密码？
          </Button>
        </Form>
        <Typography.Text type="secondary" style={{ display: 'block', marginTop: 16, fontSize: 12, textAlign: 'center' }}>
          忘记密码时可通过账号安全页绑定的邮箱找回
        </Typography.Text>
      </Card>
    </div>
  )
}