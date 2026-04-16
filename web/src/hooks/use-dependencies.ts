import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as depsApi from '../api/dependencies'

export function useDependencies(stackId: string) {
  return useQuery({
    queryKey: ['stacks', stackId, 'dependencies'],
    queryFn: () => depsApi.listDependencies(stackId),
    enabled: !!stackId,
  })
}

export function useDependents(stackId: string) {
  return useQuery({
    queryKey: ['stacks', stackId, 'dependents'],
    queryFn: () => depsApi.listDependents(stackId),
    enabled: !!stackId,
  })
}

export function useDependencyGraph() {
  return useQuery({
    queryKey: ['dependencies', 'graph'],
    queryFn: depsApi.getGraph,
  })
}

export function useAddDependency() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ stackId, dependsOnId, dependencyType }: { stackId: string; dependsOnId: string; dependencyType?: 'hard' | 'soft' }) =>
      depsApi.addDependency(stackId, dependsOnId, dependencyType),
    onSuccess: (_data, { stackId }) => {
      qc.invalidateQueries({ queryKey: ['stacks', stackId, 'dependencies'] })
      qc.invalidateQueries({ queryKey: ['dependencies', 'graph'] })
    },
  })
}

export function useRemoveDependency() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ stackId, depId }: { stackId: string; depId: string }) =>
      depsApi.removeDependency(stackId, depId),
    onSuccess: (_data, { stackId }) => {
      qc.invalidateQueries({ queryKey: ['stacks', stackId, 'dependencies'] })
      qc.invalidateQueries({ queryKey: ['dependencies', 'graph'] })
    },
  })
}
