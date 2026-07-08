import { useEffect } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Alert, Button, Card, Divider, Form, Input, Space, Typography, App as AntdApp } from 'antd'
import { LockOutlined, MailOutlined, UserOutlined } from '@ant-design/icons'
import { api, datetime, type Account } from '../api/client'

export default function AccountSecurity() {
  const { message } = AntdApp.useApp()
  const qc = useQueryClient()
  const [emailForm] = Form.useForm<{ email: string }>()
  const [passwordForm] = Form.useForm<{ current_password: string; new_password: string; confirm_password: string }>()

  const account = useQuery({
    queryKey: ['account'],
    queryFn: () => api<Account>('/api/account'),
  })

  useEffect(() => {
    if (account.data) emailForm.setFieldsValue({ email: account.data.recovery_email || '' })
  }, [account.data, emailForm])

  const saveEmail = useMutation({
    mutationFn: (values: { email: string }) => api<Account>('/api/account/recovery-email', {
      method: 'PUT', body: JSON.stringify(values),
    }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['account'] })
      message.success('找回邮箱已保存')
    },
    onError: (e: any) => message.error(e.message || '保存失败'),
  })

  const savePassword = useMutation({
    mutationFn: (values: { current_password: string; new_password: string }) => api('/api/account/password', {
      method: 'PUT', body: JSON.stringify(values),
    }),
    onSuccess: () => {
      passwordForm.resetFields()
      message.success('密码已更新')
    },
    onError: (e: any) => message.error(e.message || '密码更新失败'),
  })

  return (
    <div style={{ padding: '32px 40px 64px', maxWidth: 760 }}>
      <Typography.Title level={3} style={{ marginTop: 0 }}>账号安全</Typography.Title>
      <Typography.Paragraph type="secondary">
        设置一个你能记住的密码，并绑定找回邮箱。忘记密码时可以在登录页用邮箱验证码重置。
      </Typography.Paragraph>

      <Card loading={account.isLoading}>
        <Space direction="vertical" size={4} style={{ width: '100%' }}>
          <Typography.Text type="secondary">当前账号</Typography.Text>
          <Typography.Title level={5} style={{ margin: 0 }}>
            <UserOutlined /> {account.data?.username || 'admin'}
          </Typography.Title>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            上次更新密码：{datetime(account.data?.password_updated_at)}
          </Typography.Text>
        </Space>

        <Divider />

        {!account.data?.has_recovery_email && (
          <Alert
            type="warning"
            showIcon
            style={{ marginBottom: 18 }}
            message="还没有设置找回邮箱"
            description="设置后，忘记密码时才能通过邮箱验证码找回。"
          />
        )}

        <Typography.Title level={5}>找回邮箱</Typography.Title>
        <Form form={emailForm} layout="vertical" onFinish={values => saveEmail.mutate(values)}>
          <Form.Item name="email" label="邮箱地址" rules={[{ required: true, message: '请输入邮箱地址' }, { type: 'email', message: '邮箱格式不正确' }]}>
            <Input prefix={<MailOutlined />} placeholder="you@example.com" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={saveEmail.isPending}>保存找回邮箱</Button>
        </Form>

        <Divider />

        <Typography.Title level={5}>修改密码</Typography.Title>
        <Form form={passwordForm} layout="vertical" onFinish={values => savePassword.mutate({ current_password: values.current_password, new_password: values.new_password })}>
          <Form.Item name="current_password" label="当前密码" rules={[{ required: true, message: '请输入当前密码' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="当前密码" />
          </Form.Item>
          <Form.Item name="new_password" label="新密码" rules={[{ required: true, message: '请输入新密码' }, { min: 8, message: '新密码至少需要 8 位' }]}>
            <Input.Password prefix={<LockOutlined />} placeholder="至少 8 位" />
          </Form.Item>
          <Form.Item name="confirm_password" label="确认新密码" dependencies={['new_password']} rules={[
            { required: true, message: '请再次输入新密码' },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('new_password') === value) return Promise.resolve()
                return Promise.reject(new Error('两次输入的新密码不一致'))
              },
            }),
          ]}>
            <Input.Password prefix={<LockOutlined />} placeholder="再次输入新密码" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={savePassword.isPending}>更新密码</Button>
        </Form>
      </Card>
    </div>
  )
}