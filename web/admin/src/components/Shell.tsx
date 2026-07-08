import { Layout, Menu, Button, Typography } from 'antd'
import { DashboardOutlined, FileTextOutlined, SettingOutlined, LinkOutlined, LogoutOutlined, LockOutlined } from '@ant-design/icons'
import { NavLink, Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../hooks/auth'
import { useBrand } from '../hooks/brand'

const { Sider } = Layout

export default function Shell() {
  const { logout } = useAuth()
  const nav = useNavigate()
  const loc = useLocation()
  const brand = useBrand()

  const selectedKey = (() => {
    const p = loc.pathname
    if (p === '/admin') return '/admin'
    if (p.startsWith('/admin/posts')) return '/admin/posts'
    if (p.startsWith('/admin/account')) return '/admin/account'
    if (p.startsWith('/admin/settings')) return '/admin/settings'
    return '/admin'
  })()

  return (
    <Layout style={{ minHeight: '100dvh' }}>
      <Sider width={232} theme="light" style={{ borderRight: '1px solid var(--hair)' }}>
        <div style={{ padding: '20px 20px 8px' }}>
          <Typography.Text mono style={{ fontSize: 14 }}>
            <span style={{ color: 'var(--accent)' }}>~/</span>{brand}<span style={{ opacity: .5 }}>/admin</span>
          </Typography.Text>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          style={{ borderInlineEnd: 'none', marginTop: 8 }}
          onClick={({ key }) => nav(key)}
          items={[
            { key: '/admin', icon: <DashboardOutlined />, label: '概览' },
            { key: '/admin/posts', icon: <FileTextOutlined />, label: '文章' },
            { key: '/admin/settings', icon: <SettingOutlined />, label: '站点设置' },
            { key: '/admin/account', icon: <LockOutlined />, label: '账号安全' },
          ]}
        />
        <div style={{ position: 'absolute', bottom: 0, width: 232, padding: 16, display: 'flex', flexDirection: 'column', gap: 8 }}>
          <Button type="text" icon={<LinkOutlined />} onClick={() => window.open('/', '_blank')}
            style={{ textAlign: 'left', justifyContent: 'flex-start' }}>
            访问前台
          </Button>
          <Button type="text" danger icon={<LogoutOutlined />} onClick={() => { logout(); nav('/admin/login') }}
            style={{ textAlign: 'left', justifyContent: 'flex-start' }}>
            退出登录
          </Button>
        </div>
      </Sider>
      <Layout>
        <Outlet />
      </Layout>
    </Layout>
  )
}