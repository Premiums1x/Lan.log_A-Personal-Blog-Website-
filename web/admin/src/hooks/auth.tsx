import { createContext, useContext, useMemo, useState, type ReactNode } from 'react'
import { api, tokenStore, type User } from '../api/client'

type AuthCtx = {
  user: User | null
  token: string | null
  login: (u: string, p: string) => Promise<void>
  logout: () => void
}

const Ctx = createContext<AuthCtx>(null as any)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => tokenStore.get())
  const [user, setUser] = useState<User | null>(null)

  async function login(u: string, p: string) {
    const r = await api<{ token: string; user: User }>('/api/login', {
      method: 'POST', body: JSON.stringify({ username: u, password: p }),
    })
    tokenStore.set(r.token)
    setToken(r.token)
    setUser(r.user)
  }

  function logout() {
    tokenStore.clear()
    setToken(null)
    setUser(null)
  }

  const value = useMemo(() => ({ user, token, login, logout }), [user, token])
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}

export function useAuth() { return useContext(Ctx) }