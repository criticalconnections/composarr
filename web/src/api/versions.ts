import api from './client'
import type { CommitInfo, VersionDetail, StructuredDiff } from '../types/version'

export async function listVersions(stackId: string): Promise<CommitInfo[]> {
  const { data } = await api.get<CommitInfo[]>(`/stacks/${stackId}/versions`)
  return data
}

export async function getVersion(stackId: string, hash: string): Promise<VersionDetail> {
  const { data } = await api.get<VersionDetail>(`/stacks/${stackId}/versions/${hash}`)
  return data
}

export async function getVersionDiff(stackId: string, hash: string): Promise<StructuredDiff> {
  const { data } = await api.get<StructuredDiff>(`/stacks/${stackId}/versions/${hash}/diff`)
  return data
}

export async function getWorkingDiff(stackId: string, content?: string): Promise<StructuredDiff> {
  if (content !== undefined) {
    const { data } = await api.post<StructuredDiff>(`/stacks/${stackId}/diff`, { content })
    return data
  }
  const { data } = await api.get<StructuredDiff>(`/stacks/${stackId}/diff`)
  return data
}

export async function rollbackToVersion(stackId: string, hash: string): Promise<{ commitHash: string }> {
  const { data } = await api.post<{ commitHash: string }>(`/stacks/${stackId}/versions/${hash}/rollback`)
  return data
}

// Updated to accept commit message
export async function updateComposeWithMessage(
  stackId: string,
  content: string,
  commitMessage?: string,
): Promise<{ commitHash: string }> {
  const { data } = await api.put<{ commitHash: string }>(`/stacks/${stackId}/compose`, {
    content,
    commitMessage,
  })
  return data
}
