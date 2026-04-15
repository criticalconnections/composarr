import type { StackStatus } from '../../types/stack'

const statusConfig: Record<StackStatus, { label: string; color: string; bg: string }> = {
  running: { label: 'Running', color: 'var(--color-success)', bg: 'rgba(34, 197, 94, 0.1)' },
  stopped: { label: 'Stopped', color: 'var(--color-text-muted)', bg: 'rgba(148, 163, 184, 0.1)' },
  degraded: { label: 'Degraded', color: 'var(--color-warning)', bg: 'rgba(245, 158, 11, 0.1)' },
  unknown: { label: 'Unknown', color: 'var(--color-text-muted)', bg: 'rgba(148, 163, 184, 0.1)' },
}

interface Props {
  status: StackStatus
}

export default function StatusBadge({ status }: Props) {
  const config = statusConfig[status] ?? statusConfig.unknown
  return (
    <span
      className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium"
      style={{ color: config.color, backgroundColor: config.bg }}
    >
      <span
        className="w-1.5 h-1.5 rounded-full"
        style={{ backgroundColor: config.color }}
      />
      {config.label}
    </span>
  )
}
