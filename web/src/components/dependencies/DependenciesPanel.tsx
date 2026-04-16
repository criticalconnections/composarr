import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useDependencies, useAddDependency, useRemoveDependency } from '../../hooks/use-dependencies'
import { useStacks } from '../../hooks/use-stacks'

interface Props {
  stackId: string
}

export default function DependenciesPanel({ stackId }: Props) {
  const { data: deps } = useDependencies(stackId)
  const { data: stacks } = useStacks()
  const addDep = useAddDependency()
  const removeDep = useRemoveDependency()

  const [showAdd, setShowAdd] = useState(false)
  const [selectedStack, setSelectedStack] = useState('')
  const [depType, setDepType] = useState<'hard' | 'soft'>('hard')

  const stackMap = new Map(stacks?.map((s) => [s.id, s]) ?? [])
  const availableStacks = stacks?.filter(
    (s) => s.id !== stackId && !deps?.some((d) => d.dependsOnId === s.id),
  )

  const handleAdd = () => {
    if (!selectedStack) return
    addDep.mutate(
      { stackId, dependsOnId: selectedStack, dependencyType: depType },
      {
        onSuccess: () => {
          setShowAdd(false)
          setSelectedStack('')
        },
      },
    )
  }

  return (
    <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] p-5">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-medium text-[var(--color-text-muted)]">Dependencies</h2>
        {!showAdd && (
          <button
            onClick={() => setShowAdd(true)}
            className="text-xs text-[var(--color-primary)] hover:underline"
          >
            + Add dependency
          </button>
        )}
      </div>

      {showAdd && (
        <div className="bg-[var(--color-bg)] rounded-lg p-3 mb-4 border border-[var(--color-border)]">
          <p className="text-xs text-[var(--color-text-muted)] mb-2">This stack depends on:</p>
          <div className="flex gap-2 mb-2">
            <select
              value={selectedStack}
              onChange={(e) => setSelectedStack(e.target.value)}
              className="flex-1 px-3 py-1.5 bg-[var(--color-surface)] border border-[var(--color-border)] rounded text-sm"
            >
              <option value="">Select a stack...</option>
              {availableStacks?.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name}
                </option>
              ))}
            </select>
            <select
              value={depType}
              onChange={(e) => setDepType(e.target.value as 'hard' | 'soft')}
              className="px-3 py-1.5 bg-[var(--color-surface)] border border-[var(--color-border)] rounded text-sm"
            >
              <option value="hard">Hard</option>
              <option value="soft">Soft</option>
            </select>
          </div>
          {addDep.isError && (
            <p className="text-xs text-[var(--color-danger)] mb-2">
              {(addDep.error as Error).message}
            </p>
          )}
          <div className="flex justify-end gap-2">
            <button
              onClick={() => setShowAdd(false)}
              className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text)] px-2 py-1"
            >
              Cancel
            </button>
            <button
              onClick={handleAdd}
              disabled={!selectedStack || addDep.isPending}
              className="text-xs bg-[var(--color-primary)] text-white px-3 py-1 rounded hover:bg-[var(--color-primary-hover)] disabled:opacity-50"
            >
              Add
            </button>
          </div>
        </div>
      )}

      {!deps?.length ? (
        <p className="text-sm text-[var(--color-text-muted)]">No dependencies configured.</p>
      ) : (
        <ul className="space-y-2">
          {deps.map((dep) => {
            const depStack = stackMap.get(dep.dependsOnId)
            return (
              <li
                key={dep.id}
                className="flex items-center justify-between bg-[var(--color-bg)] rounded px-3 py-2 border border-[var(--color-border)]"
              >
                <div className="flex items-center gap-2">
                  <Link
                    to={`/stacks/${dep.dependsOnId}`}
                    className="text-sm hover:text-[var(--color-primary)]"
                  >
                    {depStack?.name ?? dep.dependsOnId.slice(0, 8)}
                  </Link>
                  <span
                    className={`text-xs px-1.5 py-0.5 rounded ${
                      dep.dependencyType === 'hard'
                        ? 'bg-[rgba(99,102,241,0.1)] text-[var(--color-primary)]'
                        : 'bg-[rgba(148,163,184,0.1)] text-[var(--color-text-muted)]'
                    }`}
                  >
                    {dep.dependencyType}
                  </span>
                </div>
                <button
                  onClick={() => removeDep.mutate({ stackId, depId: dep.id })}
                  className="text-xs text-[var(--color-danger)] hover:underline"
                >
                  Remove
                </button>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
