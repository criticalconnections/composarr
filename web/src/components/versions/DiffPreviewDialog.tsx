import { useEffect, useState } from 'react'
import DiffViewer from './DiffViewer'
import { getWorkingDiff } from '../../api/versions'
import type { StructuredDiff } from '../../types/version'

interface Props {
  stackId: string
  newContent: string
  onClose: () => void
  onConfirm: () => void
  saving: boolean
}

export default function DiffPreviewDialog({ stackId, newContent, onClose, onConfirm, saving }: Props) {
  const [diff, setDiff] = useState<StructuredDiff | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getWorkingDiff(stackId, newContent)
      .then((d) => {
        setDiff(d)
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [stackId, newContent])

  return (
    <div
      className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <div
        className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] w-full max-w-6xl max-h-[90vh] flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between p-5 border-b border-[var(--color-border)]">
          <div>
            <h2 className="text-lg font-semibold">Preview Changes</h2>
            <p className="text-sm text-[var(--color-text-muted)] mt-0.5">
              Review the diff before committing
            </p>
          </div>
          <button
            onClick={onClose}
            className="text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
          >
            ✕
          </button>
        </div>

        <div className="flex-1 overflow-auto p-5">
          {loading ? (
            <p className="text-[var(--color-text-muted)]">Computing diff...</p>
          ) : diff ? (
            <DiffViewer
              oldContent={diff.oldContent}
              newContent={diff.newContent}
              oldTitle="Current (HEAD)"
              newTitle="Proposed"
            />
          ) : (
            <p className="text-[var(--color-text-muted)]">No diff available</p>
          )}
        </div>

        <div className="flex justify-end gap-3 p-5 border-t border-[var(--color-border)]">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            disabled={saving}
            className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors text-sm font-medium disabled:opacity-50"
          >
            {saving ? 'Saving...' : 'Commit Changes'}
          </button>
        </div>
      </div>
    </div>
  )
}
