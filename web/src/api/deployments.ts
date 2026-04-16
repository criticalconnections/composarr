import api from './client'
import type { Deployment, DeploymentDetail, DeploymentLog, HealthCheckResult } from '../types/deployment'

export interface DeployOptions {
  skipPull?: boolean
  skipHealthCheck?: boolean
  trigger?: string
}

export async function deployStack(stackId: string, opts?: DeployOptions): Promise<{ deploymentId: string }> {
  const { data } = await api.post<{ deploymentId: string }>(`/stacks/${stackId}/deploy`, opts ?? {})
  return data
}

export async function listDeployments(stackId?: string, limit = 50): Promise<Deployment[]> {
  const params = new URLSearchParams()
  if (stackId) params.set('stackId', stackId)
  params.set('limit', String(limit))
  const { data } = await api.get<Deployment[]>(`/deployments?${params}`)
  return data
}

export async function getDeployment(id: string): Promise<DeploymentDetail> {
  const { data } = await api.get<DeploymentDetail>(`/deployments/${id}`)
  return data
}

export async function getDeploymentLogs(id: string): Promise<DeploymentLog[]> {
  const { data } = await api.get<DeploymentLog[]>(`/deployments/${id}/logs`)
  return data
}

export async function getDeploymentHealth(id: string): Promise<{ all: HealthCheckResult[]; latest: HealthCheckResult[] }> {
  const { data } = await api.get<{ all: HealthCheckResult[]; latest: HealthCheckResult[] }>(`/deployments/${id}/health`)
  return data
}

export async function cancelDeployment(id: string): Promise<void> {
  await api.post(`/deployments/${id}/cancel`)
}
