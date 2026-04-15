import { useParams, useNavigate } from 'react-router-dom'
import { useStack, useStackStatus, useStartStack, useStopStack, useRestartStack, useDeleteStack } from '../hooks/use-stacks'
import StatusBadge from '../components/stacks/StatusBadge'

export default function StackDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: stack, isLoading } = useStack(id!)
  const { data: containers } = useStackStatus(id!)
  const startStack = useStartStack()
  const stopStack = useStopStack()
  const restartStack = useRestartStack()
  const deleteStack = useDeleteStack()

  if (isLoading) {
    return <p className="text-[var(--color-text-muted)]">Loading stack...</p>
  }

  if (!stack) {
    return <p className="text-[var(--color-danger)]">Stack not found</p>
  }

  const handleDelete = () => {
    if (confirm(`Delete stack "${stack.name}"? This will stop all containers and remove all data.`)) {
      deleteStack.mutate(stack.id, {
        onSuccess: () => navigate('/stacks'),
      })
    }
  }

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <div className="flex items-center gap-3 mb-1">
            <h1 className="text-2xl font-bold">{stack.name}</h1>
            <StatusBadge status={stack.status} />
          </div>
          {stack.description && (
            <p className="text-[var(--color-text-muted)]">{stack.description}</p>
          )}
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => startStack.mutate(stack.id)}
            disabled={startStack.isPending || stack.status === 'running'}
            className="px-3 py-1.5 text-sm bg-[var(--color-success)] text-white rounded-lg hover:opacity-90 transition-opacity disabled:opacity-50"
          >
            {startStack.isPending ? 'Starting...' : 'Start'}
          </button>
          <button
            onClick={() => stopStack.mutate(stack.id)}
            disabled={stopStack.isPending || stack.status === 'stopped'}
            className="px-3 py-1.5 text-sm bg-[var(--color-text-muted)] text-white rounded-lg hover:opacity-90 transition-opacity disabled:opacity-50"
          >
            {stopStack.isPending ? 'Stopping...' : 'Stop'}
          </button>
          <button
            onClick={() => restartStack.mutate(stack.id)}
            disabled={restartStack.isPending}
            className="px-3 py-1.5 text-sm bg-[var(--color-warning)] text-white rounded-lg hover:opacity-90 transition-opacity disabled:opacity-50"
          >
            {restartStack.isPending ? 'Restarting...' : 'Restart'}
          </button>
          <button
            onClick={handleDelete}
            disabled={deleteStack.isPending}
            className="px-3 py-1.5 text-sm bg-[var(--color-danger)] text-white rounded-lg hover:opacity-90 transition-opacity disabled:opacity-50"
          >
            Delete
          </button>
        </div>
      </div>

      {/* Stack Info */}
      <div className="grid grid-cols-2 gap-4 mb-6">
        <InfoCard label="Slug" value={stack.slug} />
        <InfoCard label="Auto Update" value={stack.autoUpdate ? 'Enabled' : 'Disabled'} />
        <InfoCard label="Created" value={new Date(stack.createdAt).toLocaleString()} />
        <InfoCard label="Updated" value={new Date(stack.updatedAt).toLocaleString()} />
      </div>

      {/* Containers */}
      <h2 className="text-lg font-semibold mb-4">Containers</h2>
      {!containers?.length ? (
        <div className="bg-[var(--color-surface)] rounded-lg p-6 text-center border border-[var(--color-border)]">
          <p className="text-[var(--color-text-muted)]">
            No containers running. Start the stack to see containers.
          </p>
        </div>
      ) : (
        <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-[var(--color-border)]">
                <th className="text-left px-4 py-3 text-sm font-medium text-[var(--color-text-muted)]">Service</th>
                <th className="text-left px-4 py-3 text-sm font-medium text-[var(--color-text-muted)]">Name</th>
                <th className="text-left px-4 py-3 text-sm font-medium text-[var(--color-text-muted)]">State</th>
                <th className="text-left px-4 py-3 text-sm font-medium text-[var(--color-text-muted)]">Status</th>
                <th className="text-left px-4 py-3 text-sm font-medium text-[var(--color-text-muted)]">Health</th>
              </tr>
            </thead>
            <tbody>
              {containers.map((container) => (
                <tr
                  key={container.id}
                  className="border-b border-[var(--color-border)] last:border-b-0"
                >
                  <td className="px-4 py-3 font-medium text-sm">{container.service}</td>
                  <td className="px-4 py-3 text-sm text-[var(--color-text-muted)] font-mono">
                    {container.name}
                  </td>
                  <td className="px-4 py-3">
                    <span
                      className={`text-sm font-medium ${
                        container.state === 'running'
                          ? 'text-[var(--color-success)]'
                          : 'text-[var(--color-danger)]'
                      }`}
                    >
                      {container.state}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-[var(--color-text-muted)]">
                    {container.status}
                  </td>
                  <td className="px-4 py-3 text-sm text-[var(--color-text-muted)]">
                    {container.health || '---'}
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

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-[var(--color-surface)] rounded-lg p-4 border border-[var(--color-border)]">
      <p className="text-xs text-[var(--color-text-muted)] mb-1">{label}</p>
      <p className="text-sm font-medium">{value}</p>
    </div>
  )
}
