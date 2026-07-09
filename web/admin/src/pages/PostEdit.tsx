import { useState, useEffect, useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Form, Input, Button, Select, Switch, Tag, Space, Card, Typography, Row, Col, App as AntdApp, Breadcrumb,
} from 'antd'
import { ArrowLeftOutlined, EyeOutlined, EditOutlined } from '@ant-design/icons'
import { api, type Post } from '../api/client'

type FormT = {
  slug: string; title: string; excerpt: string; body_md: string;
  cover_url: string; section: string; status: 'draft' | 'published';
  pinned: boolean; tag_names: string[];
}

export default function PostEdit() {
  const { id } = useParams()
  const isEdit = !!id
  const nav = useNavigate()
  const qc = useQueryClient()
  const { message } = AntdApp.useApp()

  const { data } = useQuery({
    queryKey: ['post', id],
    queryFn: () => api<Post>(`/api/posts/${id}`),
    enabled: isEdit,
  })

  const [f, setF] = useState<FormT>({
    slug: '', title: '', excerpt: '', body_md: '', cover_url: '',
    section: 'posts', status: 'draft', pinned: false, tag_names: [],
  })
  const [tagInput, setTagInput] = useState('')
  const [showPreview, setShowPreview] = useState(false)

  useEffect(() => {
    if (data) setF({
      slug: data.slug, title: data.title, excerpt: data.excerpt, body_md: data.body_md,
      cover_url: data.cover_url, section: data.section, status: data.status,
      pinned: data.pinned, tag_names: (data.tags || []).map(t => t.name),
    })
  }, [data])

  useEffect(() => {
    if (!isEdit && !f.slug) set('slug', slugify(f.title || 'untitled'))
  }, [f.title, isEdit]) // eslint-disable-line

  const save = useMutation({
    mutationFn: async (publish: boolean) => {
      const body = JSON.stringify({ ...f, status: publish ? 'published' : f.status })
      if (isEdit) return api<Post>(`/api/posts/${id}`, { method: 'PUT', body })
      return api<Post>('/api/posts', { method: 'POST', body })
    },
    onSuccess: (_, publish) => {
      qc.invalidateQueries({ queryKey: ['posts'] })
      message.success(publish ? '已发布' : '已保存草稿')
      nav('/admin/posts')
    },
    onError: (e: any) => message.error(e.message || '保存失败'),
  })

  function set<K extends keyof FormT>(k: K, v: FormT[K]) { setF(s => ({ ...s, [k]: v })) }

  function addTag() {
    const t = tagInput.trim()
    if (t && !f.tag_names.includes(t)) set('tag_names', [...f.tag_names, t])
    setTagInput('')
  }
  function rmTag(t: string) { set('tag_names', f.tag_names.filter(x => x !== t)) }

  const wordCount = useMemo(
    () => f.body_md.replace(/\s+/g, ' ').split(' ').filter(Boolean).length,
    [f.body_md]
  )

  return (
    <div style={{ padding: '24px 40px 64px' }}>
      <Breadcrumb
        style={{ marginBottom: 16 }}
        items={[
          { title: <a onClick={() => nav('/admin/posts')}>文章</a> },
          { title: isEdit ? '编辑' : '新建' },
        ]}
      />
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
        <Typography.Title level={3} style={{ margin: 0 }}>{isEdit ? '编辑文章' : '新建文章'}</Typography.Title>
        <Space>
          <Button
            icon={showPreview ? <EditOutlined /> : <EyeOutlined />}
            onClick={() => setShowPreview(s => !s)}
          >
            {showPreview ? '编辑' : '预览'}
          </Button>
          <Button onClick={() => save.mutate(false)} loading={save.isPending}>存为草稿</Button>
          <Button type="primary" onClick={() => save.mutate(true)} loading={save.isPending}>
            {f.status === 'published' ? '更新并发布' : '发布'}
          </Button>
        </Space>
      </div>

      <Form layout="vertical">
        <Form.Item label="标题">
          <Input value={f.title} onChange={e => set('title', e.target.value)} placeholder="文章标题" autoFocus size="large" />
        </Form.Item>

        <Row gutter={16}>
          <Col span={12}>
            <Form.Item label="Slug (URL)"><Input className="mono" value={f.slug} onChange={e => set('slug', e.target.value)} placeholder="enough-is-enough" /></Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item label="分区"><Input value={f.section} onChange={e => set('section', e.target.value)} placeholder="posts / 随笔 / 工具" /></Form.Item>
          </Col>
        </Row>

        <Form.Item label="摘要（留空自动取正文开头）">
          <Input.TextArea rows={2} value={f.excerpt} onChange={e => set('excerpt', e.target.value)} placeholder="一句话摘要" />
        </Form.Item>

        <Form.Item label={`正文 (Markdown) · 约 ${wordCount} 字`}>
          {showPreview ? (
            <Card style={{ minHeight: 420 }}>
              <div className="prose-content" dangerouslySetInnerHTML={{ __html: previewHtml(f.body_md) }} />
            </Card>
          ) : (
            <Input.TextArea
              rows={18}
              value={f.body_md}
              onChange={e => set('body_md', e.target.value)}
              placeholder={'# 标题\n\n正文。支持 **粗体**、`code`、列表、> 引用、```代码块```。'}
              style={{ fontFamily: '"Geist Mono", ui-monospace, Menlo, Consolas, monospace', fontSize: 13 }}
            />
          )}
        </Form.Item>

        <Form.Item label="封面 URL（可选）">
          <Input value={f.cover_url} onChange={e => set('cover_url', e.target.value)} placeholder="https://…" />
        </Form.Item>

        <Form.Item label="标签">
          <Space wrap style={{ marginBottom: 8 }}>
            {f.tag_names.map(t => (
              <Tag key={t} closable onClose={() => rmTag(t)} color="blue">#{t}</Tag>
            ))}
          </Space>
          <Input
            value={tagInput}
            onChange={e => setTagInput(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter' || e.key === ',') { e.preventDefault(); addTag() } }}
            placeholder="输入标签后回车添加"
          />
        </Form.Item>

        <Form.Item label="置顶（显示在首页 PINNED 位）">
          <Switch checked={f.pinned} onChange={v => set('pinned', v)} />
        </Form.Item>
      </Form>
    </div>
  )
}

function slugify(s: string) {
  return s.toLowerCase().trim()
    .replace(/[^\w\u4e00-\u9fff]+/g, '-')
    .replace(/^-+|-+$/g, '') || 'untitled'
}

function previewHtml(md: string): string {
  let h = md
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    .replace(/```([\s\S]*?)```/g, (_, c) => `<pre><code>${c}</code></pre>`)
    .replace(/^### (.*)$/gm, '<h3>$1</h3>')
    .replace(/^## (.*)$/gm, '<h2>$1</h2>')
    .replace(/^# (.*)$/gm, '<h1>$1</h1>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/^> (.*)$/gm, '<blockquote>$1</blockquote>')
    .replace(/^---$/gm, '<hr/>')
    .replace(/\n\n/g, '</p><p>')
  return `<p>${h}</p>`
}