import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useStack } from '../hooks/use-stacks'
import { useVersions, useVersionDiff } from '../hooks/use-versions'
import DiffViewer from '../components/versions/DiffViewer'
import RollbackDialog from '../components/versions/RollbackDialog'
import type { CommitInfo } from '../types/version'

export default function VersionHistoryPage() {
  const { id } = useParams<{ id: string }>()
  const { data: stack } = useStack(id!)
  const { data: versions, isLoading } = useVersions(id!)
  const [selectedHash, setSelectedHash] = useState<string | null>(null)
  const [rollbackTarget, setRollbackTarget] = useState<CommitInfo | null>(null)

  if (!stack) {
    return <p className="text-[var(--color-text-muted)]">Loading...</p>
  }

  return (
    <div>
      {/* Header */}
      <div className="mb-6">
        <Link
          to={`/stacks/${id}`}
          className="text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors"
        >
          ← Back to {stack.name}
        </Link>
        <h1 className="text-2xl font-bold mt-1">Version History</h1>
      </div>

      <div className="grid grid-cols-12 gap-6">
        {/* Commit list */}
        <div className="col-span-5">
          <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] overflow-hidden">
            <div className="px-4 py-3 border-b border-[var(--color-border)] text-sm font-medium text-[var(--color-text-muted)]">
              {versions?.length ?? 0} commits
            </div>

            {isLoading ? (
              <div className="p-6 text-center text-[var(--color-text-muted)]">Loading...</div>
            ) : !versions?.length ? (
              <div className="p-6 text-center text-[var(--color-text-muted)]">
                No commits yet
              </div>
            ) : (
              <ul className="max-h-[70vh] overflow-y-auto">
                {versions.map((commit, idx) => (
                  <li
                    key={commit.hash}
                    className={`border-b border-[var(--color-border)] last:border-b-0 cursor-pointer transition-colors ${
                      selectedHash === commit.hash
                        ? 'bg-[var(--color-surface-hover)]'
                        : 'hover:bg-[var(--color-surface-hover)]'
                    }`}
                    onClick={() => setSelectedHash(commit.hash)}
                  >
                    <div className="p-4">
                      <div className="flex items-center justify-between mb-1">
                        <span className="font-medium text-sm truncate">
                          {commit.message.split('\n')[0]}
                        </span>
                        {idx === 0 && (
                          <span className="text-xs px-2 py-0.5 rounded-full bg-[var(--color-primary)] text-white">
                            HEAD
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-3 text-xs text-[var(--color-text-muted)]">
                        <span className="font-mono">{commit.shortHash}</span>
                        <span>{commit.author}</span>
                        <span>{new Date(commit.timestamp).toLocaleString()}</span>
                      </div>
                      {idx > 0 && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation()
                            setRollbackTarget(commit)
                          }}
                          className="mt-2 text-xs text-[var(--color-warning)] hover:underline"
                        >
                          ↶ Rollback to this version
                        </button>
                      )}
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>

        {/* Diff view */}
        <div className="col-span-7">
          {selectedHash ? (
            <CommitDiff stackId={id!} hash={selectedHash} />
          ) : (
            <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] p-12 text-center text-[var(--color-text-muted)]">
              Select a commit to view its changes
            </div>
          )}
        </div>
      </div>

      {rollbackTarget && id && (
        <RollbackDialog
          stackId={id}
          commit={rollbackTarget}
          onClose={() => setRollbackTarget(null)}
        />
      )}
    </div>
  )
}

function CommitDiff({ stackId, hash }: { stackId: string; hash: string }) {
  const { data: diff, isLoading } = useVersionDiff(stackId, hash)

  if (isLoading) {
    return (
      <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] p-12 text-center text-[var(--color-text-muted)]">
        Loading diff...
      </div>
    )
  }

  if (!diff) {
    return (
      <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] p-12 text-center text-[var(--color-text-muted)]">
        No diff available
      </div>
    )
  }

  return (
    <DiffViewer
      oldContent={diff.oldContent}
      newContent={diff.newContent}
      oldTitle={diff.oldHash ? `Parent (${diff.oldHash.slice(0, 8)})` : 'Initial'}
      newTitle={`This commit (${hash.slice(0, 8)})`}
    />
  )
}
