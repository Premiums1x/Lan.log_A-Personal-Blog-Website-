import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button, Card, Form, Input, Typography, App as AntdApp } from 'antd'
import { LockOutlined, MailOutlined, UserOutlined } from '@ant-design/icons'
import { api } from '../api/client'
import { useBrand } from '../hooks/brand'

export default function ResetPassword() {
  const nav = useNavigate()
  const brand = useBrand()
  const { message } = AntdApp.useApp()
  const [username, setUsername] = useState('')
  const [requested, setRequested] = useState(false)
  const [requesting, setRequesting] = useState(false)
  const [saving, setSaving] = useState(false)

  async function requestCode(values: { username: string }) {
    setRequesting(true)
    try {
      const name = values.username.trim()
      const r = await api<{ message: string }>('/api/password-reset/request', {
        method: 'POST', body: JSON.stringify({ username: name }),
      })
      setUsername(name)
      setRequested(true)
      message.success(r.message || '如果账号已配置邮箱，验证码已发送')
    } catch (e: any) {
      message.error(e.message || '验证码发送失败')
    } finally {
      setRequesting(false)
    }
  }

  async function confirm(values: { code: string; new_password: string }) {
    setSaving(true)
    try {
      await api('/api/password-reset/confirm', {
        method: 'POST', body: JSON.stringify({ username, code: values.code, new_password: values.new_password }),
      })
      message.success('密码已重置，请重新登录')
      nav('/admin/login', { replace: true })
    } catch (e: any) {
      message.error(e.message || '密码重置失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div style={{ minHeight: '100dvh', display: 'grid', placeItems: 'center', padding: 16, background: 'var(--bg)' }}>
      <Card style={{ width: '100%', maxWidth: 420 }}>
        <div style={{ marginBottom: 24, textAlign: 'center' }}>
          <Typography.Text className="mono" style={{ color: 'var(--accent)', display: 'block', marginBottom: 6 }}>
            $ reset-password
          </Typography.Text>
          <Typography.Title level={4} style={{ margin: 0 }}>{brand} 密码找回</Typography.Title>
          <Typography.Paragraph type="secondary" style={{ margin: '10px 0 0', fontSize: 13 }}>
            验证码会发送到账号安全页里设置的找回邮箱。
          </Typography.Paragraph>
        </div>

        {!requested ? (
          <Form onFinish={requestCode} size="large" autoComplete="off">
            <Form.Item name="username" rules={[{ required: true, message: '请输入用户名' }]}>
              <Input prefix={<UserOutlined />} placeholder="用户名" autoFocus />
            </Form.Item>
            <Button type="primary" htmlType="submit" block loading={requesting}>发送验证码</Button>
          </Form>
        ) : (
          <Form onFinish={confirm} size="large" autoComplete="off">
            <Form.Item>
              <Input prefix={<UserOutlined />} value={username} disabled />
            </Form.Item>
            <Form.Item name="code" rules={[{ required: true, message: '请输入验证码' }, { len: 6, message: '验证码是 6 位数字' }]}>
              <Input prefix={<MailOutlined />} placeholder="邮箱验证码" inputMode="numeric" />
            </Form.Item>
            <Form.Item name="new_password" rules={[{ required: true, message: '请输入新密码' }, { min: 8, message: '新密码至少需要 8 位' }]}>
              <Input.Password prefix={<LockOutlined />} placeholder="新密码" />
            </Form.Item>
            <Form.Item name="confirm_password" dependencies={['new_password']} rules={[
              { required: true, message: '请再次输入新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('new_password') === value) return Promise.resolve()
                  return Promise.reject(new Error('两次输入的新密码不一致'))
                },
              }),
            ]}>
              <Input.Password prefix={<LockOutlined />} placeholder="确认新密码" />
            </Form.Item>
            <Button type="primary" htmlType="submit" block loading={saving}>重置密码</Button>
            <Button type="link" block onClick={() => setRequested(false)} style={{ marginTop: 8 }}>重新发送验证码</Button>
          </Form>
        )}

        <Typography.Text type="secondary" style={{ display: 'block', marginTop: 16, fontSize: 12, textAlign: 'center' }}>
          <Link to="/admin/login">返回登录</Link>
        </Typography.Text>
      </Card>
    </div>
  )
}