export const EventTypes = {
  DeployStarted: 'deploy.started',
  DeployValidating: 'deploy.validating',
  DeployPulling: 'deploy.pulling',
  DeployStarting: 'deploy.starting',
  DeployHealthChecking: 'deploy.health_checking',
  DeploySucceeded: 'deploy.succeeded',
  DeployFailed: 'deploy.failed',
  DeployRollingBack: 'deploy.rolling_back',
  DeployRolledBack: 'deploy.rolled_back',
  DeployLog: 'deploy.log',
  HealthUpdate: 'health.update',
  StackStatusChanged: 'stack.status_changed',
} as const

export type EventType = (typeof EventTypes)[keyof typeof EventTypes]

export interface ServerEvent<T = unknown> {
  type: EventType | string
  stackId?: string
  deploymentId?: string
  payload?: T
  timestamp: string
}
