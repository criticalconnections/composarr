import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useStacks, useDeleteStack } from '../hooks/use-stacks'
import StatusBadge from '../components/stacks/StatusBadge'
import CreateStackDialog from '../components/stacks/CreateStackDialog'

export default function StackListPage() {
  const { data: stacks, isLoading } = useStacks()
  const [showCreate, setShowCreate] = useState(false)

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Stacks</h1>
        <button
          onClick={() => setShowCreate(true)}
          className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors text-sm font-medium"
        >
          + New Stack
        </button>
      </div>

      {isLoading ? (
        <p className="text-[var(--color-text-muted)]">Loading stacks...</p>
      ) : !stacks?.length ? (
        <div className="bg-[var(--color-surface)] rounded-lg p-12 text-center border border-[var(--color-border)]">
          <p className="text-lg text-[var(--color-text-muted)] mb-2">No stacks yet</p>
          <p className="text-sm text-[var(--color-text-muted)] mb-6">
            Create a stack to get started managing your Docker Compose services
          </p>
          <button
            onClick={() => setShowCreate(true)}
            className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors text-sm font-medium"
          >
            Create your first stack
          </button>
        </div>
      ) : (
        <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="border-b border-[var(--color-border)]">
                <th className="text-left px-4 py-3 text-sm font-medium text-[var(--color-text-muted)]">Name</th>
                <th className="text-left px-4 py-3 text-sm font-medium text-[var(--color-text-muted)]">Status</th>
                <th className="text-left px-4 py-3 text-sm font-medium text-[var(--color-text-muted)]">Description</th>
                <th className="text-left px-4 py-3 text-sm font-medium text-[var(--color-text-muted)]">Created</th>
                <th className="text-right px-4 py-3 text-sm font-medium text-[var(--color-text-muted)]">Actions</th>
              </tr>
            </thead>
            <tbody>
              {stacks.map((stack) => (
                <StackRow key={stack.id} stack={stack} />
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showCreate && (
        <CreateStackDialog onClose={() => setShowCreate(false)} />
      )}
    </div>
  )
}

function StackRow({ stack }: { stack: { id: string; name: string; status: 'running' | 'stopped' | 'degraded' | 'unknown'; description: string; createdAt: string } }) {
  const deleteStack = useDeleteStack()

  return (
    <tr className="border-b border-[var(--color-border)] last:border-b-0 hover:bg-[var(--color-surface-hover)] transition-colors">
      <td className="px-4 py-3">
        <Link to={`/stacks/${stack.id}`} className="font-medium hover:text-[var(--color-primary)] transition-colors">
          {stack.name}
        </Link>
      </td>
      <td className="px-4 py-3">
        <StatusBadge status={stack.status} />
      </td>
      <td className="px-4 py-3 text-sm text-[var(--color-text-muted)] max-w-xs truncate">
        {stack.description || '---'}
      </td>
      <td className="px-4 py-3 text-sm text-[var(--color-text-muted)]">
        {new Date(stack.createdAt).toLocaleDateString()}
      </td>
      <td className="px-4 py-3 text-right">
        <button
          onClick={() => {
            if (confirm(`Delete stack "${stack.name}"? This will stop all containers.`)) {
              deleteStack.mutate(stack.id)
            }
          }}
          className="text-sm text-[var(--color-danger)] hover:underline"
        >
          Delete
        </button>
      </td>
    </tr>
  )
}
