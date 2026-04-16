import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import Editor from '@monaco-editor/react'
import { useStack } from '../hooks/use-stacks'
import { useUpdateCompose } from '../hooks/use-versions'
import { getCompose } from '../api/stacks'
import DiffPreviewDialog from '../components/versions/DiffPreviewDialog'

export default function StackEditorPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { data: stack } = useStack(id!)

  const [content, setContent] = useState<string>('')
  const [originalContent, setOriginalContent] = useState<string>('')
  const [commitMessage, setCommitMessage] = useState('')
  const [showDiff, setShowDiff] = useState(false)
  const [loading, setLoading] = useState(true)

  const updateCompose = useUpdateCompose()

  useEffect(() => {
    if (!id) return
    getCompose(id).then((c) => {
      setContent(c)
      setOriginalContent(c)
      setLoading(false)
    })
  }, [id])

  const hasChanges = content !== originalContent

  const handleSave = () => {
    if (!id) return
    updateCompose.mutate(
      { stackId: id, content, commitMessage: commitMessage.trim() || undefined },
      {
        onSuccess: () => {
          setOriginalContent(content)
          setCommitMessage('')
          setShowDiff(false)
          navigate(`/stacks/${id}`)
        },
      },
    )
  }

  if (!stack) {
    return <p className="text-[var(--color-text-muted)]">Loading...</p>
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div>
          <Link
            to={`/stacks/${id}`}
            className="text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors"
          >
            ← Back to {stack.name}
          </Link>
          <h1 className="text-2xl font-bold mt-1">Edit Compose File</h1>
        </div>

        <div className="flex items-center gap-3">
          <input
            type="text"
            value={commitMessage}
            onChange={(e) => setCommitMessage(e.target.value)}
            placeholder="Commit message (optional)"
            className="w-80 px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-sm focus:outline-none focus:border-[var(--color-primary)]"
          />
          <button
            onClick={() => setShowDiff(true)}
            disabled={!hasChanges}
            className="px-4 py-2 text-sm text-[var(--color-text)] bg-[var(--color-surface-hover)] rounded-lg hover:bg-[var(--color-border)] transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Preview Diff
          </button>
          <button
            onClick={handleSave}
            disabled={!hasChanges || updateCompose.isPending}
            className="px-4 py-2 text-sm bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          >
            {updateCompose.isPending ? 'Saving...' : 'Save & Commit'}
          </button>
        </div>
      </div>

      {/* Editor */}
      <div className="flex-1 bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] overflow-hidden min-h-[500px]">
        {loading ? (
          <div className="flex items-center justify-center h-full text-[var(--color-text-muted)]">
            Loading...
          </div>
        ) : (
          <Editor
            height="70vh"
            defaultLanguage="yaml"
            value={content}
            onChange={(v) => setContent(v || '')}
            theme="vs-dark"
            options={{
              minimap: { enabled: false },
              fontSize: 14,
              lineNumbers: 'on',
              tabSize: 2,
              scrollBeyondLastLine: false,
              wordWrap: 'on',
            }}
          />
        )}
      </div>

      {updateCompose.isError && (
        <p className="mt-3 text-sm text-[var(--color-danger)]">
          {(updateCompose.error as Error).message}
        </p>
      )}

      {showDiff && id && (
        <DiffPreviewDialog
          stackId={id}
          newContent={content}
          onClose={() => setShowDiff(false)}
          onConfirm={handleSave}
          saving={updateCompose.isPending}
        />
      )}
    </div>
  )
}
