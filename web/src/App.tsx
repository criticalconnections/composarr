import { Routes, Route } from 'react-router-dom'
import AppShell from './components/layout/AppShell'
import DashboardPage from './pages/DashboardPage'
import StackListPage from './pages/StackListPage'
import StackDetailPage from './pages/StackDetailPage'

export default function App() {
  return (
    <AppShell>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/stacks" element={<StackListPage />} />
        <Route path="/stacks/:id" element={<StackDetailPage />} />
      </Routes>
    </AppShell>
  )
}
