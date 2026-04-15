import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as stacksApi from '../api/stacks'
import type { CreateStackRequest, UpdateStackRequest } from '../types/stack'

export function useStacks() {
  return useQuery({
    queryKey: ['stacks'],
    queryFn: stacksApi.listStacks,
  })
}

export function useStack(id: string) {
  return useQuery({
    queryKey: ['stacks', id],
    queryFn: () => stacksApi.getStack(id),
    enabled: !!id,
  })
}

export function useStackStatus(id: string) {
  return useQuery({
    queryKey: ['stacks', id, 'status'],
    queryFn: () => stacksApi.getStackStatus(id),
    enabled: !!id,
    refetchInterval: 10_000,
  })
}

export function useCreateStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (req: CreateStackRequest) => stacksApi.createStack(req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
  })
}

export function useUpdateStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateStackRequest }) =>
      stacksApi.updateStack(id, req),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
      queryClient.invalidateQueries({ queryKey: ['stacks', id] })
    },
  })
}

export function useDeleteStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => stacksApi.deleteStack(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['stacks'] })
    },
  })
}

export function useStartStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => stacksApi.startStack(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['stacks', id] })
      queryClient.invalidateQueries({ queryKey: ['stacks', id, 'status'] })
    },
  })
}

export function useStopStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => stacksApi.stopStack(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['stacks', id] })
      queryClient.invalidateQueries({ queryKey: ['stacks', id, 'status'] })
    },
  })
}

export function useRestartStack() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => stacksApi.restartStack(id),
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['stacks', id, 'status'] })
    },
  })
}
