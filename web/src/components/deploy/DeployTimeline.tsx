import { useEffect, useState } from 'react'
import { EventTypes } from '../../types/events'
import { useWebSocket } from '../../hooks/use-websocket'

type StepStatus = 'pending' | 'active' | 'done' | 'failed' | 'skipped'

interface Step {
  key: string
  label: string
  status: StepStatus
}

const initialSteps: Step[] = [
  { key: 'started', label: 'Deployment started', status: 'pending' },
  { key: 'validating', label: 'Validating compose file', status: 'pending' },
  { key: 'pulling', label: 'Pulling images', status: 'pending' },
  { key: 'starting', label: 'Starting containers', status: 'pending' },
  { key: 'health', label: 'Verifying container health', status: 'pending' },
  { key: 'done', label: 'Deployment complete', status: 'pending' },
]

interface Props {
  deploymentId: string
  initialStatus?: string
}

export default function DeployTimeline({ deploymentId, initialStatus }: Props) {
  const [steps, setSteps] = useState<Step[]>(initialSteps)
  const [terminal, setTerminal] = useState<'succeeded' | 'failed' | 'rolled_back' | null>(
    initialStatus === 'succeeded' ? 'succeeded'
      : initialStatus === 'failed' ? 'failed'
      : initialStatus === 'rolled_back' ? 'rolled_back'
      : null,
  )
  const [rollbackReason, setRollbackReason] = useState<string>('')

  // If the deployment was already in a terminal state, mark all steps done
  useEffect(() => {
    if (initialStatus === 'succeeded') {
      setSteps((s) => s.map((step) => ({ ...step, status: 'done' })))
    } else if (initialStatus === 'failed' || initialStatus === 'rolled_back') {
      setSteps((s) =>
        s.map((step) =>
          step.key === 'done'
            ? { ...step, label: initialStatus === 'rolled_back' ? 'Rolled back' : 'Failed', status: 'failed' }
            : step,
        ),
      )
    }
  }, [initialStatus])

  useWebSocket((event) => {
    if (event.deploymentId !== deploymentId) return

    setSteps((current) => {
      const next = [...current]
      const setStatus = (key: string, status: StepStatus, label?: string) => {
        const idx = next.findIndex((s) => s.key === key)
        if (idx >= 0) {
          next[idx] = { ...next[idx], status, label: label ?? next[idx].label }
        }
      }
      const completePrior = (currentKey: string) => {
        const idx = next.findIndex((s) => s.key === currentKey)
        for (let i = 0; i < idx; i++) {
          if (next[i].status === 'pending' || next[i].status === 'active') {
            next[i] = { ...next[i], status: 'done' }
          }
        }
      }

      switch (event.type) {
        case EventTypes.DeployStarted:
          setStatus('started', 'done')
          break
        case EventTypes.DeployValidating:
          completePrior('validating')
          setStatus('validating', 'active')
          break
        case EventTypes.DeployPulling:
          completePrior('pulling')
          setStatus('pulling', 'active')
          break
        case EventTypes.DeployStarting:
          completePrior('starting')
          setStatus('starting', 'active')
          break
        case EventTypes.DeployHealthChecking:
          completePrior('health')
          setStatus('health', 'active')
          break
        case EventTypes.DeploySucceeded:
          for (let i = 0; i < next.length; i++) {
            if (next[i].status !== 'done') next[i] = { ...next[i], status: 'done' }
          }
          setTerminal('succeeded')
          break
        case EventTypes.DeployFailed: {
          const failed = next.findIndex((s) => s.status === 'active')
          if (failed >= 0) next[failed] = { ...next[failed], status: 'failed' }
          setStatus('done', 'failed', 'Deployment failed')
          setTerminal('failed')
          break
        }
        case EventTypes.DeployRollingBack:
          setStatus('done', 'active', 'Rolling back...')
          if (event.payload && typeof event.payload === 'object' && 'reason' in event.payload) {
            setRollbackReason(String((event.payload as { reason?: string }).reason ?? ''))
          }
          break
        case EventTypes.DeployRolledBack:
          setStatus('done', 'failed', 'Rolled back')
          setTerminal('rolled_back')
          break
      }

      return next
    })
  })

  return (
    <div>
      <ol className="space-y-3">
        {steps.map((step, idx) => (
          <li key={step.key} className="flex items-start gap-3">
            <StatusIcon status={step.status} index={idx + 1} />
            <div className="pt-0.5">
              <p
                className={`text-sm ${
                  step.status === 'failed'
                    ? 'text-[var(--color-danger)]'
                    : step.status === 'done'
                    ? 'text-[var(--color-text)]'
                    : step.status === 'active'
                    ? 'text-[var(--color-primary)] font-medium'
                    : 'text-[var(--color-text-muted)]'
                }`}
              >
                {step.label}
              </p>
            </div>
          </li>
        ))}
      </ol>
      {terminal === 'rolled_back' && rollbackReason && (
        <p className="mt-4 text-sm text-[var(--color-warning)]">
          Reason: {rollbackReason}
        </p>
      )}
    </div>
  )
}

function StatusIcon({ status, index }: { status: StepStatus; index: number }) {
  const base = 'w-6 h-6 rounded-full flex items-center justify-center text-xs font-medium flex-shrink-0'

  if (status === 'done') {
    return (
      <span className={`${base} bg-[var(--color-success)] text-white`}>✓</span>
    )
  }
  if (status === 'failed') {
    return (
      <span className={`${base} bg-[var(--color-danger)] text-white`}>✕</span>
    )
  }
  if (status === 'active') {
    return (
      <span className={`${base} bg-[var(--color-primary)] text-white animate-pulse`}>•</span>
    )
  }
  return (
    <span className={`${base} bg-[var(--color-surface-hover)] text-[var(--color-text-muted)]`}>
      {index}
    </span>
  )
}
