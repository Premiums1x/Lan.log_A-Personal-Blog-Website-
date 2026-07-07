import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Table, Button, Input, Select, Space, Tag, Typography, Popconfirm, App as AntdApp } from 'antd'
import { PlusOutlined, SearchOutlined } from '@ant-design/icons'
import { Link, useNavigate } from 'react-router-dom'
import { api, type Post, datetime } from '../api/client'

export default function Posts() {
  const nav = useNavigate()
  const qc = useQueryClient()
  const { message } = AntdApp.useApp()
  const [q, setQ] = useState('')
  const [filter, setFilter] = useState<'all' | 'draft' | 'published'>('all')

  const { data, isLoading } = useQuery({
    queryKey: ['posts'],
    queryFn: () => api<{ items: Post[] }>('/api/posts'),
  })
  const items = (data?.items || []).filter(p =>
    (filter === 'all' || p.status === filter) &&
    (q.trim() === '' || p.title.toLowerCase().includes(q.toLowerCase()) || p.slug.includes(q))
  )

  const del = useMutation({
    mutationFn: (id: string) => api(`/api/posts/${id}`, { method: 'DELETE' }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['posts'] }); message.success('已删除') },
    onError: (e: any) => message.error(e.message || '删除失败'),
  })

  return (
    <div style={{ padding: '32px 40px 64px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
        <Typography.Title level={3} style={{ margin: 0 }}>文章</Typography.Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => nav('/admin/posts/new')}>新建文章</Button>
      </div>

      <Space style={{ marginBottom: 16 }}>
        <Input
          prefix={<SearchOutlined />}
          placeholder="搜索标题或 slug"
          allowClear
          value={q}
          onChange={e => setQ(e.target.value)}
          style={{ width: 280 }}
        />
        <Select
          value={filter}
          onChange={v => setFilter(v)}
          style={{ width: 120 }}
          options={[
            { value: 'all', label: '全部' },
            { value: 'published', label: '已发布' },
            { value: 'draft', label: '草稿' },
          ]}
        />
      </Space>

      <Table
        rowKey="id"
        loading={isLoading}
        dataSource={items}
        pagination={{ pageSize: 20, showSizeChanger: false, hideOnSinglePage: true }}
        columns={[
          {
            title: 'Commit',
            dataIndex: 'commit_hash',
            width: 90,
            render: h => <Typography.Text mono style={{ color: 'var(--accent)', fontSize: 12 }}>{(h || '').slice(0, 7)}</Typography.Text>,
          },
          {
            title: '标题',
            dataIndex: 'title',
            render: (t, p) => (
              <Link to={`/admin/posts/${p.id}/edit`}>
                <div>{t || '(未命名)'}</div>
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>/{p.slug} · {p.section}</Typography.Text>
              </Link>
            ),
          },
          {
            title: '状态',
            dataIndex: 'status',
            width: 90,
            render: s => <Tag color={s === 'published' ? 'green' : 'default'}>{s === 'published' ? '已发布' : '草稿'}</Tag>,
          },
          {
            title: '置顶',
            dataIndex: 'pinned',
            width: 70,
            render: p => p ? <Tag color="blue">置顶</Tag> : <span style={{ color: 'var(--muted-soft)' }}>—</span>,
          },
          {
            title: '更新时间',
            dataIndex: 'updated_at',
            width: 160,
            render: t => <Typography.Text type="secondary" style={{ fontSize: 12 }}>{datetime(t)}</Typography.Text>,
          },
          {
            title: '操作',
            width: 80,
            render: (_, p) => (
              <Popconfirm
                title="删除文章"
                description={`确定删除《${p.title}》？该操作不可撤销。`}
                onConfirm={() => del.mutate(p.id)}
                okText="删除"
                cancelText="取消"
                okButtonProps={{ danger: true }}
              >
                <Button type="link" danger size="small">删除</Button>
              </Popconfirm>
            ),
          },
        ]}
      />
    </div>
  )
}