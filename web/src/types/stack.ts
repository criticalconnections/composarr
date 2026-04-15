export interface Stack {
  id: string
  name: string
  slug: string
  description: string
  composePath: string
  status: StackStatus
  autoUpdate: boolean
  createdAt: string
  updatedAt: string
}

export type StackStatus = 'running' | 'stopped' | 'degraded' | 'unknown'

export interface ContainerStatus {
  id: string
  name: string
  service: string
  state: string
  status: string
  health: string
}

export interface CreateStackRequest {
  name: string
  description: string
  composeContent: string
}

export interface UpdateStackRequest {
  name?: string
  description?: string
  autoUpdate?: boolean
}
