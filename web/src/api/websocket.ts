import type { ServerEvent } from '../types/events'

type Listener = (event: ServerEvent) => void

class WSManager {
  private socket: WebSocket | null = null
  private listeners = new Set<Listener>()
  private reconnectDelay = 1000
  private maxReconnectDelay = 30_000
  private explicitlyClosed = false

  connect() {
    if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
      return
    }
    this.explicitlyClosed = false

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/api/v1/ws/events`

    try {
      this.socket = new WebSocket(url)
    } catch {
      this.scheduleReconnect()
      return
    }

    this.socket.onopen = () => {
      this.reconnectDelay = 1000
    }

    this.socket.onmessage = (msg) => {
      try {
        const event: ServerEvent = JSON.parse(msg.data)
        this.listeners.forEach((l) => l(event))
      } catch {
        // ignore malformed
      }
    }

    this.socket.onclose = () => {
      this.socket = null
      if (!this.explicitlyClosed) {
        this.scheduleReconnect()
      }
    }

    this.socket.onerror = () => {
      this.socket?.close()
    }
  }

  private scheduleReconnect() {
    setTimeout(() => this.connect(), this.reconnectDelay)
    this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay)
  }

  disconnect() {
    this.explicitlyClosed = true
    this.socket?.close()
    this.socket = null
  }

  subscribe(listener: Listener): () => void {
    this.listeners.add(listener)
    if (this.listeners.size === 1) {
      this.connect()
    }
    return () => {
      this.listeners.delete(listener)
    }
  }
}

export const wsManager = new WSManager()
