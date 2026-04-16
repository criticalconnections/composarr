import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useSchedules, useDeleteSchedule, useUpdateSchedule, useQueuedUpdates, useCancelQueuedUpdate } from '../hooks/use-schedules'
import { useStacks } from '../hooks/use-stacks'
import ScheduleForm from '../components/schedules/ScheduleForm'

export default function SchedulesPage() {
  const { data: schedules, isLoading } = useSchedules()
  const { data: stacks } = useStacks()
  const { data: queued } = useQueuedUpdates()
  const [showForm, setShowForm] = useState(false)

  const stackMap = new Map(stacks?.map((s) => [s.id, s.name]) ?? [])

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Schedules</h1>
        <button
          onClick={() => setShowForm(true)}
          className="px-4 py-2 bg-[var(--color-primary)] text-white rounded-lg hover:bg-[var(--color-primary-hover)] text-sm font-medium"
        >
          + New Schedule
        </button>
      </div>

      {/* Schedules list */}
      <section className="mb-8">
        <h2 className="text-lg font-semibold mb-3">Maintenance Windows</h2>
        {isLoading ? (
          <p className="text-[var(--color-text-muted)]">Loading...</p>
        ) : !schedules?.length ? (
          <div className="bg-[var(--color-surface)] rounded-lg p-8 text-center border border-[var(--color-border)]">
            <p className="text-[var(--color-text-muted)] mb-2">No schedules yet</p>
            <p className="text-sm text-[var(--color-text-muted)]">
              Create a maintenance window to automatically deploy queued updates on a schedule.
            </p>
          </div>
        ) : (
          <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-[var(--color-border)]">
                  <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Name</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Stack</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Cron</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Duration</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Timezone</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Status</th>
                  <th className="text-right px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Actions</th>
                </tr>
              </thead>
              <tbody>
                {schedules.map((s) => (
                  <ScheduleRow
                    key={s.id}
                    schedule={s}
                    stackName={stackMap.get(s.stackId) ?? '(deleted stack)'}
                  />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {/* Queued updates */}
      <section>
        <h2 className="text-lg font-semibold mb-3">Queued Updates</h2>
        {!queued?.length ? (
          <div className="bg-[var(--color-surface)] rounded-lg p-6 text-center border border-[var(--color-border)]">
            <p className="text-[var(--color-text-muted)] text-sm">
              No pending updates. Queue an update from a stack's editor to stage it for the next window.
            </p>
          </div>
        ) : (
          <div className="bg-[var(--color-surface)] rounded-lg border border-[var(--color-border)] overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-[var(--color-border)]">
                  <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Stack</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Commit Message</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Status</th>
                  <th className="text-left px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Deploy After</th>
                  <th className="text-right px-4 py-3 text-xs font-medium text-[var(--color-text-muted)]">Actions</th>
                </tr>
              </thead>
              <tbody>
                {queued.map((q) => (
                  <QueuedUpdateRow
                    key={q.id}
                    update={q}
                    stackName={stackMap.get(q.stackId) ?? q.stackId.slice(0, 8)}
                  />
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {showForm && <ScheduleForm onClose={() => setShowForm(false)} />}
    </div>
  )
}

function ScheduleRow({ schedule, stackName }: { schedule: { id: string; name: string; stackId: string; cronExpr: string; duration: number; timezone: string; enabled: boolean }; stackName: string }) {
  const deleteSchedule = useDeleteSchedule()
  const updateSchedule = useUpdateSchedule()

  return (
    <tr className="border-b border-[var(--color-border)] last:border-b-0">
      <td className="px-4 py-3 font-medium">{schedule.name}</td>
      <td className="px-4 py-3">
        <Link to={`/stacks/${schedule.stackId}`} className="text-[var(--color-primary)] hover:underline">
          {stackName}
        </Link>
      </td>
      <td className="px-4 py-3 font-mono text-xs">{schedule.cronExpr}</td>
      <td className="px-4 py-3 text-xs text-[var(--color-text-muted)]">
        {Math.floor(schedule.duration / 60)} min
      </td>
      <td className="px-4 py-3 text-xs text-[var(--color-text-muted)]">{schedule.timezone}</td>
      <td className="px-4 py-3">
        <button
          onClick={() =>
            updateSchedule.mutate({
              id: schedule.id,
              req: { enabled: !schedule.enabled },
            })
          }
          className={`text-xs px-2 py-0.5 rounded-full font-medium ${
            schedule.enabled
              ? 'bg-[rgba(34,197,94,0.1)] text-[var(--color-success)]'
              : 'bg-[rgba(148,163,184,0.1)] text-[var(--color-text-muted)]'
          }`}
        >
          {schedule.enabled ? 'Enabled' : 'Disabled'}
        </button>
      </td>
      <td className="px-4 py-3 text-right">
        <button
          onClick={() => {
            if (confirm(`Delete schedule "${schedule.name}"?`)) {
              deleteSchedule.mutate(schedule.id)
            }
          }}
          className="text-xs text-[var(--color-danger)] hover:underline"
        >
          Delete
        </button>
      </td>
    </tr>
  )
}

function QueuedUpdateRow({ update, stackName }: { update: { id: string; stackId: string; commitMessage: string; status: string; deployAfter: string | null }; stackName: string }) {
  const cancel = useCancelQueuedUpdate()

  return (
    <tr className="border-b border-[var(--color-border)] last:border-b-0">
      <td className="px-4 py-3 font-medium">
        <Link to={`/stacks/${update.stackId}`} className="hover:text-[var(--color-primary)]">
          {stackName}
        </Link>
      </td>
      <td className="px-4 py-3 text-xs text-[var(--color-text-muted)] max-w-xs truncate">
        {update.commitMessage}
      </td>
      <td className="px-4 py-3">
        <span className={statusClass(update.status)}>{update.status}</span>
      </td>
      <td className="px-4 py-3 text-xs text-[var(--color-text-muted)]">
        {update.deployAfter ? new Date(update.deployAfter).toLocaleString() : '—'}
      </td>
      <td className="px-4 py-3 text-right">
        {update.status === 'queued' && (
          <button
            onClick={() => cancel.mutate(update.id)}
            className="text-xs text-[var(--color-danger)] hover:underline"
          >
            Cancel
          </button>
        )}
      </td>
    </tr>
  )
}

function statusClass(status: string): string {
  switch (status) {
    case 'queued':
      return 'text-xs font-medium text-[var(--color-warning)]'
    case 'deploying':
    case 'deployed':
      return 'text-xs font-medium text-[var(--color-success)]'
    case 'failed':
      return 'text-xs font-medium text-[var(--color-danger)]'
    case 'cancelled':
      return 'text-xs font-medium text-[var(--color-text-muted)]'
    default:
      return 'text-xs font-medium text-[var(--color-text-muted)]'
  }
}
