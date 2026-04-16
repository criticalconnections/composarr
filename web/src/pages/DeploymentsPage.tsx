import { Link } from 'react-router-dom'
import { useDeployments } from '../hooks/use-deployments'
import { useStacks } from '../hooks/use-stacks'

export default function DeploymentsPage() {
  const { data: deployments, isLoading } = useDeployments()
  const { data: stacks } = useStacks()

  const stackMap = new Map(stacks?.map((s) => [s.id, s.name]) ?? [])

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Deployments</h1>

      {isLoading ? (
        <p className="text-[var(--color-text-muted)]">Loading...</p>
      ) : !deployments?.length ? (
        <div className="bg-[var(--color-surface)] rounded-lg p-12 text-center border border-[var(--color-border)]">
          <p className="text-[var(--color-text-muted)]">No deployments yet</p>
        </div>
      ) : (
        <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-[var(--color-border)]">
                <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Stack</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Commit</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Trigger</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Status</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Started</th>
              </tr>
            </thead>
            <tbody>
              {deployments.map((d) => (
                <tr key={d.id} className="border-b border-[var(--color-border)] last:border-b-0 hover:bg-[var(--color-surface-hover)]">
                  <td className="px-4 py-3">
                    <Link to={`/stacks/${d.stackId}`} className="hover:text-[var(--color-primary)]">
                      {stackMap.get(d.stackId) ?? d.stackId.slice(0, 8)}
                    </Link>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-[var(--color-text-muted)]">
                    <Link to={`/deployments/${d.id}`} className="hover:text-[var(--color-primary)]">
                      {d.commitHash.slice(0, 8)}
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-xs text-[var(--color-text-muted)]">{d.triggerType}</td>
                  <td className="px-4 py-3">
                    <span className={statusColor(d.status)}>{d.status}</span>
                  </td>
                  <td className="px-4 py-3 text-xs text-[var(--color-text-muted)]">
                    {new Date(d.createdAt).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

function statusColor(status: string): string {
  switch (status) {
    case 'succeeded':
      return 'text-[var(--color-success)] text-xs font-medium'
    case 'failed':
      return 'text-[var(--color-danger)] text-xs font-medium'
    case 'rolled_back':
      return 'text-[var(--color-warning)] text-xs font-medium'
    case 'running':
    case 'health_checking':
      return 'text-[var(--color-primary)] text-xs font-medium'
    default:
      return 'text-[var(--color-text-muted)] text-xs font-medium'
  }
}
