import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useDeployment, useCancelDeployment } from '../hooks/use-deployments'
import { useWebSocket } from '../hooks/use-websocket'
import { useQueryClient } from '@tanstack/react-query'
import DeployTimeline from '../components/deploy/DeployTimeline'
import HealthCheckResults from '../components/deploy/HealthCheckResults'
import type { DeploymentLog, HealthCheckResult } from '../types/deployment'

export default function DeploymentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { data, isLoading } = useDeployment(id!)
  const cancel = useCancelDeployment()
  const queryClient = useQueryClient()

  // Live log + health updates from WebSocket, layered on top of fetched data
  const [liveLogs, setLiveLogs] = useState<DeploymentLog[]>([])
  const [liveHealth, setLiveHealth] = useState<HealthCheckResult[]>([])

  useEffect(() => {
    setLiveLogs([])
    setLiveHealth([])
  }, [id])

  useWebSocket((event) => {
    if (event.deploymentId !== id) return

    if (event.type === 'deploy.log' && event.payload) {
      const p = event.payload as { level: string; message: string; time: string }
      setLiveLogs((prev) => [
        ...prev,
        {
          id: `live-${prev.length}`,
          deploymentId: id!,
          level: p.level as DeploymentLog['level'],
          message: p.message,
          timestamp: p.time,
        },
      ])
    }

    if (event.type === 'health.update' && event.payload) {
      const p = event.payload as { results: HealthCheckResult[] }
      if (p.results) setLiveHealth(p.results)
    }

    if (
      event.type === 'deploy.succeeded' ||
      event.type === 'deploy.failed' ||
      event.type === 'deploy.rolled_back'
    ) {
      queryClient.invalidateQueries({ queryKey: ['deployments', 'detail', id] })
    }
  })

  if (isLoading) {
    return <p className="text-[var(--color-text-muted)]">Loading...</p>
  }

  if (!data) {
    return <p className="text-[var(--color-danger)]">Deployment not found</p>
  }

  const { deployment } = data
  const allLogs = [...data.logs, ...liveLogs]
  const healthResults = liveHealth.length > 0 ? liveHealth : data.healthResults
  const isInFlight =
    deployment.status === 'pending' ||
    deployment.status === 'running' ||
    deployment.status === 'health_checking'

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <Link
            to={`/stacks/${deployment.stackId}`}
            className="text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
          >
            ← Back to stack
          </Link>
          <div className="flex items-center gap-3 mt-1">
            <h1 className="text-2xl font-bold">Deployment</h1>
            <DeployStatusBadge status={deployment.status} />
          </div>
          <p className="text-sm text-[var(--color-text-muted)] mt-1">
            <span className="font-mono">{deployment.commitHash.slice(0, 8)}</span>
            {' · '}
            triggered {deployment.triggerType}
            {' · '}
            {new Date(deployment.createdAt).toLocaleString()}
          </p>
        </div>
        {isInFlight && (
          <button
            onClick={() => cancel.mutate(deployment.id)}
            disabled={cancel.isPending}
            className="px-3 py-1.5 text-sm bg-[var(--color-danger)] text-white rounded-lg hover:opacity-90 disabled:opacity-50"
          >
            Cancel
          </button>
        )}
      </div>

      {/* Two-column layout: Timeline + Logs */}
      <div className="grid grid-cols-12 gap-6 mb-6">
        <div className="col-span-5">
          <div className="bg-[var(--color-surface)] rounded-lg p-5 border border-[var(--color-border)]">
            <h2 className="text-sm font-medium mb-4 text-[var(--color-text-muted)]">Pipeline</h2>
            <DeployTimeline deploymentId={deployment.id} initialStatus={deployment.status} />
          </div>
        </div>

        <div className="col-span-7">
          <div className="bg-[var(--color-surface)] rounded-lg p-5 border border-[var(--color-border)] h-full">
            <h2 className="text-sm font-medium mb-3 text-[var(--color-text-muted)]">Logs</h2>
            <div className="font-mono text-xs space-y-1 max-h-[60vh] overflow-y-auto bg-[var(--color-bg)] rounded p-3">
              {!allLogs.length ? (
                <p className="text-[var(--color-text-muted)]">No logs yet</p>
              ) : (
                allLogs.map((log) => (
                  <div key={log.id} className="flex gap-3">
                    <span className="text-[var(--color-text-muted)] flex-shrink-0">
                      {new Date(log.timestamp).toLocaleTimeString()}
                    </span>
                    <span className={logLevelColor(log.level)}>{log.level.toUpperCase()}</span>
                    <span className="text-[var(--color-text)] whitespace-pre-wrap break-words">
                      {log.message}
                    </span>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Health Results */}
      <h2 className="text-lg font-semibold mb-3">Health Checks</h2>
      <HealthCheckResults results={healthResults} />

      {deployment.errorMessage && (
        <div className="mt-6 p-4 bg-[var(--color-danger)]/10 border border-[var(--color-danger)]/30 rounded-lg">
          <p className="text-sm font-medium text-[var(--color-danger)] mb-1">Error</p>
          <p className="text-sm text-[var(--color-text)]">{deployment.errorMessage}</p>
        </div>
      )}
    </div>
  )
}

function DeployStatusBadge({ status }: { status: string }) {
  const config: Record<string, { color: string; bg: string; label: string }> = {
    pending: { color: 'var(--color-text-muted)', bg: 'rgba(148, 163, 184, 0.1)', label: 'Pending' },
    running: { color: 'var(--color-primary)', bg: 'rgba(99, 102, 241, 0.1)', label: 'Running' },
    health_checking: { color: 'var(--color-warning)', bg: 'rgba(245, 158, 11, 0.1)', label: 'Checking Health' },
    succeeded: { color: 'var(--color-success)', bg: 'rgba(34, 197, 94, 0.1)', label: 'Succeeded' },
    failed: { color: 'var(--color-danger)', bg: 'rgba(239, 68, 68, 0.1)', label: 'Failed' },
    rolled_back: { color: 'var(--color-warning)', bg: 'rgba(245, 158, 11, 0.1)', label: 'Rolled Back' },
  }
  const c = config[status] ?? config.pending

  return (
    <span
      className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium"
      style={{ color: c.color, backgroundColor: c.bg }}
    >
      <span className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: c.color }} />
      {c.label}
    </span>
  )
}

function logLevelColor(level: string): string {
  switch (level) {
    case 'error':
      return 'text-[var(--color-danger)] flex-shrink-0 w-12'
    case 'warn':
      return 'text-[var(--color-warning)] flex-shrink-0 w-12'
    case 'debug':
      return 'text-[var(--color-text-muted)] flex-shrink-0 w-12'
    default:
      return 'text-[var(--color-primary)] flex-shrink-0 w-12'
  }
}
