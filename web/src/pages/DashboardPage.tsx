import { Link } from 'react-router-dom'
import { useStacks } from '../hooks/use-stacks'
import StatusBadge from '../components/stacks/StatusBadge'

export default function DashboardPage() {
  const { data: stacks, isLoading } = useStacks()

  const running = stacks?.filter((s) => s.status === 'running').length ?? 0
  const stopped = stacks?.filter((s) => s.status === 'stopped').length ?? 0
  const degraded = stacks?.filter((s) => s.status === 'degraded').length ?? 0
  const total = stacks?.length ?? 0

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4 mb-8">
        <StatCard label="Total Stacks" value={total} />
        <StatCard label="Running" value={running} color="var(--color-success)" />
        <StatCard label="Stopped" value={stopped} color="var(--color-text-muted)" />
        <StatCard label="Degraded" value={degraded} color="var(--color-warning)" />
      </div>

      {/* Recent Stacks */}
      <h2 className="text-lg font-semibold mb-4">Stacks</h2>
      {isLoading ? (
        <p className="text-[var(--color-text-muted)]">Loading...</p>
      ) : !stacks?.length ? (
        <div className="bg-[var(--color-surface)] rounded-lg p-8 text-center border border-[var(--color-border)]">
          <p className="text-[var(--color-text-muted)] mb-4">No stacks yet</p>
          <Link
            to="/stacks"
            className="inline-block px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors"
          >
            Create your first stack
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {stacks.map((stack) => (
            <Link
              key={stack.id}
              to={`/stacks/${stack.id}`}
              className="bg-[var(--color-surface)] rounded-lg p-4 border border-[var(--color-border)] hover:border-[var(--color-primary)] transition-colors"
            >
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-medium">{stack.name}</h3>
                <StatusBadge status={stack.status} />
              </div>
              {stack.description && (
                <p className="text-sm text-[var(--color-text-muted)] truncate">
                  {stack.description}
                </p>
              )}
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

function StatCard({ label, value, color }: { label: string; value: number; color?: string }) {
  return (
    <div className="bg-[var(--color-surface)] rounded-lg p-4 border border-[var(--color-border)]">
      <p className="text-sm text-[var(--color-text-muted)]">{label}</p>
      <p className="text-3xl font-bold mt-1" style={color ? { color } : undefined}>
        {value}
      </p>
    </div>
  )
}
