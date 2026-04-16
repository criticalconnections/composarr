import { Link } from 'react-router-dom'
import { useStacks } from '../hooks/use-stacks'
import { useDeployments } from '../hooks/use-deployments'
import StatusBadge from '../components/stacks/StatusBadge'

export default function DashboardPage() {
  const { data: stacks, isLoading } = useStacks()
  const { data: deployments } = useDeployments()

  const running = stacks?.filter((s) => s.status === 'running').length ?? 0
  const stopped = stacks?.filter((s) => s.status === 'stopped').length ?? 0
  const degraded = stacks?.filter((s) => s.status === 'degraded').length ?? 0
  const total = stacks?.length ?? 0

  const recentDeployments = deployments?.slice(0, 5) ?? []
  const stackNameMap = new Map(stacks?.map((s) => [s.id, s.name]) ?? [])

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

      <div className="grid grid-cols-2 gap-6">
        {/* Stacks */}
        <div>
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
            <div className="space-y-2">
              {stacks.map((stack) => (
                <Link
                  key={stack.id}
                  to={`/stacks/${stack.id}`}
                  className="flex items-center justify-between bg-[var(--color-surface)] rounded-lg p-3 border border-[var(--color-border)] hover:border-[var(--color-primary)] transition-colors"
                >
                  <span className="font-medium">{stack.name}</span>
                  <StatusBadge status={stack.status} />
                </Link>
              ))}
            </div>
          )}
        </div>

        {/* Recent Deployments */}
        <div>
          <h2 className="text-lg font-semibold mb-4">Recent Deployments</h2>
          {!recentDeployments.length ? (
            <div className="bg-[var(--color-surface)] rounded-lg p-8 text-center border border-[var(--color-border)]">
              <p className="text-[var(--color-text-muted)]">No deployments yet</p>
            </div>
          ) : (
            <div className="space-y-2">
              {recentDeployments.map((d) => (
                <Link
                  key={d.id}
                  to={`/deployments/${d.id}`}
                  className="block bg-[var(--color-surface)] rounded-lg p-3 border border-[var(--color-border)] hover:border-[var(--color-primary)] transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <span className="font-medium">
                      {stackNameMap.get(d.stackId) ?? 'Unknown stack'}
                    </span>
                    <span className={`text-xs ${deployStatusColor(d.status)}`}>{d.status}</span>
                  </div>
                  <div className="flex items-center gap-3 mt-1 text-xs text-[var(--color-text-muted)]">
                    <span className="font-mono">{d.commitHash.slice(0, 8)}</span>
                    <span>{new Date(d.createdAt).toLocaleString()}</span>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>
      </div>
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

function deployStatusColor(status: string): string {
  switch (status) {
    case 'succeeded':
      return 'text-[var(--color-success)]'
    case 'failed':
      return 'text-[var(--color-danger)]'
    case 'rolled_back':
      return 'text-[var(--color-warning)]'
    case 'running':
    case 'health_checking':
      return 'text-[var(--color-primary)]'
    default:
      return 'text-[var(--color-text-muted)]'
  }
}
