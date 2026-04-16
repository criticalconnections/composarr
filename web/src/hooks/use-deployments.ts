import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as deploysApi from '../api/deployments'

export function useDeployments(stackId?: string) {
  return useQuery({
    queryKey: ['deployments', stackId ?? 'all'],
    queryFn: () => deploysApi.listDeployments(stackId),
  })
}

export function useDeployment(id: string) {
  return useQuery({
    queryKey: ['deployments', 'detail', id],
    queryFn: () => deploysApi.getDeployment(id),
    enabled: !!id,
  })
}

export function useDeployStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ stackId, opts }: { stackId: string; opts?: deploysApi.DeployOptions }) =>
      deploysApi.deployStack(stackId, opts),
    onSuccess: (_data, { stackId }) => {
      queryClient.invalidateQueries({ queryKey: ['deployments'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', stackId] })
    },
  })
}

export function useCancelDeployment() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deploysApi.cancelDeployment(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['deployments'] })
    },
  })
}
