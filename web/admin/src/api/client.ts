const TOKEN_KEY = 'xu_admin_token'

export const tokenStore = {
  get(): string | null { return localStorage.getItem(TOKEN_KEY) },
  set(t: string) { localStorage.setItem(TOKEN_KEY, t) },
  clear() { localStorage.removeItem(TOKEN_KEY) },
}

export async function api<T = any>(
  path: string,
  opts: RequestInit = {},
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(opts.headers as any),
  }
  const tok = tokenStore.get()
  if (tok) headers.Authorization = 'Bearer ' + tok
  const res = await fetch(path, { ...opts, headers })
  const text = await res.text()
  let data: any = null
  if (text) {
    try { data = JSON.parse(text) } catch { data = text }
  }
  if (!res.ok) {
    const msg = (data && data.error) || res.statusText || 'request failed'
    const err = new Error(typeof msg === 'string' ? msg : JSON.stringify(msg))
    ;(err as any).status = res.status
    throw err
  }
  return data as T
}

export type Post = {
  id: string
  slug: string
  title: string
  excerpt: string
  body_md: string
  body_html: string
  cover_url: string
  section: string
  status: 'draft' | 'published'
  commit_hash: string
  read_minutes: number
  words: number
  pinned: boolean
  published_at: string | null
  created_at: string
  updated_at: string
  tags: { id: string; slug: string; name: string }[]
}

export type User = { id: string; username: string; display_name?: string }

export type Account = {
  id: string
  username: string
  display_name?: string
  recovery_email: string
  has_recovery_email: boolean
  password_updated_at: string
}

// Format an ISO timestamp into a readable Chinese date string.
export function datetime(iso?: string | null): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (isNaN(d.getTime())) return iso
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}