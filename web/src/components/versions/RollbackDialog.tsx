import { useRollback } from '../../hooks/use-versions'
import type { CommitInfo } from '../../types/version'

interface Props {
  stackId: string
  commit: CommitInfo
  onClose: () => void
}

export default function RollbackDialog({ stackId, commit, onClose }: Props) {
  const rollback = useRollback()

  const handleConfirm = () => {
    rollback.mutate(
      { stackId, hash: commit.hash },
      {
        onSuccess: () => onClose(),
      },
    )
  }

  return (
    <div
      className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <div
        className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] w-full max-w-lg p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold mb-2">Rollback to this version?</h2>
        <p className="text-sm text-[var(--color-text-muted)] mb-4">
          This creates a new commit that restores the compose file to the state below.
          History is preserved (no destructive operations).
        </p>

        <div className="bg-[var(--color-bg)] rounded-lg p-4 border border-[var(--color-border)] mb-5">
          <div className="flex items-center justify-between mb-2">
            <span className="font-mono text-sm text-[var(--color-text-muted)]">
              {commit.shortHash}
            </span>
            <span className="text-xs text-[var(--color-text-muted)]">
              {new Date(commit.timestamp).toLocaleString()}
            </span>
          </div>
          <p className="font-medium">{commit.message.split('\n')[0]}</p>
          <p className="text-xs text-[var(--color-text-muted)] mt-1">by {commit.author}</p>
        </div>

        {rollback.isError && (
          <p className="text-sm text-[var(--color-danger)] mb-3">
            {(rollback.error as Error).message}
          </p>
        )}

        <div className="flex justify-end gap-3">
          <button
            onClick={onClose}
            disabled={rollback.isPending}
            className="px-4 py-2 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            disabled={rollback.isPending}
            className="px-4 py-2 bg-[var(--color-warning)] text-white rounded-lg hover:opacity-90 transition-opacity text-sm font-medium disabled:opacity-50"
          >
            {rollback.isPending ? 'Rolling back...' : 'Confirm Rollback'}
          </button>
        </div>
      </div>
    </div>
  )
}
