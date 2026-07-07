import { useQuery } from '@tanstack/react-query'
import { Card, Typography, Tag, Row, Col } from 'antd'
import { ArrowRightOutlined } from '@ant-design/icons'
import { Link } from 'react-router-dom'
import { api } from '../api/client'

const SECTIONS: { key: string; label: string; desc: string }[] = [
  { key: 'branding', label: '品牌与页脚', desc: '站点名 / footer tag / 起始年 / commit hash / 底部徽章' },
  { key: 'nav', label: '导航', desc: '顶部导航链接，CTA 可选按钮和地址' },
  { key: 'hero', label: '首页主视觉', desc: 'eyebrow 终端命令 / 标题 / 副文 / 状态角标' },
  { key: 'stack', label: '技术栈', desc: '首页 “built with” 四个单元格' },
  { key: 'about', label: '关于页', desc: '个人介绍 / bio.yml / uptime' },
  { key: 'now', label: 'now.txt', desc: '关于页终端里展示的当前焦点信息' },
  { key: 'archive', label: '归档页', desc: '归档页标题 / 说明 / meta 信息' },
  { key: 'shelf', label: '书架页', desc: '学习资料、工具、课程等合集' },
  { key: 'footer', label: '页脚列', desc: 'browse / about 等页脚分组链接' },
]

export default function Settings() {
  const { data } = useQuery({
    queryKey: ['settings'],
    queryFn: () => api<{ keys: string[] }>('/api/settings'),
  })
  const known = new Set(data?.keys || [])

  return (
    <div style={{ padding: '32px 40px 64px' }}>
      <Typography.Title level={3} style={{ marginTop: 0 }}>站点内容分区</Typography.Title>
      <Typography.Paragraph type="secondary" style={{ marginBottom: 28 }}>
        每个分区对应前台某一块展示内容。点进去编辑 JSON，保存即时生效，前台无需重新部署。
      </Typography.Paragraph>

      <Row gutter={[16, 16]}>
        {SECTIONS.map(s => (
          <Col key={s.key} xs={24} sm={12}>
            <Link to={`/admin/settings/${s.key}`}>
              <Card hoverable size="small" styles={{ body: { padding: 16 } }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                  <Tag color={known.has(s.key) ? 'green' : 'default'} style={{ margin: 0 }}>
                    {known.has(s.key) ? '已配置' : '未配置'}
                  </Tag>
                  <Typography.Text mono type="secondary" style={{ fontSize: 12 }}>{s.key}</Typography.Text>
                  <ArrowRightOutlined style={{ marginLeft: 'auto', color: 'var(--muted-soft)' }} />
                </div>
                <Typography.Title level={5} style={{ margin: '10px 0 4px' }}>{s.label}</Typography.Title>
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>{s.desc}</Typography.Text>
              </Card>
            </Link>
          </Col>
        ))}
      </Row>
    </div>
  )
}