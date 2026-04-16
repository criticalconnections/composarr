import api from './client'
import type { Schedule, QueuedUpdate, CreateScheduleRequest, UpdateScheduleRequest, QueueUpdateRequest } from '../types/schedule'

export async function listSchedules(): Promise<Schedule[]> {
  const { data } = await api.get<Schedule[]>('/schedules')
  return data
}

export async function listSchedulesByStack(stackId: string): Promise<Schedule[]> {
  const { data } = await api.get<Schedule[]>(`/stacks/${stackId}/schedules`)
  return data
}

export async function createSchedule(stackId: string, req: CreateScheduleRequest): Promise<Schedule> {
  const { data } = await api.post<Schedule>(`/stacks/${stackId}/schedules`, req)
  return data
}

export async function updateSchedule(id: string, req: UpdateScheduleRequest): Promise<Schedule> {
  const { data } = await api.put<Schedule>(`/schedules/${id}`, req)
  return data
}

export async function deleteSchedule(id: string): Promise<void> {
  await api.delete(`/schedules/${id}`)
}

export async function getNextWindow(id: string): Promise<{ nextWindow: string }> {
  const { data } = await api.get<{ nextWindow: string }>(`/schedules/${id}/next`)
  return data
}

export interface UpcomingWindow {
  schedule: Schedule
  nextWindow: string
}

export async function getUpcomingWindows(limit = 10): Promise<UpcomingWindow[]> {
  const { data } = await api.get<UpcomingWindow[]>(`/schedules/upcoming?limit=${limit}`)
  return data
}

export async function listQueuedUpdates(stackId?: string): Promise<QueuedUpdate[]> {
  if (stackId) {
    const { data } = await api.get<QueuedUpdate[]>(`/stacks/${stackId}/queue`)
    return data
  }
  const { data } = await api.get<QueuedUpdate[]>('/queued-updates')
  return data
}

export async function queueUpdate(stackId: string, req: QueueUpdateRequest): Promise<QueuedUpdate> {
  const { data } = await api.post<QueuedUpdate>(`/stacks/${stackId}/queue`, req)
  return data
}

export async function cancelQueuedUpdate(id: string): Promise<void> {
  await api.delete(`/queued-updates/${id}`)
}
