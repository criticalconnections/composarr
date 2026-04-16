import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { useStack } from '../hooks/use-stacks'
import { useDeployStack } from '../hooks/use-deployments'
import { getWorkingDiff } from '../api/versions'
import type { StructuredDiff } from '../types/version'
import DiffViewer from '../components/versions/DiffViewer'

export default function DeployPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: stack } = useStack(id!)
  const deploy = useDeployStack()

  const [diff, setDiff] = useState<StructuredDiff | null>(null)
  const [skipPull, setSkipPull] = useState(false)
  const [skipHealthCheck, setSkipHealthCheck] = useState(false)
  const [diffLoading, setDiffLoading] = useState(true)

  useEffect(() => {
    if (!id) return
    getWorkingDiff(id)
      .then((d) => {
        setDiff(d)
        setDiffLoading(false)
      })
      .catch(() => setDiffLoading(false))
  }, [id])

  const handleDeploy = () => {
    if (!id) return
    deploy.mutate(
      { stackId: id, opts: { skipPull, skipHealthCheck } },
      {
        onSuccess: ({ deploymentId }) => {
          navigate(`/deployments/${deploymentId}`)
        },
      },
    )
  }

  if (!stack) {
    return <p className="text-[var(--color-text-muted)]">Loading...</p>
  }

  const noChanges =
    diff && diff.oldContent === diff.newContent

  return (
    <div>
      <div className="mb-6">
        <Link
          to={`/stacks/${id}`}
          className="text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors"
        >
          ← Back to {stack.name}
        </Link>
        <h1 className="text-2xl font-bold mt-1">Deploy {stack.name}</h1>
        <p className="text-[var(--color-text-muted)] mt-1">
          Review the changes that will be applied, then click Deploy to apply them with health verification.
        </p>
      </div>

      {/* Options */}
      <div className="bg-[var(--color-surface)] rounded-lg p-4 border border-[var(--color-border)] mb-4">
        <h3 className="text-sm font-medium mb-3">Options</h3>
        <div className="space-y-2">
          <label className="flex items-center gap-2 text-sm cursor-pointer">
            <input
              type="checkbox"
              checked={skipPull}
              onChange={(e) => setSkipPull(e.target.checked)}
              className="accent-[var(--color-primary)]"
            />
            Skip image pull (faster, uses local images only)
          </label>
          <label className="flex items-center gap-2 text-sm cursor-pointer">
            <input
              type="checkbox"
              checked={skipHealthCheck}
              onChange={(e) => setSkipHealthCheck(e.target.checked)}
              className="accent-[var(--color-primary)]"
            />
            Skip health check (no auto-rollback)
          </label>
        </div>
      </div>

      {/* Diff */}
      <div className="mb-6">
        <h3 className="text-sm font-medium mb-3">Changes since last deploy</h3>
        {diffLoading ? (
          <p className="text-sm text-[var(--color-text-muted)]">Computing diff...</p>
        ) : !diff || noChanges ? (
          <div className="bg-[var(--color-surface)] rounded-lg p-6 text-center border border-[var(--color-border)]">
            <p className="text-[var(--color-text-muted)]">
              No changes since the last commit. Deploying will re-apply the current compose file.
            </p>
          </div>
        ) : (
          <DiffViewer
            oldContent={diff.oldContent}
            newContent={diff.newContent}
            oldTitle="Currently deployed (HEAD)"
            newTitle="Working copy"
          />
        )}
      </div>

      {deploy.isError && (
        <p className="mb-4 text-sm text-[var(--color-danger)]">
          {(deploy.error as Error).message}
        </p>
      )}

      <div className="flex justify-end gap-3">
        <Link
          to={`/stacks/${id}`}
          className="px-4 py-2 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
        >
          Cancel
        </Link>
        <button
          onClick={handleDeploy}
          disabled={deploy.isPending}
          className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors text-sm font-medium disabled:opacity-50"
        >
          {deploy.isPending ? 'Starting deployment...' : 'Deploy'}
        </button>
      </div>
    </div>
  )
}
