import { useState } from 'react'
import { useCreateStack } from '../../hooks/use-stacks'

interface Props {
  onClose: () => void
}

const defaultCompose = `services:
  app:
    image: nginx:latest
    ports:
      - "80:80"
    restart: unless-stopped
`

export default function CreateStackDialog({ onClose }: Props) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [composeContent, setComposeContent] = useState(defaultCompose)
  const createStack = useCreateStack()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    createStack.mutate(
      { name, description, composeContent },
      {
        onSuccess: () => onClose(),
      },
    )
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={onClose}>
      <div
        className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] w-full max-w-2xl max-h-[90vh] overflow-auto p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold mb-4">Create New Stack</h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-muted)] mb-1">
              Stack Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., media-stack"
              required
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] focus:outline-none focus:border-[var(--color-primary)]"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-muted)] mb-1">
              Description
            </label>
            <input
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] focus:outline-none focus:border-[var(--color-primary)]"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-muted)] mb-1">
              Docker Compose File
            </label>
            <textarea
              value={composeContent}
              onChange={(e) => setComposeContent(e.target.value)}
              rows={12}
              required
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg text-[var(--color-text)] font-mono text-sm focus:outline-none focus:border-[var(--color-primary)] resize-y"
            />
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={createStack.isPending || !name || !composeContent}
              className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] transition-colors text-sm font-medium disabled:opacity-50"
            >
              {createStack.isPending ? 'Creating...' : 'Create Stack'}
            </button>
          </div>

          {createStack.isError && (
            <p className="text-sm text-[var(--color-danger)]">
              {(createStack.error as Error).message}
            </p>
          )}
        </form>
      </div>
    </div>
  )
}
