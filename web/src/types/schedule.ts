export interface Schedule {
  id: string
  stackId: string
  name: string
  cronExpr: string
  duration: number
  timezone: string
  enabled: boolean
  createdAt: string
  updatedAt: string
}

export interface QueuedUpdate {
  id: string
  stackId: string
  scheduleId: string | null
  composeContent: string
  commitMessage: string
  status: 'queued' | 'deploying' | 'deployed' | 'failed' | 'cancelled'
  queuedAt: string
  deployAfter: string | null
  deployedAt: string | null
}

export interface CreateScheduleRequest {
  name: string
  cronExpr: string
  duration: number
  timezone: string
  enabled: boolean
}

export interface UpdateScheduleRequest {
  name?: string
  cronExpr?: string
  duration?: number
  timezone?: string
  enabled?: boolean
}

export interface QueueUpdateRequest {
  composeContent: string
  commitMessage?: string
  scheduleId?: string
}
