import { useWebSocket } from '../../hooks/use-websocket'
import { useStacks } from '../../hooks/use-stacks'
import { useToast } from './ToastContext'
import { EventTypes } from '../../types/events'

/**
 * Subscribes to WebSocket deploy events and shows toasts for terminal states.
 * Renders nothing visible — just a side-effect component mounted once at app root.
 */
export default function DeployEventToasts() {
  const { push } = useToast()
  const { data: stacks } = useStacks()

  useWebSocket((event) => {
    const stackName =
      stacks?.find((s) => s.id === event.stackId)?.name ?? 'Stack'

    switch (event.type) {
      case EventTypes.DeploySucceeded:
        push({
          kind: 'success',
          title: `${stackName} deployed`,
          message: 'All containers are healthy.',
        })
        break
      case EventTypes.DeployFailed: {
        const payload = event.payload as { error?: string } | undefined
        push({
          kind: 'error',
          title: `${stackName} deploy failed`,
          message: payload?.error ?? 'See deployment detail for logs.',
          duration: 8000,
        })
        break
      }
      case EventTypes.DeployRolledBack: {
        const payload = event.payload as { reason?: string } | undefined
        push({
          kind: 'warning',
          title: `${stackName} rolled back`,
          message: payload?.reason ?? 'Restored to the previous successful deployment.',
          duration: 8000,
        })
        break
      }
      case EventTypes.DeployRollingBack:
        push({
          kind: 'warning',
          title: `${stackName} rolling back`,
          message: 'Restoring previous deployment...',
        })
        break
    }
  })

  return null
}
