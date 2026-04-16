import { Link } from 'react-router-dom'
import { useUpcomingWindows } from '../../hooks/use-schedules'
import { useStacks } from '../../hooks/use-stacks'

export default function UpcomingWindows() {
  const { data: upcoming, isLoading } = useUpcomingWindows(5)
  const { data: stacks } = useStacks()
  const stackMap = new Map(stacks?.map((s) => [s.id, s.name]) ?? [])

  if (isLoading) {
    return <p className="text-sm text-[var(--color-text-muted)]">Loading schedules...</p>
  }

  if (!upcoming?.length) {
    return (
      <div className="bg-[var(--color-surface)] rounded-lg p-6 text-center border border-[var(--color-border)]">
        <p className="text-sm text-[var(--color-text-muted)]">No upcoming maintenance windows.</p>
        <Link
          to="/schedules"
          className="inline-block mt-2 text-xs text-[var(--color-primary)] hover:underline"
        >
          Create a schedule →
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-2">
      {upcoming.map((w) => (
        <div
          key={w.schedule.id}
          className="bg-[var(--color-surface)] rounded-lg p-3 border border-[var(--color-border)] flex items-center justify-between"
        >
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <Link
                to={`/stacks/${w.schedule.stackId}`}
                className="font-medium text-sm hover:text-[var(--color-primary)] truncate"
              >
                {stackMap.get(w.schedule.stackId) ?? 'Unknown'}
              </Link>
              <span className="text-xs text-[var(--color-text-muted)]">·</span>
              <span className="text-xs text-[var(--color-text-muted)] truncate">
                {w.schedule.name}
              </span>
            </div>
            <p className="text-xs text-[var(--color-text-muted)] mt-0.5 font-mono">
              {w.schedule.cronExpr}
            </p>
          </div>
          <div className="text-right flex-shrink-0 ml-3">
            <p className="text-xs font-medium text-[var(--color-warning)]">
              {formatCountdown(w.nextWindow)}
            </p>
            <p className="text-xs text-[var(--color-text-muted)]">
              {new Date(w.nextWindow).toLocaleString([], {
                month: 'short',
                day: 'numeric',
                hour: '2-digit',
                minute: '2-digit',
              })}
            </p>
          </div>
        </div>
      ))}
    </div>
  )
}

function formatCountdown(iso: string): string {
  const diff = new Date(iso).getTime() - Date.now()
  if (diff <= 0) return 'now'

  const seconds = Math.floor(diff / 1000)
  if (seconds < 60) return `in ${seconds}s`

  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `in ${minutes}m`

  const hours = Math.floor(minutes / 60)
  if (hours < 48) return `in ${hours}h`

  const days = Math.floor(hours / 24)
  return `in ${days}d`
}
