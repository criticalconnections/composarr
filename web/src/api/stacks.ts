import api from './client'
import type { Stack, ContainerStatus, CreateStackRequest, UpdateStackRequest } from '../types/stack'

export async function listStacks(): Promise<Stack[]> {
  const { data } = await api.get<Stack[]>('/stacks')
  return data
}

export async function getStack(id: string): Promise<Stack> {
  const { data } = await api.get<Stack>(`/stacks/${id}`)
  return data
}

export async function createStack(req: CreateStackRequest): Promise<Stack> {
  const { data } = await api.post<Stack>('/stacks', req)
  return data
}

export async function updateStack(id: string, req: UpdateStackRequest): Promise<Stack> {
  const { data } = await api.put<Stack>(`/stacks/${id}`, req)
  return data
}

export async function deleteStack(id: string): Promise<void> {
  await api.delete(`/stacks/${id}`)
}

export async function getCompose(id: string): Promise<string> {
  const { data } = await api.get<{ content: string }>(`/stacks/${id}/compose`)
  return data.content
}

export async function updateCompose(id: string, content: string): Promise<void> {
  await api.put(`/stacks/${id}/compose`, { content })
}

export async function startStack(id: string): Promise<void> {
  await api.post(`/stacks/${id}/start`)
}

export async function stopStack(id: string): Promise<void> {
  await api.post(`/stacks/${id}/stop`)
}

export async function restartStack(id: string): Promise<void> {
  await api.post(`/stacks/${id}/restart`)
}

export async function getStackStatus(id: string): Promise<ContainerStatus[]> {
  const { data } = await api.get<ContainerStatus[]>(`/stacks/${id}/status`)
  return data
}

export async function getStackLogs(id: string, service?: string, tail?: number): Promise<string> {
  const params = new URLSearchParams()
  if (service) params.set('service', service)
  if (tail) params.set('tail', String(tail))
  const { data } = await api.get<{ logs: string }>(`/stacks/${id}/logs?${params}`)
  return data.logs
}
