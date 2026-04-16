import { useState } from 'react'
import { useCreateSchedule } from '../../hooks/use-schedules'
import { useStacks } from '../../hooks/use-stacks'

interface Props {
  defaultStackId?: string
  onClose: () => void
}

const presets = [
  { label: 'Every day at 2am', expr: '0 0 2 * * *' },
  { label: 'Every Tuesday at 2am', expr: '0 0 2 * * 2' },
  { label: 'Every Sunday at 3am', expr: '0 0 3 * * 0' },
  { label: 'First of the month at midnight', expr: '0 0 0 1 * *' },
  { label: 'Every hour', expr: '0 0 * * * *' },
]

export default function ScheduleForm({ defaultStackId, onClose }: Props) {
  const { data: stacks } = useStacks()
  const [stackId, setStackId] = useState(defaultStackId ?? '')
  const [name, setName] = useState('')
  const [cronExpr, setCronExpr] = useState('0 0 2 * * 2')
  const [duration, setDuration] = useState(7200)
  const [timezone, setTimezone] = useState(Intl.DateTimeFormat().resolvedOptions().timeZone)
  const [enabled, setEnabled] = useState(true)

  const createSchedule = useCreateSchedule()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!stackId) return
    createSchedule.mutate(
      {
        stackId,
        req: { name, cronExpr, duration, timezone, enabled },
      },
      {
        onSuccess: () => onClose(),
      },
    )
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div
        className="bg-[var(--color-surface)] rounded-xl border border-[var(--color-border)] w-full max-w-lg p-6 max-h-[90vh] overflow-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold mb-4">Create Maintenance Window</h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          {!defaultStackId && (
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-muted)] mb-1">
                Stack
              </label>
              <select
                value={stackId}
                onChange={(e) => setStackId(e.target.value)}
                required
                className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg focus:outline-none focus:border-[var(--color-primary)]"
              >
                <option value="">Select a stack...</option>
                {stacks?.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-muted)] mb-1">
              Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., Tuesday maintenance"
              required
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg focus:outline-none focus:border-[var(--color-primary)]"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-[var(--color-text-muted)] mb-1">
              Cron Expression
            </label>
            <input
              type="text"
              value={cronExpr}
              onChange={(e) => setCronExpr(e.target.value)}
              placeholder="sec min hour dom month dow"
              required
              className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg font-mono text-sm focus:outline-none focus:border-[var(--color-primary)]"
            />
            <div className="mt-2 flex flex-wrap gap-1">
              {presets.map((p) => (
                <button
                  key={p.expr}
                  type="button"
                  onClick={() => setCronExpr(p.expr)}
                  className="text-xs px-2 py-1 bg-[var(--color-bg)] border border-[var(--color-border)] rounded hover:border-[var(--color-primary)] transition-colors"
                >
                  {p.label}
                </button>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-sm font-medium text-[var(--color-text-muted)] mb-1">
                Window Duration (seconds)
              </label>
              <input
                type="number"
                value={duration}
                onChange={(e) => setDuration(parseInt(e.target.value) || 0)}
                min={60}
                required
                className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg focus:outline-none focus:border-[var(--color-primary)]"
              />
              <p className="text-xs text-[var(--color-text-muted)] mt-1">
                {Math.floor(duration / 60)} minutes
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-[var(--color-text-muted)] mb-1">
                Timezone
              </label>
              <input
                type="text"
                value={timezone}
                onChange={(e) => setTimezone(e.target.value)}
                placeholder="UTC"
                required
                className="w-full px-3 py-2 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg focus:outline-none focus:border-[var(--color-primary)]"
              />
            </div>
          </div>

          <label className="flex items-center gap-2 text-sm cursor-pointer">
            <input
              type="checkbox"
              checked={enabled}
              onChange={(e) => setEnabled(e.target.checked)}
              className="accent-[var(--color-primary)]"
            />
            Enabled
          </label>

          {createSchedule.isError && (
            <p className="text-sm text-[var(--color-danger)]">
              {(createSchedule.error as Error).message}
            </p>
          )}

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={createSchedule.isPending || !stackId}
              className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] text-sm font-medium disabled:opacity-50"
            >
              {createSchedule.isPending ? 'Creating...' : 'Create Schedule'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
