import type { HealthCheckResult } from '../../types/deployment'

interface Props {
  results: HealthCheckResult[]
}

export default function HealthCheckResults({ results }: Props) {
  if (!results.length) {
    return (
      <p className="text-sm text-[var(--color-text-muted)]">No health checks recorded yet.</p>
    )
  }

  return (
    <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] overflow-hidden">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-[var(--color-border)]">
            <th className="text-left px-3 py-2 text-xs font-medium text-[var(--color-text-muted)]">Service</th>
            <th className="text-left px-3 py-2 text-xs font-medium text-[var(--color-text-muted)]">Container</th>
            <th className="text-left px-3 py-2 text-xs font-medium text-[var(--color-text-muted)]">Status</th>
            <th className="text-left px-3 py-2 text-xs font-medium text-[var(--color-text-muted)]">Detail</th>
            <th className="text-left px-3 py-2 text-xs font-medium text-[var(--color-text-muted)]">Checked</th>
          </tr>
        </thead>
        <tbody>
          {results.map((r) => (
            <tr key={r.id} className="border-b border-[var(--color-border)] last:border-b-0">
              <td className="px-3 py-2 font-medium">{r.serviceName}</td>
              <td className="px-3 py-2 font-mono text-xs text-[var(--color-text-muted)]">
                {r.containerName}
              </td>
              <td className="px-3 py-2">
                <HealthBadge status={r.status} />
              </td>
              <td className="px-3 py-2 text-xs text-[var(--color-text-muted)] max-w-md truncate">
                {r.checkOutput || '—'}
              </td>
              <td className="px-3 py-2 text-xs text-[var(--color-text-muted)]">
                {new Date(r.checkedAt).toLocaleTimeString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function HealthBadge({ status }: { status: HealthCheckResult['status'] }) {
  const config = {
    healthy: { color: 'var(--color-success)', bg: 'rgba(34, 197, 94, 0.1)' },
    unhealthy: { color: 'var(--color-danger)', bg: 'rgba(239, 68, 68, 0.1)' },
    starting: { color: 'var(--color-warning)', bg: 'rgba(245, 158, 11, 0.1)' },
    none: { color: 'var(--color-text-muted)', bg: 'rgba(148, 163, 184, 0.1)' },
  }[status] ?? { color: 'var(--color-text-muted)', bg: 'rgba(148, 163, 184, 0.1)' }

  return (
    <span
      className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium"
      style={{ color: config.color, backgroundColor: config.bg }}
    >
      <span className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: config.color }} />
      {status}
    </span>
  )
}
