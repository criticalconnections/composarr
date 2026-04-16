import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as versionsApi from '../api/versions'

export function useVersions(stackId: string) {
  return useQuery({
    queryKey: ['stacks', stackId, 'versions'],
    queryFn: () => versionsApi.listVersions(stackId),
    enabled: !!stackId,
  })
}

export function useVersion(stackId: string, hash: string) {
  return useQuery({
    queryKey: ['stacks', stackId, 'versions', hash],
    queryFn: () => versionsApi.getVersion(stackId, hash),
    enabled: !!stackId && !!hash,
  })
}

export function useVersionDiff(stackId: string, hash: string) {
  return useQuery({
    queryKey: ['stacks', stackId, 'versions', hash, 'diff'],
    queryFn: () => versionsApi.getVersionDiff(stackId, hash),
    enabled: !!stackId && !!hash,
  })
}

export function useRollback() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ stackId, hash }: { stackId: string; hash: string }) =>
      versionsApi.rollbackToVersion(stackId, hash),
    onSuccess: (_data, { stackId }) => {
      queryClient.invalidateQueries({ queryKey: ['stacks', stackId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackId, 'versions'] })
    },
  })
}

export function useUpdateCompose() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ stackId, content, commitMessage }: { stackId: string; content: string; commitMessage?: string }) =>
      versionsApi.updateComposeWithMessage(stackId, content, commitMessage),
    onSuccess: (_data, { stackId }) => {
      queryClient.invalidateQueries({ queryKey: ['stacks', stackId] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackId, 'versions'] })
    },
  })
}
