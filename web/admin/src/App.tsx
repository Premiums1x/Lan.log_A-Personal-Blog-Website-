import { Routes, Route, Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from './hooks/auth'
import Shell from './components/Shell'
import Login from './pages/Login'
import ResetPassword from './pages/ResetPassword'
import Dashboard from './pages/Dashboard'
import Posts from './pages/Posts'
import PostEdit from './pages/PostEdit'
import Settings from './pages/Settings'
import SettingsSection from './pages/SettingsSection'
import AccountSecurity from './pages/AccountSecurity'

function Guarded() {
  const { token } = useAuth()
  const loc = useLocation()
  if (!token) return <Navigate to="/admin/login" replace state={{ from: loc }} />
  return <Shell />
}

const RedirectHome = () => <Navigate to="/admin" replace />

export default function App() {
  return (
    <Routes>
      <Route path="/admin/login" element={<Login />} />
      <Route path="/admin/forgot-password" element={<ResetPassword />} />
      <Route path="/admin" element={<Guarded />}>
        <Route index element={<Dashboard />} />
        <Route path="posts" element={<Posts />} />
        <Route path="posts/new" element={<PostEdit />} />
        <Route path="posts/:id/edit" element={<PostEdit />} />
        <Route path="account" element={<AccountSecurity />} />
        <Route path="settings" element={<Settings />} />
        <Route path="settings/:key" element={<SettingsSection />} />
      </Route>
      <Route path="*" element={<RedirectHome />} />
    </Routes>
  )
}