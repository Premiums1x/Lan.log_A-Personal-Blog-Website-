import { useState, useEffect, useMemo } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Form, Input, Button, Switch, Tag, Space, Card, Typography, Row, Col,
  App as AntdApp, Breadcrumb, Modal, Alert,
} from 'antd'
import { api, type Post } from '../api/client'
import MDEditor from '@uiw/react-md-editor'

type ExcerptSource = 'manual' | 'ai' | 'empty'

type FormT = {
  slug: string; title: string; excerpt: string; excerpt_source: ExcerptSource; body_md: string;
  cover_url: string; section: string; status: 'draft' | 'published';
  pinned: boolean; tag_names: string[];
}

type SaveIntent = {
  publish: boolean
  form?: FormT
  excerptReviewed?: boolean
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
    slug: '', title: '', excerpt: '', excerpt_source: 'empty', body_md: '', cover_url: '',
    section: 'posts', status: 'draft', pinned: false, tag_names: [],
  })
  const [tagInput, setTagInput] = useState('')

  const [excerptModalOpen, setExcerptModalOpen] = useState(false)
  const [generatedExcerpt, setGeneratedExcerpt] = useState<string | null>(null)
  const [aiLoading, setAILoading] = useState(false)

  useEffect(() => {
    if (data) setF({
      slug: data.slug, title: data.title, excerpt: data.excerpt,
      excerpt_source: data.excerpt_source || (data.excerpt ? 'manual' : 'empty'),
      body_md: data.body_md, cover_url: data.cover_url, section: data.section, status: data.status,
      pinned: data.pinned, tag_names: (data.tags || []).map(t => t.name),
    })
  }, [data])

  useEffect(() => {
    if (!isEdit && !f.slug) set('slug', slugify(f.title || 'untitled'))
  }, [f.title, isEdit]) // eslint-disable-line

  const save = useMutation({
    mutationFn: async (intent: SaveIntent) => {
      const form = intent.form || f
      const body = JSON.stringify({
        ...form,
        status: intent.publish ? 'published' : form.status,
        excerpt_reviewed: !!intent.excerptReviewed,
      })
      if (isEdit) return api<Post>(`/api/posts/${id}`, { method: 'PUT', body })
      return api<Post>('/api/posts', { method: 'POST', body })
    },
    onSuccess: (_, intent) => {
      qc.invalidateQueries({ queryKey: ['posts'] })
      message.success(intent.publish ? '已发布' : '已保存草稿')
      nav('/admin/posts')
    },
    onError: (e: any) => message.error(e.message || '保存失败'),
  })

  function set<K extends keyof FormT>(k: K, v: FormT[K]) { setF(s => ({ ...s, [k]: v })) }

  function setExcerpt(value: string) {
    setF(s => ({ ...s, excerpt: value, excerpt_source: value.trim() ? 'manual' : 'empty' }))
  }

  function addTag() {
    const t = tagInput.trim()
    if (t && !f.tag_names.includes(t)) set('tag_names', [...f.tag_names, t])
    setTagInput('')
  }
  function rmTag(t: string) { set('tag_names', f.tag_names.filter(x => x !== t)) }

  function needsExcerptReview() {
    if (!isEdit) return !f.excerpt.trim()
    const bodyChanged = !!data && f.body_md !== data.body_md
    return bodyChanged || !!data?.excerpt_stale
  }

  function startPublish() {
    if (needsExcerptReview()) {
      setGeneratedExcerpt(null)
      setExcerptModalOpen(true)
      return
    }
    save.mutate({ publish: true, excerptReviewed: true })
  }

  function publishWith(form: FormT) {
    setExcerptModalOpen(false)
    setGeneratedExcerpt(null)
    save.mutate({ publish: true, form, excerptReviewed: true })
  }

  async function generateLatestExcerpt() {
    setAILoading(true)
    try {
      const result = await api<{ excerpt: string }>('/api/ai/excerpt', {
        method: 'POST',
        body: JSON.stringify({ title: f.title, body_md: f.body_md }),
      })
      setGeneratedExcerpt(result.excerpt)
    } catch (e: any) {
      message.error(e.message || 'AI 摘要生成失败')
    } finally {
      setAILoading(false)
    }
  }

  const currentExcerptSource: ExcerptSource = f.excerpt.trim()
    ? (f.excerpt_source === 'ai' ? 'ai' : 'manual')
    : 'empty'

  const wordCount = useMemo(() => {
    const text = f.body_md || ''
    const chineseChars = text.match(/[\u4e00-\u9fa5]/g) || []
    const englishWords = text.replace(/[\u4e00-\u9fa5]/g, ' ').split(/\s+/).filter(Boolean)
    return chineseChars.length + englishWords.length
  }, [f.body_md])

  const isBodyReview = isEdit && (!!data?.excerpt_stale || f.body_md !== data?.body_md)

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
          <Button onClick={() => save.mutate({ publish: false })} loading={save.isPending}>存为草稿</Button>
          <Button type="primary" onClick={startPublish} loading={save.isPending}>
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

        <Form.Item label="摘要（可选）" extra={f.excerpt_source === 'ai' && f.excerpt ? '当前摘要由 AI 生成，仍可手动修改。' : undefined}>
          <Input.TextArea rows={2} value={f.excerpt} onChange={e => setExcerpt(e.target.value)} placeholder="留空则不展示" />
        </Form.Item>

        <Form.Item label={`正文 (Markdown) · 约 ${wordCount} 字`}>
          <div data-color-mode="light">
            <MDEditor
              value={f.body_md}
              onChange={val => set('body_md', val || '')}
              height={500}
            />
          </div>
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

      <Modal
        open={excerptModalOpen}
        title={generatedExcerpt === null ? '发布前确认摘要' : '确认 AI 生成的摘要'}
        onCancel={() => { if (!aiLoading) setExcerptModalOpen(false) }}
        closable={!aiLoading}
        maskClosable={false}
        footer={generatedExcerpt === null ? [
          <Button key="cancel" onClick={() => setExcerptModalOpen(false)} disabled={aiLoading}>取消</Button>,
          <Button key="clear" onClick={() => publishWith({ ...f, excerpt: '', excerpt_source: 'empty' })} disabled={aiLoading}>清空摘要</Button>,
          <Button key="keep" onClick={() => publishWith({ ...f, excerpt_source: currentExcerptSource })} disabled={aiLoading}>保留当前摘要</Button>,
          <Button key="ai" type="primary" onClick={generateLatestExcerpt} loading={aiLoading}>AI 生成最新摘要</Button>,
        ] : [
          <Button key="back" onClick={() => setGeneratedExcerpt(null)}>返回选择</Button>,
          <Button
            key="publish"
            type="primary"
            onClick={() => publishWith({ ...f, excerpt: generatedExcerpt, excerpt_source: 'ai' })}
            disabled={!generatedExcerpt.trim()}
          >
            使用并发布
          </Button>,
        ]}
      >
        {generatedExcerpt === null ? (
          <Space direction="vertical" size={16} style={{ width: '100%' }}>
            <Alert
              type="info"
              showIcon
              message={isBodyReview ? '正文已修改，当前摘要可能已经过期' : '摘要为空，发布后标题下方将不展示摘要'}
              description="你可以生成与最新正文匹配的摘要，也可以保留、清空或不使用 AI。"
            />
            {f.excerpt.trim() && (
              <div>
                <Typography.Text type="secondary">当前摘要</Typography.Text>
                <Typography.Paragraph style={{ margin: '6px 0 0', whiteSpace: 'pre-wrap' }}>{f.excerpt}</Typography.Paragraph>
              </div>
            )}
          </Space>
        ) : (
          <Space direction="vertical" size={14} style={{ width: '100%' }}>
            {f.excerpt.trim() && (
              <div>
                <Typography.Text type="secondary">原摘要</Typography.Text>
                <Typography.Paragraph style={{ margin: '6px 0 0', color: 'var(--muted)' }}>{f.excerpt}</Typography.Paragraph>
              </div>
            )}
            <div>
              <Typography.Text strong>最新摘要（可编辑）</Typography.Text>
              <Input.TextArea
                style={{ marginTop: 8 }}
                rows={4}
                value={generatedExcerpt}
                onChange={e => setGeneratedExcerpt(e.target.value)}
              />
            </div>
          </Space>
        )}
      </Modal>
    </div>
  )
}

function slugify(s: string) {
  return s.toLowerCase().trim()
    .replace(/[^\w\u4e00-\u9fff]+/g, '-')
    .replace(/^-+|-+$/g, '') || 'untitled'
}
