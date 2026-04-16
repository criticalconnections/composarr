import { lazy, Suspense } from 'react'
import { Routes, Route } from 'react-router-dom'
import AppShell from './components/layout/AppShell'
import DashboardPage from './pages/DashboardPage'
import StackListPage from './pages/StackListPage'
import StackDetailPage from './pages/StackDetailPage'
import DeploymentsPage from './pages/DeploymentsPage'
import SchedulesPage from './pages/SchedulesPage'
import DependencyGraphPage from './pages/DependencyGraphPage'

// Heavy pages loaded lazily to keep initial bundle small
const StackEditorPage = lazy(() => import('./pages/StackEditorPage'))
const VersionHistoryPage = lazy(() => import('./pages/VersionHistoryPage'))
const DeployPage = lazy(() => import('./pages/DeployPage'))
const DeploymentDetailPage = lazy(() => import('./pages/DeploymentDetailPage'))

function LoadingFallback() {
  return (
    <div className="flex items-center justify-center min-h-[50vh]">
      <p className="text-[var(--color-text-muted)]">Loading...</p>
    </div>
  )
}

export default function App() {
  return (
    <AppShell>
      <Suspense fallback={<LoadingFallback />}>
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/stacks" element={<StackListPage />} />
          <Route path="/stacks/:id" element={<StackDetailPage />} />
          <Route path="/stacks/:id/editor" element={<StackEditorPage />} />
          <Route path="/stacks/:id/versions" element={<VersionHistoryPage />} />
          <Route path="/stacks/:id/deploy" element={<DeployPage />} />
          <Route path="/deployments" element={<DeploymentsPage />} />
          <Route path="/deployments/:id" element={<DeploymentDetailPage />} />
          <Route path="/schedules" element={<SchedulesPage />} />
          <Route path="/dependencies" element={<DependencyGraphPage />} />
        </Routes>
      </Suspense>
    </AppShell>
  )
}
