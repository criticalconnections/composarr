import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as schedulesApi from '../api/schedules'
import type { CreateScheduleRequest, UpdateScheduleRequest, QueueUpdateRequest } from '../types/schedule'

export function useSchedules() {
  return useQuery({
    queryKey: ['schedules'],
    queryFn: schedulesApi.listSchedules,
  })
}

export function useUpcomingWindows(limit = 5) {
  return useQuery({
    queryKey: ['schedules', 'upcoming', limit],
    queryFn: () => schedulesApi.getUpcomingWindows(limit),
    refetchInterval: 30_000,
  })
}

export function useSchedulesByStack(stackId: string) {
  return useQuery({
    queryKey: ['stacks', stackId, 'schedules'],
    queryFn: () => schedulesApi.listSchedulesByStack(stackId),
    enabled: !!stackId,
  })
}

export function useCreateSchedule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ stackId, req }: { stackId: string; req: CreateScheduleRequest }) =>
      schedulesApi.createSchedule(stackId, req),
    onSuccess: (_data, { stackId }) => {
      qc.invalidateQueries({ queryKey: ['schedules'] })
      qc.invalidateQueries({ queryKey: ['stacks', stackId, 'schedules'] })
    },
  })
}

export function useUpdateSchedule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateScheduleRequest }) =>
      schedulesApi.updateSchedule(id, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['schedules'] })
    },
  })
}

export function useDeleteSchedule() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => schedulesApi.deleteSchedule(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['schedules'] })
    },
  })
}

export function useQueuedUpdates(stackId?: string) {
  return useQuery({
    queryKey: ['queued-updates', stackId ?? 'all'],
    queryFn: () => schedulesApi.listQueuedUpdates(stackId),
  })
}

export function useQueueUpdate() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ stackId, req }: { stackId: string; req: QueueUpdateRequest }) =>
      schedulesApi.queueUpdate(stackId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['queued-updates'] })
    },
  })
}

export function useCancelQueuedUpdate() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => schedulesApi.cancelQueuedUpdate(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['queued-updates'] })
    },
  })
}
