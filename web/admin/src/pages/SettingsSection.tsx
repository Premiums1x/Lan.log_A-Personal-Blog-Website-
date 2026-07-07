import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Card, Typography, Input, Button, Space, App as AntdApp, Breadcrumb, Alert } from 'antd'
import { SaveOutlined, RollbackOutlined } from '@ant-design/icons'
import { api } from '../api/client'

const HELP: Record<string, string> = {
  branding: '根级字段：brand, footer_tag, since_year(int), commit_hash, build_badge。',
  nav: '{"links":[{label,href}], "cta_label", "cta_href"}',
  hero: '{"eyebrow_cmd","title","title_accent","title_tail","sub","meta":[{k,v}],"corner":[{label,val}]}',
  stack: '{"cells":[{ic,title,desc}]}，ic 可用 01/02/...',
  about: '{"title","title_accent","intro":[字符串],"meta":[{k,v}],"bio_yml":{name,role,stack,writes,based,hosting},"uptime","body_md"}',
  now: '{"lines":[{is_cmd:true,f,args,c} 或 {is_cmd:false,arrow,k,v,is_string?,c?}]}',
  archive: '{"eyebrow_cmd","title","title_accent","intro","meta":[{k,v}]}',
  shelf: '{"eyebrow_cmd","title","title_accent","intro","meta":[{k,v}],"groups":[{title,eyebrow,desc,items:[{title,desc,meta,href,status,tags}]}]}',
  footer: '{"cols":[{h,links:[{label,href}]}]}',
}

export default function SettingsSection() {
  const { key = '' } = useParams()
  const nav = useNavigate()
  const qc = useQueryClient()
  const { message } = AntdApp.useApp()
  const { data, isLoading } = useQuery({
    queryKey: ['setting', key],
    queryFn: () => api<{ section_key: string; value: any }>(`/api/settings/${key}`),
  })
  const [text, setText] = useState('')

  useEffect(() => { if (data) setText(JSON.stringify(data.value, null, 2)) }, [data])

  const save = useMutation({
    mutationFn: () => {
      let value: any
      try { value = JSON.parse(text) } catch (e: any) { throw new Error('JSON 格式错误：' + e.message) }
      return api(`/api/settings/${key}`, { method: 'PUT', body: JSON.stringify({ value }) })
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['settings'] })
      qc.invalidateQueries({ queryKey: ['setting', key] })
      message.success('已保存，前台即时生效')
    },
    onError: (e: any) => message.error(e.message || '保存失败'),
  })

  return (
    <div style={{ padding: '24px 40px 64px' }}>
      <Breadcrumb
        style={{ marginBottom: 16 }}
        items={[
          { title: <a onClick={() => nav('/admin/settings')}>站点设置</a> },
          { title: key },
        ]}
      />
      <Typography.Title level={3} style={{ marginTop: 0 }}>
        <Typography.Text mono style={{ color: 'var(--accent)', fontSize: 16 }}>{key}</Typography.Text>
      </Typography.Title>

      {HELP[key] && (
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 20 }}
          message="该分区字段结构"
          description={<Typography.Text mono style={{ fontSize: 12, whiteSpace: 'pre-wrap' }}>{HELP[key]}</Typography.Text>}
        />
      )}

      <Card loading={isLoading} styles={{ body: { padding: 16 } }}>
        <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
          内容（JSON 格式）
        </Typography.Text>
        <Input.TextArea
          rows={20}
          value={text}
          onChange={e => setText(e.target.value)}
          spellCheck={false}
          style={{
            fontFamily: '"Geist Mono", ui-monospace, Menlo, Consolas, monospace',
            fontSize: 13,
          }}
        />
        <Space style={{ marginTop: 16 }}>
          <Button type="primary" icon={<SaveOutlined />} onClick={() => save.mutate()} loading={save.isPending}>
            保存
          </Button>
          <Button icon={<RollbackOutlined />} onClick={() => nav('/admin/settings')}>返回</Button>
        </Space>
        <Typography.Text type="secondary" style={{ display: 'block', marginTop: 16, fontSize: 12 }}>
          保存后即生效，前台服务端渲染会在下次请求时读到新值。
        </Typography.Text>
      </Card>
    </div>
  )
}


