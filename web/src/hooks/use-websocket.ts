import { useEffect, useRef } from 'react'
import { wsManager } from '../api/websocket'
import type { ServerEvent } from '../types/events'

/**
 * Subscribe to server events. The handler is invoked for every event;
 * filter inside the handler if you only care about specific events.
 */
export function useWebSocket(handler: (event: ServerEvent) => void) {
  const handlerRef = useRef(handler)
  handlerRef.current = handler

  useEffect(() => {
    const unsub = wsManager.subscribe((event) => handlerRef.current(event))
    return unsub
  }, [])
}
