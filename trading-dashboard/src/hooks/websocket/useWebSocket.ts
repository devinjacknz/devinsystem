import { useEffect, useRef, useCallback } from 'react'

interface WebSocketOptions {
  onOpen?: () => void
  onMessage?: (data: any) => void
  onClose?: () => void
  onError?: (error: Event) => void
}

export const useWebSocket = (url: string, options: WebSocketOptions = {}) => {
  const ws = useRef<WebSocket | null>(null)
  const reconnectTimeout = useRef<NodeJS.Timeout>()

  const connect = useCallback(() => {
    ws.current = new WebSocket(url)

    ws.current.addEventListener('open', () => {
      options.onOpen?.()
    })

    ws.current.addEventListener('message', (event) => {
      try {
        const data = event.data instanceof Blob ? event.data : JSON.parse(event.data)
        options.onMessage?.(data)
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    })

    ws.current.addEventListener('close', () => {
      options.onClose?.()
      reconnectTimeout.current = setTimeout(connect, 5000)
    })

    ws.current.addEventListener('error', (error) => {
      options.onError?.(error)
    })
  }, [url, options])

  useEffect(() => {
    connect()
    return () => {
      ws.current?.close()
      if (reconnectTimeout.current) {
        clearTimeout(reconnectTimeout.current)
      }
    }
  }, [connect])

  const send = useCallback((message: any) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(message))
    }
  }, [])

  return {
    send,
    readyState: ws.current?.readyState,
    lastMessage: null
  }
}
