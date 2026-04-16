import api from './client'
import type { StackDependency, DependencyGraph } from '../types/dependency'

export async function listDependencies(stackId: string): Promise<StackDependency[]> {
  const { data } = await api.get<StackDependency[]>(`/stacks/${stackId}/dependencies`)
  return data
}

export async function listDependents(stackId: string): Promise<StackDependency[]> {
  const { data } = await api.get<StackDependency[]>(`/stacks/${stackId}/dependents`)
  return data
}

export async function addDependency(
  stackId: string,
  dependsOnId: string,
  dependencyType: 'hard' | 'soft' = 'hard',
): Promise<StackDependency> {
  const { data } = await api.post<StackDependency>(`/stacks/${stackId}/dependencies`, {
    dependsOnId,
    dependencyType,
  })
  return data
}

export async function removeDependency(stackId: string, depId: string): Promise<void> {
  await api.delete(`/stacks/${stackId}/dependencies/${depId}`)
}

export async function getGraph(): Promise<DependencyGraph> {
  const { data } = await api.get<DependencyGraph>('/dependencies/graph')
  return data
}
