import { useQuery } from '@tanstack/react-query'
import { Card, Col, Row, Statistic, Typography, List, Tag, Button, Space, Skeleton } from 'antd'
import { PlusOutlined, FileTextOutlined, SettingOutlined } from '@ant-design/icons'
import { Link, useNavigate } from 'react-router-dom'
import { api, type Post, datetime } from '../api/client'

export default function Dashboard() {
  const nav = useNavigate()
  const { data, isLoading } = useQuery({
    queryKey: ['posts'],
    queryFn: () => api<{ items: Post[]; total: number }>('/api/posts'),
  })
  const items = data?.items || []
  const published = items.filter(p => p.status === 'published')
  const drafts = items.filter(p => p.status === 'draft')
  const pinned = items.filter(p => p.pinned)

  return (
    <div style={{ padding: '32px 40px 64px' }}>
      <Typography.Title level={3} style={{ marginTop: 0 }}>概览</Typography.Title>

      <Row gutter={16} style={{ marginBottom: 28 }}>
        <Col span={8}>
          <Card><Statistic title="已发布" value={published.length} /></Card>
        </Col>
        <Col span={8}>
          <Card><Statistic title="草稿" value={drafts.length} /></Card>
        </Col>
        <Col span={8}>
          <Card><Statistic title="置顶" value={pinned.length} suffix="篇" /></Card>
        </Col>
      </Row>

      <Space style={{ marginBottom: 28 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => nav('/admin/posts/new')}>新建文章</Button>
        <Button icon={<FileTextOutlined />} onClick={() => nav('/admin/posts')}>管理文章</Button>
        <Button icon={<SettingOutlined />} onClick={() => nav('/admin/settings')}>站点设置</Button>
      </Space>

      <Typography.Title level={5}>最近文章</Typography.Title>
      <Card>
        {isLoading ? <Skeleton active /> : items.length === 0 ? (
          <Typography.Text type="secondary">还没有文章，点"新建文章"开始写第一篇。</Typography.Text>
        ) : (
          <List
            dataSource={items.slice(0, 6)}
            renderItem={p => (
              <List.Item
                actions={[
                  <Tag key="s" color={p.status === 'published' ? 'green' : 'default'}>
                    {p.status === 'published' ? '已发布' : '草稿'}
                  </Tag>,
                ].filter(Boolean) as any}
              >
                <List.Item.Meta
                  avatar={<Typography.Text mono type="secondary" style={{ fontSize: 12 }}>
                    {p.commit_hash?.slice(0, 7) || '·······'}
                  </Typography.Text>}
                  title={<Link to={`/admin/posts/${p.id}/edit`}>{p.title || '(未命名)'}</Link>}
                  description={<Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    /{p.slug} · {p.section} · {p.read_minutes} 分钟阅读 · {datetime(p.updated_at)}
                  </Typography.Text>}
                />
              </List.Item>
            )}
          />
        )}
      </Card>
    </div>
  )
}