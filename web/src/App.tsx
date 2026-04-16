import { Routes, Route } from 'react-router-dom'
import AppShell from './components/layout/AppShell'
import DashboardPage from './pages/DashboardPage'
import StackListPage from './pages/StackListPage'
import StackDetailPage from './pages/StackDetailPage'
import StackEditorPage from './pages/StackEditorPage'
import VersionHistoryPage from './pages/VersionHistoryPage'

export default function App() {
  return (
    <AppShell>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/stacks" element={<StackListPage />} />
        <Route path="/stacks/:id" element={<StackDetailPage />} />
        <Route path="/stacks/:id/editor" element={<StackEditorPage />} />
        <Route path="/stacks/:id/versions" element={<VersionHistoryPage />} />
      </Routes>
    </AppShell>
  )
}
