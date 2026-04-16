export type DeploymentStatus =
  | 'pending'
  | 'running'
  | 'health_checking'
  | 'succeeded'
  | 'failed'
  | 'rolled_back'

export type DeploymentTrigger = 'manual' | 'scheduled' | 'auto_update' | 'rollback'

export interface Deployment {
  id: string
  stackId: string
  commitHash: string
  previousCommit: string
  status: DeploymentStatus
  triggerType: DeploymentTrigger
  startedAt: string | null
  completedAt: string | null
  errorMessage: string
  createdAt: string
}

export interface DeploymentLog {
  id: string
  deploymentId: string
  level: 'info' | 'warn' | 'error' | 'debug'
  message: string
  timestamp: string
}

export interface HealthCheckResult {
  id: string
  deploymentId: string
  containerName: string
  serviceName: string
  status: 'healthy' | 'unhealthy' | 'starting' | 'none'
  checkOutput: string
  checkedAt: string
}

export interface DeploymentDetail {
  deployment: Deployment
  logs: DeploymentLog[]
  healthResults: HealthCheckResult[]
}
